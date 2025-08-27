package interfaces

import (
	"MTConnect/internal/domain/entities"
	"MTConnect/internal/domain/models"
	"time"
)

// Usecases - это агрегирующий интерфейс для всех use cases
type Usecases interface {
	CreateConnection(req models.ConnectionRequest) (*models.ConnectionInfo, error)
	RestoreConnection(machine entities.CncMachine) (*models.ConnectionInfo, error) // Новый метод
	GetAllConnections() []*models.ConnectionInfo
	DeleteConnection(sessionID string) error
	CheckConnection(sessionID string) (*models.ConnectionInfo, error)
	StartPolling(sessionID string, interval time.Duration) error
	StopPolling(sessionID string) error
}
