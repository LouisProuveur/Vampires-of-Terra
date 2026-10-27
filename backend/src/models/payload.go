// src/models/payload.go
package models

import "encoding/json"

// ---- Generic Payload interface for all event payloads ----

type Payload interface {
	isPayload()
	ToJSON() ([]byte, error)
}

func NewPayload(etype EventType, msg json.RawMessage) (Payload, error) {
	var payload Payload
	switch etype {
	case PrivateMessage:
		payload = &PrivateMessagePayload{}
	case BroadcastMessage:
		payload = &BroadcastMessagePayload{}
	default:
		return nil, nil
	}
	if err := json.Unmarshal(msg, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

// ---- Payload for PrivateMessage Event ----

type PrivateMessagePayload struct {
	To      string `json:"to"`
	From    string `json:"from,omitempty"`
	Message string `json:"message"`
}

func (_ *PrivateMessagePayload) isPayload() {}

func (pmp *PrivateMessagePayload) ToJSON() ([]byte, error) {
	return json.Marshal(pmp)
}

// ---- Payload for BroadcastMessage Event ----

type BroadcastMessagePayload struct {
	From    string `json:"from,omitempty"`
	Message string `json:"message"`
}

func (_ *BroadcastMessagePayload) isPayload() {}

func (bmp *BroadcastMessagePayload) ToJSON() ([]byte, error) {
	return json.Marshal(bmp)
}
