package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"

	"github.com/hoyci/fakeflix/internal/modules/video"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gorm_postgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	baseAPIURL string
	db         *gorm.DB
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, _, gormDB, err := setupTestDatabase(ctx)
	if err != nil {
		log.Fatalf("Failed to setup test database: %v", err)
	}
	db = gormDB
	defer pgContainer.Terminate(ctx)

	dbHost, err := pgContainer.Host(ctx)
	if err != nil {
		log.Fatalf("Failed to get container host: %v", err)
	}
	dbPort, err := pgContainer.MappedPort(ctx, "5432")
	if err != nil {
		log.Fatalf("Failed to get container port: %v", err)
	}

	const apiPort = "3030"
	baseAPIURL = fmt.Sprintf("http://localhost:%s", apiPort)

	apiPath := buildAPI(m)
	defer os.Remove(apiPath)

	cmd := exec.Command(apiPath)
	cmd.Dir = "../.."
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("APP_PORT=%s", apiPort),
		"DB_HOST="+dbHost,
		"DB_PORT="+dbPort.Port(),
		"DB_USERNAME=user",
		"DB_PASSWORD=password",
		"DB_DATABASE=test-db-e2e",
		"APP_ENV=testing",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start API: %v", err)
	}
	log.Printf("API started with PID %d, connected to Testcontainer", cmd.Process.Pid)

	defer func() {
		log.Printf("Terminating API process (PID %d)...", cmd.Process.Pid)
		if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
			log.Printf("Failed to send SIGINT, forcing shutdown...")
			cmd.Process.Kill()
		}
		cmd.Wait()
	}()

	waitForAPIReady(baseAPIURL + "/videos")

	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestAddVideoE2E(t *testing.T) {
	t.Run("should create video successfully with valid data", func(t *testing.T) {
		var createdVideoID string
		videoFilePath := filepath.Join("..", "..", "testdata", "sample.mp4")
		thumbFilePath := filepath.Join("..", "..", "testdata", "sample.jpg")

		t.Cleanup(func() {
			teardown(t, db, createdVideoID)
		})

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("title", "E2E Test Video")
		_ = writer.WriteField("description", "A description for the test video.")
		addFileToMultipart(t, writer, "video", videoFilePath)
		addFileToMultipart(t, writer, "thumbnail", thumbFilePath)
		if err := writer.Close(); err != nil {
			t.Fatalf("Failed to close multipart writer: %v", err)
		}

		req, _ := http.NewRequest(http.MethodPost, baseAPIURL+"/videos", body)
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
		if msg, ok := respBody["message"]; !ok || msg != "created successfully" {
			t.Errorf("Expected success message, but got '%s'", msg)
		}

		var createdVideo video.VideoModel
		result := db.Order("created_at desc").First(&createdVideo)
		if result.Error != nil {
			t.Fatalf("Failed to find created video in the database: %v", result.Error)
		}

		createdVideoID = createdVideo.ID

		if createdVideo.Title != "E2E Test Video" {
			t.Errorf("Expected title 'E2E Test Video', but got '%s'", createdVideo.Title)
		}
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

		req, _ := http.NewRequest(http.MethodPost, baseAPIURL+"/videos", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("Expected status code 400, but got %d", resp.StatusCode)
		}
	})
}

func TestGetVideoE2E(t *testing.T) {
	videoID := uuid.NewString()
	sourceVideoPath := filepath.Join("..", "..", "testdata", "sample.mp4")
	sourceThumbPath := filepath.Join("..", "..", "testdata", "sample.jpg")

	sourceFileInfo, err := os.Stat(sourceVideoPath)
	if err != nil {
		t.Fatalf("Failed to get stats for source video file: %v", err)
	}
	videoSize := int(sourceFileInfo.Size())

	uploadVideosDir := filepath.Join("..", "..", "upload", "videos")
	uploadThumbsDir := filepath.Join("..", "..", "upload", "thumbs")
	if err := os.MkdirAll(uploadVideosDir, os.ModePerm); err != nil {
		t.Fatalf("Failed to create upload/videos directory: %v", err)
	}
	if err := os.MkdirAll(uploadThumbsDir, os.ModePerm); err != nil {
		t.Fatalf("Failed to create upload/thumbs directory: %v", err)
	}

	videoURLPath := filepath.Join("upload", "videos", fmt.Sprintf("%s.mp4", videoID))
	destVideoPath := filepath.Join("..", "..", videoURLPath)

	thumbURLPath := filepath.Join("upload", "thumbs", fmt.Sprintf("%s.jpg", videoID))
	destThumbPath := filepath.Join("..", "..", thumbURLPath)

	if err := copyFile(sourceVideoPath, destVideoPath); err != nil {
		t.Fatalf("Failed to copy test video file: %v", err)
	}
	if err := copyFile(sourceThumbPath, destThumbPath); err != nil {
		t.Fatalf("Failed to copy test thumbnail file: %v", err)
	}

	thumbURLForDB := "/" + thumbURLPath
	videoToCreate := &video.VideoModel{
		ID:           videoID,
		Title:        "Test Video for Download",
		Description:  "A video to test streaming and full download.",
		URL:          "/" + videoURLPath,
		ThumbnailURL: &thumbURLForDB,
		SizeInKB:     videoSize / 1024,
		Duration:     30,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := db.Create(videoToCreate).Error; err != nil {
		t.Fatalf("Failed to seed video in test database: %v", err)
	}

	t.Cleanup(func() {
		os.Remove(destVideoPath)
		os.Remove(destThumbPath)
		db.Unscoped().Delete(&video.VideoModel{}, "id = ?", videoID)
		log.Printf("Cleaned up resources for video download test with ID: %s", videoID)
	})

	t.Run("should download the full video file without range header", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, baseAPIURL+"/videos/"+videoID, nil)
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
		if resp.Header.Get("Content-Type") != "video/mp4" {
			t.Errorf("expected Content-Type video/mp4, but got %s", resp.Header.Get("Content-Type"))
		}
		if resp.Header.Get("Accept-Ranges") != "bytes" {
			t.Errorf("expected Accept-Ranges header to be 'bytes', but got %s", resp.Header.Get("Accept-Ranges"))
		}
	})

	t.Run("should stream a partial chunk of the video using Range header", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, baseAPIURL+"/videos/"+videoID, nil)
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

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		if len(bodyBytes) != 100 {
			t.Errorf("expected response body to have 100 bytes, but got %d", len(bodyBytes))
		}
	})
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

func setupTestDatabase(ctx context.Context) (testcontainers.Container, string, *gorm.DB, error) {
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("test-db-e2e"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to start container: %w", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return pgContainer, "", nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	db, err := gorm.Open(gorm_postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		return pgContainer, "", nil, fmt.Errorf("failed to connect with gorm: %w", err)
	}

	migrationsPath := "file://../../internal/infra/database/migrate/migrations"
	migrator, err := migrate.New(migrationsPath, connStr)
	if err != nil {
		return pgContainer, "", db, fmt.Errorf("failed to create migrate instance: %w", err)
	}
	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		return pgContainer, "", db, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("PostgreSQL container started and database migrated.")
	return pgContainer, connStr, db, nil
}

func buildAPI(_ *testing.M) string {
	apiPath := filepath.Join(os.TempDir(), "test-api")
	log.Println("Building the API for testing...")
	buildCmd := exec.Command("go", "build", "-o", apiPath, "../../cmd/api")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		log.Fatalf("Failed to build API: %s\n%s", err, output)
	}
	log.Println("API built successfully at", apiPath)
	return apiPath
}

func waitForAPIReady(healthCheckURL string) {
	log.Println("Waiting for the API to be ready...")
	client := &http.Client{Timeout: 1 * time.Second}
	for range 20 {
		if resp, err := client.Get(healthCheckURL); err == nil && resp.StatusCode < 500 {
			log.Println("API is ready.")
			resp.Body.Close()
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	log.Fatal("API was not ready in time.")
}

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

	var videoModel video.VideoModel
	if err := db.First(&videoModel, "id = ?", videoID).Error; err == nil {
		if videoModel.URL != "" {
			videoPath := filepath.Join("..", "..", videoModel.URL)
			if err := os.Remove(videoPath); err != nil && !os.IsNotExist(err) {
				t.Logf("Failed to remove video file during cleanup: %s", videoPath)
			}
		}
		if videoModel.ThumbnailURL != nil && *videoModel.ThumbnailURL != "" {
			thumbPath := filepath.Join("..", "..", *videoModel.ThumbnailURL)
			if err := os.Remove(thumbPath); err != nil && !os.IsNotExist(err) {
				t.Logf("Failed to remove thumbnail file during cleanup: %s", thumbPath)
			}
		}
	}

	if err := db.Unscoped().Delete(&video.VideoModel{}, "id = ?", videoID).Error; err != nil {
		t.Errorf("Failed to clean up video from database with ID %s: %v", videoID, err)
	}

	t.Logf("Cleaned up resources for video with ID: %s", videoID)
}
