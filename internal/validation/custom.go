package validation

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func validateMethod(fl validator.FieldLevel) bool {
	method := strings.ToUpper(fl.Field().String())
	return method == "GET" || method == "POST"
}

func validateUUID(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	_, err := uuid.Parse(value)
	return err == nil
}
