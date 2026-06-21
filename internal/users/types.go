package users

type CreateUserDto struct {
	Email          string       `json:"email"  validate:"required,email"`
	Username       string       `json:"username" validate:"required"`
	Provider       ProviderType `json:"provider" validate:"required,provider"`
	Password       string       `json:"password" validate:"required_if=Provider LOCAL,excluded_unless=Provider LOCAL"`
	ProviderUserID string       `json:"provider_user_id" validate:"required_unless=Provider LOCAL,excluded_if=Provider LOCAL"`
}

type ProviderType string

const (
	Local  ProviderType = "LOCAL"
	Google ProviderType = "GOOGLE"
	Github ProviderType = "GITHUB"
)

func (p ProviderType) IsValid() bool {
	switch p {
	case Local, Google, Github:
		return true
	}
	return false
}

func (p ProviderType) String() string {
	return string(p)
}
