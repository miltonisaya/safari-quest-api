package websocket

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Controller struct{}

// Connect godoc
//
//	@Summary		Open WebSocket connection
//	@Description	Upgrades the HTTP connection to WebSocket. The client will receive real-time events whenever a menu, role, or user is created, updated, or deleted. Pass the JWT as a query parameter: ?token=<jwt>
//	@Tags			WebSocket
//	@Produce		json
//	@Param			token	query		string	true	"Bearer token"
//	@Success		101		{string}	string	"Switching Protocols"
//	@Failure		401		{object}	response.CustomApiResponse
//	@Router			/ws [get]
func (controller Controller) Connect(c *gin.Context) {
	// Auth() has validated the token and stored the user.
	userID, _ := c.Get("userUUID")
	id, _ := userID.(uuid.UUID)
	ServeWS(GlobalHub, c.Writer, c.Request, id)
}
