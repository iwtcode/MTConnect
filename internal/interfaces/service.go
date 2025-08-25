package interfaces

import (
	"MTConnect/internal/domain/entities"
	"time"
)

// PollingService определяет контракт для сервиса, опрашивающего эндпоинты
type PollingService interface {
	StartPollingForMachine(conn *entities.ConnectionInfo, interval time.Duration) error
	StopPollingForMachine(sessionID string) error
	CheckMachineConnection(endpointURL string) error
	StartAllPolling(connections []*entities.ConnectionInfo, interval time.Duration) error
	StopAllPolling()
	LoadMetadataForEndpoint(endpointURL string) error
	// Новый метод для запуска опроса для нового подключения, если опрос уже активен
	StartPollingForNewConnectionIfNeeded(conn *entities.ConnectionInfo) error
}
