package main_test

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"

	"github.com/hoyci/fakeflix/internal/infra/db/postgres"
	"gorm.io/gorm"
)

func addFileToMultipart(t *testing.T, writer *multipart.Writer, fieldName, filePath string) {
	t.Helper()

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open file %s: %v", filePath, err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatalf("Failed to copy file content: %v", err)
	}
}

func teardown(t *testing.T, db *gorm.DB, videoID string) {
	t.Helper()

	if videoID == "" {
		return
	}

	var movieModel postgres.MovieModel
	err := db.Preload("Video").Preload("Thumbnail").First(&movieModel, "video_id = ?", videoID).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			t.Logf("Failed to find movie for cleanup by videoID %s: %v", videoID, err)
		}
		if err := db.Unscoped().Delete(&postgres.VideoModel{}, "id = ?", videoID).Error; err != nil {
			t.Errorf("Failed to clean up video from database with ID %s: %v", videoID, err)
		}
		return
	}

	if movieModel.Video.URL != "" {
		videoPath := filepath.Join("..", "..", movieModel.Video.URL)
		if err := os.Remove(videoPath); err != nil && !os.IsNotExist(err) {
			t.Logf("Failed to remove video file during cleanup: %s", videoPath)
		}
	}
	if movieModel.Thumbnail != nil && movieModel.Thumbnail.URL != "" {
		thumbPath := filepath.Join("..", "..", movieModel.Thumbnail.URL)
		if err := os.Remove(thumbPath); err != nil && !os.IsNotExist(err) {
			t.Logf("Failed to remove thumbnail file during cleanup: %s", thumbPath)
		}
	}

	if err := db.Unscoped().Delete(&postgres.ContentModel{}, "id = ?", movieModel.ContentID).Error; err != nil {
		t.Errorf("Failed to clean up content from database with ID %s: %v", movieModel.ContentID, err)
	}

	t.Logf("Cleaned up resources for video with ID: %s", videoID)
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
