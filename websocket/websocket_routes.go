package websocket

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts the WebSocket endpoints on rg.
// rg must carry Auth(): the JWT arrives as ?token=<jwt>. It is not behind
// Authorize() because a socket is a persistent connection, not a resource action.
func RegisterRoutes(rg *gin.RouterGroup) {
	controller := Controller{}
	rg.GET("/ws", controller.Connect)
}
