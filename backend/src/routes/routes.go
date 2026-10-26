// src/routes/routes.go
package routes

import (
	"Vampires-of-Terra/controllers"

	"github.com/gin-gonic/gin"
)

// TODO: rate limit on /new and /join to prevent DoS
// TODO: size limit on /new and /join as well
// TODO: set trusted proxies (currently trusting all proxies)
func SetupRouter(hub *controllers.Hub) *gin.Engine {

	router := gin.Default()
	router.POST("/new", hub.CreateGame)
	router.POST("/join", hub.JoinGame)
	router.GET("/connect/:id", hub.Connect)

	return router

}
