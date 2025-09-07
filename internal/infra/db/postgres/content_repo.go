package postgres

import (
	"context"
	"errors"

	"github.com/hoyci/fakeflix/internal/domain/content"
	"github.com/hoyci/fakeflix/internal/domain/movie"
	"github.com/hoyci/fakeflix/internal/domain/thumbnail"
	"github.com/hoyci/fakeflix/internal/domain/video"
	"gorm.io/gorm"
)

type contentRepository struct {
	db *gorm.DB
}

func NewContentRepository(db *gorm.DB) content.Repository {
	return &contentRepository{db: db}
}

func (r *contentRepository) Save(ctx context.Context, contentEntity *content.Content) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer tx.Rollback()

	movieEntity, _ := contentEntity.Movie()
	videoEntity := movieEntity.Video()
	thumbnailEntity := movieEntity.Thumbnail()

	videoModel := VideoModel{
		ID:        videoEntity.ID(),
		URL:       videoEntity.URL(),
		SizeInKb:  videoEntity.SizeInKB(),
		Duration:  videoEntity.Duration(),
		CreatedAt: videoEntity.CreatedAt(),
		UpdatedAt: videoEntity.UpdatedAt(),
	}
	if err := tx.Create(&videoModel).Error; err != nil {
		return err
	}

	var thumbnailID *string
	if thumbnailEntity != nil {
		thumbnailModel := ThumbnailModel{
			ID:        thumbnailEntity.ID(),
			URL:       thumbnailEntity.URL(),
			CreatedAt: thumbnailEntity.CreatedAt(),
			UpdatedAt: thumbnailEntity.UpdatedAt(),
		}
		if err := tx.Create(&thumbnailModel).Error; err != nil {
			return err
		}
		thumbnailID = &thumbnailModel.ID
	}

	contentModel := ContentModel{
		ID:          contentEntity.ID(),
		Title:       contentEntity.Title(),
		Description: contentEntity.Description(),
		ContentType: contentEntity.ContentType(),
		CreatedAt:   contentEntity.CreatedAt(),
		UpdatedAt:   contentEntity.UpdatedAt(),
	}
	if err := tx.Create(&contentModel).Error; err != nil {
		return err
	}

	movieModel := MovieModel{
		ID:          movieEntity.ID(),
		ContentID:   contentEntity.ID(),
		VideoID:     videoEntity.ID(),
		ThumbnailID: thumbnailID,
		CreatedAt:   movieEntity.CreatedAt(),
		UpdatedAt:   movieEntity.UpdatedAt(),
	}
	if err := tx.Create(&movieModel).Error; err != nil {
		return err
	}

	return tx.Commit().Error
}

func (r *contentRepository) FindByID(ctx context.Context, id string) (*content.Content, error) {
	var model ContentModel

	err := r.db.WithContext(ctx).
		Preload("Movie.Video").
		Preload("Movie.Thumbnail").
		// Preload("TvShow.Episodes.Video"). // Exemplo para quando implementar TvShow
		// Preload("TvShow.Episodes.Thumbnail").
		First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, content.ErrNotFound
		}
		return nil, err
	}

	return toDomainContent(&model)
}

func toDomainContent(model *ContentModel) (*content.Content, error) {
	var media content.Media
	var err error

	switch model.ContentType {
	case content.MovieType:
		if model.Movie != nil {
			media, err = toDomainMovie(model.Movie)
			if err != nil {
				return nil, err
			}
		}
	case content.TvShowType:
		// Lógica para TvShow viria aqui
	}

	return content.HydrateContent(
		model.ID,
		model.Title,
		model.Description,
		model.ContentType,
		media,
		model.CreatedAt,
		model.UpdatedAt,
	), nil
}

func toDomainMovie(model *MovieModel) (*movie.Movie, error) {
	videoEntity := toDomainVideo(&model.Video)

	var thumbnailEntity *thumbnail.Thumbnail
	if model.Thumbnail != nil {
		thumbnailEntity = toDomainThumbnail(model.Thumbnail)
	}

	return movie.HydrateMovie(
		model.ID,
		videoEntity,
		thumbnailEntity,
		model.CreatedAt,
		model.UpdatedAt,
	), nil
}

func toDomainVideo(model *VideoModel) *video.Video {
	return video.HydrateVideo(
		model.ID,
		model.URL,
		model.SizeInKb,
		model.Duration,
		model.CreatedAt,
		model.UpdatedAt,
	)
}

func toDomainThumbnail(model *ThumbnailModel) *thumbnail.Thumbnail {
	return thumbnail.HydrateThumbnail(
		model.ID,
		model.URL,
		model.CreatedAt,
		model.UpdatedAt,
	)
}
