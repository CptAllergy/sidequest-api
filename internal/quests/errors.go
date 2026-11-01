package quests

import (
	"errors"
	"net/http"

	"github.com/cptallergy/sidequest-api/internal/lib/httpx"
	"github.com/cptallergy/sidequest-api/internal/media"
)

var (
	ErrQuestNotFound = errors.New("quest not found")
	ErrForbidden     = errors.New("user does not have permission to perform this action")
	ErrBadRequest    = errors.New("invalid parameters")
)

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrQuestNotFound):
		httpx.Error(w, http.StatusNotFound, "quest not found")
	case errors.Is(err, ErrForbidden):
		// TODO maybe return not found to hide existing entries
		httpx.Error(w, http.StatusForbidden, "forbidden")
	case errors.Is(err, ErrBadRequest):
		httpx.Error(w, http.StatusBadRequest, "invalid parameters")
	case errors.Is(err, media.ErrInvalidImage):
		httpx.Error(w, http.StatusBadRequest, "invalid image")
	default:
		httpx.InternalError(w, r, err)
	}
}
