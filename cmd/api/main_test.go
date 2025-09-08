package main_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"testing"

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

	dbHost, _ := pgContainer.Host(ctx)
	dbPort, _ := pgContainer.MappedPort(ctx, "5432")

	const apiPort = "3030"
	baseAPIURL = fmt.Sprintf("http://localhost:%s", apiPort)

	apiPath := buildAPI()
	defer os.Remove(apiPath)

	apiCmd := exec.Command(apiPath)
	apiCmd.Dir = "../.."
	apiCmd.Env = append(os.Environ(),
		fmt.Sprintf("APP_PORT=%s", apiPort),
		"DB_HOST="+dbHost,
		"DB_PORT="+dbPort.Port(),
		"DB_USERNAME=user",
		"DB_PASSWORD=password",
		"DB_DATABASE=test-db-e2e",
		"APP_ENV=testing",
	)
	apiCmd.Stdout = os.Stdout
	apiCmd.Stderr = os.Stderr

	if err := apiCmd.Start(); err != nil {
		log.Fatalf("Failed to start API: %v", err)
	}
	log.Printf("API started with PID %d, connected to Testcontainer", apiCmd.Process.Pid)

	waitForAPIReady(baseAPIURL + "/videos")

	exitCode := m.Run()

	log.Printf("Terminating API process (PID %d)...", apiCmd.Process.Pid)
	if err := apiCmd.Process.Signal(syscall.SIGINT); err != nil {
		log.Printf("Failed to send SIGINT, forcing shutdown...")
		apiCmd.Process.Kill()
	}
	apiCmd.Wait()

	os.Exit(exitCode)
}
