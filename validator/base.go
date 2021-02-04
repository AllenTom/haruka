package validator

import (
	"errors"
	vvalidator "github.com/go-playground/validator/v10"
)

type Validator interface {
	Check() (string, bool)
}

// StructValidator support validator from github.com/go-playground/validator
type StructValidator struct {
	Struct interface{}
}

func (v *StructValidator) Check() (string, bool) {
	va := vvalidator.New()
	err := va.Struct(v.Struct)
	validationErrors := err.(vvalidator.ValidationErrors)
	if err != nil {
		return validationErrors.Error(), false

	}
	return "", true
}

func RunValidators(validators ...Validator) error {
	for _, validator := range validators {
		info, isValidate := validator.Check()
		if !isValidate {
			return errors.New(info)
		}
	}
	return nil
}
