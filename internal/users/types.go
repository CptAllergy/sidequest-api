package users

// TODO create return type DTOs

type CreateUserRequest struct {
	Username string `json:"username" validate:"required"`
}
