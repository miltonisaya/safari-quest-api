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

	// WebSocket — requires a valid JWT passed as ?token=<jwt> query param.
	// Not behind Authorize() because it is a persistent connection, not a
	// discrete resource action.
	wsController := controllers.WSController{}
	rg.GET("/ws", middlewares.Auth(), wsController.Connect)

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
			roles.GET("/:uuid/authorities", roleController.Authorities)
			roles.PUT("/:uuid/authorities", roleController.UpdateAuthorities)
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

		menuController := controllers.MenuController{}
		menus := protected.Group("/menus")
		{
			menus.GET("", menuController.Index)
			menus.POST("", menuController.Create)
			menus.GET("/:uuid", menuController.Show)
			menus.PUT("/:uuid", menuController.Update)
			menus.DELETE("/:uuid", menuController.Delete)
		}

		geographyTypeController := controllers.GeographyTypeController{}
		geographyTypes := protected.Group("/geography-types")
		{
			geographyTypes.GET("", geographyTypeController.Index)
			geographyTypes.POST("", geographyTypeController.Create)
			geographyTypes.GET("/:uuid", geographyTypeController.Show)
			geographyTypes.PUT("/:uuid", geographyTypeController.Update)
			geographyTypes.DELETE("/:uuid", geographyTypeController.Delete)
		}

		geographyController := controllers.GeographyController{}
		geographies := protected.Group("/geographies")
		{
			geographies.GET("", geographyController.Index)
			geographies.POST("", geographyController.Create)
			geographies.GET("/:uuid", geographyController.Show)
			geographies.PUT("/:uuid", geographyController.Update)
			geographies.DELETE("/:uuid", geographyController.Delete)
		}

		authorityController := controllers.AuthorityController{}
		authorities := protected.Group("/authorities")
		{
			authorities.GET("", authorityController.Index)
		}
	}
}

func healthCheck(c *gin.Context) {
	response.Success(c, http.StatusOK, "SafariQuest API is running", nil)
}
