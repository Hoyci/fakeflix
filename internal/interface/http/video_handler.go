package http

import (
	"net/http"
	"path/filepath"

	"github.com/charmbracelet/log"

	"github.com/go-chi/chi"
	"github.com/hoyci/fakeflix/internal/infra/media"
	"github.com/hoyci/fakeflix/internal/usecase/video"
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

	h.logger.Info("Received streaming request", "videoID", videoID, "range_header", r.Header.Get("Range"))

	input := video.GetStreamInfoInputDTO{VideoID: videoID}

	output, err := h.getStreamInfoUseCase.Execute(r.Context(), input)
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
