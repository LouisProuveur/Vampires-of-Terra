// src/models/event.go

// Package models defines the data structures for the application.
// It defines methods and functions to handle the data, and structures for JSON Marshalling and Unmarshalling
package models

import (
	"Vampires-of-Terra/apperrors"
	"encoding/json"
)

// EventType defines the different types of events that are acceptable within the game
type EventType int

const (
	Unknown EventType = iota
	PrivateMessage
	BroadcastMessage
)

// ChannelEvent represents the structure of an event on the wire. An event has a type and a payload.
type ChannelEvent struct {
	Type    EventType       `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// ToJSON returns the JSON encoding of the [ChannelEvent].
func (ce *ChannelEvent) ToJSON() ([]byte, error) {
	return json.Marshal(ce)
}

// toEvent is a helper function for [ToEvent], allowing the same behaviour for different kinds of [models.Payload]
func (ce *ChannelEvent) toEvent[P Payload](payload P, playerID string) (Event, error) {
	if err := payload.FromJSON(ce.Payload); err != nil {
		return nil, err
	}
	return NewEvent(payload, playerID), nil
}

// ToEvent casts a [ChannelEvent] to an [Event] that can be handled by the event loop and the game state machine.
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

// Event defines an interface to interact with the different kinds of events.
type Event interface {
	ToChannelEvent() (*ChannelEvent, error)
	isEvent()
}

type typedEvent[P Payload] struct {
	PlayerID string
	Payload  P
}

func NewEvent[P Payload](p P, playerID string) *typedEvent[P] {

	return &typedEvent[P]{
		PlayerID: playerID,
		Payload:  p,
	}
}

func (e *typedEvent[P]) ToChannelEvent() (*ChannelEvent, error) {
	payloadJSON, err := e.Payload.ToJSON()
	if err != nil {
		return nil, err
	}
	return &ChannelEvent{
		Type:    e.Payload.EventType(),
		Payload: payloadJSON,
	}, nil
}

func (e *typedEvent[P]) isEvent() {}

type (
	// PrivateMessageEvent represents a message sent only to a set of players
	PrivateMessageEvent = *typedEvent[*PrivateMessagePayload]
	// BroadcastMessageEvent represents a message sent to all players
	BroadcastMessageEvent = *typedEvent[*BroadcastMessagePayload]
)
