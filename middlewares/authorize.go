package middlewares

import (
	"net/http"

	"safari-quest-api/pkg/authority"
	"safari-quest-api/pkg/response"
	"safari-quest-api/repositories"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Authorize derives the required authority code from the matched route pattern
// and HTTP method, then checks the authenticated user holds it via their roles.
// It reads "userUUID" from the context set by Auth() so it must run after Auth().
func Authorize() gin.HandlerFunc {
	return func(c *gin.Context) {
		userUUID, exists := c.Get("userUUID")
		if !exists {
			response.Fail(c, http.StatusUnauthorized, "Unauthorized", nil)
			c.Abort()
			return
		}

		authorityCode := authority.DeriveCode(c.Request.Method, c.FullPath())
		if authorityCode == "" {
			response.Fail(c, http.StatusForbidden, "Cannot determine required authority for this route", nil)
			c.Abort()
			return
		}

		hasAuthority, err := repositories.UserHasAuthority(userUUID.(uuid.UUID), authorityCode)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Authorization check failed")
			c.Abort()
			return
		}

		if !hasAuthority {
			response.Fail(c, http.StatusForbidden, "You do not have permission to perform this action", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
