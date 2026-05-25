package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/Qaid-Danial/password-manager/backend/utils"
)

// AuthRequired validates the JWT Bearer token in the Authorization header.
// On success it sets "userID" and "userRole" in the Gin context.
// On failure it aborts with 401 and a generic message — no detail about
// whether the token is expired, malformed, or signed with the wrong key.
func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := utils.ValidateToken(tokenStr, jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Next()
	}
}

// GetUserID extracts the authenticated user's ID from the Gin context.
// Must only be called from handlers protected by AuthRequired.
func GetUserID(c *gin.Context) string {
	return c.GetString("userID")
}

// GetUserRole extracts the authenticated user's role from the Gin context.
func GetUserRole(c *gin.Context) string {
	return c.GetString("userRole")
}
