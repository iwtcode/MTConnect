package mtconnect_service

import (
	"time"

	"github.com/iwtcode/MTConnect/internal/domain/entities"
	"github.com/iwtcode/MTConnect/internal/domain/models"
	"github.com/iwtcode/MTConnect/internal/interfaces"
	"github.com/iwtcode/MTConnect/internal/middleware/logging"
	"github.com/iwtcode/MTConnect/internal/services/mtconnect_service/connector"
	"github.com/iwtcode/MTConnect/internal/services/mtconnect_service/poller"
)

type mtconnectService struct {
	connMgr *connector.ConnectionManager
	pollMgr *poller.PollingManager
	logger  *logging.Logger
}

func NewMTConnectService(repo interfaces.CncMachineRepository, producer interfaces.KafkaService, logger *logging.Logger) interfaces.MTConnectService {
	pollingManager := poller.NewPollingManager(repo, producer, logger)
	connectionManager := connector.NewConnectionManager(repo, logger)

	return &mtconnectService{
		connMgr: connectionManager,
		pollMgr: pollingManager,
		logger:  logger,
	}
}

func (s *mtconnectService) CreateConnection(req models.ConnectionRequest) (*models.ConnectionInfo, error) {
	connInfo, err := s.connMgr.CreateConnection(req)
	if err != nil {
		return nil, err
	}

	return connInfo, nil
}

func (s *mtconnectService) RestoreConnection(machine entities.CncMachine) (*models.ConnectionInfo, error) {
	connInfo, err := s.connMgr.RestoreConnection(machine)
	if err != nil {
		return nil, err
	}

	return connInfo, nil
}

func (s *mtconnectService) GetConnection(sessionID string) (*models.ConnectionInfo, bool) {
	return s.connMgr.GetConnection(sessionID)
}

func (s *mtconnectService) GetAllConnections() []*models.ConnectionInfo {
	return s.connMgr.GetAllConnections()
}

func (s *mtconnectService) DeleteConnection(sessionID string) error {
	if s.pollMgr.IsPollingActive(sessionID) {
		if err := s.pollMgr.StopPolling(sessionID); err != nil {
			s.logger.Error("Error stopping polling before deleting connection", "sessionID", sessionID, "error", err)
		}
	}

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
