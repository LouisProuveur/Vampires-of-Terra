// src/models/payload.go
package models

import "encoding/json"

// ---- Generic Payload interface for all event payloads ----

type Payload interface {
	isPayload()
	FromJSON(json.RawMessage) error
	ToJSON() ([]byte, error)
	EventType() EventType
}

// ---- Payload for PrivateMessage Event ----

type PrivateMessagePayload struct {
	To      []string `json:"to"`
	From    string   `json:"from,omitempty"`
	Message string   `json:"message"`
}

func (*PrivateMessagePayload) isPayload() {}

func (pmp *PrivateMessagePayload) ToJSON() ([]byte, error) {
	return json.Marshal(pmp)
}

func (pmp *PrivateMessagePayload) FromJSON(data json.RawMessage) error {
	return json.Unmarshal(data, pmp)
}

func (*PrivateMessagePayload) EventType() EventType {
	return PrivateMessage
}

// ---- Payload for BroadcastMessage Event ----

type BroadcastMessagePayload struct {
	From    string `json:"from,omitempty"`
	Message string `json:"message"`
}

func (*BroadcastMessagePayload) isPayload() {}

func (bmp *BroadcastMessagePayload) ToJSON() ([]byte, error) {
	return json.Marshal(bmp)
}

func (bmp *BroadcastMessagePayload) FromJSON(data json.RawMessage) error {
	return json.Unmarshal(data, bmp)
}

func (*BroadcastMessagePayload) EventType() EventType {
	return BroadcastMessage
}
