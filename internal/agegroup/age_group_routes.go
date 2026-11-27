package agegroup

import "github.com/gin-gonic/gin"

// RegisterPublicRoutes lists the age groups without sign-in (booking forms need them).
func RegisterPublicRoutes(rg *gin.RouterGroup) {
	rg.GET("/catalog/age-groups", Controller{}.Index)
}

// RegisterRoutes mounts the administrators' editor (age-group_put).
func RegisterRoutes(rg *gin.RouterGroup) {
	rg.PUT("/age-groups", Controller{}.Replace)
}
