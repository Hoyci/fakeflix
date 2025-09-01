package video

import (
	"fmt"
	"mime/multipart"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type AddVideoDTO struct {
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	VideoFile   []*multipart.FileHeader `form:"video"`
	ThumbFile   []*multipart.FileHeader `form:"thumbnail"`
}

func (req AddVideoDTO) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Title,
			validation.Required.Error("the title is required"),
			validation.Length(1, 255),
		),
		validation.Field(&req.Description,
			validation.Required.Error("the description is required"),
			validation.Length(0, 500),
		),
		validation.Field(&req.VideoFile,
			validation.Required.Error("a video file is required"),
			validation.Length(1, 1).Error("only one video file is allowed"),
			withMimeTypes("video/mp4").Error("the video must be an MP4 file"),
		),
		validation.Field(&req.ThumbFile,
			validation.Required.Error("a thumbnail file is required"),
			validation.Length(1, 1).Error("only one thumbnail image file is allowed"),
			withMimeTypes("image/jpeg", "image/png").Error("the thumbnail must be a JPEG or PNG file"),
		),
	)
}

type mimeTypeRule struct {
	allowedTypes []string
	message      string
}

func withMimeTypes(allowedTypes ...string) *mimeTypeRule {
	return &mimeTypeRule{
		allowedTypes: allowedTypes,
	}
}

func (r *mimeTypeRule) Error(message string) *mimeTypeRule {
	r.message = message
	return r
}

func (r *mimeTypeRule) Validate(value interface{}) error {
	files, ok := value.([]*multipart.FileHeader)
	if !ok {
		return fmt.Errorf("field must be a slice of *multipart.FileHeader")
	}

	if len(files) == 0 {
		return nil
	}
	fileHeader := files[0]

	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return fmt.Errorf("could not read file: %w", err)
	}

	contentType := http.DetectContentType(buffer)
	for _, allowedType := range r.allowedTypes {
		if contentType == allowedType {
			return nil // Validation successful
		}
	}

	if r.message != "" {
		return fmt.Errorf(r.message)
	}
	return fmt.Errorf("invalid file type. Received: %s", contentType)
}
