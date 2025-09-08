package main_test

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/google/uuid"
	// CORREÇÃO: Importar o pacote de modelos de banco de dados correto.
	"github.com/hoyci/fakeflix/internal/infra/db/postgres"
)

func TestGetVideoE2E(t *testing.T) {
	videoID := uuid.NewString()
	thumbID := uuid.NewString()
	contentID := uuid.NewString()
	movieID := uuid.NewString()

	sourceVideoPath := filepath.Join("..", "..", "testdata", "sample.mp4")
	sourceThumbPath := filepath.Join("..", "..", "testdata", "sample.jpg")

	sourceFileInfo, err := os.Stat(sourceVideoPath)
	if err != nil {
		t.Fatalf("Failed to get stats for source video file: %v", err)
	}
	videoSize := int(sourceFileInfo.Size())

	uploadVideosDir := filepath.Join("..", "..", "upload", "videos")
	uploadThumbsDir := filepath.Join("..", "..", "upload", "thumbs")
	os.MkdirAll(uploadVideosDir, os.ModePerm)
	os.MkdirAll(uploadThumbsDir, os.ModePerm)

	videoURLPath := filepath.Join("upload", "videos", fmt.Sprintf("%s.mp4", videoID))
	destVideoPath := filepath.Join("..", "..", videoURLPath)

	thumbURLPath := filepath.Join("upload", "thumbs", fmt.Sprintf("%s.jpg", thumbID))
	destThumbPath := filepath.Join("..", "..", thumbURLPath)

	if err := copyFile(sourceVideoPath, destVideoPath); err != nil {
		t.Fatalf("Failed to copy test video file: %v", err)
	}
	if err := copyFile(sourceThumbPath, destThumbPath); err != nil {
		t.Fatalf("Failed to copy test thumbnail file: %v", err)
	}

	videoModel := postgres.VideoModel{
		ID:       videoID,
		URL:      "/" + videoURLPath,
		SizeInKb: videoSize / 1024,
		Duration: 30,
	}
	thumbModel := postgres.ThumbnailModel{
		ID:  thumbID,
		URL: "/" + thumbURLPath,
	}
	thumbModelIDStr := thumbModel.ID
	contentModel := postgres.ContentModel{
		ID:          contentID,
		Title:       "Test Video for Download",
		Description: "A video to test streaming and full download.",
		ContentType: "MOVIE",
	}
	movieModel := postgres.MovieModel{
		ID:          movieID,
		ContentID:   contentID,
		VideoID:     videoID,
		ThumbnailID: &thumbModelIDStr,
	}

	if err := db.Create(&videoModel).Error; err != nil {
		t.Fatalf("Failed to seed video in test database: %v", err)
	}
	if err := db.Create(&thumbModel).Error; err != nil {
		t.Fatalf("Failed to seed thumbnail in test database: %v", err)
	}
	if err := db.Create(&contentModel).Error; err != nil {
		t.Fatalf("Failed to seed content in test database: %v", err)
	}
	if err := db.Create(&movieModel).Error; err != nil {
		t.Fatalf("Failed to seed movie in test database: %v", err)
	}

	t.Cleanup(func() {
		os.Remove(destVideoPath)
		os.Remove(destThumbPath)
		db.Unscoped().Delete(&postgres.ContentModel{}, "id = ?", contentID)
		db.Unscoped().Delete(&postgres.MovieModel{}, "id = ?", movieID)
		db.Unscoped().Delete(&postgres.VideoModel{}, "id = ?", videoID)
		db.Unscoped().Delete(&postgres.ThumbnailModel{}, "id = ?", thumbID)
		log.Printf("Cleaned up resources for video download test with ID: %s", videoID)
	})

	t.Run("should download the full video file without range header", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, baseAPIURL+"/videos/"+videoID+"/stream", nil)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status code 200 OK, but got %d", resp.StatusCode)
		}
		if resp.Header.Get("Content-Length") != strconv.Itoa(videoSize) {
			t.Errorf("expected Content-Length %d, but got %s", videoSize, resp.Header.Get("Content-Length"))
		}
	})

	t.Run("should stream a partial chunk of the video using Range header", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, baseAPIURL+"/videos/"+videoID+"/stream", nil)
		req.Header.Set("Range", "bytes=50-149")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusPartialContent {
			t.Errorf("expected status code 206 Partial Content, but got %d", resp.StatusCode)
		}
		if resp.Header.Get("Content-Length") != "100" {
			t.Errorf("expected Content-Length of 100 for the range, but got %s", resp.Header.Get("Content-Length"))
		}

		expectedContentRange := fmt.Sprintf("bytes 50-149/%d", videoSize)
		if resp.Header.Get("Content-Range") != expectedContentRange {
			t.Errorf("expected Content-Range '%s', but got '%s'", expectedContentRange, resp.Header.Get("Content-Range"))
		}
	})
}
