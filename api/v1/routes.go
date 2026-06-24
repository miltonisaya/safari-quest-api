package v1

import (
	"net/http"

	"safari-quest-api/controllers"
	"safari-quest-api/pkg/response"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/health", healthCheck)

	roleController := controllers.RoleController{}
	roles := rg.Group("/roles")
	{
		roles.GET("", roleController.Index)
		roles.POST("", roleController.Create)
		roles.GET("/:uuid", roleController.Show)
		roles.PUT("/:uuid", roleController.Update)
		roles.DELETE("/:uuid", roleController.Delete)
	}

	userController := controllers.UserController{}
	users := rg.Group("/users")
	{
		users.GET("", userController.Index)
		users.POST("", userController.Create)
		users.GET("/:uuid", userController.Show)
		users.PUT("/:uuid", userController.Update)
		users.DELETE("/:uuid", userController.Delete)
	}
}

func healthCheck(c *gin.Context) {
	response.Success(c, http.StatusOK, "SafariQuest API is running", nil)
}
