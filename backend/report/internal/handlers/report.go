package handlers

import (
	"net/http"
	"report/internal/services"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	reportService *services.ReportService
}

func NewReportHandler(reportService *services.ReportService) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
	}
}

func (h *ReportHandler) GetReports(c *gin.Context) {
	userID, valid := h.getUserID(c)
	if !valid {
		return
	}

	report, err := h.reportService.GetUserReport(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get report"})
		return
	}

	c.JSON(http.StatusOK, report)
}

func (h *ReportHandler) GenerateReports(c *gin.Context) {
	userID, valid := h.getUserID(c)
	if !valid {
		return
	}

	url, err := h.reportService.GenerateUserReport(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate report"})
		return
	}

	if url == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Данные для отчета пока не найдены"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (h *ReportHandler) getUserID(c *gin.Context) (int, bool) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found"})
		return 0, false
	}

	userID, ok := userIDStr.(int)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return 0, false
	}

	return userID, true
}
