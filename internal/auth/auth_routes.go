package auth

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts the authentication endpoints on rg.
// These routes are public: rg must not require a token.
func RegisterRoutes(rg *gin.RouterGroup) {
	controller := Controller{}
	rg.POST("/auth/login", controller.Login)
	rg.POST("/auth/register", controller.Register)
}
