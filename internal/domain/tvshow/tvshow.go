package tvshow

import (
	"errors"
	"fmt"
	"time"

	"github.com/hoyci/fakeflix/internal/domain/episode"
	"github.com/hoyci/fakeflix/internal/domain/thumbnail"

	"github.com/google/uuid"
)

type TvShow struct {
	id        string
	thumbnail *thumbnail.Thumbnail
	episodes  []*episode.Episode
	createdAt time.Time
	updatedAt time.Time
}

func NewTvShow() (*TvShow, error) {
	return &TvShow{
		id:        uuid.NewString(),
		episodes:  make([]*episode.Episode, 0),
		createdAt: time.Now().UTC(),
		updatedAt: time.Now().UTC(),
	}, nil
}

func HydrateTvShow(id string, thumb *thumbnail.Thumbnail, episodes []*episode.Episode, createdAt, updatedAt time.Time) *TvShow {
	return &TvShow{
		id:        id,
		thumbnail: thumb,
		episodes:  episodes,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (t *TvShow) AddEpisode(newEpisode *episode.Episode) error {
	if newEpisode == nil {
		return errors.New("cannot add a nil episode")
	}

	for _, ep := range t.episodes {
		if ep.Season() == newEpisode.Season() && ep.Number() == newEpisode.Number() {
			return fmt.Errorf("episode S%02dE%02d already exists", newEpisode.Season(), newEpisode.Number())
		}
	}

	t.episodes = append(t.episodes, newEpisode)
	t.updatedAt = time.Now().UTC()
	return nil
}

func (t *TvShow) AddThumbnail(thumb *thumbnail.Thumbnail) error {
	if thumb == nil {
		return errors.New("cannot add a nil thumbnail")
	}
	t.thumbnail = thumb
	t.updatedAt = time.Now().UTC()
	return nil
}

func (t *TvShow) ID() string                      { return t.id }
func (t *TvShow) Thumbnail() *thumbnail.Thumbnail { return t.thumbnail }
func (t *TvShow) Episodes() []*episode.Episode    { return t.episodes }
func (t *TvShow) IsMedia()                        {}
