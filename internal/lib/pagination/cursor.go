package pagination

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
)

var ErrInvalidCursor = errors.New("invalid cursor")

const (
	DefaultLimit = 20
	MaxLimit     = 50
)

type Page[T any] struct {
	Items   []T
	HasMore bool
}

type PageResponse[T any] struct {
	Items      []T     `json:"items"`
	NextCursor *string `json:"nextCursor"`
}

func Encode[T any](cursor T) (string, error) {
	data, err := json.Marshal(cursor)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cursor: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

func Decode[T any](encoded string) (*T, error) {
	var cursor T
	if encoded == "" {
		return nil, nil
	}
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, ErrInvalidCursor
	}
	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil, ErrInvalidCursor
	}

	return &cursor, nil
}

func ClampLimit(n int) int {
	if n <= 0 {
		return DefaultLimit
	}
	return min(n, MaxLimit)
}

// Trim takes rows fetched with limit+1 and splits them into items and hasMore
func Trim[T any](rows []T, limit int) Page[T] {
	if len(rows) > limit {
		return Page[T]{
			Items:   rows[:limit],
			HasMore: true,
		}

	}
	return Page[T]{
		Items:   rows,
		HasMore: false,
	}
}
