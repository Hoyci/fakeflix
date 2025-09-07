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
)

type StoredFileInfo struct {
	URL      string
	SizeInKb int
	Duration int
}

type MediaService interface {
	Store(fileHeader *multipart.FileHeader, destFolder string) (*StoredFileInfo, error)
}

type localMediaService struct{}

func NewLocalMediaService() MediaService {
	return &localMediaService{}
}

func (s *localMediaService) Store(fileHeader *multipart.FileHeader, destFolder string) (*StoredFileInfo, error) {
	if err := os.MkdirAll(destFolder, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	destPath := filepath.Join(destFolder, fileHeader.Filename)
	if err := saveFile(fileHeader, destPath); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	duration, err := getVideoDuration(destPath)
	if err != nil {
		duration = 0
	}

	info := &StoredFileInfo{
		URL:      "/" + destPath,
		SizeInKb: int(fileHeader.Size / 1024),
		Duration: duration,
	}

	return info, nil
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
	if !strings.HasSuffix(strings.ToLower(filePath), ".mp4") {
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
		return 0, fmt.Errorf("failed to run ffprobe: %w", err)
	}

	durationStr := strings.TrimSpace(string(output))
	durationFloat, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return int(durationFloat), nil
}
