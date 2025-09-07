package movie

import (
	"errors"
	"time"

	"github.com/hoyci/fakeflix/internal/domain/thumbnail"
	"github.com/hoyci/fakeflix/internal/domain/video"

	"github.com/google/uuid"
)

type Movie struct {
	id        string
	video     *video.Video
	thumbnail *thumbnail.Thumbnail
	createdAt time.Time
	updatedAt time.Time
}

func NewMovie(vid *video.Video) (*Movie, error) {
	if vid == nil {
		return nil, errors.New("a movie must have a video")
	}

	return &Movie{
		id:        uuid.NewString(),
		video:     vid,
		createdAt: time.Now().UTC(),
		updatedAt: time.Now().UTC(),
	}, nil
}

func HydrateMovie(id string, vid *video.Video, thumb *thumbnail.Thumbnail, createdAt, updatedAt time.Time) *Movie {
	return &Movie{
		id:        id,
		video:     vid,
		thumbnail: thumb,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (m *Movie) AddThumbnail(thumb *thumbnail.Thumbnail) error {
	if thumb == nil {
		return errors.New("cannot add a nil thumbnail")
	}
	m.thumbnail = thumb
	m.updatedAt = time.Now().UTC()
	return nil
}

func (m *Movie) ID() string {
	return m.id
}

func (m *Movie) Video() *video.Video {
	return m.video
}

func (m *Movie) Thumbnail() *thumbnail.Thumbnail {
	return m.thumbnail
}

func (m *Movie) CreatedAt() time.Time {
	return m.createdAt
}

func (m *Movie) UpdatedAt() time.Time {
	return m.updatedAt
}
func (m *Movie) IsMedia() {}
