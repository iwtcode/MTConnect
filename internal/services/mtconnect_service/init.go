package mtconnect_service

import (
	"MTConnect/internal/domain/models"
	"MTConnect/internal/interfaces"
	"MTConnect/internal/services/mtconnect_service/connector"
	"MTConnect/internal/services/mtconnect_service/poller"
	"time"
)

// mtconnectService - это конкретная реализация интерфейса MTConnectService.
type mtconnectService struct {
	connMgr *connector.ConnectionManager
	pollMgr *poller.PollingManager
}

// NewMTConnectService создает новый экземпляр сервиса.
func NewMTConnectService(repo interfaces.Repository, producer interfaces.KafkaService) interfaces.MTConnectService {
	// Создаем PollingManager
	pollingManager := poller.NewPollingManager(repo, producer)
	// Создаем ConnectionManager, передавая ему PollingManager
	connectionManager := connector.NewConnectionManager(pollingManager, repo)

	return &mtconnectService{
		connMgr: connectionManager,
		pollMgr: pollingManager,
	}
}

// --- Реализация методов интерфейса MTConnectService ---

func (s *mtconnectService) CreateConnection(req models.ConnectionRequest) (*models.ConnectionInfo, error) {
	return s.connMgr.CreateConnection(req)
}

func (s *mtconnectService) GetConnection(sessionID string) (*models.ConnectionInfo, bool) {
	return s.connMgr.GetConnection(sessionID)
}

func (s *mtconnectService) GetAllConnections() []*models.ConnectionInfo {
	return s.connMgr.GetAllConnections()
}

func (s *mtconnectService) DeleteConnection(sessionID string) error {
	return s.connMgr.DeleteConnection(sessionID)
}

func (s *mtconnectService) CheckConnection(sessionID string) (*models.ConnectionInfo, error) {
	return s.connMgr.CheckConnection(sessionID)
}

func (s *mtconnectService) StartPolling(interval time.Duration) error {
	// Получаем активные подключения и запускаем опрос
	connections := s.connMgr.GetAllConnections()
	return s.pollMgr.StartAllPolling(connections, interval)
}

func (s *mtconnectService) StopPolling() error {
	s.pollMgr.StopAllPolling()
	return nil
}

func (s *mtconnectService) IsPollingActive() bool {
	return s.pollMgr.IsPollingActive()
}
