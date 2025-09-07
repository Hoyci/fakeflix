package http

import (
	"net/http"

	"github.com/hoyci/fakeflix/internal/usecase/movie"
	"github.com/hoyci/fakeflix/pkg/fault"
	"github.com/hoyci/fakeflix/pkg/httputils"
)

type CreateMovieRequest struct {
	Title       string
	Description string
}

type MovieHandler struct {
	createMovieUseCase *movie.CreateMovieUseCase
}

func NewMovieHandler(createMovieUseCase *movie.CreateMovieUseCase) *MovieHandler {
	return &MovieHandler{
		createMovieUseCase: createMovieUseCase,
	}
}

func (h *MovieHandler) CreateMovie(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max memory
		httputils.RespondWithError(w, fault.New("invalid form data", fault.WithError(err)))
		return
	}

	_, videoHeader, _ := r.FormFile("video")
	_, thumbHeader, _ := r.FormFile("thumbnail")

	requestDTO := movie.CreateMovieInputDTO{
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		Video:       videoHeader,
		Thumbnail:   thumbHeader,
	}

	if err := requestDTO.Validate(); err != nil {
		httputils.RespondWithError(w, fault.New(
			err.Error(),
			fault.WithKind(fault.KindValidation),
		))
		return
	}

	output, err := h.createMovieUseCase.Execute(r.Context(), requestDTO)
	if err != nil {
		httputils.RespondWithError(w, err)
		return
	}

	httputils.RespondWithJSON(w, http.StatusCreated, output)
}
