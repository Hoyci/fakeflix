package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/hoyci/fakeflix/internal/infra/config"
	"github.com/hoyci/fakeflix/internal/infra/db/postgres"
	"github.com/hoyci/fakeflix/internal/infra/logger"
	"github.com/hoyci/fakeflix/internal/infra/media"
	httphandler "github.com/hoyci/fakeflix/internal/interface/http"
	"github.com/hoyci/fakeflix/internal/usecase/movie"
	videousecase "github.com/hoyci/fakeflix/internal/usecase/video"
)

func main() {
	cfg := config.GetConfig()

	appLogger := logger.NewLogger(cfg)
	appLogger.Info(fmt.Sprintf("starting %s application", cfg.AppName), "env", cfg.Environment)

	db, err := postgres.NewConnection(cfg)
	if err != nil {
		appLogger.Fatal("could not connect to the database", "error", err)
	}
	appLogger.Info("database connection established")

	sqlDB, err := db.DB()
	if err != nil {
		appLogger.Fatal("could not get underlying sql.DB from gorm", "error", err)
	}
	defer sqlDB.Close()

	contentRepo := postgres.NewContentRepository(db, appLogger)
	videoRepo := postgres.NewVideoRepository(db, appLogger)
	mediaService := media.NewLocalMediaService(appLogger)

	createMovieUseCase := movie.NewCreateMovieUseCase(contentRepo, mediaService, appLogger)
	getStreamInfoUseCase := videousecase.NewGetStreamInfoUseCase(videoRepo, appLogger)

	movieHandler := httphandler.NewMovieHandler(createMovieUseCase, appLogger)
	videoHandler := httphandler.NewVideoHandler(getStreamInfoUseCase, mediaService, appLogger)

	router := chi.NewRouter()
	router.Post("/movies", movieHandler.CreateMovie)
	router.Get("/videos/{videoID}/stream", videoHandler.StreamVideo)

	listenAddr := fmt.Sprintf(":%d", cfg.Port)
	appLogger.Info("server is starting", "address", listenAddr)
	if err := http.ListenAndServe(listenAddr, router); err != nil {
		appLogger.Fatal("failed to start server", "error", err)
	}
}
