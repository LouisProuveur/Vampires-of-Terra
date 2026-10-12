package models

import "encoding/json"

type EventType int

const (
	ErrorMessage EventType = iota
	PrivateMessage
	BroadcastMessage
	Other
)

type Event struct {
	Type    EventType       `json:"type"`
	Token   string          `json:"token"`
	Payload json.RawMessage `json:"payload,omitempty"`
}
