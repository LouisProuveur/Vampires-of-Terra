// src/controllers/hub.go
package controllers

import (
	"Vampires-of-Terra/errors"
	"Vampires-of-Terra/models"
	"context"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

const charBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

/*
Hub manages all game controllers and lifecycles of games.
It handles creation of games and mapping of game IDs to their respective GameController instances.
*/
type Hub struct {
	mutex sync.RWMutex
	ctx   context.Context
	games map[string]*GameController
}

/*
Creates a new Hub instance with the provided context.
*/
func NewHub(ctx context.Context) *Hub {
	return &Hub{
		ctx:   ctx,
		games: make(map[string]*GameController),
	}
}

/*
Creates a new game managed by the Hub.
It expects a JSON payload with the game ID and password for the game as well as the player's name.
If the game ID is unique, a GameController instance is created and the associated goroutine is started.
If the player's name is unique, the player is added to the game.
On success, returns the player's token
*/
func (hub *Hub) CreateGame(c *gin.Context) {

	var req models.GameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	if _, exists := hub.games[req.ID]; exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Game with this ID already exists"})
		return
	}

	gameCtx, cancel := context.WithCancel(hub.ctx)
	gameController, err := CreateGameController(req.ID, req.Password, cancel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	hub.games[req.ID] = gameController

	go gameController.Run(gameCtx)

	token, err := gameController.JoinGame(req.PlayerName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Game created", "token": token})
}

/*
Used to add a player to a game.
It expects a JSON payload with the game ID and password for the game as well as the player's name.
If the player's ID is unique and the password correct, the player is added to the game.
On success, returns the player's token
*/
func (hub *Hub) JoinGame(c *gin.Context) {

	var req models.GameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hub.mutex.RLock()
	defer hub.mutex.RUnlock()

	gameController, exists := hub.games[req.ID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}

	passwordValid := gameController.VerifyPassword(req.Password)
	if !passwordValid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	token, err := gameController.JoinGame(req.PlayerName)
	if err == errors.ErrPlayerNameTaken {
		c.JSON(http.StatusConflict, gin.H{"error": "Name Already Taken"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Joined game", "token": token})

}

func (hub *Hub) Connect(c *gin.Context) {

	gameID := c.Param("id")

	gameController, exists := hub.games[gameID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game ID not found"})
		return
	}

	gameController.Connect(c)

}
