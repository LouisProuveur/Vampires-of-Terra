// src/models/event.go
package models

import "encoding/json"

type EventType int

const (
	ErrorMessage EventType = iota
	PrivateMessage
	BroadcastMessage
)

type ChannelEvent struct {
	Type    EventType       `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type Event struct {
	Type     EventType
	PlayerID string
	Payload  json.RawMessage
}

func (ie *ChannelEvent) ToEvent(playerName string) *Event {
	return &Event{
		Type:     ie.Type,
		PlayerID: playerName,
		Payload:  ie.Payload,
	}
}

func (ie *ChannelEvent) ToJSON() ([]byte, error) {
	return json.Marshal(ie)
}
