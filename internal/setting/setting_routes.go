package setting

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts the system settings; rg is the authorised group, so
// only roles holding system-setting_* (administrators) reach them.
func RegisterRoutes(rg *gin.RouterGroup) {
	controller := Controller{}
	rg.GET("/system-settings", controller.Index)
	rg.PUT("/system-settings", controller.Update)
}
