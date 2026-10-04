package media

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/cptallergy/sidequest-api/internal/lib/json"
)

type Service interface {
	CreatePresignedPutUrl(
		ctx context.Context,
		presignedUploadDto PresignedUploadDto) (
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

// TODO have more helpful error messages sent to the user
// TODO can have some generic error helper to help reduce the code duplication

// TODO need to think about the error handling here, maybe create some custom error types and use those to determine the status code and message to return
func (h *Handler) PresignedUpload(w http.ResponseWriter, r *http.Request) {
	var presignedUploadDto PresignedUploadDto
	if err := json.Read(r, &presignedUploadDto); err != nil {
		slog.Error("Error reading request body", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	uploadInfo, objectKey, err := h.srv.CreatePresignedPutUrl(r.Context(), presignedUploadDto)
	if err != nil {
		slog.Error("Error creating presigned PUT url", "error", err)
		if errors.Is(r.Context().Err(), context.DeadlineExceeded) {
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	res := PresignedUploadResponseDto{
		URL:          uploadInfo.URL,
		ImageKey:     objectKey,
		SignedHeader: uploadInfo.SignedHeader,
	}

	err = json.Write(w, http.StatusOK, res)
	if err != nil {
		slog.Error("Error writing response", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
