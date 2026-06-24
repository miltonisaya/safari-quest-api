package v1

import (
	"net/http"

	"safari-quest-api/controllers"
	"safari-quest-api/middlewares"
	"safari-quest-api/pkg/response"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup) {
	// Public routes — no authentication required.
	rg.GET("/health", healthCheck)

	authController := controllers.AuthController{}
	rg.POST("/auth/login", authController.Login)

	// Protected routes — Auth() validates the JWT and Authorize() automatically
	// derives the required authority code from the route pattern and HTTP method,
	// then checks the user holds it via their roles. No per-route configuration needed.
	protected := rg.Group("", middlewares.Auth(), middlewares.Authorize())
	{
		roleController := controllers.RoleController{}
		roles := protected.Group("/roles")
		{
			roles.GET("", roleController.Index)
			roles.POST("", roleController.Create)
			roles.GET("/:uuid", roleController.Show)
			roles.PUT("/:uuid", roleController.Update)
			roles.DELETE("/:uuid", roleController.Delete)
		}

		userController := controllers.UserController{}
		users := protected.Group("/users")
		{
			users.GET("", userController.Index)
			users.POST("", userController.Create)
			users.GET("/:uuid", userController.Show)
			users.PUT("/:uuid", userController.Update)
			users.DELETE("/:uuid", userController.Delete)
		}
	}
}

func healthCheck(c *gin.Context) {
	response.Success(c, http.StatusOK, "SafariQuest API is running", nil)
}
