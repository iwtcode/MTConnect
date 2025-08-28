package handlers

import (
	"MTConnect/internal/config"
	"MTConnect/internal/interfaces"
	"MTConnect/internal/middleware/logging"
	"MTConnect/internal/middleware/swagger"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler - структура для обработчиков HTTP-запросов
type Handler struct {
	usecase interfaces.Usecases
	logger  *logging.Logger
}

// NewHandler создает новый экземпляр Handler
func NewHandler(usecase interfaces.Usecases, logger *logging.Logger) *Handler {
	return &Handler{
		usecase: usecase,
		logger:  logger.WithPrefix("HANDLER"),
	}
}

// ProvideRouter настраивает и возвращает HTTP-роутер
func ProvideRouter(h *Handler, cfg *config.AppConfig, swagCfg *swagger.Config) http.Handler {
	gin.SetMode(cfg.GinMode)

	router := gin.Default()

	// Swagger
	swagger.Setup(router, swagCfg)

	// Logger Middleware
	router.Use(LoggingMiddleware(h.logger))

	// Группа API v1
	v1 := router.Group("/api/v1")
	{
		connections := v1.Group("/connect")
		{
			connections.POST("", h.CreateConnection)
			connections.GET("", h.GetConnections)
			connections.DELETE("", h.DeleteConnection)
			connections.POST("/check", h.CheckConnection)
		}

		polling := v1.Group("/polling")
		{
			polling.POST("/start", h.StartPolling)
			polling.POST("/stop", h.StopPolling)
		}
	}

	return router
}
