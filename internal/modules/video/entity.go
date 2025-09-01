package video

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Video struct {
	id           string `gorm:"type:uuid;primary_key"`
	title        string
	description  string
	url          string
	sizeInKB     int
	duration     int
	thumbnailURL *string
	createdAt    time.Time
	updatedAt    time.Time
}

func NewVideo(id, title, description, url string, thumbnailURL *string, sizeInKB, duration int) (*Video, error) {
	video := &Video{
		id:           id,
		title:        title,
		description:  description,
		url:          url,
		thumbnailURL: thumbnailURL,
		sizeInKB:     sizeInKB,
		duration:     duration,
		createdAt:    time.Now().UTC(),
		updatedAt:    time.Now().UTC(),
	}

	if err := video.validate(); err != nil {
		return nil, err
	}

	return video, nil
}

func (v *Video) validate() error {
	err := validation.ValidateStruct(v,
		validation.Field(&v.title, validation.Required.Error("title is required"), validation.Length(1, 255)),
		validation.Field(&v.description, validation.Required.Error("description is required"), validation.Length(1, 500)),
		validation.Field(&v.url, validation.Required.Error("url is required"), validation.Length(1, 255)),
		validation.Field(&v.sizeInKB, validation.Required.Error("sizeInKB is required"), validation.Min(1)),
		validation.Field(&v.duration, validation.Required.Error("duration is required"), validation.Min(1)),
	)

	return err
}

func (v *Video) ID() string            { return v.id }
func (v *Video) Title() string         { return v.title }
func (v *Video) Description() string   { return v.description }
func (v *Video) URL() string           { return v.url }
func (v *Video) ThumbnailURL() *string { return v.thumbnailURL }
func (v *Video) SizeInKB() int         { return v.sizeInKB }
func (v *Video) Duration() int         { return v.duration }
func (v *Video) CreatedAt() time.Time  { return v.createdAt }
func (v *Video) UpdatedAt() time.Time  { return v.updatedAt }
