package interfaces

import (
	"MTConnect/internal/domain/models"
	"time"
)

// MTConnectService - это агрегирующий интерфейс для всей бизнес-логики.
type MTConnectService interface {
	ConnectionManager
	PollingManager
}

// ConnectionManager определяет контракт для управления пулом подключений.
type ConnectionManager interface {
	CreateConnection(req models.ConnectionRequest) (*models.ConnectionInfo, error)
	GetConnection(sessionID string) (*models.ConnectionInfo, bool)
	GetAllConnections() []*models.ConnectionInfo
	DeleteConnection(sessionID string) error
	CheckConnection(sessionID string) (*models.ConnectionInfo, error)
}

// PollingManager определяет контракт для сервиса, опрашивающего эндпоинты.
type PollingManager interface {
	StartPolling(interval time.Duration) error
	StopPolling() error
	IsPollingActive() bool
}
