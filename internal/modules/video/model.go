package video

import "time"

type VideoModel struct {
	ID           string `gorm:"type:uuid;primary_key"`
	Title        string
	Description  string
	URL          string
	SizeInKB     int
	Duration     int
	ThumbnailURL *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (VideoModel) TableName() string {
	return "videos"
}
