package video

import (
	"context"

	"gorm.io/gorm"
)

type videoRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &videoRepository{db: db}
}

func (repo *videoRepository) CreateVideo(ctx context.Context, video *Video) error {
	videoModel := VideoModel{
		ID:           video.ID(),
		Title:        video.Title(),
		Description:  video.Description(),
		URL:          video.URL(),
		SizeInKB:     video.SizeInKB(),
		Duration:     video.Duration(),
		ThumbnailURL: video.ThumbnailURL(),
		CreatedAt:    video.CreatedAt(),
		UpdatedAt:    video.UpdatedAt(),
	}

	return repo.db.WithContext(ctx).Create(&videoModel).Error
}

func (repo *videoRepository) GetVideoByID(ctx context.Context, videoID string) (*VideoModel, error) {
	var video VideoModel
	result := repo.db.WithContext(ctx).Where("id = ?", videoID).First(&video)
	if result.Error != nil {
		return nil, result.Error
	}

	return &video, nil
}
