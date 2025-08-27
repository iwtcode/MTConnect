package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// StartPolling запускает опрос всех активных подключений
func (h *Handler) StartPolling(c *gin.Context) {
	intervalStr := c.DefaultQuery("interval", "1000") // Интервал по умолчанию 1000мс
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Status": "error", "Message": "неверный параметр 'interval', ожидается целое число (миллисекунды)"})
		return
	}
	duration := time.Duration(interval) * time.Millisecond

	if err := h.usecase.StartPolling(duration); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Status": "error", "Message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Status": "monitoring started"})
}

// StopPolling останавливает опрос
func (h *Handler) StopPolling(c *gin.Context) {
	_ = h.usecase.StopPolling()
	c.JSON(http.StatusOK, gin.H{"Status": "monitoring stopped"})
}
