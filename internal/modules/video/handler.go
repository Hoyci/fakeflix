package video

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/go-chi/chi"
	"github.com/hoyci/fakeflix/packages/fault"
	"github.com/hoyci/fakeflix/packages/httputils"
	"gorm.io/gorm"
)

type Handler struct {
	service Service
	log     *log.Logger
}

func NewHTTPHandler(s Service, log *log.Logger) *Handler {
	return &Handler{service: s, log: log}
}

func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Post("/videos", h.AddVideo)
	router.Get("/videos/{videoID}", h.GetVideo)
}

func (h *Handler) AddVideo(w http.ResponseWriter, r *http.Request) {
	var dto AddVideoDTO
	dto.Title = r.FormValue("title")
	dto.Description = r.FormValue("description")
	dto.VideoFile = r.MultipartForm.File["video"]
	dto.ThumbFile = r.MultipartForm.File["thumbnail"]

	if err := dto.Validate(); err != nil {
		h.log.Warn("validation failed for add video DTO", err)
		httputils.RespondWithError(w, fault.New(
			"invalid input to add video",
			fault.WithHTTPCode(http.StatusBadRequest),
			fault.WithKind(fault.KindValidation),
			fault.WithError(err),
		),
		)
		return
	}

	err := h.service.AddVideo(r.Context(), dto)
	if err != nil {
		httputils.RespondWithError(w, err)
		return
	}

	httputils.RespondWithJSON(w, http.StatusCreated, map[string]string{"message": "created successfully"})
}

func (h *Handler) GetVideo(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "videoID")
	if videoID == "" {
		httputils.RespondWithError(w, fault.New("videoID is required", fault.WithHTTPCode(http.StatusBadRequest)))
		return
	}

	video, err := h.service.GetVideo(r.Context(), videoID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			httputils.RespondWithError(w, fault.New("video not found", fault.WithKind(fault.KindNotFound), fault.WithHTTPCode(http.StatusNotFound)))
			return
		}
		httputils.RespondWithError(w, fault.New("failed to find video", fault.WithKind(fault.KindUnexpected), fault.WithHTTPCode(http.StatusInternalServerError)))
		return
	}

	videoPath := strings.TrimPrefix(video.URL, "/")
	file, err := os.Open(videoPath)
	if err != nil {
		h.log.Error("cannot open video file", "path", videoPath, "error", err)
		httputils.RespondWithError(w, fault.New("could not open video file",
			fault.WithHTTPCode(http.StatusInternalServerError),
			fault.WithKind(fault.KindUnexpected),
		))
		return
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		h.log.Error("cannot stat video file", "path", videoPath, "error", err)
		httputils.RespondWithError(w, fault.New("could not read video file properties",
			fault.WithHTTPCode(http.StatusInternalServerError),
			fault.WithKind(fault.KindUnexpected),
		))
		return
	}

	http.ServeContent(w, r, file.Name(), fileStat.ModTime(), file)
}
