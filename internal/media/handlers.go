package media

import (
	"net/http"
	"strings"

	"github.com/cptallergy/sidequest-api/internal/lib/auth"
	"github.com/cptallergy/sidequest-api/internal/lib/httpx"
)

type Handler struct {
	srv Service
}

func NewHandler(srv Service) *Handler {
	return &Handler{srv}
}

func (h *Handler) PresignedUpload(w http.ResponseWriter, r *http.Request) {
	userId, ok := auth.UserId(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var presignedUploadDto PresignedUploadRequest
	if ok = httpx.Read(w, r, &presignedUploadDto); !ok {
		return
	}

	uploadInfo, objectKey, err := h.srv.CreatePresignedPutUrl(r.Context(), userId, presignedUploadDto)
	if err != nil {
		writeError(w, r, err)
		return
	}

	// Flatten the signed headers into a map[string]string and remove "Host" header
	headers := make(map[string]string, len(uploadInfo.SignedHeader))
	for name, values := range uploadInfo.SignedHeader {
		if strings.EqualFold(name, "Host") {
			continue
		}
		headers[name] = strings.Join(values, ",")
	}

	res := PresignedUploadResponse{
		URL:      uploadInfo.URL,
		ImageKey: objectKey,
		Headers:  headers,
	}

	httpx.Write(w, http.StatusOK, res)
}
