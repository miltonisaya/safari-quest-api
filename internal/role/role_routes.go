package role

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts the role endpoints on rg.
// rg is expected to carry Auth() and Authorize(); see api/v1/routes.go.
func RegisterRoutes(rg *gin.RouterGroup) {
	controller := Controller{}
	group := rg.Group("/roles")
	{
		group.GET("", controller.Index)
		group.POST("", controller.Create)
		group.GET("/:uuid", controller.Show)
		group.PUT("/:uuid", controller.Update)
		group.DELETE("/:uuid", controller.Delete)
		group.GET("/:uuid/authorities", controller.Authorities)
		group.PUT("/:uuid/authorities", controller.UpdateAuthorities)
	}
}
