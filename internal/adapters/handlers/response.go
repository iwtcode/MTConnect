package handlers

import (
	"net/http"

	"github.com/iwtcode/MTConnect/pkg/errors"

	"github.com/gin-gonic/gin"
)

// ErrorResponse возвращает стандартизированный ответ с ошибкой
func (h *Handler) ErrorResponse(c *gin.Context, err error, statusCode int, message string, showError bool) {
	errorMessage := message
	if showError && err != nil {
		errorMessage = message + ": " + err.Error()
	}

	h.logger.Error(message, "error", err, "statusCode", statusCode)
	c.AbortWithStatusJSON(statusCode, gin.H{
		"status": "error",
		"error": gin.H{
			"code":    statusCode,
			"message": errorMessage,
		},
	})
}

// BadRequest возвращает ошибку 400
func (h *Handler) BadRequest(c *gin.Context, err error, message string) {
	if message == "" {
		message = errors.BadRequest
	}
	h.ErrorResponse(c, err, http.StatusBadRequest, message, true)
}

// InternalError возвращает ошибку 500
func (h *Handler) InternalError(c *gin.Context, err error) {
	h.ErrorResponse(c, err, http.StatusInternalServerError, errors.InternalServerError, false)
}

// NotFound возвращает ошибку 404
func (h *Handler) NotFound(c *gin.Context, err error) {
	h.ErrorResponse(c, err, http.StatusNotFound, errors.NotFound, true)
}
