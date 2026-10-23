package users

import (
	"errors"
	"net/http"

	"github.com/cptallergy/sidequest-api/internal/lib/httpx"
)

var (
	ErrNotFound = errors.New("user not found")
	// TODO add this error where needed
	ErrForbidden = errors.New("user does not have permission to perform this action")
)

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpx.Error(w, http.StatusNotFound, "user not found")
	case errors.Is(err, ErrForbidden):
		httpx.Error(w, http.StatusForbidden, "forbidden")
	default:
		httpx.InternalError(w, r, err)
	}
}
