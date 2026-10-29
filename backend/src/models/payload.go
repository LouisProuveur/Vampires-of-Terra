// src/models/payload.go

// Package models defines the data structures for the application.
// It defines methods and functions to handle the data, and structures for JSON Marshalling and Unmarshalling
package models

import "encoding/json"

// ---- Generic Payload interface for all event payloads ----

// Payload is the data inside an event
type Payload interface {
	isPayload()
	// FromJSON fills a payload based on the provided encoding
	FromJSON(json.RawMessage) error
	// ToJSON returns the JSON encoding of a payload
	ToJSON() ([]byte, error)
	// EventType returns the [EventType] a payload is associated with
	EventType() EventType
}

// ---- Payload for PrivateMessage Event ----

// PrivateMessagePayload is the data inside a [PrivateMessageEvent]
type PrivateMessagePayload struct {
	To      []string `json:"to"`
	From    string   `json:"from,omitempty"`
	Message string   `json:"message"`
}

func (*PrivateMessagePayload) isPayload() {}

// ToJSON returns the JSON encoding of a payload
func (pmp *PrivateMessagePayload) ToJSON() ([]byte, error) {
	return json.Marshal(pmp)
}

// FromJSON fills a payload based on the provided encoding
func (pmp *PrivateMessagePayload) FromJSON(data json.RawMessage) error {
	return json.Unmarshal(data, pmp)
}

// EventType returns the [EventType] a payload is associated with
func (*PrivateMessagePayload) EventType() EventType {
	return PrivateMessage
}

// ---- Payload for BroadcastMessage Event ----

// BroadcastMessagePayload is the data inside a [BroadcastMessageEvent]
type BroadcastMessagePayload struct {
	From    string `json:"from,omitempty"`
	Message string `json:"message"`
}

func (*BroadcastMessagePayload) isPayload() {}

// ToJSON returns the JSON encoding of a payload
func (bmp *BroadcastMessagePayload) ToJSON() ([]byte, error) {
	return json.Marshal(bmp)
}

// FromJSON fills a payload based on the provided encoding
func (bmp *BroadcastMessagePayload) FromJSON(data json.RawMessage) error {
	return json.Unmarshal(data, bmp)
}

// EventType returns the [EventType] a payload is associated with
func (*BroadcastMessagePayload) EventType() EventType {
	return BroadcastMessage
}
