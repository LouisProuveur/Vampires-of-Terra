// src/models/gameRequest.go
package models

type GameRequest struct {
	ID       string `json:"game_id" binding:"required,alphanum,min=1,max=10"`
	Password string `json:"password" binding:"required,gamePasswordLength"`
	PlayerID string `json:"player_id" binding:"required,playerID,min=1,max=24"`
}

type CreateGameRequest struct {
	GameRequest
	MaxPlayers int `json:"max_players" binding:"required,min=1,max=24"`
}

type JoinGameRequest struct {
	GameRequest
}
