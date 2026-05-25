package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck pings the database and reports overall service health.
// Used by Docker's healthcheck and any uptime monitoring.
func HealthCheck(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "degraded",
				"database": "disconnected",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"database": "connected",
		})
	}
}
