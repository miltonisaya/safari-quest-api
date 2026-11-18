package menu

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts the menu endpoints on rg.
// rg is expected to carry Auth() and Authorize(); see api/v1/routes.go.
func RegisterRoutes(rg *gin.RouterGroup) {
	controller := Controller{}
	group := rg.Group("/menus")
	{
		group.GET("", controller.Index)
		group.POST("", controller.Create)
		group.GET("/:uuid", controller.Show)
		group.PUT("/:uuid", controller.Update)
		group.DELETE("/:uuid", controller.Delete)
	}
}
