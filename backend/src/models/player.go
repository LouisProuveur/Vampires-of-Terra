// src/models/player.go

// Package models defines the data structures for the application.
// It defines methods and functions to handle the data, and structures for JSON Marshalling and Unmarshalling
package models

// Player represents the game data associated with a given player in the game
type Player struct {
	Name      string
	Host      bool
	Connected bool
}

// NewPlayer creates a new [Player] instance that is not the host
func NewPlayer(name string) *Player {
	return &Player{Name: name, Host: false, Connected: false}
}

// HostPlayer creates a new [Player] instance that is the game host
func HostPlayer(name string) *Player {
	return &Player{Name: name, Host: true, Connected: false}
}
