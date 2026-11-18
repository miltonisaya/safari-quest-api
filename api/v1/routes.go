package v1

import (
	"net/http"
	attractionType "safari-quest-api/internal/attractiontype"

	"safari-quest-api/internal/auth"
	"safari-quest-api/internal/authority"
	"safari-quest-api/internal/geography"
	"safari-quest-api/internal/geographytype"
	"safari-quest-api/internal/menu"
	"safari-quest-api/internal/role"
	"safari-quest-api/internal/user"
	"safari-quest-api/middlewares"
	"safari-quest-api/pkg/response"
	"safari-quest-api/websocket"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes decides which middleware each feature sits behind; the
// features themselves declare their own paths in their <struct>_routes.go files.
func RegisterRoutes(rg *gin.RouterGroup) {
	// Public routes — no authentication required.
	rg.GET("/health", healthCheck)
	auth.RegisterRoutes(rg)

	// WebSocket — authenticated, but not behind Authorize().
	websocket.RegisterRoutes(rg.Group("", middlewares.Auth()))

	// Protected routes — Auth() validates the JWT and Authorize() automatically
	// derives the required authority code from the route pattern and HTTP method,
	// then checks the user holds it via their roles. No per-route configuration needed.
	protected := rg.Group("", middlewares.Auth(), middlewares.Authorize())
	role.RegisterRoutes(protected)
	user.RegisterRoutes(protected)
	menu.RegisterRoutes(protected)
	geographytype.RegisterRoutes(protected)
	geography.RegisterRoutes(protected)
	authority.RegisterRoutes(protected)
	attractionType.RegisterRoutes(protected)
}

func healthCheck(c *gin.Context) {
	response.Success(c, http.StatusOK, "SafariQuest API is running", nil)
}
