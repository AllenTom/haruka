package validator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type StructValidatorExample struct {
	FirstName      string `validate:"required"`
	LastName       string `validate:"required"`
	Age            uint8  `validate:"gte=0,lte=130"`
	Email          string `validate:"required,email"`
	FavouriteColor string `validate:"iscolor"`
}

func TestStructValidator_Check(t *testing.T) {
	example := &StructValidatorExample{
		FirstName:      "Badger",
		LastName:       "Smith",
		Age:            135,
		Email:          "Badger.Smith@gmail.com",
		FavouriteColor: "#000-",
	}
	v := StructValidator{Struct: example}
	message, isOk := v.Check()
	assert.Equal(t, false, isOk, "must invalidate")
	assert.NotEqual(t, 0, len(message), "must has error message")
}
