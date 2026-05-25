package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Qaid-Danial/password-manager/backend/models"
	"github.com/Qaid-Danial/password-manager/backend/services"
)

type AuthHandler struct {
	authSvc *services.AuthService
}

func NewAuthHandler(authSvc *services.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	token, user, err := h.authSvc.Register(req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		if err.Error() == "username or email already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "registration failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"token": token, "user": user})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	token, user, err := h.authSvc.Login(req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		// Always return the same message regardless of failure reason
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}
