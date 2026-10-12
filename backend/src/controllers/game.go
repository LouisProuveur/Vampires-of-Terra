// src/controller/game.go
package controllers

import (
	"Vampires-of-Terra/models"
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type PlayerClaims struct {
	PlayerName string `json:"pid"`
	GameID     string `json:"gid"`
	jwt.RegisteredClaims
}

type GameController struct {
	game          *models.Game
	id            string
	password_hash []byte
	cancel        context.CancelFunc
}

/*
Creates a new gamecontroller instance with password and id
Returns the gamecontroller instance
*/
func CreateGameController(id string, password string, cancel context.CancelFunc) (*GameController, error) {

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	game := models.NewGame()

	return &GameController{game: game, id: id, password_hash: passwordHash, cancel: cancel}, nil
}

/*
Verifies the provided password against the stored hash
*/
func (gameController *GameController) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword(gameController.password_hash, []byte(password))
	return err == nil
}

/*
Produces a JWT holding the player's name and associated game ID.
These tokens are only valid for the game they were created for.
*/
func (gameController *GameController) signToken(playerName string) (string, error) {
	claims := PlayerClaims{
		PlayerName: playerName,
		GameID:     gameController.id,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString("hello there")
}

/*
Parses and validates a player's token.
A token is only valid for the given game instance they were produced for.
*/
func (gameController *GameController) parseToken(tokenStr string) (*PlayerClaims, error) {
	claims := &PlayerClaims{}

	token, err := jwt.ParseWithClaims(
		tokenStr,
		claims,
		func(t *jwt.Token) (any, error) { return "hello there", nil },
		jwt.WithValidMethods([]string{"HS256"}),
	)
	if err != nil {
		return nil, err
	}

	if !token.Valid || claims.GameID != gameController.id {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

/*
Registers a player into the game.
Returns the player's token
*/
func (gameController *GameController) JoinGame(playerName string) (string, error) {

	if err := gameController.game.AddPlayer(playerName); err != nil {
		return "", err
	}

	token, err := gameController.signToken(playerName)
	if err != nil {
		return "", err
	}

	return token, nil

}

func (gameController *GameController) Run(ctx context.Context) {
	println("GameController is started for game ID:", gameController.id)
}
