package media

import "net/http"

type PresignedUploadDto struct {
	ContentType string `json:"contentType" validate:"required"`
}

type PresignedUploadResponseDto struct {
	URL          string      `json:"url"`
	ImageKey     string      `json:"imageKey"`
	SignedHeader http.Header `json:"signedHeader"`
}

var AllowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}
