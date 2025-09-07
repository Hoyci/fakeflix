package content

import (
	"errors"
	"fmt"
	"time"

	"github.com/hoyci/fakeflix/internal/domain/movie"
	"github.com/hoyci/fakeflix/internal/domain/tvshow"

	"github.com/google/uuid"
)

type ContentType string

const (
	MovieType  ContentType = "MOVIE"
	TvShowType ContentType = "TV_SHOW"
)

func (ct ContentType) IsValid() bool {
	switch ct {
	case MovieType, TvShowType:
		return true
	}
	return false
}

type Media interface {
	IsMedia()
}

type Content struct {
	id          string
	title       string
	description string
	contentType ContentType
	media       Media
	createdAt   time.Time
	updatedAt   time.Time
}

func NewContent(title, description string, contentType ContentType, media Media) (*Content, error) {
	if title == "" {
		return nil, errors.New("content title is required")
	}
	if !contentType.IsValid() {
		return nil, errors.New("invalid content type")
	}
	if media == nil {
		return nil, errors.New("content media is required")
	}

	switch contentType {
	case MovieType:
		if _, ok := media.(*movie.Movie); !ok {
			return nil, fmt.Errorf("media type mismatch: expected %s", MovieType)
		}
	case TvShowType:
		if _, ok := media.(*tvshow.TvShow); !ok {
			return nil, fmt.Errorf("media type mismatch: expected %s", TvShowType)
		}
	}

	return &Content{
		id:          uuid.NewString(),
		title:       title,
		description: description,
		contentType: contentType,
		media:       media,
		createdAt:   time.Now().UTC(),
		updatedAt:   time.Now().UTC(),
	}, nil
}

func HydrateContent(id, title, description string, contentType ContentType, media Media, createdAt, updatedAt time.Time) *Content {
	return &Content{
		id:          id,
		title:       title,
		description: description,
		contentType: contentType,
		media:       media,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

func (c *Content) ChangeTitle(newTitle string) error {
	if newTitle == "" {
		return errors.New("title cannot be empty")
	}
	c.title = newTitle
	c.updatedAt = time.Now().UTC()
	return nil
}

func (c *Content) ID() string               { return c.id }
func (c *Content) Title() string            { return c.title }
func (c *Content) Description() string      { return c.description }
func (c *Content) ContentType() ContentType { return c.contentType }
func (c *Content) CreatedAt() time.Time     { return c.createdAt }
func (c *Content) UpdatedAt() time.Time     { return c.updatedAt }

func (c *Content) Movie() (*movie.Movie, error) {
	if c.contentType != MovieType {
		return nil, errors.New("content is not a movie")
	}
	mov, _ := c.media.(*movie.Movie)
	return mov, nil
}

func (c *Content) TvShow() (*tvshow.TvShow, error) {
	if c.contentType != TvShowType {
		return nil, errors.New("content is not a tv show")
	}
	tv, _ := c.media.(*tvshow.TvShow)
	return tv, nil
}
