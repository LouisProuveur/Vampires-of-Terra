// src/models/event.go
package models

import (
	"Vampires-of-Terra/apperrors"
	"encoding/json"
)

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

func (ce *ChannelEvent) ToJSON() ([]byte, error) {
	return json.Marshal(ce)
}

func (ce *ChannelEvent) toEvent[P Payload](payload P, playerID string) (Event, error) {
	if err := payload.FromJSON(ce.Payload); err != nil {
		return nil, err
	}
	return NewEvent(payload, playerID), nil
}

func (ce *ChannelEvent) ToEvent(playerID string) (Event, error) {

	switch ce.Type {
	case PrivateMessage:
		return ce.toEvent(&PrivateMessagePayload{}, playerID)
	case BroadcastMessage:
		return ce.toEvent(&BroadcastMessagePayload{}, playerID)
	default:
		return nil, apperrors.ErrUnkownPayloadType
	}
}

type Event interface {
	ToChannelEvent() (*ChannelEvent, error)
	isEvent()
}

type TypedEvent[P Payload] struct {
	PlayerID string
	Payload  P
}

func NewEvent[P Payload](p P, playerID string) *TypedEvent[P] {

	return &TypedEvent[P]{
		PlayerID: playerID,
		Payload:  p,
	}
}

func (e *TypedEvent[P]) ToChannelEvent() (*ChannelEvent, error) {
	payloadJSON, err := e.Payload.ToJSON()
	if err != nil {
		return nil, err
	}
	return &ChannelEvent{
		Type:    e.Payload.EventType(),
		Payload: payloadJSON,
	}, nil
}

func (e *TypedEvent[P]) isEvent() {}

type (
	PrivateMessageEvent   = *TypedEvent[*PrivateMessagePayload]
	BroadcastMessageEvent = *TypedEvent[*BroadcastMessagePayload]
)
