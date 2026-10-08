// src/controller/game.go
package controllers

import (
	"Vampires-of-Terra/models"
	"context"

	"golang.org/x/crypto/bcrypt"
)

type GameController struct {
	game          *models.Game
	id            string
	password_hash []byte
	cancel        context.CancelFunc
}

func CreateGameController(id string, password string, cancel context.CancelFunc) (*GameController, error) {

	password_hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	game := models.NewGame()

	return &GameController{game: game, id: id, password_hash: password_hash, cancel: cancel}, nil
}

func (gameController *GameController) verifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword(gameController.password_hash, []byte(password))
	return err == nil
}

func (gameController *GameController) Run(ctx context.Context) {
	println("GameController is started for game ID:", gameController.id)
}
