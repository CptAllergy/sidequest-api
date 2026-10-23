package media

import (
	"errors"
	"net/http"

	"github.com/cptallergy/sidequest-api/internal/lib/httpx"
)

var (
	ErrUnsupportedMediaType = errors.New("unsupported or invalid image content type")
)

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrUnsupportedMediaType):
		httpx.Error(w, http.StatusBadRequest, "unsupported media type")
	default:
		httpx.InternalError(w, r, err)
	}
}
