package video

import (
	"context"
	"strings"

	"github.com/hoyci/fakeflix/internal/domain/video"
	"github.com/hoyci/fakeflix/pkg/fault"
)

type GetStreamInfoInputDTO struct {
	VideoID string
}

type GetStreamInfoOutputDTO struct {
	FilePath string
	FileSize int
}

type GetStreamInfoUseCase struct {
	videoRepo video.Repository
}

func NewGetStreamInfoUseCase(videoRepo video.Repository) *GetStreamInfoUseCase {
	return &GetStreamInfoUseCase{
		videoRepo: videoRepo,
	}
}

func (uc *GetStreamInfoUseCase) Execute(ctx context.Context, input GetStreamInfoInputDTO) (*GetStreamInfoOutputDTO, error) {
	videoEntity, err := uc.videoRepo.FindByID(ctx, input.VideoID)
	if err != nil {
		return nil, fault.New(
			"video not found",
			fault.WithKind(fault.KindNotFound),
			fault.WithError(err),
		)
	}

	filePath := strings.TrimPrefix(videoEntity.URL(), "/")

	return &GetStreamInfoOutputDTO{
		FilePath: filePath,
		FileSize: videoEntity.SizeInKB(),
	}, nil
}
