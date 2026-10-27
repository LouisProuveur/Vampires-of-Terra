// src/models/event.go
package models

import "encoding/json"

/*
TODO: use an event interface with Event.(type) rather than relying on an Type field and assertion checks
This will also make payload management easier since a payload and event types are currently not bound together
*/

type EventType int

const (
	Unknown EventType = iota
	PrivateMessage
	BroadcastMessage
)

type ChannelEvent struct {
	Type    EventType       `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

func (ce *ChannelEvent) ToEvent(playerID string) (*Event, error) {
	payload, err := NewPayload(ce.Type, ce.Payload)
	if err != nil {
		return nil, err
	}
	return &Event{
		Type:     ce.Type,
		PlayerID: playerID,
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
