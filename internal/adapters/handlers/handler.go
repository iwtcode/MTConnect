package handlers

import (
	"MTConnect/internal/interfaces"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	usecase interfaces.MachineUsecase
}

func NewHandler(usecase interfaces.Usecases) *Handler {
	return &Handler{usecase: usecase}
}

// GetCurrentData обрабатывает запрос на получение текущих данных станка
func (h *Handler) GetCurrentData(c *gin.Context) {
	machineId := c.Param("machineId")
	machineData, err := h.usecase.GetMachineData(machineId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, machineData)
}

// CheckConnection обрабатывает запрос на проверку доступности станка
func (h *Handler) CheckConnection(c *gin.Context) {
	machineId := c.Param("machineId")
	if err := h.usecase.CheckMachineConnection(machineId); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "станок " + machineId + " доступен"})
}

// StartPolling обрабатывает запрос на запуск опроса
func (h *Handler) StartPolling(c *gin.Context) {
	machineId := c.Param("machineId")
	intervalStr := c.DefaultQuery("interval", "1000") // Интервал в миллисекундах
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный параметр 'interval', ожидается целое число (миллисекунды)"})
		return
	}

	duration := time.Duration(interval) * time.Millisecond
	if err := h.usecase.StartPolling(machineId, duration); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "опрос для станка " + machineId + " запущен"})
}

// StopPolling обрабатывает запрос на остановку опроса
func (h *Handler) StopPolling(c *gin.Context) {
	machineId := c.Param("machineId")
	if err := h.usecase.StopPolling(machineId); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "опрос для станка " + machineId + " остановлен"})
}

// GetControlProgram (заглушка)
func (h *Handler) GetControlProgram(c *gin.Context) {
	machineId := c.Param("machineId")
	c.JSON(http.StatusNotImplemented, gin.H{
		"machineId": machineId,
		"message":   "функционал получения управляющей программы пока не реализован",
		"program":   "G0 X0 Y0...",
	})
}
