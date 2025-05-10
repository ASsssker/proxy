package validation

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

func validateMethod(fl validator.FieldLevel) bool {
	method := strings.ToUpper(fl.Field().String())
	return method == "GET" || method == "POST"
}
