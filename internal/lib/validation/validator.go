package validation

import (
	"github.com/go-playground/validator/v10"
)

func SetupValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate
}
