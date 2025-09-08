package main_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gorm_postgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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

	migrationsPath := "file://../../internal/infra/db/postgres/migrate/migrations"
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

func buildAPI() string {
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
