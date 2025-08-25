package usecases

import (
	"MTConnect/internal/domain/entities"
	"MTConnect/internal/interfaces"
	"time"
)

type ConnectionUsecase struct {
	connSvc interfaces.ConnectionService
	pollSvc interfaces.PollingService
}

func NewConnectionUsecase(connSvc interfaces.ConnectionService, pollSvc interfaces.PollingService) interfaces.ConnectionUsecase {
	return &ConnectionUsecase{
		connSvc: connSvc,
		pollSvc: pollSvc,
	}
}

func (u *ConnectionUsecase) CreateConnection(req entities.ConnectionRequest) (*entities.ConnectionInfo, error) {
	return u.connSvc.CreateConnection(req)
}

func (u *ConnectionUsecase) GetAllConnections() []*entities.ConnectionInfo {
	return u.connSvc.GetAllConnections()
}

func (u *ConnectionUsecase) DeleteConnection(sessionID string) error {
	return u.connSvc.DeleteConnection(sessionID)
}

func (u *ConnectionUsecase) CheckConnection(sessionID string) (*entities.ConnectionInfo, error) {
	return u.connSvc.CheckConnection(sessionID)
}

func (u *ConnectionUsecase) StartPolling(interval time.Duration) error {
	connections := u.connSvc.GetAllConnections()
	return u.pollSvc.StartAllPolling(connections, interval)
}

func (u *ConnectionUsecase) StopPolling() error {
	u.pollSvc.StopAllPolling()
	return nil
}
