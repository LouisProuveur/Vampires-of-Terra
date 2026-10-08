// src/routes/routes.go
package routes

import (
	"Vampires-of-Terra/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRouter(hub *controllers.Hub) *gin.Engine {

	router := gin.Default()
	router.POST("/new", hub.CreateGame)

	return router

}
