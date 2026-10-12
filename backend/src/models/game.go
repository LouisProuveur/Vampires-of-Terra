package models

import (
	"Vampires-of-Terra/errors"
)

type Game struct {
	players map[string]Player
}

/*
Creates a new instance of Game
*/
func NewGame() *Game {
	game := new(Game)
	game.players = make(map[string]Player)

	return game
}

/*
Adds a player to the game.
Produces an error if the name is already taken.
*/
func (g *Game) AddPlayer(playerName string) error {

	if len(g.players) == 0 {
		g.players[playerName] = HostPlayer(playerName)
	} else if _, exists := g.players[playerName]; !exists {
		g.players[playerName] = NewPlayer(playerName)
	} else {
		return errors.ErrPlayerNameTaken
	}

	return nil

}
