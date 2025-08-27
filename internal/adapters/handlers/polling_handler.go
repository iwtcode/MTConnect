package handlers

import (
	"MTConnect/internal/domain/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// StartPolling запускает опрос для конкретного подключения
func (h *Handler) StartPolling(c *gin.Context) {
	var req models.PollingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Status": "error", "Message": err.Error()})
		return
	}

	duration := time.Duration(req.Interval) * time.Millisecond

	if err := h.usecase.StartPolling(req.SessionID, duration); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Status": "error", "Message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "ok",
		"Message": fmt.Sprintf("Polling started for session %s", req.SessionID),
	})
}

// StopPolling останавливает опрос для конкретного подключения
func (h *Handler) StopPolling(c *gin.Context) {
	var req models.SessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Status": "error", "Message": err.Error()})
		return
	}

	if err := h.usecase.StopPolling(req.SessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Status": "error", "Message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "ok",
		"Message": fmt.Sprintf("Polling stopped for session %s", req.SessionID),
	})
}
