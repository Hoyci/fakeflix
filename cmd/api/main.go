package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/hoyci/fakeflix/internal/config"
	pg "github.com/hoyci/fakeflix/internal/infra/database"
	"github.com/hoyci/fakeflix/internal/infra/logger"
	"github.com/hoyci/fakeflix/internal/modules/video"
)

func main() {
	cfg := config.GetConfig()

	appLogger := logger.NewLogger(cfg)
	appLogger.Info("starting bookday application", "app_name", cfg.AppName, "env", cfg.Environment)

	db, err := pg.NewConnection(cfg)
	if err != nil {
		appLogger.Fatal("could not connect to the database", "error", err)
	}
	appLogger.Info("database connection established")

	sqlDB, err := db.DB()
	if err != nil {
		appLogger.Fatal("could not get underlying sql.DB from gorm", "error", err)
	}
	defer sqlDB.Close()

	router := chi.NewRouter()

	repo := video.NewRepository(db)
	service := video.NewService(repo, appLogger)
	handler := video.NewHTTPHandler(service, appLogger)

	handler.RegisterRoutes(router)

	listenAddr := fmt.Sprintf(":%d", cfg.Port)
	appLogger.Info("server is starting", "address", listenAddr)
	if err := http.ListenAndServe(listenAddr, router); err != nil {
		appLogger.Fatal("failed to start server", "error", err)
	}
}
