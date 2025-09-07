package video

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Video struct {
	id        string
	url       string
	sizeInKB  int
	duration  int
	createdAt time.Time
	updatedAt time.Time
}

func NewVideo(url string, sizeInKB, duration int) (*Video, error) {
	if url == "" {
		return nil, errors.New("video url is required")
	}

	if sizeInKB <= 0 {
		return nil, errors.New("video size must be positive")
	}

	if duration <= 0 {
		return nil, errors.New("video duration must be positive")
	}

	return &Video{
		id:        uuid.NewString(),
		url:       url,
		sizeInKB:  sizeInKB,
		duration:  duration,
		createdAt: time.Now().UTC(),
		updatedAt: time.Now().UTC(),
	}, nil
}

func HydrateVideo(id, url string, sizeInKB, duration int, createdAt, updatedAt time.Time) *Video {
	return &Video{
		id:        id,
		url:       url,
		sizeInKB:  sizeInKB,
		duration:  duration,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (v *Video) ID() string {
	return v.id
}

func (v *Video) URL() string {
	return v.url
}

func (v *Video) SizeInKB() int {
	return v.sizeInKB
}

func (v *Video) Duration() int {
	return v.duration
}

func (v *Video) CreatedAt() time.Time {
	return v.createdAt
}

func (v *Video) UpdatedAt() time.Time {
	return v.updatedAt
}
