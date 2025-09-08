package main_test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/hoyci/fakeflix/internal/infra/db/postgres"
)

func TestAddVideoE2E(t *testing.T) {
	t.Run("should create movie successfully with valid data", func(t *testing.T) {
		var createdVideoID string
		videoFilePath := filepath.Join("..", "..", "testdata", "sample.mp4")
		thumbFilePath := filepath.Join("..", "..", "testdata", "sample.jpg")

		t.Cleanup(func() {
			teardown(t, db, createdVideoID)
		})

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("title", "E2E Test Movie")
		_ = writer.WriteField("description", "A description for the test movie.")
		addFileToMultipart(t, writer, "video", videoFilePath)
		addFileToMultipart(t, writer, "thumbnail", thumbFilePath)
		if err := writer.Close(); err != nil {
			t.Fatalf("Failed to close multipart writer: %v", err)
		}

		req, _ := http.NewRequest(http.MethodPost, baseAPIURL+"/movies", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status code 201, but got %d. Response: %s", resp.StatusCode, string(bodyBytes))
		}

		var respBody map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}

		var createdVideo postgres.VideoModel
		result := db.Order("created_at desc").First(&createdVideo)
		if result.Error != nil {
			t.Fatalf("Failed to find created video in the database: %v", result.Error)
		}

		createdVideoID = createdVideo.ID

		if createdVideo.Duration <= 0 {
			t.Errorf("Expected video duration to be positive, but got %d", createdVideo.Duration)
		}
	})

	t.Run("should fail when the video file is not an mp4", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("title", "Video with Invalid Format")
		_ = writer.WriteField("description", "Test description.")
		addFileToMultipart(t, writer, "video", filepath.Join("..", "..", "testdata", "sample.mp3"))
		addFileToMultipart(t, writer, "thumbnail", filepath.Join("..", "..", "testdata", "sample.jpg"))
		writer.Close()

		req, _ := http.NewRequest(http.MethodPost, baseAPIURL+"/movies", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("Expected status code 422, but got %d", resp.StatusCode)
		}
	})
}
