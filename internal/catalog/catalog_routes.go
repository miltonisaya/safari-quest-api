package catalog

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts the public catalog on rg. rg must NOT require a
// token: these routes feed the public website. The seeder skips /catalog/
// paths, so they create no authority codes.
func RegisterRoutes(rg *gin.RouterGroup) {
	controller := Controller{}
	group := rg.Group("/catalog/packages")
	{
		group.GET("", controller.Index)
		group.GET("/:uuid", controller.Show)
		group.GET("/:uuid/reviews", controller.Reviews)
	}
}
