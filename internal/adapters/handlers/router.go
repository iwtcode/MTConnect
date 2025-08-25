package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ProvideRouter настраивает и возвращает HTTP-роутер
func ProvideRouter(h *Handler) http.Handler {
	router := gin.Default()

	// Новая группа API v1
	v1 := router.Group("/api/v1")
	{
		// Управление подключениями
		v1.POST("/connect", h.CreateConnection)
		v1.GET("/connect", h.GetConnections)
		v1.DELETE("/connect", h.DeleteConnection)
		v1.POST("/connect/check", h.CheckConnection)

		// Управление опросом
		v1.GET("/polling/start", h.StartPolling)
		v1.GET("/polling/stop", h.StopPolling)
	}

	return router
}
