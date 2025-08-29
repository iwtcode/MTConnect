package handlers

import (
	"MTConnect/internal/domain/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// StartPolling запускает опрос данных для указанного подключения.
// @Summary Запустить опрос
// @Description Запускает периодический сбор данных для подключения по SessionID с заданным интервалом.
// @Tags Polling
// @Accept json
// @Produce json
// @Param input body models.PollingRequest true "Параметры для запуска опроса"
// @Success 200 {object} models.MessageResponse "Сообщение об успешном запуске"
// @Failure 400 {object} models.ErrorResponse "Неверный формат запроса"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /polling/start [post]
func (h *Handler) StartPolling(c *gin.Context) {
	var req models.PollingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, err, "Invalid request payload")
		return
	}

	duration := time.Duration(req.Interval) * time.Millisecond
	h.logger.Info("Attempting to start polling", "sessionID", req.SessionID, "interval", duration)

	if err := h.usecase.StartPolling(req.SessionID, duration); err != nil {
		h.InternalError(c, err)
		return
	}

	h.logger.Info("Polling started successfully", "sessionID", req.SessionID)
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": fmt.Sprintf("Polling started for session %s", req.SessionID),
	})
}

// StopPolling останавливает опрос данных для указанного подключения.
// @Summary Остановить опрос
// @Description Останавливает сбор данных для подключения по SessionID.
// @Tags Polling
// @Accept json
// @Produce json
// @Param input body models.SessionRequest true "ID сессии для остановки опроса"
// @Success 200 {object} models.MessageResponse "Сообщение об успешной остановке"
// @Failure 400 {object} models.ErrorResponse "Неверный формат запроса"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /polling/stop [post]
func (h *Handler) StopPolling(c *gin.Context) {
	var req models.SessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, err, "Missing or invalid SessionID")
		return
	}

	h.logger.Info("Attempting to stop polling", "sessionID", req.SessionID)

	if err := h.usecase.StopPolling(req.SessionID); err != nil {
		h.InternalError(c, err)
		return
	}

	h.logger.Info("Polling stopped successfully", "sessionID", req.SessionID)
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": fmt.Sprintf("Polling stopped for session %s", req.SessionID),
	})
}
