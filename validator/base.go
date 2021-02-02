package validator

import (
	"errors"
)

type Validator interface {
	Check() (string, bool)
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
