package content

import (
	"context"
	"errors"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type Repository interface {
	Save(ctx context.Context, content *Content) error
	// FindByID(ctx context.Context, id string) (*Content, error)
}
