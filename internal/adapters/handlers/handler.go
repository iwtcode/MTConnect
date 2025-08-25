package handlers

import (
	"MTConnect/internal/domain/entities"
	"MTConnect/internal/interfaces"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	usecase interfaces.Usecases
}

func NewHandler(usecase interfaces.Usecases) *Handler {
	return &Handler{usecase: usecase}
}

// --- V1 API Управления Подключениями ---

func (h *Handler) CreateConnection(c *gin.Context) {
	var req entities.ConnectionRequest
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

func (h *Handler) GetConnections(c *gin.Context) {
	connections := h.usecase.GetAllConnections()
	c.JSON(http.StatusOK, gin.H{
		"Status":      "ok",
		"PoolSize":    len(connections),
		"Connections": connections,
	})
}

func (h *Handler) DeleteConnection(c *gin.Context) {
	var req entities.SessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Status": "error", "Message": err.Error()})
		return
	}

	if err := h.usecase.DeleteConnection(req.SessionID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Status": "error", "Message": err.Error()})
		return
	}

	// ИЗМЕНЕНИЕ: Добавлен "Status": "ok"
	c.JSON(http.StatusOK, gin.H{
		"Status":  "ok",
		"Message": "Session " + req.SessionID + " disconnected successfully",
	})
}

func (h *Handler) CheckConnection(c *gin.Context) {
	var req entities.SessionRequest
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

// --- V1 API Управления Опросом ---

func (h *Handler) StartPolling(c *gin.Context) {
	intervalStr := c.DefaultQuery("interval", "1000") // Интервал по умолчанию 1000мс
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		// ИЗМЕНЕНИЕ: Приведено к единому формату "Status", "Message"
		c.JSON(http.StatusBadRequest, gin.H{"Status": "error", "Message": "неверный параметр 'interval', ожидается целое число (миллисекунды)"})
		return
	}
	duration := time.Duration(interval) * time.Millisecond

	if err := h.usecase.StartPolling(duration); err != nil {
		// ИЗМЕНЕНИЕ: Приведено к единому формату "Status", "Message"
		c.JSON(http.StatusInternalServerError, gin.H{"Status": "error", "Message": err.Error()})
		return
	}

	// ИЗМЕНЕНИЕ: "status" -> "Status"
	c.JSON(http.StatusOK, gin.H{"Status": "monitoring started"})
}

func (h *Handler) StopPolling(c *gin.Context) {
	_ = h.usecase.StopPolling()
	// ИЗМЕНЕНИЕ: "status" -> "Status"
	c.JSON(http.StatusOK, gin.H{"Status": "monitoring stopped"})
}
