package govalidator

import (
	"github.com/go-playground/validator/v10"
)

func New() (*validator.Validate, error) {
	validate := validator.New()
	for i, v := range validators {
		err := validate.RegisterValidation(i, v)
		if err != nil {
			return validate, err
		}
	}
	return validate, nil
}
