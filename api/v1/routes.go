package v1

import (
	"net/http"
	attractionType "safari-quest-api/internal/attractiontype"

	"safari-quest-api/internal/account"
	"safari-quest-api/internal/agegroup"
	"safari-quest-api/internal/attraction"
	"safari-quest-api/internal/attractioncategory"
	"safari-quest-api/internal/auth"
	"safari-quest-api/internal/authority"
	"safari-quest-api/internal/booking"
	"safari-quest-api/internal/catalog"
	"safari-quest-api/internal/fileresource"
	"safari-quest-api/internal/fileresourcetype"
	"safari-quest-api/internal/geography"
	"safari-quest-api/internal/geographytype"
	"safari-quest-api/internal/menu"
	"safari-quest-api/internal/review"
	"safari-quest-api/internal/role"
	"safari-quest-api/internal/setting"
	"safari-quest-api/internal/touroperator"
	"safari-quest-api/internal/tourpackage"
	"safari-quest-api/internal/user"
	"safari-quest-api/middlewares"
	"safari-quest-api/pkg/response"
	"safari-quest-api/websocket"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes decides which middleware each feature sits behind; the
// features themselves declare their own paths in their <struct>_routes.go files.
func RegisterRoutes(rg *gin.RouterGroup) {
	// Public routes — no authentication required.
	rg.GET("/health", healthCheck)
	rg.GET("/uploads/:name", fileresource.Serve)
	auth.RegisterRoutes(rg)
	account.RegisterPublicRoutes(rg)  // forgot / reset password
	agegroup.RegisterPublicRoutes(rg) // age groups for booking forms
	catalog.RegisterRoutes(rg)        // public package catalog for the website

	// Signed in, but not behind Authorize(): every account uses these.
	signedIn := rg.Group("", middlewares.Auth())
	websocket.RegisterRoutes(signedIn)
	account.RegisterRoutes(signedIn) // own profile and password

	// Protected routes — Auth() validates the JWT and Authorize() automatically
	// derives the required authority code from the route pattern and HTTP method,
	// then checks the user holds it via their roles. No per-route configuration needed.
	protected := rg.Group("", middlewares.Auth(), middlewares.Authorize())
	role.RegisterRoutes(protected)
	user.RegisterRoutes(protected)
	menu.RegisterRoutes(protected)
	geographytype.RegisterRoutes(protected)
	geography.RegisterRoutes(protected)
	authority.RegisterRoutes(protected)
	attractionType.RegisterRoutes(protected)
	attractioncategory.RegisterRoutes(protected)
	touroperator.RegisterRoutes(protected)
	attraction.RegisterRoutes(protected)
	fileresourcetype.RegisterRoutes(protected)
	fileresource.RegisterRoutes(protected)
	tourpackage.RegisterRoutes(protected)
	booking.RegisterRoutes(protected)
	review.RegisterRoutes(protected)
	setting.RegisterRoutes(protected)
	agegroup.RegisterRoutes(protected)
}

func healthCheck(c *gin.Context) {
	response.Success(c, http.StatusOK, "SafariQuest API is running", nil)
}
