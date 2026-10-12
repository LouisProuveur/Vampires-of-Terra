package errors

import "errors"

var (
	ErrPlayerNameTaken = errors.New("Player with this name already exists")
)
