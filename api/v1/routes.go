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
		roles.GET("/:id", roleController.Show)
		roles.PUT("/:id", roleController.Update)
		roles.DELETE("/:id", roleController.Delete)
	}
}

func healthCheck(c *gin.Context) {
	response.Success(c, http.StatusOK, "SafariQuest API is running", nil)
}
