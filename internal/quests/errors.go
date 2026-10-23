package quests

import (
	"errors"
	"net/http"

	"github.com/cptallergy/sidequest-api/internal/lib/httpx"
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
		httpx.Error(w, http.StatusForbidden, "forbidden")
	case errors.Is(err, ErrBadRequest):
		httpx.Error(w, http.StatusBadRequest, "invalid parameters")
	default:
		httpx.InternalError(w, r, err)
	}
}
