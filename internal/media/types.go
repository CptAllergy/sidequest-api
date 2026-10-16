package media

import "net/http"

type PresignedUploadRequest struct {
	ContentType string `json:"contentType" validate:"required"`
}

type PresignedUploadResponse struct {
	URL          string      `json:"url"`
	ImageKey     string      `json:"imageKey"`
	SignedHeader http.Header `json:"signedHeader"`
}

var AllowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

const QUEST_PATH = "quests/"
const TMP_PATH = "tmp/"
