package validation

import "github.com/go-playground/validator/v10"

func NewValidator() (*validator.Validate, error) {
	v := validator.New()
	if err := v.RegisterValidation("httpmethod", validateMethod); err != nil {
		return nil, err
	}
	if err := v.RegisterValidation("uuid", validateUUID); err != nil {
		return nil, err
	}
	return v, nil
}
