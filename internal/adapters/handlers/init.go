package handlers

import (
	"MTConnect/internal/config"
	"MTConnect/internal/interfaces"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler - структура для обработчиков HTTP-запросов
type Handler struct {
	usecase interfaces.Usecases
}

// NewHandler создает новый экземпляр Handler
func NewHandler(usecase interfaces.Usecases) *Handler {
	return &Handler{usecase: usecase}
}

// ProvideRouter настраивает и возвращает HTTP-роутер
func ProvideRouter(h *Handler, cfg *config.AppConfig) http.Handler {
	gin.SetMode(cfg.GinMode)

	router := gin.Default()

	// Группа API v1
	v1 := router.Group("/api/v1")
	{
		// Управление подключениями
		v1.POST("/connect", h.CreateConnection)
		v1.GET("/connect", h.GetConnections)
		v1.DELETE("/connect", h.DeleteConnection)
		v1.POST("/connect/check", h.CheckConnection)

		// Управление опросом
		v1.POST("/polling/start", h.StartPolling) // Изменено на POST
		v1.POST("/polling/stop", h.StopPolling)   // Изменено на POST
	}

	return router
}
