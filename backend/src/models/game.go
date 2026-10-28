// src/models/game.go
package models

import (
	"Vampires-of-Terra/apperrors"
)

type Game struct {
	players    map[string]*Player
	maxPlayers int
}

/*
Creates a new instance of Game
*/
func NewGame(maxPlayers int) *Game {
	return &Game{
		players:    make(map[string]*Player, maxPlayers),
		maxPlayers: maxPlayers,
	}
}

/*
Adds a player to the game.
Produces an error if the ID is already taken.
*/
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

func (g *Game) GetPlayer(playerID string) *Player {
	return g.players[playerID]
}

// TODO: add game JSON formatting to allow resync

// TODO: implement game state machine
