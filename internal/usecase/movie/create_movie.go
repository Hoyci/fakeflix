package movie

import (
	"context"
	"mime/multipart"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/hoyci/fakeflix/internal/domain/content"
	"github.com/hoyci/fakeflix/internal/domain/movie"
	"github.com/hoyci/fakeflix/internal/domain/thumbnail"
	"github.com/hoyci/fakeflix/internal/domain/video"
	"github.com/hoyci/fakeflix/internal/infra/media"
	"github.com/hoyci/fakeflix/pkg/fault"
)

type CreateMovieInputDTO struct {
	Title       string
	Description string
	Video       *multipart.FileHeader
	Thumbnail   *multipart.FileHeader
}

func (req CreateMovieInputDTO) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Title, validation.Required.Error("title is required"), validation.Length(1, 255)),
		validation.Field(&req.Description, validation.Required.Error("description is required")),
		validation.Field(&req.Video, validation.Required.Error("video file is required")),
	)
}

type CreateMovieOutputDTO struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

type CreateMovieUseCase struct {
	contentRepo  content.Repository
	mediaService media.MediaService
}

func NewCreateMovieUseCase(contentRepo content.Repository, mediaService media.MediaService) *CreateMovieUseCase {
	return &CreateMovieUseCase{
		contentRepo:  contentRepo,
		mediaService: mediaService,
	}
}

func (uc *CreateMovieUseCase) Execute(ctx context.Context, input CreateMovieInputDTO) (*CreateMovieOutputDTO, error) {
	videoInfo, err := uc.mediaService.Store(input.Video, "upload/videos")
	if err != nil {
		return nil, fault.New(
			"error while saving video",
			fault.WithKind(fault.KindUnexpected),
			fault.WithError(err),
		)
	}

	var thumbInfo *media.StoredFileInfo
	if input.Thumbnail != nil {
		thumbInfo, err = uc.mediaService.Store(input.Thumbnail, "upload/thumbs")
		if err != nil {
			return nil, fault.New(
				"error while saving thumbnail",
				fault.WithKind(fault.KindUnexpected),
				fault.WithError(err),
			)
		}
	}

	videoEntity, err := video.NewVideo(videoInfo.URL, videoInfo.SizeInKb, videoInfo.Duration)
	if err != nil {
		return nil, fault.New(
			"invalid input for video",
			fault.WithKind(fault.KindValidation),
			fault.WithError(err))
	}

	movieEntity, err := movie.NewMovie(videoEntity)
	if err != nil {
		return nil, fault.New(
			"failed to create movie entity",
			fault.WithKind(fault.KindValidation),
			fault.WithError(err),
		)
	}

	if thumbInfo != nil {
		thumbnailEntity, err := thumbnail.NewThumbnail(thumbInfo.URL)
		if err != nil {
			return nil, fault.New(
				"invalid input for thumbnail",
				fault.WithKind(fault.KindValidation),
				fault.WithError(err),
			)
		}
		if err := movieEntity.AddThumbnail(thumbnailEntity); err != nil {
			return nil, fault.New(
				"failed to add thumbnail to movie",
				fault.WithKind(fault.KindValidation),
				fault.WithError(err),
			)
		}
	}

	contentEntity, err := content.NewContent(input.Title, input.Description, content.MovieType, movieEntity)
	if err != nil {
		return nil, fault.New(
			"failed to create content entity",
			fault.WithKind(fault.KindValidation),
			fault.WithError(err),
		)
	}

	err = uc.contentRepo.Save(ctx, contentEntity)
	if err != nil {
		return nil, fault.New(
			"failed to save movie",
			fault.WithKind(fault.KindUnexpected),
			fault.WithError(err),
		)
	}

	return &CreateMovieOutputDTO{
		ID:          contentEntity.ID(),
		Title:       contentEntity.Title(),
		Description: contentEntity.Description(),
		CreatedAt:   contentEntity.CreatedAt().String(),
	}, nil
}
