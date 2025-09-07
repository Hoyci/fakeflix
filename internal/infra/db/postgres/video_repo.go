package postgres

import (
	"context"
	"errors"

	"github.com/hoyci/fakeflix/internal/domain/video"
	"gorm.io/gorm"
)

type videoRepository struct {
	db *gorm.DB
}

func NewVideoRepository(db *gorm.DB) video.Repository {
	return &videoRepository{db: db}
}

func (r *videoRepository) FindByID(ctx context.Context, id string) (*video.Video, error) {
	var model VideoModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, video.ErrNotFound
		}
		return nil, err
	}

	return toDomainVideo(&model), nil
}
