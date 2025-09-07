package http

import (
	"net/http"
	"path/filepath"

	"github.com/charmbracelet/log"

	"github.com/go-chi/chi"
	"github.com/hoyci/fakeflix/internal/infra/media"
	"github.com/hoyci/fakeflix/internal/usecase/video"
	"github.com/hoyci/fakeflix/pkg/fault"
	"github.com/hoyci/fakeflix/pkg/httputils"
)

type VideoHandler struct {
	getStreamInfoUseCase *video.GetStreamInfoUseCase
	mediaService         media.MediaService
	logger               *log.Logger
}

func NewVideoHandler(uc *video.GetStreamInfoUseCase, ms media.MediaService, logger *log.Logger) *VideoHandler {
	return &VideoHandler{
		getStreamInfoUseCase: uc,
		mediaService:         ms,
		logger:               logger,
	}
}

func (h *VideoHandler) StreamVideo(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "videoID")

	requestDTO := video.GetStreamInfoInputDTO{
		VideoID: videoID,
	}

	if err := requestDTO.Validate(); err != nil {
		h.logger.Warn("Request validation failed", "error", err)
		httputils.RespondWithError(w, fault.New(
			err.Error(),
			fault.WithKind(fault.KindValidation),
		))
		return
	}

	h.logger.Info("Received streaming request", "videoID", requestDTO.VideoID, "range_header", r.Header.Get("Range"))

	output, err := h.getStreamInfoUseCase.Execute(r.Context(), requestDTO)
	if err != nil {
		httputils.RespondWithError(w, err)
		return
	}

	file, fileStat, err := h.mediaService.GetStream(output.FilePath)
	if err != nil {
		httputils.RespondWithError(w, err)
		return
	}
	defer file.Close()

	http.ServeContent(w, r, filepath.Base(output.FilePath), fileStat.ModTime(), file)
}
