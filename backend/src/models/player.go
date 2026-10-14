// src/models/player.go
package models

type Player struct {
	Name      string
	Host      bool
	Connected bool
}

func NewPlayer(name string) *Player {
	return &Player{Name: name, Host: false, Connected: false}
}

func HostPlayer(name string) *Player {
	return &Player{Name: name, Host: true, Connected: false}
}
