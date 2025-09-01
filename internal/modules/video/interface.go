package video

import "context"

type Repository interface {
	CreateVideo(ctx context.Context, video *Video) error
	GetVideoByID(ctx context.Context, videoID string) (*VideoModel, error)
}

type Service interface {
	AddVideo(ctx context.Context, dto AddVideoDTO) error
	GetVideo(ctx context.Context, videoID string) (*VideoModel, error)
}
