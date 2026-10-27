// src/controllers/game.go
package controllers

import (
	"Vampires-of-Terra/apperrors"
	"Vampires-of-Terra/models"
	"context"
	"crypto/rand"
	"encoding/json"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
)

const (
	GameIdleTimeout = 1 * time.Minute
	GracePeriod     = 30 * time.Second
)

type playerClaims struct {
	PlayerName string `json:"pid"`
	GameID     string `json:"gid"`
	jwt.RegisteredClaims
}

type playerJoinRequest struct {
	playerName string
	result     chan error
}

type GameController struct {
	game           *models.Game
	id             string
	passwordHash   []byte
	secret         []byte
	cancel         context.CancelFunc
	clients        map[string]*Client // maps player IDs to clients
	playerJoin     chan *playerJoinRequest
	incomingEvents chan *models.Event
	register       chan *Client
	unregister     chan *Client
	done           chan struct{}
}

/*
Creates a new gamecontroller instance with password and id
Returns the gamecontroller instance
*/
func CreateGameController(id string, passwordHash []byte, maxPlayers int, cancel context.CancelFunc) (*GameController, error) {

	game := models.NewGame(maxPlayers)
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}

	return &GameController{
		game:           game,
		id:             id,
		passwordHash:   passwordHash,
		secret:         secret,
		cancel:         cancel,
		clients:        make(map[string]*Client),
		incomingEvents: make(chan *models.Event),
		playerJoin:     make(chan *playerJoinRequest),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		done:           make(chan struct{}),
	}, nil
}

/*
Hashes a password in a way that is expected by the game controller.
*/
func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

/*
Verifies the provided password against the stored hash
*/
func (gc *GameController) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword(gc.passwordHash, []byte(password))
	return err == nil
}

/*
Produces a JWT holding the player's name and associated game ID.
These tokens are only valid for the game they were created for.
*/
func (gc *GameController) signToken(playerName string) (string, error) {
	claims := playerClaims{
		PlayerName: playerName,
		GameID:     gc.id,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(gc.secret)
}

/*
Parses and validates a player's token.
A token is only valid for the given game instance they were produced for.
Returns the player ID
*/
func (gc *GameController) parseToken(tokenStr string) (string, error) {
	claims := &playerClaims{}

	token, err := jwt.ParseWithClaims(
		tokenStr,
		claims,
		func(t *jwt.Token) (any, error) { return gc.secret, nil },
		jwt.WithValidMethods([]string{"HS256"}),
	)
	if err != nil {
		return "", err
	}

	if !token.Valid || claims.GameID != gc.id {
		return "", jwt.ErrTokenInvalidClaims
	}

	return claims.PlayerName, nil
}

/*
Registers a player into the game.
Returns the player's token
*/
func (gc *GameController) JoinGame(playerName string) (string, error) {

	token, err := gc.signToken(playerName)
	if err != nil {
		return "", err
	}

	req := &playerJoinRequest{
		playerName: playerName,
		result:     make(chan error, 1),
	}

	select {
	case gc.playerJoin <- req:
	case <-gc.done:
		return "", apperrors.ErrGameClosed
	}

	select {
	case err := <-req.result:
		if err != nil {
			return "", err
		}
	case <-gc.done:
		return "", apperrors.ErrGameClosed
	}

	return token, nil

}

func (gc *GameController) Connect(c *gin.Context, upgrader *websocket.Upgrader) {

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("websocker error failed", "game_id", gc.id, "error", err)
		return
	}

	// Wait 5 secs for client to send auth message
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	conn.SetReadLimit(512)
	_, message, err := conn.ReadMessage()
	if err != nil {
		slog.Warn("client authenication failed", "ip", conn.RemoteAddr(), "error", err)
		conn.Close()
		return
	}

	id, err := gc.parseToken(string(message))
	if err != nil {
		slog.Warn("invalid token was used", "ip", c.ClientIP())
		conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "unauthorized access"),
			time.Now().Add(time.Second),
		)
		conn.Close()
		return
	}

	slog.Debug("new client was created", "player_id", id)
	client := NewClient(id, gc, conn)
	gc.Register(client)

}

func (gc *GameController) Register(client *Client) {
	select {
	case gc.register <- client:
		go client.readPump()
		go client.writePump()
	case <-gc.done:
		client.conn.Close()
	}
}

func (gc *GameController) Unregister(client *Client) {
	select {
	case gc.unregister <- client:
	case <-gc.done:
	}
}

func (gc *GameController) QueueEvent(ev *models.ChannelEvent, playerName string) {
	event, err := ev.ToEvent(playerName)
	if err != nil {
		slog.Error("failed to convert channel event to event", "game_id", gc.id, "player_id", playerName, "error", err)
		gc.sendToPlayer(playerName, &models.ChannelEvent{
			Type:    models.ErrorMessage,
			Payload: json.RawMessage(`{"message":"failed to process event"}`),
		})
		return
	}

	select {
	case gc.incomingEvents <- event:
	case <-gc.done:
	}
}

func (gc *GameController) Run(ctx context.Context, callback func(string, *GameController)) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("event loop crashed", "type", "game_loop_crash", "game_id", gc.id, "panic", r, "stack", debug.Stack())
		}
		slog.Debug("game has been closed", "game_id", gc.id)
	}()
	defer callback(gc.id, gc)
	defer close(gc.done)
	defer func() {
		for id, c := range gc.clients {
			close(c.send)
			delete(gc.clients, id)
		}
	}()
	defer gc.cancel()

	slog.Debug("game loop has been started", "game_id", gc.id)

	// TODO: implement resync
	timer := time.NewTimer(GameIdleTimeout)
	grace := false
	for {
		select {
		case <-ctx.Done():
			slog.Debug("game has been closed", "game_id", gc.id)
			return
		case p := <-gc.playerJoin:
			p.result <- gc.game.AddPlayer(p.playerName)
			slog.Debug("player added to the game", "game_id", gc.id, "player_id", p.playerName)
		case c := <-gc.register:
			if old, exists := gc.clients[c.playerID]; exists {
				close(old.send)
			}
			gc.clients[c.playerID] = c
			if player := gc.game.GetPlayer(c.playerID); player != nil {
				player.Connected = true
			}
			slog.Debug("player connected to the game", "game_id", gc.id, "player_id", c.playerID)
		case c := <-gc.unregister:
			if gc.clients[c.playerID] == c {
				delete(gc.clients, c.playerID)
				close(c.send)
				if player := gc.game.GetPlayer(c.playerID); player != nil {
					player.Connected = false
				}
			}
			slog.Debug("player disconnected from the game", "game_id", gc.id, "player_id", c.playerID)
			continue // Leaving does not reset grace or timer
		case e := <-gc.incomingEvents:
			gc.HandleEvent(e)

		// If noone is connected or grace period is over, return. Else, warn clients and enter grace period
		case <-timer.C:
			if grace || len(gc.clients) == 0 {
				return
			} else {
				grace = true
				timer.Reset(GracePeriod)
				// warn clients that game is about to close
				slog.Debug("game is idle, entering grace period", "game_id", gc.id)
				continue
			}
		}
		timer.Reset(GameIdleTimeout)
		grace = false
	}
}

func (gc *GameController) HandleEvent(event *models.Event) {

	switch ev_type := event.Type; ev_type {

	case models.BroadcastMessage:

		payload := event.Payload.(*models.BroadcastMessagePayload)
		// make sure "from" field is set to correct value
		payload.From = event.PlayerID

		cevent, err := event.ToChannelEvent()
		if err != nil {
			slog.Error("failed to convert event to channel event", "game_id", gc.id, "player_id", event.PlayerID, "error", err)
			return
		}

		gc.sendToAll(cevent)

	case models.PrivateMessage:

		payload := event.Payload.(*models.PrivateMessagePayload)

		cevent, err := event.ToChannelEvent()
		if err != nil {
			slog.Error("failed to convert event to channel event", "game_id", gc.id, "player_id", event.PlayerID, "error", err)
			return
		}

		gc.sendToPlayer(payload.To, cevent)

	default:
		// TODO: send all other events to game state machine
		slog.Warn("unknown event type", "game_id", gc.id, "player_id", event.PlayerID, "event_type", ev_type)

	}
}

/*
sends a channel event to all connected clients
*/
func (gc *GameController) sendToAll(cevent *models.ChannelEvent) {

	data, err := cevent.ToJSON()
	if err != nil {
		slog.Error("failed to serialize event", "game_id", gc.id, "error", err)
		return
	}

	for _, c := range gc.clients {
		select {
		case c.send <- data:
		default:
			slog.Warn("client send buffer full, dropping message", "player_id", c.playerID, "game_id", gc.id)
		}
	}
}

/*
sends a channel event to the client with a given player ID
*/
func (gc *GameController) sendToPlayer(playerID string, cevent *models.ChannelEvent) {

	client, exists := gc.clients[playerID]
	if !exists {
		slog.Debug("sending message to non-existent client", "player_id", playerID, "game_id", gc.id)
		return
	}

	data, err := cevent.ToJSON()
	if err != nil {
		slog.Error("failed to serialize event", "game_id", gc.id, "error", err)
		return
	}

	select {
	case client.send <- data:
	default:
		slog.Warn("client send buffer full, dropping message", "player_id", client.playerID, "game_id", gc.id)
	}
}
