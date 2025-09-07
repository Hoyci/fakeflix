package thumbnail

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Thumbnail struct {
	id        string
	url       string
	createdAt time.Time
	updatedAt time.Time
}

func NewThumbnail(url string) (*Thumbnail, error) {
	if url == "" {
		return nil, errors.New("thumbnail url is required")
	}

	return &Thumbnail{
		id:        uuid.NewString(),
		url:       url,
		createdAt: time.Now().UTC(),
		updatedAt: time.Now().UTC(),
	}, nil
}

func HydrateThumbnail(id, url string, createdAt, updatedAt time.Time) *Thumbnail {
	return &Thumbnail{
		id:        id,
		url:       url,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (v *Thumbnail) ID() string {
	return v.id
}

func (v *Thumbnail) URL() string {
	return v.url
}

func (v *Thumbnail) CreatedAt() time.Time {
	return v.createdAt
}

func (v *Thumbnail) UpdatedAt() time.Time {
	return v.updatedAt
}
