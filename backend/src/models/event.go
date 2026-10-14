// src/models/event.go
package models

import "encoding/json"

type EventType int

const (
	Unknown EventType = iota
	ErrorMessage
	PrivateMessage
	BroadcastMessage
)

type IncomingEvent struct {
	Type    EventType       `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type Event struct {
	Type     EventType
	PlayerID string
	Payload  json.RawMessage
}

func (ie *IncomingEvent) ToEvent(playerName string) *Event {
	return &Event{
		Type:     ie.Type,
		PlayerID: playerName,
		Payload:  ie.Payload,
	}
}
