package media

import (
	"context"
	"errors"
	"net/http"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/cptallergy/sidequest-api/internal/lib/httpx"
)

type Service interface {
	CreatePresignedPutUrl(
		ctx context.Context,
		presignedUploadDto PresignedUploadRequest) (
		request *v4.PresignedHTTPRequest,
		objectKey string,
		err error,
	)
}

type Handler struct {
	srv Service
}

func NewHandler(srv Service) *Handler {
	return &Handler{srv}
}

// TODO can have some generic error helper to help reduce the code duplication

// TODO need to think about the error handling here, maybe create some custom error types and use those to determine the status code and message to return
func (h *Handler) PresignedUpload(w http.ResponseWriter, r *http.Request) {
	var presignedUploadDto PresignedUploadRequest
	if err := httpx.Read(r, &presignedUploadDto); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	uploadInfo, objectKey, err := h.srv.CreatePresignedPutUrl(r.Context(), presignedUploadDto)
	if err != nil {
		if errors.Is(err, ErrUnsupportedMediaType) {
			httpx.Error(w, http.StatusBadRequest, "unsupported media type")
			return
		}

		if errors.Is(r.Context().Err(), context.DeadlineExceeded) {
			return
		}

		httpx.InternalError(w, r, err)
		return
	}

	res := PresignedUploadResponse{
		URL:          uploadInfo.URL,
		ImageKey:     objectKey,
		SignedHeader: uploadInfo.SignedHeader,
	}

	httpx.Write(w, http.StatusOK, res)
}
