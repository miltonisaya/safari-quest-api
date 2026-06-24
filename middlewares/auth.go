package middlewares

import (
	"net/http"
	"strings"

	"safari-quest-api/config"
	"safari-quest-api/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Auth validates a Bearer JWT in the Authorization header.
// On success it stores two values on the Gin context:
//   - "claims"   → jwt.MapClaims  (full token payload)
//   - "userUUID" → uuid.UUID      (the "sub" claim, used by Authorize middleware)
//
// On failure it aborts with 401 so the handler is never reached.
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Fail(c, http.StatusUnauthorized, "Authorization header missing or malformed", nil)
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(config.App.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			response.Fail(c, http.StatusUnauthorized, "Invalid or expired token", nil)
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Fail(c, http.StatusUnauthorized, "Invalid token claims", nil)
			c.Abort()
			return
		}

		// The "sub" claim holds the user UUID. Parse it once here so every
		// downstream middleware and handler can use it without re-parsing.
		sub, ok := claims["sub"].(string)
		if !ok {
			response.Fail(c, http.StatusUnauthorized, "Token is missing subject claim", nil)
			c.Abort()
			return
		}

		userUUID, err := uuid.Parse(sub)
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, "Token subject is not a valid UUID", nil)
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Set("userUUID", userUUID)
		c.Next()
	}
}
