package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminRequired enforces role-based access control.
// Must be chained AFTER AuthRequired so that userRole is already set.
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if GetUserRole(c) != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}
