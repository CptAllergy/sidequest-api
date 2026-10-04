package media

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/cptallergy/sidequest-api/internal/lib/auth"
	"github.com/cptallergy/sidequest-api/internal/lib/json"
	"github.com/cptallergy/sidequest-api/internal/lib/storage"
)

type Service interface {
}

type Handler struct {
	storage storage.Storage
}

func NewHandler(storage storage.Storage) *Handler {

	return &Handler{storage}
}

// TODO have more helpful error messages sent to the user
// TODO can have some generic error helper to help reduce the code duplication

// TODO need to think about the error handling here, maybe create some custom error types and use those to determine the status code and message to return
func (h *Handler) PresignedUpload(w http.ResponseWriter, r *http.Request) {
	identity, ok := auth.GetIdentityFromContext(r.Context())
	if !ok {
		slog.Error("Error getting identity from context")
		http.Error(w, "Error getting identity from context", http.StatusUnauthorized)
		return
	}

	var presignedUploadDto PresignedUploadDto
	if err := json.Read(r, &presignedUploadDto); err != nil {
		slog.Error("Error reading request body", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	users, err := h.storage.Request(r.Context())
	if err != nil {
		slog.Error("Error listing users", "error", err)
		if errors.Is(r.Context().Err(), context.DeadlineExceeded) {
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	// TODO should I use an envelope for this **helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"quests": all})*** what's the point of the envelope?
	err = json.Write(w, http.StatusOK, users)
	if err != nil {
		slog.Error("Error writing response", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
