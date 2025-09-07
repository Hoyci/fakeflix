package episode

import (
	"errors"
	"time"

	"github.com/hoyci/fakeflix/internal/domain/thumbnail"
	"github.com/hoyci/fakeflix/internal/domain/video"

	"github.com/google/uuid"
)

type Episode struct {
	id          string
	title       string
	description string
	season      int
	number      int
	video       *video.Video
	thumbnail   *thumbnail.Thumbnail
	createdAt   time.Time
	updatedAt   time.Time
}

func NewEpisode(title, description string, season, number int, vid *video.Video) (*Episode, error) {
	if title == "" {
		return nil, errors.New("episode title is required")
	}
	if season <= 0 {
		return nil, errors.New("episode season must be a positive number")
	}
	if number <= 0 {
		return nil, errors.New("episode number must be a positive number")
	}
	if vid == nil {
		return nil, errors.New("episode must have a video")
	}

	return &Episode{
		id:          uuid.NewString(),
		title:       title,
		description: description,
		season:      season,
		number:      number,
		video:       vid,
		createdAt:   time.Now().UTC(),
		updatedAt:   time.Now().UTC(),
	}, nil
}

func HydrateEpisode(id, title, description string, season, number int, vid *video.Video, thumb *thumbnail.Thumbnail, createdAt, updatedAt time.Time) *Episode {
	return &Episode{
		id:          id,
		title:       title,
		description: description,
		season:      season,
		number:      number,
		video:       vid,
		thumbnail:   thumb,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

func (e *Episode) AddThumbnail(thumb *thumbnail.Thumbnail) error {
	if thumb == nil {
		return errors.New("cannot add a nil thumbnail")
	}
	e.thumbnail = thumb
	e.updatedAt = time.Now().UTC()
	return nil
}

func (e *Episode) ID() string                      { return e.id }
func (e *Episode) Title() string                   { return e.title }
func (e *Episode) Description() string             { return e.description }
func (e *Episode) Season() int                     { return e.season }
func (e *Episode) Number() int                     { return e.number }
func (e *Episode) Video() *video.Video             { return e.video }
func (e *Episode) Thumbnail() *thumbnail.Thumbnail { return e.thumbnail }
