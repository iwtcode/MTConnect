package handlers

import (
	"net/http"

	"MTConnect/internal/interfaces"

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
