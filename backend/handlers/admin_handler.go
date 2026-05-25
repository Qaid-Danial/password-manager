package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Qaid-Danial/password-manager/backend/services"
)

type AdminHandler struct {
	adminSvc *services.AdminService
}

func NewAdminHandler(adminSvc *services.AdminService) *AdminHandler {
	return &AdminHandler{adminSvc: adminSvc}
}

func (h *AdminHandler) GetAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	// Cap page_size to prevent resource exhaustion
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := h.adminSvc.GetAuditLogs(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
