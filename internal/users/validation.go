package users

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// TODO helpful message errors?
func RegisterValidations(v *validator.Validate) error {

	usernameRegex := regexp.MustCompile(`^[a-z0-9_]+$`)
	err := v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		return usernameRegex.MatchString(fl.Field().String())
	})

	return err
}
