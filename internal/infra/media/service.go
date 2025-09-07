package media

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
)

type StoredFileInfo struct {
	URL      string
	SizeInKb int
	Duration int
}

type MediaService interface {
	Store(fileHeader *multipart.FileHeader, destFolder string) (*StoredFileInfo, error)
	GetStream(filePath string) (io.ReadSeekCloser, os.FileInfo, error)
}

type localMediaService struct {
	logger *log.Logger
}

func NewLocalMediaService(logger *log.Logger) MediaService {
	return &localMediaService{
		logger: logger,
	}
}

func (s *localMediaService) Store(fileHeader *multipart.FileHeader, destFolder string) (*StoredFileInfo, error) {
	log := s.logger.With("filename", fileHeader.Filename, "destFolder", destFolder)
	log.Debug("Starting file store operation")

	if err := os.MkdirAll(destFolder, os.ModePerm); err != nil {
		log.Error("Failed to create destination directory", "error", err)
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	destPath := filepath.Join(destFolder, fileHeader.Filename)
	if err := saveFile(fileHeader, destPath); err != nil {
		log.Error("Failed to save file to disk", "path", destPath, "error", err)
		return nil, fmt.Errorf("failed to save file: %w", err)
	}
	log.Debug("File saved successfully to disk", "path", destPath)

	duration, err := getVideoDuration(s.logger, destPath)
	if err != nil {
		log.Warn("Could not get video duration", "path", destPath, "error", err)
		duration = 0
	}
	log.Debug("Video duration retrieved", "duration_sec", duration)

	info := &StoredFileInfo{
		URL:      "/" + destPath,
		SizeInKb: int(fileHeader.Size / 1024),
		Duration: duration,
	}

	log.Info("File stored and processed successfully", "url", info.URL, "sizeKB", info.SizeInKb)
	return info, nil
}

func (s *localMediaService) GetStream(filePath string) (io.ReadSeekCloser, os.FileInfo, error) {
	log := s.logger.With("filePath", filePath)
	log.Debug("Attempting to get file stream")

	file, err := os.Open(filePath)
	if err != nil {
		log.Error("Failed to open file for streaming", "error", err)
		return nil, nil, err
	}

	fileStat, err := file.Stat()
	if err != nil {
		log.Error("Failed to get file stats", "error", err)
		file.Close()
		return nil, nil, err
	}

	log.Debug("File stream opened successfully")
	return file, fileStat, nil
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

func getVideoDuration(logger *log.Logger, filePath string) (int, error) {
	log := logger.With("filePath", filePath)
	if !strings.HasSuffix(strings.ToLower(filePath), ".mp4") {
		log.Debug("File is not an mp4, skipping duration check")
		return 0, nil
	}

	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath,
	)

	output, err := cmd.Output()
	if err != nil {
		log.Error("Failed to run ffprobe command", "error", err, "output", string(output))
		return 0, fmt.Errorf("failed to run ffprobe: %w", err)
	}

	durationStr := strings.TrimSpace(string(output))
	durationFloat, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		log.Warn("Failed to parse ffprobe duration output", "output", durationStr, "error", err)
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return int(durationFloat), nil
}
