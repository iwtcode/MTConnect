package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ProvideRouter настраивает и возвращает HTTP-роутер
func ProvideRouter(h *Handler) http.Handler {
	router := gin.Default()
	api := router.Group("/api/:machineId")
	{
		// Существующий эндпоинт
		api.GET("/current", h.GetCurrentData)

		// 1.2.1 Проверка соединения
		api.GET("/check", h.CheckConnection)

		// 1.2.2 Запуск опроса
		api.POST("/polling/start", h.StartPolling) // Используем POST для изменения состояния

		// 1.2.3 Остановка опроса
		api.POST("/polling/stop", h.StopPolling) // Используем POST для изменения состояния

		// 1.2.4 Получение УП (заглушка)
		api.GET("/program", h.GetControlProgram)
	}
	return router
}
