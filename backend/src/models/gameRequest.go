// src/models/gameRequest.go

// Package models defines the data structures for the application.
// It defines methods and functions to handle the data, and structures for JSON Marshalling and Unmarshalling
package models

// gameRequest is the common part of [CreateGameRequest] and [JoinGameRequest]
type gameRequest struct {
	ID       string `json:"game_id" binding:"required,alphanum,min=1,max=10"`
	Password string `json:"password" binding:"required,gamePasswordLength"`
	PlayerID string `json:"player_id" binding:"required,playerID,min=1,max=24"`
}

// CreateGameRequest represents a request to create a new game
type CreateGameRequest struct {
	gameRequest
	MaxPlayers int `json:"max_players" binding:"required,min=1,max=24"`
}

// JoinGameRequest represents a request to join a game
type JoinGameRequest struct {
	gameRequest
}
