// src/routes/routes.go

// Package routes defines the http endpoints for the game
package routes

import (
	"Vampires-of-Terra/controllers"

	"github.com/gin-gonic/gin"
)

// TODO: rate limit on /new, /join and /connect to prevent DoS
// TODO: size limit on /new, /join and /connect as well
// TODO: or find a way to check that only fronted or app can access the endpoints
// TODO: set trusted proxies (currently trusting all proxies)

// SetupRouter sets up a gin router associated with a given [controllers.Hub] and returns it.
func SetupRouter(hub *controllers.Hub) *gin.Engine {

	router := gin.Default()
	router.POST("/new", hub.CreateGame)
	router.POST("/join", hub.JoinGame)
	router.GET("/connect/:id", hub.Connect)

	return router

}
