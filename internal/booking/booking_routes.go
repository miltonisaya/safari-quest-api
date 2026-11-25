package booking

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts the booking endpoints on rg.
// rg is expected to carry Auth() and Authorize(); see api/v1/routes.go.
//
// Each path has its own last segment so its derived authority codes stay
// separate: company staff use booking_* and traveler_get; travelers use
// bookable-package_get and my-booking_* (DELETE /my-bookings/:uuid cancels,
// it does not delete).
func RegisterRoutes(rg *gin.RouterGroup) {
	controller := Controller{}

	staff := rg.Group("/bookings")
	{
		staff.GET("", controller.Index)
		staff.POST("", controller.Create)
		staff.GET("/:uuid", controller.Show)
		staff.PUT("/:uuid", controller.Update)
		staff.DELETE("/:uuid", controller.Delete)
	}
	rg.GET("/travelers", controller.Travelers)

	rg.GET("/bookable-packages", controller.Bookable)
	rg.GET("/bookable-packages/:uuid", controller.BookableDetail)
	mine := rg.Group("/my-bookings")
	{
		mine.GET("", controller.Mine)
		mine.POST("", controller.Book)
		mine.DELETE("/:uuid", controller.Cancel)
	}
}
