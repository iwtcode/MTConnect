package interfaces

import "github.com/iwtcode/MTConnect/internal/domain/models"

// ConnectionService определяет контракт для управления пулом подключений.
type ConnectionService interface {
	CreateConnection(req models.ConnectionRequest) (*models.ConnectionInfo, error)
	GetConnection(sessionID string) (*models.ConnectionInfo, bool)
	GetAllConnections() []*models.ConnectionInfo
	DeleteConnection(sessionID string) error
	CheckConnection(sessionID string) (*models.ConnectionInfo, error)
}
