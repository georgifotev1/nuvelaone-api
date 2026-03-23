package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func Validate(s any) error {
	if err := validate.Struct(s); err != nil {
		verr, ok := err.(validator.ValidationErrors)
		if !ok {
			return err
		}
		for _, e := range verr {
			return fmt.Errorf("%s: %s", e.Field(), e.Tag())
		}
	}
	return nil
}
