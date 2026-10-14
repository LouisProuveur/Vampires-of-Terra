// src/apperrors/errors.go
package apperrors

import "errors"

var (
	ErrPlayerNameTaken = errors.New("player with this name already exists")
	ErrGameClosed      = errors.New("game server is closed")
	ErrGameFull        = errors.New("game is full of players")
)
