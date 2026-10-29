// src/controllers/hub.go

// Package controllers is responsible for managing the interactions between the users and the game itself.
// It creates game instances, allows to connect users to games and manages transfer of information between the game and the user.
package controllers

import (
	"Vampires-of-Terra/apperrors"
	"Vampires-of-Terra/models"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func makeCheckOrigin(allowed string) func(r *http.Request) bool {
	if allowed == "" {
		slog.Warn("no allowlist set")
		return func(r *http.Request) bool {
			return true
		}
	}
	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "" || origin == allowed
	}
}

// Hub manages all game controllers and lifecycles of games.
// It handles creation of games and mapping of game IDs to their respective GameController instances.
type Hub struct {
	mutex    sync.RWMutex
	ctx      context.Context
	games    map[string]*GameController
	upgrader websocket.Upgrader
}

// NewHub creates a new Hub instance with the provided context.
func NewHub(ctx context.Context, allowedOrigin string) *Hub {

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     makeCheckOrigin(allowedOrigin),
	}

	return &Hub{
		ctx:      ctx,
		games:    make(map[string]*GameController),
		upgrader: upgrader,
	}
}

// CreateGame creates a new game managed by the Hub.
// It expects a JSON payload that matches [models.CreateGameRequest].
// If the game ID is unique, a GameController instance is created and the associated goroutine is started.
// On success, returns the player's token
func (h *Hub) CreateGame(c *gin.Context) {

	// TODO: cap number of games

	var req models.CreateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		slog.Error("error while hashing password", "game_id", req.ID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error while hashing password"})
		return
	}

	gameCtx, cancel := context.WithCancel(h.ctx)
	gameController, err := CreateGameController(req.ID, passwordHash, req.MaxPlayers, cancel)
	if err != nil {
		cancel()
		slog.Error("failed to create gamecontroller", "game_id", req.ID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initialize game server"})
		return
	}

	locked := true
	h.mutex.Lock()
	defer func() {
		if locked {
			h.mutex.Unlock()
		}
	}()

	if _, exists := h.games[req.ID]; exists {
		slog.Debug("duplicate game detected", "game_id", req.ID)
		c.JSON(http.StatusConflict, gin.H{"error": "Game with this ID already exists"})
		return
	}
	h.games[req.ID] = gameController

	h.mutex.Unlock()
	locked = false

	go gameController.Run(gameCtx, h.removeGameCallback)

	token, err := gameController.JoinGame(req.PlayerID)
	if err != nil {
		cancel()
		respondJoinError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Game created", "token": token})
}

// JoinGame is used to add a player to a game.
// It expects a JSON payload that matches [models.JoinGameRequest].
// If the player's ID is unique and the password correct, the player is added to the game.
// On success, returns the player's token
func (h *Hub) JoinGame(c *gin.Context) {

	var req models.JoinGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.Password) > 72 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at most 72 bytes"})
		return
	}

	h.mutex.RLock()
	gameController, exists := h.games[req.ID]
	h.mutex.RUnlock()

	if !exists {
		slog.Debug("game not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}

	passwordValid := gameController.VerifyPassword(req.Password)
	if !passwordValid {
		slog.Debug("invalid password")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	token, err := gameController.JoinGame(req.PlayerID)
	if err != nil {
		respondJoinError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Joined game", "token": token})

}

// respondJoinError sends a response appropriate to the returned error type
func respondJoinError(c *gin.Context, err error) {
	switch {

	case errors.Is(err, apperrors.ErrPlayerNameTaken):
		slog.Debug("duplicate name detected")
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})

	case errors.Is(err, apperrors.ErrGameFull):
		slog.Debug("game full")
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})

	case errors.Is(err, apperrors.ErrGameClosed):
		slog.Debug("game closed")
		c.JSON(http.StatusGone, gin.H{"error": err.Error()})

	default:
		slog.Error("error on player join", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "an unexpected error occurred"})
	}
}

// Connect upgrades the underlying connection to a websocket and starts client authentication.
func (h *Hub) Connect(c *gin.Context) {

	gameID := c.Param("id")

	h.mutex.RLock()
	gameController, exists := h.games[gameID]
	h.mutex.RUnlock()

	if !exists {
		slog.Debug("game not found", "game_id", gameID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Game ID not found"})
		return
	}

	gameController.Connect(c, &h.upgrader)

}

// removeGameCallback is a callback function that the Run() goroutine calls upon exit, effectively removing itself from the map
// This function locks the mutex, don't use when already holding the lock
func (h *Hub) removeGameCallback(id string, gc *GameController) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.games[id] == gc {
		delete(h.games, id)
	}
}
