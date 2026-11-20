package touroperator

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts the tour operator endpoints on rg.
// rg is expected to carry Auth() and Authorize(); see api/v1/routes.go.
// The /staff routes (users linked through tour_operator_users) additionally
// require the caller to be staff of that operator or a platform administrator,
// checked in the service. They use their own "staff" path segment so their
// derived authority codes (staff_get, staff_post, staff_delete) are separate
// from the platform-wide user_* codes.
func RegisterRoutes(rg *gin.RouterGroup) {
	controller := Controller{}
	group := rg.Group("/tour-operators")
	{
		group.GET("", controller.Index)
		group.POST("", controller.Create)
		group.GET("/:uuid", controller.Show)
		group.PUT("/:uuid", controller.Update)
		group.DELETE("/:uuid", controller.Delete)
		group.GET("/:uuid/staff", controller.Users)
		group.POST("/:uuid/staff", controller.CreateUser)
		group.DELETE("/:uuid/staff/:user_uuid", controller.RemoveUser)
	}
}
