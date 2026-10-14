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
Produces an error if the name is already taken.
*/
func (g *Game) AddPlayer(playerName string) error {

	length := len(g.players)

	if length == 0 {
		g.players[playerName] = HostPlayer(playerName)
		return nil
	}

	if length >= g.maxPlayers {
		return apperrors.ErrGameFull
	}

	if _, exists := g.players[playerName]; !exists {
		g.players[playerName] = NewPlayer(playerName)
	} else {
		return apperrors.ErrPlayerNameTaken
	}

	return nil

}

func (g *Game) GetPlayer(playerName string) *Player {
	return g.players[playerName]
}

// TODO: add game JSON formatting to allow resync

// TODO: implement game state machine
