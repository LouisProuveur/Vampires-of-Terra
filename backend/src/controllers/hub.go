// src/controllers/hub.go
package controllers

import (
	"context"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

const charBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

type NewGameRequest struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}

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
It expects a JSON payload with the game ID and password for the game.
If the ID is unique, a GameController instance is created and the associated gorouting is started.
*/
func (hub *Hub) CreateGame(c *gin.Context) {

	var req NewGameRequest
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

	c.JSON(http.StatusCreated, gin.H{"message": "Game created", "game_id": req.ID})
}
