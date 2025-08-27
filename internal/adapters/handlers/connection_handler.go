package handlers

import (
	"MTConnect/internal/domain/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateConnection создает новое подключение
func (h *Handler) CreateConnection(c *gin.Context) {
	var req models.ConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Status": "error", "Message": err.Error()})
		return
	}

	connInfo, err := h.usecase.CreateConnection(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Status": "error", "Message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Status": "ok", "connectionInfo": connInfo})
}

// GetConnections возвращает список активных подключений
func (h *Handler) GetConnections(c *gin.Context) {
	connections := h.usecase.GetAllConnections()
	c.JSON(http.StatusOK, gin.H{
		"Status":      "ok",
		"PoolSize":    len(connections),
		"Connections": connections,
	})
}

// DeleteConnection удаляет подключение
func (h *Handler) DeleteConnection(c *gin.Context) {
	var req models.SessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Status": "error", "Message": err.Error()})
		return
	}

	if err := h.usecase.DeleteConnection(req.SessionID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Status": "error", "Message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "ok",
		"Message": "Session " + req.SessionID + " disconnected successfully",
	})
}

// CheckConnection проверяет состояние подключения
func (h *Handler) CheckConnection(c *gin.Context) {
	var req models.SessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Status": "error", "Message": err.Error()})
		return
	}

	connInfo, err := h.usecase.CheckConnection(req.SessionID)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"Status": "unhealthy", "Message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Status": "healthy", "connectionInfo": connInfo})
}
