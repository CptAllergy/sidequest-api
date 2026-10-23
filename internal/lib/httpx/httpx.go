package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

type errorResponse struct {
	Error string `json:"error"`
}

func Write(w http.ResponseWriter, status int, data any) {
	body, err := json.Marshal(data)
	if err != nil {
		slog.Error("Error marshaling response", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err = w.Write(body); err != nil {
		slog.Error("Error writing response", "err", err)
	}
}

// Read decodes the json request body into data. On failure it writes the error response and returns false
func Read(w http.ResponseWriter, r *http.Request, data any) (ok bool) {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(data); err != nil {
		// TODO add a RequestEntityTooLarge check for chi middleware
		Error(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}

func Error(w http.ResponseWriter, status int, msg string) {
	Write(w, status, errorResponse{Error: msg})
}

func InternalError(w http.ResponseWriter, r *http.Request, err error) {
	switch ctxErr := r.Context().Err(); {
	case errors.Is(ctxErr, context.DeadlineExceeded):
		// Chi timeout
		slog.WarnContext(r.Context(), "request timed out",
			"err", err,
			"path", r.URL.Path)
		return

	case errors.Is(ctxErr, context.Canceled):
		// Client disconnected
		return
	}
	slog.ErrorContext(r.Context(), "internal error",
		"err", err,
		"method", r.Method,
		"path", r.URL.Path,
		"request_id", middleware.GetReqID(r.Context()))
	Error(w, http.StatusInternalServerError, "internal server error")
}
