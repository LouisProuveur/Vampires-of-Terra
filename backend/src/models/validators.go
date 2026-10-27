// src/models/validators.go
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

	// Player ID validator
	if err := v.RegisterValidation("playerID", func(fl validator.FieldLevel) bool {
		return nameRe.MatchString(fl.Field().String())
	}); err != nil {
		return err
	}

	// Game Password Validator
	if err := v.RegisterValidation("gamePasswordLength", func(fl validator.FieldLevel) bool {
		length := len(fl.Field().String())
		return length >= 1 && length <= 72
	}); err != nil {
		return err
	}

	return nil
}
