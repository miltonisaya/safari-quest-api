package review

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts the review endpoints on rg.
// rg is expected to carry Auth() and Authorize(); see api/v1/routes.go.
//
// Each path's last segment yields its own authority code:
// package-review_get (read a package's reviews), my-review_* (travelers),
// review_get and review-reply_put (company staff), and
// review-moderation_put (administrators only).
func RegisterRoutes(rg *gin.RouterGroup) {
	controller := Controller{}

	rg.GET("/package-reviews", controller.PackageReviews)

	mine := rg.Group("/my-reviews")
	{
		mine.GET("", controller.Mine)
		mine.POST("", controller.Create)
		mine.PUT("/:uuid", controller.Update)
		mine.DELETE("/:uuid", controller.Delete)
	}

	rg.GET("/reviews", controller.Index)
	rg.PUT("/review-replies/:uuid", controller.Reply)
	rg.PUT("/review-moderation/:uuid", controller.Moderate)
}
