package account

import "github.com/gin-gonic/gin"

// RegisterPublicRoutes mounts the password reset endpoints; rg must not require a token.
func RegisterPublicRoutes(rg *gin.RouterGroup) {
	controller := Controller{}
	rg.POST("/auth/forgot-password", controller.ForgotPassword)
	rg.POST("/auth/reset-password", controller.ResetPassword)
}

// RegisterRoutes mounts the signed-in user's own profile endpoints. rg must
// require a token but no authority: every account may manage itself.
func RegisterRoutes(rg *gin.RouterGroup) {
	controller := Controller{}
	rg.GET("/me", controller.Me)
	rg.PUT("/me", controller.UpdateMe)
	rg.PUT("/me/password", controller.ChangeMyPassword)
}
