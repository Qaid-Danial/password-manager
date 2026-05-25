package routes

import (
	"database/sql"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/Qaid-Danial/password-manager/backend/config"
	"github.com/Qaid-Danial/password-manager/backend/handlers"
	"github.com/Qaid-Danial/password-manager/backend/middleware"
	"github.com/Qaid-Danial/password-manager/backend/services"
)

// Setup wires all middleware, services, and handlers into the Gin router.
// Middleware order matters: security headers → CORS → per-route auth.
func Setup(db *sql.DB, cfg *config.Config) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.Use(middleware.SecurityHeaders())

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:80", "http://localhost"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// Services
	auditSvc := services.NewAuditService(db)
	authSvc := services.NewAuthService(db, cfg, auditSvc)
	vaultSvc := services.NewVaultService(db, cfg, auditSvc)
	adminSvc := services.NewAdminService(db)

	// Handlers
	authHandler := handlers.NewAuthHandler(authSvc)
	vaultHandler := handlers.NewVaultHandler(vaultSvc)
	adminHandler := handlers.NewAdminHandler(adminSvc)

	api := router.Group("/api")
	{
		api.GET("/health", handlers.HealthCheck(db))

		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			// Rate limit applied only to login to mitigate brute-force attacks
			auth.POST("/login",
				middleware.RateLimitMiddleware(cfg.RateLimitRequests, cfg.RateLimitWindow),
				authHandler.Login,
			)
		}

		// All vault routes require a valid JWT
		vault := api.Group("/vault")
		vault.Use(middleware.AuthRequired(cfg.JWTSecret))
		{
			vault.GET("", vaultHandler.List)
			vault.POST("", vaultHandler.Create)
			vault.GET("/:id", vaultHandler.GetByID)
			vault.PUT("/:id", vaultHandler.Update)
			vault.DELETE("/:id", vaultHandler.Delete)
		}

		// Admin routes require JWT + admin role
		admin := api.Group("/admin")
		admin.Use(middleware.AuthRequired(cfg.JWTSecret))
		admin.Use(middleware.AdminRequired())
		{
			admin.GET("/audit-logs", adminHandler.GetAuditLogs)
		}
	}

	return router
}
