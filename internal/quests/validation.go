package quests

import "github.com/go-playground/validator/v10"

// TODO ensure we get helpful error messages when some validation fails
func RegisterValidations(v *validator.Validate) error {
	err := v.RegisterValidation("quest_status", func(fl validator.FieldLevel) bool {
		return StatusType(fl.Field().String()).IsValid()
	})

	if err != nil {
		return err
	}

	err = v.RegisterValidation("quest_type", func(fl validator.FieldLevel) bool {
		return QuestType(fl.Field().String()).IsValid()
	})

	return err
}
