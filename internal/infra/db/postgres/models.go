package postgres

import (
	"time"

	"github.com/hoyci/fakeflix/internal/domain/content"

	"gorm.io/gorm"
)

type ContentModel struct {
	ID          string `gorm:"type:uuid;primary_key"`
	Title       string
	Description string
	ContentType content.ContentType `gorm:"type:varchar(50)"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	Movie  *MovieModel  `gorm:"foreignKey:ContentID"`
	TvShow *TvShowModel `gorm:"foreignKey:ContentID"`
}

type MovieModel struct {
	ID          string  `gorm:"type:uuid;primary_key"`
	ContentID   string  `gorm:"type:uuid;unique;not null"`
	VideoID     string  `gorm:"type:uuid;unique;not null"`
	ThumbnailID *string `gorm:"type:uuid;unique"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	Video     VideoModel      `gorm:"foreignKey:VideoID"`
	Thumbnail *ThumbnailModel `gorm:"foreignKey:ThumbnailID"`
}

type TvShowModel struct {
	ID          string  `gorm:"type:uuid;primary_key"`
	ContentID   string  `gorm:"type:uuid;unique;not null"`
	ThumbnailID *string `gorm:"type:uuid;unique"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	Thumbnail *ThumbnailModel `gorm:"foreignKey:ThumbnailID"`
	Episodes  []*EpisodeModel `gorm:"foreignKey:TvShowID"`
}

type EpisodeModel struct {
	ID          string  `gorm:"type:uuid;primary_key"`
	TvShowID    string  `gorm:"type:uuid;not null"`
	ThumbnailID *string `gorm:"type:uuid;unique"`
	VideoID     string  `gorm:"type:uuid;unique;not null"`
	Title       string
	Description string
	Season      int
	Number      int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	Thumbnail *ThumbnailModel `gorm:"foreignKey:ThumbnailID"`
	Video     VideoModel      `gorm:"foreignKey:VideoID"`
}

type VideoModel struct {
	ID        string `gorm:"type:uuid;primary_key"`
	URL       string
	SizeInKb  int
	Duration  int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type ThumbnailModel struct {
	ID        string `gorm:"type:uuid;primary_key"`
	URL       string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (ContentModel) TableName() string {
	return "contents"
}

func (MovieModel) TableName() string {
	return "movies"
}

func (TvShowModel) TableName() string {
	return "tv_shows"
}

func (EpisodeModel) TableName() string {
	return "episodes"
}

func (VideoModel) TableName() string {
	return "videos"
}

func (ThumbnailModel) TableName() string {
	return "thumbnails"
}
