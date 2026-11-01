package media

import (
	"fmt"
	"mime"
	"path"
	"strings"

	"github.com/google/uuid"
)

var allowedImageExtensions = map[string]bool{
	".jpg":  true,
	".png":  true,
	".webp": true,
}

func isValidTempImageKey(key string) bool {
	name, ok := strings.CutPrefix(key, TmpPath)
	if !ok {
		return false
	}

	ext := path.Ext(name)
	if !allowedImageExtensions[ext] {
		return false
	}

	id := strings.TrimSuffix(name, ext)
	parsed, err := uuid.Parse(id)
	if err != nil {
		return false
	}

	// Ensures no weird uuid format passes
	return parsed.String() == id
}

func validateImageType(mediaType string) (parsedType string, ext string, err error) {
	if mediaType == "" {
		return "", "", fmt.Errorf("%w: media-type header is required", ErrUnsupportedMediaType)
	}

	parsedType, _, err = mime.ParseMediaType(mediaType)
	if err != nil {
		return "", "", fmt.Errorf("%w: invalid media type format: %v", ErrUnsupportedMediaType, err)
	}

	ext, ok := AllowedMimeTypes[parsedType]
	if !ok {
		return "", "", fmt.Errorf("%w: '%s' is not allowed", ErrUnsupportedMediaType, parsedType)
	}

	return parsedType, ext, nil
}
