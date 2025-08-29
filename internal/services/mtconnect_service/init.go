package mtconnect_service

import (
	"MTConnect/internal/domain/entities"
	"MTConnect/internal/domain/models"
	"MTConnect/internal/interfaces"
	"MTConnect/internal/middleware/logging"
	"MTConnect/internal/services/mtconnect_service/connector"
	"MTConnect/internal/services/mtconnect_service/poller"
	"time"
)

type mtconnectService struct {
	connMgr *connector.ConnectionManager
	pollMgr *poller.PollingManager
}

func NewMTConnectService(repo interfaces.CncMachineRepository, producer interfaces.KafkaService, logger *logging.Logger) interfaces.MTConnectService {
	pollingManager := poller.NewPollingManager(repo, producer, logger)
	connectionManager := connector.NewConnectionManager(pollingManager, repo, logger)

	return &mtconnectService{
		connMgr: connectionManager,
		pollMgr: pollingManager,
	}
}

// --- Реализация методов интерфейса MTConnectService ---

func (s *mtconnectService) CreateConnection(req models.ConnectionRequest) (*models.ConnectionInfo, error) {
	return s.connMgr.CreateConnection(req)
}

func (s *mtconnectService) RestoreConnection(machine entities.CncMachine) (*models.ConnectionInfo, error) {
	return s.connMgr.RestoreConnection(machine)
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

func (s *mtconnectService) StartPolling(conn *models.ConnectionInfo, interval time.Duration) error {
	return s.pollMgr.StartPolling(conn, interval)
}

func (s *mtconnectService) StopPolling(sessionID string) error {
	return s.pollMgr.StopPolling(sessionID)
}

func (s *mtconnectService) IsPollingActive(sessionID string) bool {
	return s.pollMgr.IsPollingActive(sessionID)
}
