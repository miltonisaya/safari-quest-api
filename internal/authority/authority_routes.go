package authority

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts the authority endpoints on rg.
// rg is expected to carry Auth() and Authorize(); see api/v1/routes.go.
func RegisterRoutes(rg *gin.RouterGroup) {
	controller := Controller{}
	rg.GET("/authorities", controller.Index)
}
