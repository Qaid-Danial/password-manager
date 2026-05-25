package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Qaid-Danial/password-manager/backend/middleware"
	"github.com/Qaid-Danial/password-manager/backend/models"
	"github.com/Qaid-Danial/password-manager/backend/services"
)

type VaultHandler struct {
	vaultSvc *services.VaultService
}

func NewVaultHandler(vaultSvc *services.VaultService) *VaultHandler {
	return &VaultHandler{vaultSvc: vaultSvc}
}

func (h *VaultHandler) List(c *gin.Context) {
	entries, err := h.vaultSvc.List(middleware.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch vault"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"entries": entries})
}

func (h *VaultHandler) Create(c *gin.Context) {
	var req models.VaultEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	entry, err := h.vaultSvc.Create(middleware.GetUserID(c), req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create entry"})
		return
	}
	c.JSON(http.StatusCreated, entry)
}

func (h *VaultHandler) GetByID(c *gin.Context) {
	entry, err := h.vaultSvc.GetByID(c.Param("id"), middleware.GetUserID(c), c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch entry"})
		return
	}
	// Return 404 whether the entry doesn't exist OR belongs to another user —
	// returning 403 would confirm the entry exists and enable ID enumeration.
	if entry == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
		return
	}
	c.JSON(http.StatusOK, entry)
}

func (h *VaultHandler) Update(c *gin.Context) {
	var req models.VaultEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	entry, err := h.vaultSvc.Update(c.Param("id"), middleware.GetUserID(c), req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update entry"})
		return
	}
	if entry == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
		return
	}
	c.JSON(http.StatusOK, entry)
}

func (h *VaultHandler) Delete(c *gin.Context) {
	found, err := h.vaultSvc.Delete(c.Param("id"), middleware.GetUserID(c), c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete entry"})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
		return
	}
	c.Status(http.StatusNoContent)
}
