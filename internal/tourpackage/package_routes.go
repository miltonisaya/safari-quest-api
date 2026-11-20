package tourpackage

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts the package endpoints on rg.
func RegisterRoutes(rg *gin.RouterGroup) {
	controller := Controller{}
	group := rg.Group("/packages")
	{
		group.GET("", controller.Index)
		group.POST("", controller.Create)
		group.GET("/:uuid", controller.Show)
		group.PUT("/:uuid", controller.Update)
		group.DELETE("/:uuid", controller.Delete)
	}
}
