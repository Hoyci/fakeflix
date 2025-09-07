package video

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("video not found")

type Repository interface {
	FindByID(ctx context.Context, id string) (*Video, error)
}
