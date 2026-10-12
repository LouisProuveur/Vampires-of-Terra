// src/controller/game.go
package controllers

import (
	"Vampires-of-Terra/models"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
)

// TODO: change origin checker
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type PlayerClaims struct {
	PlayerName string `json:"pid"`
	GameID     string `json:"gid"`
	jwt.RegisteredClaims
}

type GameController struct {
	game           *models.Game
	id             string
	password_hash  []byte
	cancel         context.CancelFunc
	clients        map[string]*Client // maps player IDs to clients
	incomingEvents chan models.Event
	register       chan *Client
	unregister     chan *Client
	done           chan struct{}
}

/*
Creates a new gamecontroller instance with password and id
Returns the gamecontroller instance
*/
func CreateGameController(id string, password string, cancel context.CancelFunc) (*GameController, error) {

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	game := models.NewGame()

	return &GameController{
		game:           game,
		id:             id,
		password_hash:  passwordHash,
		cancel:         cancel,
		clients:        make(map[string]*Client),
		incomingEvents: make(chan models.Event),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		done:           make(chan struct{}),
	}, nil
}

/*
Verifies the provided password against the stored hash
*/
func (gc *GameController) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword(gc.password_hash, []byte(password))
	return err == nil
}

/*
Produces a JWT holding the player's name and associated game ID.
These tokens are only valid for the game they were created for.
*/
func (gc *GameController) signToken(playerName string) (string, error) {
	claims := PlayerClaims{
		PlayerName: playerName,
		GameID:     gc.id,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString("hello there")
}

/*
Parses and validates a player's token.
A token is only valid for the given game instance they were produced for.
Returns the player ID
*/
func (gc *GameController) parseToken(tokenStr string) (string, error) {
	claims := &PlayerClaims{}

	token, err := jwt.ParseWithClaims(
		tokenStr,
		claims,
		func(t *jwt.Token) (any, error) { return "hello there", nil },
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

	if err := gc.game.AddPlayer(playerName); err != nil {
		return "", err
	}

	token, err := gc.signToken(playerName)
	if err != nil {
		return "", err
	}

	return token, nil

}

func (gc *GameController) Connect(c *gin.Context) {

	id, err := gc.parseToken(c.Query("token"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := NewClient(id, gc, conn)
	gc.Register(client, c.Query("token"))

}

func (gc *GameController) Register(client *Client, token string) {
	select {
	case gc.register <- client:
	case <-gc.done:
	}
}

func (gc *GameController) Unregister(client *Client) {
	select {
	case gc.unregister <- client:
	case <-gc.done:
	}
}

func (gc *GameController) QueueEvent(ev models.Event) {
	select {
	case gc.incomingEvents <- ev:
	case <-gc.done:
	}
}

func (gc *GameController) Run(ctx context.Context) {
	defer close(gc.done)
	defer func() {
		for id, c := range gc.clients {
			close(c.send)
			delete(gc.clients, id)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case c := <-gc.register:
			if old, exists := gc.clients[c.playerID]; exists {
				close(old.send)
			}
			gc.clients[c.playerID] = c
		case c := <-gc.unregister:
			if gc.clients[c.playerID] == c {
				delete(gc.clients, c.playerID)
				close(c.send)
			}
		case e := <-gc.incomingEvents:
			gc.HandleEvent(e)
		}
	}
}

func (gc *GameController) HandleEvent(event models.Event) {

}
