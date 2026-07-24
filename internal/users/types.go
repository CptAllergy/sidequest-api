package users

type CreateUserDto struct {
	Username string `json:"username" validate:"required"`
}
