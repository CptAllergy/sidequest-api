package media

import (
	"context"
	"errors"
	"fmt"
	"mime"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/cptallergy/sidequest-api/internal/lib/storage"
	"github.com/google/uuid"
)

const tempDir = "tmp"

type srv struct {
	storage storage.Storage
}

var (
	ErrPresigned            = errors.New("failed to create presigned URL")
	ErrUnsupportedMediaType = errors.New("unsupported or invalid image content type")
)

func NewService(storage storage.Storage) Service {
	return &srv{storage}
}

func (s *srv) CreatePresignedPutUrl(
	ctx context.Context,
	presignedUploadDto PresignedUploadDto) (
	request *v4.PresignedHTTPRequest,
	objectKey string,
	err error,
) {
	mediaType, ext, err := validateMediaType(presignedUploadDto.ContentType)
	if err != nil {
		return nil, "", fmt.Errorf("%w: invalid media-type", ErrUnsupportedMediaType)
	}

	objectKey = fmt.Sprintf("%s/%s%s", tempDir, uuid.New().String(), ext)
	uploadInfo, err := s.storage.PresignedPutUrl(ctx, objectKey, mediaType, 3600)
	if err != nil {
		return nil, "", ErrPresigned
	}
	return uploadInfo, objectKey, nil
}

func validateMediaType(mediaType string) (parsedType string, ext string, err error) {
	if mediaType == "" {
		return "", "", fmt.Errorf("%w: media-type header is required", ErrUnsupportedMediaType)
	}

	parsedType, _, err = mime.ParseMediaType(mediaType)
	if err != nil {
		return "", "", fmt.Errorf("%w: invalid media type format: %v", ErrUnsupportedMediaType, err)
	}

	ext, ok := AllowedImageTypes[parsedType]
	if !ok {
		return "", "", fmt.Errorf("%w: '%s' is not allowed", ErrUnsupportedMediaType, parsedType)
	}

	return parsedType, ext, nil
}
