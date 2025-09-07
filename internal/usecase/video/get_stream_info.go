package video

import (
	"context"
	"strings"

	"github.com/charmbracelet/log"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/hoyci/fakeflix/internal/domain/video"
	"github.com/hoyci/fakeflix/pkg/fault"
)

type GetStreamInfoInputDTO struct {
	VideoID string
}

func (req GetStreamInfoInputDTO) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.VideoID, validation.Required.Error("videoID is required")),
	)
}

type GetStreamInfoOutputDTO struct {
	FilePath string
	FileSize int
}

type GetStreamInfoUseCase struct {
	videoRepo video.Repository
	logger    *log.Logger
}

func NewGetStreamInfoUseCase(videoRepo video.Repository, logger *log.Logger) *GetStreamInfoUseCase {
	return &GetStreamInfoUseCase{
		videoRepo: videoRepo,
		logger:    logger,
	}
}

func (uc *GetStreamInfoUseCase) Execute(ctx context.Context, input GetStreamInfoInputDTO) (*GetStreamInfoOutputDTO, error) {
	uc.logger.Debug("Starting stream execution", "videoID", input.VideoID)

	videoEntity, err := uc.videoRepo.FindByID(ctx, input.VideoID)
	if err != nil {
		uc.logger.Error("Failed to find video by videoID", "videoID", input.VideoID, "error", err)
		return nil, fault.New(
			"video not found",
			fault.WithKind(fault.KindNotFound),
			fault.WithError(err),
		)
	}
	uc.logger.Debug("Video founded", "url", videoEntity.URL())

	filePath := strings.TrimPrefix(videoEntity.URL(), "/")

	return &GetStreamInfoOutputDTO{
		FilePath: filePath,
		FileSize: videoEntity.SizeInKB(),
	}, nil
}
