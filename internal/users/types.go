package users

// TODO create return type DTOs

type CreateUserDto struct {
	Username string `json:"username" validate:"required"`
}
