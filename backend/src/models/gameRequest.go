package models

type GameRequest struct {
	ID         string `json:"game_id"`
	Password   string `json:"password"`
	PlayerName string `json:"player_name"`
}
