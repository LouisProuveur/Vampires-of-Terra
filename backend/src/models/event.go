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

func (ce *ChannelEvent) ToEvent(playerName string) (*Event, error) {
	payload, err := NewPayload(ce.Type, ce.Payload)
	if err != nil {
		return nil, err
	}
	return &Event{
		Type:     ce.Type,
		PlayerID: playerName,
		Payload:  payload,
	}, nil
}

func (ce *ChannelEvent) ToJSON() ([]byte, error) {
	return json.Marshal(ce)
}

type Event struct {
	Type     EventType
	PlayerID string
	Payload  Payload
}

func (e *Event) ToChannelEvent() (*ChannelEvent, error) {
	payloadJSON, err := e.Payload.ToJSON()
	if err != nil {
		return nil, err
	}
	return &ChannelEvent{
		Type:    e.Type,
		Payload: payloadJSON,
	}, nil
}
