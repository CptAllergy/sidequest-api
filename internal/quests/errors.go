package quests

import (
	"errors"
	"net/http"

	"github.com/cptallergy/sidequest-api/internal/lib/httpx"
	"github.com/cptallergy/sidequest-api/internal/media"
)

var (
	ErrQuestNotFound = errors.New("quest not found")
	ErrEntryNotFound = errors.New("entry not found")
	ErrForbidden     = errors.New("user does not have permission to perform this action")
	ErrBadRequest    = errors.New("invalid parameters")
)

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrQuestNotFound):
		httpx.Error(w, http.StatusNotFound, "quest not found")
	case errors.Is(err, ErrEntryNotFound):
		httpx.Error(w, http.StatusNotFound, "entry not found")
	case errors.Is(err, ErrForbidden):
		// Return 404 instead of 403 to avoid leaking information about the existence of the quest
		httpx.Error(w, http.StatusNotFound, "quest not found")
	case errors.Is(err, ErrBadRequest):
		httpx.Error(w, http.StatusBadRequest, "invalid parameters")
	case errors.Is(err, media.ErrInvalidImage):
		httpx.Error(w, http.StatusBadRequest, "invalid image")
	default:
		httpx.InternalError(w, r, err)
	}
}
