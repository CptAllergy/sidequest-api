package httpx

import (
	"encoding/json"
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

func Read(r *http.Request, data any) error {
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(data)
}

func Error(w http.ResponseWriter, status int, msg string) {
	Write(w, status, errorResponse{Error: msg})
}

// TODO USE LOWERCASE LETTERS FOR ERROR RESPONSES, NO PERIODS
func InternalError(w http.ResponseWriter, r *http.Request, err error) {
	slog.ErrorContext(r.Context(), "internal error",
		"err", err,
		"method", r.Method,
		"path", r.URL.Path,
		"request_id", middleware.GetReqID(r.Context()))
	Error(w, http.StatusInternalServerError, "internal server error")
}
