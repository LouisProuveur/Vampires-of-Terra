package models

type Player struct {
	name      string
	host      bool
	connected bool
}

func NewPlayer(name string) Player {
	return Player{name: name, host: false, connected: false}
}

func HostPlayer(name string) Player {
	return Player{name: name, host: true, connected: false}
}
