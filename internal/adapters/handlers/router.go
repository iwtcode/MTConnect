package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ProvideRouter настраивает и возвращает HTTP-роутер
func ProvideRouter(h *Handler) http.Handler {
	router := gin.Default()
	api := router.Group("/api")
	{
		api.GET("/:machineId/current", h.GetCurrentData)
	}
	return router
}
