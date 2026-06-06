package validator

import (
	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	return &Validator{
		validate: validator.New(validator.WithRequiredStructEnabled()),
	}
}

func (v *Validator) Validate(input interface{}) map[string]string {
	err := v.validate.Struct(input)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return map[string]string{
			"request": err.Error(),
		}
	}

	result := make(map[string]string, len(validationErrors))
	for _, validationError := range validationErrors {
		result[validationError.Field()] = validationError.Tag()
	}

	return result
}
