// src/controllers/client.go

// Package controllers is responsible for managing the interactions between the users and the game itself.
// It creates game instances, allows to connect users to games and manages transfer of information between the game and the user.
package controllers

import (
	"Vampires-of-Terra/models"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
	sendBufferSize = 64
)

type client struct {
	playerID string
	game     *GameController
	conn     *websocket.Conn
	send     chan []byte
}

func newClient(id string, gc *GameController, conn *websocket.Conn) *client {
	return &client{
		playerID: id,
		game:     gc,
		conn:     conn,
		send:     make(chan []byte, sendBufferSize),
	}
}

// TODO: event rate limiter
func (c *client) readPump() {
	defer func() {
		c.game.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			slog.Debug("error when reading client message", "player_id", c.playerID, "game_id", c.game.id, "error", err)
			return
		}
		event := &models.ChannelEvent{}
		if err := json.Unmarshal(message, event); err != nil {
			slog.Debug("error when parsing incoming event", "player_id", c.playerID, "game_id", c.game.id, "error", err)
			continue
		}

		c.game.QueueEvent(event, c.playerID)
	}
}

func (c *client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		slog.Debug("client connection closed", "player_id", c.playerID, "game_id", c.game.id)
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteControl(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseGoingAway, ""),
					time.Now().Add(time.Second),
				)
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
