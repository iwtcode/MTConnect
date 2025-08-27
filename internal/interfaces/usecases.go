package interfaces

import (
	"MTConnect/internal/domain/models"
	"time"
)

// Usecases - это агрегирующий интерфейс для всех use cases
type Usecases interface {
	CreateConnection(req models.ConnectionRequest) (*models.ConnectionInfo, error)
	GetAllConnections() []*models.ConnectionInfo
	DeleteConnection(sessionID string) error
	CheckConnection(sessionID string) (*models.ConnectionInfo, error)
	StartPolling(interval time.Duration) error
	StopPolling() error
}
