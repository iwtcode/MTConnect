package interfaces

import "MTConnect/internal/domain/entities"

// ConnectionService определяет контракт для управления пулом подключений.
type ConnectionService interface {
	CreateConnection(req entities.ConnectionRequest) (*entities.ConnectionInfo, error)
	GetConnection(sessionID string) (*entities.ConnectionInfo, bool)
	GetAllConnections() []*entities.ConnectionInfo
	DeleteConnection(sessionID string) error
	CheckConnection(sessionID string) (*entities.ConnectionInfo, error)
}
