package media

type PresignedUploadRequest struct {
	ContentType string `json:"contentType" validate:"required"`
}

type PresignedUploadResponse struct {
	URL      string            `json:"url"`
	ImageKey string            `json:"imageKey"`
	Headers  map[string]string `json:"headers"`
}

var AllowedMimeTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

const GetImageUrlExpiration = 3600 // 1 hour

const QuestPath = "quests/"
const TmpPath = "tmp/"
const MetadataUserId = "user-id"
const MaxImageSize int64 = 5 * 1024 * 1024 // 5MB
