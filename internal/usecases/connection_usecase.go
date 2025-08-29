package usecases

import (
	"fmt"
	"time"

	"github.com/iwtcode/MTConnect/internal/domain/entities"
	"github.com/iwtcode/MTConnect/internal/domain/models"
	"github.com/iwtcode/MTConnect/internal/interfaces"
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

func (u *Usecase) RestoreConnection(machine entities.CncMachine) (*models.ConnectionInfo, error) {
	return u.mtconnectSvc.RestoreConnection(machine)
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

func (u *Usecase) StartPolling(sessionID string, interval time.Duration) error {
	conn, found := u.mtconnectSvc.GetConnection(sessionID)
	if !found {
		return fmt.Errorf("не удалось запустить опрос: сессия '%s' не найдена в активном пуле", sessionID)
	}
	return u.mtconnectSvc.StartPolling(conn, interval)
}

func (u *Usecase) StopPolling(sessionID string) error {
	return u.mtconnectSvc.StopPolling(sessionID)
}
