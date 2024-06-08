package utils

import "github.com/go-playground/validator/v10"

// ValidateStruct validates a struct based on the `validate` tags.
func ValidateStruct(s interface{}) error {
    validate := validator.New()
    return validate.Struct(s)
}