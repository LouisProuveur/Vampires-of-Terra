package controllers

import (
	"Vampires-of-Terra/models"
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

type Client struct {
	playerID string
	game     *GameController
	conn     *websocket.Conn
	send     chan []byte
}

func NewClient(id string, gc *GameController, conn *websocket.Conn) *Client {
	return &Client{
		playerID: id,
		game:     gc,
		conn:     conn,
		send:     make(chan []byte),
	}
}

func (c *Client) getName() string {
	return c.playerID
}

func (c *Client) readPump() {
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
		var event models.Event
		if err := c.conn.ReadJSON(&event); err != nil {
			return
		}
		c.game.QueueEvent(event)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
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
