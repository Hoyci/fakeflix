// Package httputil provides utility functions for handling common web-related tasks,
// such as writing standardized JSON responses and handling errors.
package httputils

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hoyci/fakeflix/pkg/fault"
)

func RespondWithError(w http.ResponseWriter, err error) {
	var f *fault.Error
	if errors.As(err, &f) {
		statusCode := mapKindToStatusCode(f.Kind)
		RespondWithJSON(w, statusCode, map[string]string{"error": f.Message})
		return
	}
	RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "an unexpected error occurred"})
}

func mapKindToStatusCode(kind string) int {
	switch kind {
	case fault.KindNotFound:
		return http.StatusNotFound
	case fault.KindValidation:
		return http.StatusUnprocessableEntity
	case fault.KindConflict:
		return http.StatusConflict
	case fault.KindUnauthenticated:
		return http.StatusUnauthorized
	case fault.KindForbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

func RespondWithJSON(w http.ResponseWriter, code int, payload any) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Error marshaling JSON response"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
