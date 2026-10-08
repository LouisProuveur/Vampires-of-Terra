package models

type Game struct {
	Players []Player
}

func NewGame() *Game {
	game := new(Game)
	game.Players = make([]Player, 0)

	return game
}

func (g *Game) AddPlayer(player Player) {
	g.Players = append(g.Players, player)
}
