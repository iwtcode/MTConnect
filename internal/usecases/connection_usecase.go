package usecases

import (
	"MTConnect/internal/domain/models"
	"MTConnect/internal/interfaces"
	"time"
)

type Usecase struct {
	mtconnectSvc interfaces.MTConnectService
}

func NewUsecase(mtconnectSvc interfaces.MTConnectService) interfaces.Usecases {
	return &Usecase{
		mtconnectSvc: mtconnectSvc,
	}
}

func (u *Usecase) CreateConnection(req models.ConnectionRequest) (*models.ConnectionInfo, error) {
	return u.mtconnectSvc.CreateConnection(req)
}

func (u *Usecase) GetAllConnections() []*models.ConnectionInfo {
	return u.mtconnectSvc.GetAllConnections()
}

func (u *Usecase) DeleteConnection(sessionID string) error {
	return u.mtconnectSvc.DeleteConnection(sessionID)
}

func (u *Usecase) CheckConnection(sessionID string) (*models.ConnectionInfo, error) {
	return u.mtconnectSvc.CheckConnection(sessionID)
}

func (u *Usecase) StartPolling(interval time.Duration) error {
	return u.mtconnectSvc.StartPolling(interval)
}

func (u *Usecase) StopPolling() error {
	return u.mtconnectSvc.StopPolling()
}
