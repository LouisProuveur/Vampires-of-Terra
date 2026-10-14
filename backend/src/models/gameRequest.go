// src/models/gameRequest.go
package models

type CreateGameRequest struct {
	ID         string `json:"game_id" binding:"required,alphanum,min=1,max=10"`
	Password   string `json:"password" binding:"required,min=1,max=72"`
	PlayerName string `json:"player_name" binding:"required,playername,min=1,max=24"`
	MaxPlayers int    `json:"max_players" binding:"required,min=1,max=24"`
}

type JoinGameRequest struct {
	ID         string `json:"game_id" binding:"required,alphanum,min=1,max=10"`
	Password   string `json:"password" binding:"required,min=1,max=72"`
	PlayerName string `json:"player_name" binding:"required,playername,min=1,max=24"`
}
