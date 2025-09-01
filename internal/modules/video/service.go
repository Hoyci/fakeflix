package video

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/hoyci/fakeflix/packages/fault"
)

type service struct {
	videoRepo Repository
	log       *log.Logger
}

func NewService(videoRepo Repository, log *log.Logger) Service {
	return &service{
		videoRepo: videoRepo,
		log:       log,
	}
}

func (s *service) AddVideo(ctx context.Context, dto AddVideoDTO) error {
	s.log.Info("starting video addition process")

	videoHeader := dto.VideoFile[0]
	thumbHeader := dto.ThumbFile[0]

	videoID := uuid.NewString()

	videoFilename := fmt.Sprintf("%s-%s-%s", time.Now(), &videoID, filepath.Ext(videoHeader.Filename))
	videoPath := filepath.Join("upload", "videos", videoFilename)

	thumbFilename := fmt.Sprintf("%s-%s-%s", time.Now(), &videoID, filepath.Ext(thumbHeader.Filename))
	thumbPath := filepath.Join("upload", "thumbs", thumbFilename)

	if err := os.MkdirAll(filepath.Dir(videoPath), os.ModePerm); err != nil {
		return fault.New("failed to create video directory", fault.WithError(err))
	}

	if err := os.MkdirAll(filepath.Dir(thumbPath), os.ModePerm); err != nil {
		return fault.New("failed to create thumbnail directory", fault.WithError(err))
	}

	if err := saveFile(videoHeader, videoPath); err != nil {
		return fault.New("failed to save video file", fault.WithError(err))
	}
	if err := saveFile(thumbHeader, thumbPath); err != nil {
		return fault.New("failed to save thumbnail file", fault.WithError(err))
	}

	duration, err := getVideoDuration(videoPath)
	if err != nil {
		return fault.New("failed to get video duration", fault.WithError(err))
	}

	videoSizeKB := int(videoHeader.Size / 1024)
	videoURL := "/" + videoPath
	thumbURL := "/" + thumbPath

	videoEntity, err := NewVideo(videoID, dto.Title, dto.Description, videoURL, &thumbURL, videoSizeKB, duration)
	if err != nil {
		return fault.New(
			"failed to create video entity",
			fault.WithHTTPCode(http.StatusUnprocessableEntity),
			fault.WithKind(fault.KindValidation),
			fault.WithError(err),
		)
	}

	s.log.Info("video entity created successfully", "videoID", videoEntity.ID(), "duration", videoEntity.Duration())

	if err := s.videoRepo.CreateVideo(ctx, videoEntity); err != nil {
		s.log.Error("failed to persist book and initial stock ledger entry", "error", err)
		return fault.New(
			"failed to save video to database",
			fault.WithHTTPCode(http.StatusInternalServerError),
			fault.WithKind(fault.KindUnexpected),
			fault.WithError(err),
		)
	}

	return nil
}

func (s *service) GetVideo(ctx context.Context, videoID string) (*VideoModel, error) {
	s.log.Info("starting get video process")

	video, err := s.videoRepo.GetVideoByID(ctx, videoID)
	if err != nil {
		var f *fault.Error
		if errors.As(err, &f) && f.Kind == fault.KindNotFound {
			s.log.Info("video not found", "videoID", videoID)
			return nil, err
		}

		s.log.Error("failed to get video by id", "videoID", videoID, "error", err)
		return nil, err
	}

	return video, nil
}

func saveFile(fileHeader *multipart.FileHeader, destPath string) error {
	src, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func getVideoDuration(filePath string) (int, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to run ffprobe: %w", err)
	}

	durationStr := strings.TrimSpace(string(output))
	durationFloat, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return int(durationFloat), nil
}
