package users

import (
	"time"
)

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=4,max=30,username"`
}

type UserResponse struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"displayName"`
	AvatarUrl   *string   `json:"avatarUrl"`
	Bio         *string   `json:"bio"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
