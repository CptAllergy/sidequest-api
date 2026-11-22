package media

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/cptallergy/sidequest-api/internal/lib/storage"
	"github.com/google/uuid"
)

type Service interface {
	CreatePresignedPutUrl(
		ctx context.Context,
		userId string,
		presignedUploadDto PresignedUploadRequest,
	) (
		request *v4.PresignedHTTPRequest,
		objectKey string,
		err error,
	)
	CreatePresignedGetUrl(ctx context.Context, imageKey string) (string, error)
	PromoteTemporaryImage(ctx context.Context, userId string, imageKey string, dest string) (*string, error)
	DeleteImage(ctx context.Context, imageKey string) error
	DeleteQuestImages(ctx context.Context, questId string) error
	DeleteTemporaryImage(ctx context.Context, imageKey string)
	DeleteOrphanedImage(ctx context.Context, imageKey string)
}
type srv struct {
	storage storage.Storage
}

func NewService(storage storage.Storage) Service {
	return &srv{storage}
}

func (s *srv) CreatePresignedPutUrl(
	ctx context.Context,
	userId string,
	presignedUploadDto PresignedUploadRequest,
) (
	request *v4.PresignedHTTPRequest,
	objectKey string,
	err error,
) {
	mediaType, ext, err := validateImageType(presignedUploadDto.ContentType)
	if err != nil {
		return nil, "", fmt.Errorf("%w: invalid media-type", ErrUnsupportedMediaType)
	}

	objectKey = fmt.Sprintf("%s%s%s", TmpPath, uuid.New().String(), ext)
	metadata := map[string]string{
		MetadataUserId: userId,
	}
	uploadInfo, err := s.storage.PresignPutUrl(ctx, objectKey, mediaType, metadata, GetImageUrlExpiration)
	if err != nil {
		return nil, "", err
	}
	return uploadInfo, objectKey, nil
}

func (s *srv) CreatePresignedGetUrl(ctx context.Context, imageKey string) (string, error) {
	url, err := s.storage.PresignGetUrl(ctx, imageKey, GetImageUrlExpiration)
	if err != nil {
		return "", fmt.Errorf("%w: failed to create presigned get URL", err)
	}
	return url, nil
}

// PromoteTemporaryImage moves an image from the temporary folder to a permanent location and returns the new image key.
func (s *srv) PromoteTemporaryImage(ctx context.Context, userId string, imageKey string, dest string) (*string, error) {
	if !isValidTempImageKey(imageKey) {
		return nil, fmt.Errorf("invalid image source: key %q must be valid and in the temporary folder", imageKey)
	}

	info, err := s.storage.HeadObject(ctx, imageKey)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotFound) {
			return nil, fmt.Errorf("%w: image key %q not found", ErrInvalidImage, imageKey)
		}
		return nil, fmt.Errorf("error occurred while fetching image info: %w", err)
	}

	if info.Size == 0 || info.Size > MaxImageSize {
		return nil, fmt.Errorf("%w: invalid image size %q", ErrInvalidImage, imageKey)
	}

	// Check if the image belongs to the user by checking metadata
	if info.Metadata[MetadataUserId] != userId {
		return nil, fmt.Errorf("%w: image key %q does not belong to user %q", ErrInvalidImage, imageKey, userId)
	}

	// Detect the content type of the actual image and validate it against allowed mime types
	var detectedType string
	detectedType, err = s.storage.ReadObjectContentType(ctx, imageKey)
	if err != nil {
		return nil, fmt.Errorf("error occurred while reading image content type: %w", err)
	}
	var detectedExt string
	detectedType, detectedExt, err = validateImageType(detectedType)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid media-type", ErrInvalidImage)
	}

	if detectedType != info.ContentType {
		return nil, fmt.Errorf("%w: image content type mismatch", ErrInvalidImage)
	}

	// Create new image location
	newImageUUID := uuid.New().String()
	newImageKey := new(fmt.Sprintf("%s/%s%s", dest, newImageUUID, detectedExt))

	err = s.storage.CopyObject(ctx, imageKey, *newImageKey)
	if err != nil {
		return nil, fmt.Errorf("moving image from tmp to final location: %w", err)
	}

	return newImageKey, nil
}

func (s *srv) DeleteImage(ctx context.Context, imageKey string) error {
	err := s.storage.DeleteObject(ctx, imageKey)
	if err != nil {
		return err
	}

	return nil
}

// DeleteQuestImages deletes all images associated with a quest by deleting all objects with the prefix "quests/{questId}/".
func (s *srv) DeleteQuestImages(ctx context.Context, questId string) error {
	if questId == "" {
		return fmt.Errorf("questId cannot be empty")
	}

	err := s.storage.DeleteObjectsWithPrefix(ctx, QuestPath+questId+"/")
	if err != nil {
		return err
	}

	return nil
}

func (s *srv) DeleteTemporaryImage(ctx context.Context, imageKey string) {
	err := s.storage.DeleteObject(ctx, imageKey)
	// Failure here is fine because tmp/ is cleaned periodically
	if err != nil {
		slog.Warn("Failed to delete tmp/ image", "key", imageKey, "err", err)
	}
}

func (s *srv) DeleteOrphanedImage(ctx context.Context, imageKey string) {
	err := s.storage.DeleteObject(ctx, imageKey)
	if err != nil {
		slog.Error("Failed to delete orphaned image", "key", imageKey, "err", err)
	} else {
		slog.Info("Deleted orphaned image", "key", imageKey)
	}
}
