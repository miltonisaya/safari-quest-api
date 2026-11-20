package fileresource

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts the file resource endpoints on rg.
func RegisterRoutes(rg *gin.RouterGroup) {
	controller := Controller{}
	group := rg.Group("/file-resources")
	{
		group.GET("", controller.Index)
		group.POST("", controller.Create)
		group.GET("/:uuid", controller.Show)
		group.PUT("/:uuid", controller.Update)
		group.DELETE("/:uuid", controller.Delete)
	}
}
