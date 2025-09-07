package postgres

import (
	"context"
	"errors"

	"github.com/charmbracelet/log"
	"github.com/hoyci/fakeflix/internal/domain/video"
	"gorm.io/gorm"
)

type videoRepository struct {
	db     *gorm.DB
	logger *log.Logger
}

func NewVideoRepository(db *gorm.DB, logger *log.Logger) video.Repository {
	return &videoRepository{db: db, logger: logger}
}

func (r *videoRepository) FindByID(ctx context.Context, id string) (*video.Video, error) {
	log := r.logger.With("videoID", id)
	log.Debug("Start video search")
	var model VideoModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		log.Error("Failed to find video by ID", "error", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, video.ErrNotFound
		}
		return nil, err
	}

	log.Debug("Video founded", "videoURL", model.URL)
	return toDomainVideo(&model), nil
}
