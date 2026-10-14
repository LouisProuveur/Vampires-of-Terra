// models/validators.go
package models

import (
	"errors"
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var nameRe = regexp.MustCompile(`^[\p{Latin}\p{N}_-]+$`)

func RegisterValidators() error {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return errors.New("unexpected validator engine")
	}
	if err := v.RegisterValidation("playername", func(fl validator.FieldLevel) bool {
		return nameRe.MatchString(fl.Field().String())
	}); err != nil {
		return err
	}
	return nil
}
