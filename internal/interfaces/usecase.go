package interfaces

import (
	"MTConnect/internal/domain/entities"
	"time"
)

// Usecases - это агрегирующий интерфейс для всех use cases
type Usecases interface {
	ConnectionUsecase
}

// ConnectionUsecase определяет контракт для логики управления подключениями
type ConnectionUsecase interface {
	CreateConnection(req entities.ConnectionRequest) (*entities.ConnectionInfo, error)
	GetAllConnections() []*entities.ConnectionInfo
	DeleteConnection(sessionID string) error
	CheckConnection(sessionID string) (*entities.ConnectionInfo, error)
	StartPolling(interval time.Duration) error
	StopPolling() error
}
