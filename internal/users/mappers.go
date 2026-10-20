package users

import (
	db "github.com/cptallergy/sidequest-api/internal/db/sqlc"
)

func toUserResponse(u db.User) UserResponse {
	return UserResponse{
		ID:          u.ID,
		Username:    u.Username,
		DisplayName: u.DisplayName,
		AvatarUrl:   u.AvatarUrl,
		Bio:         u.Bio,
		CreatedAt:   u.CreatedAt.Time,
		UpdatedAt:   u.UpdatedAt.Time,
	}
}
