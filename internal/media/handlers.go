package media

import (
	"context"
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

func (h *Handler) PresignedUpload(w http.ResponseWriter, r *http.Request) {
	var presignedUploadDto PresignedUploadRequest
	if ok := httpx.Read(w, r, &presignedUploadDto); !ok {
		return
	}

	uploadInfo, objectKey, err := h.srv.CreatePresignedPutUrl(r.Context(), presignedUploadDto)
	if err != nil {
		writeError(w, r, err)
		return
	}

	res := PresignedUploadResponse{
		URL:          uploadInfo.URL,
		ImageKey:     objectKey,
		SignedHeader: uploadInfo.SignedHeader,
	}

	httpx.Write(w, http.StatusOK, res)
}
