package users

// TODO add required ifs for the correct fields based on what provider is set
type CreateUserDTO struct {
	Email          string `json:"email"  validate:"required,email"`
	Username       string `json:"username" validate:"required"`
	Provider       string `json:"provider" validate:"required,provider"`
	Password       string `json:"password"`
	ProviderUserID string `json:"provider_user_id"`
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
