package validation

import (
	"github.com/cptallergy/sidequest-api/internal/quests"
	"github.com/cptallergy/sidequest-api/internal/users"
	"github.com/go-playground/validator/v10"
)

func SetupValidator() (*validator.Validate, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := quests.RegisterValidations(validate); err != nil {
		return nil, err
	}
	if err := users.RegisterValidations(validate); err != nil {
		return nil, err
	}
	return validate, nil
}
