// src/models/game.go

// Package models defines the data structures for the application.
// It defines methods and functions to handle the data, and structures for JSON Marshalling and Unmarshalling
package models

import (
	"Vampires-of-Terra/apperrors"
)

// Game represents the state of the game itself.
type Game struct {
	players    map[string]*Player
	maxPlayers int
}

// NewGame creates a new instance of a [Game]
func NewGame(maxPlayers int) *Game {
	return &Game{
		players:    make(map[string]*Player, maxPlayers),
		maxPlayers: maxPlayers,
	}
}

// AddPlayer adds a new player to the game.
// It produces an error if the player name is invalid or already exists.
// Name "server" is reserved for the server.
func (g *Game) AddPlayer(playerID string) error {

	if _, exists := g.players[playerID]; exists || playerID == "server" {
		return apperrors.ErrPlayerNameTaken
	}

	length := len(g.players)

	if length == 0 {
		g.players[playerID] = HostPlayer(playerID)
		return nil
	}

	if length >= g.maxPlayers {
		return apperrors.ErrGameFull
	}

	g.players[playerID] = NewPlayer(playerID)

	return nil

}

// GetPlayer returns a [Player] struct associated with the given ID
func (g *Game) GetPlayer(playerID string) *Player {
	return g.players[playerID]
}

// TODO: add game JSON formatting to allow resync

// TODO: implement game state machine
