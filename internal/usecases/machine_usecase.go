package usecases

import (
	"MTConnect/internal/domain/entities"
	"MTConnect/internal/interfaces"
	"fmt"
	"time"
)

type MachineUsecase struct {
	repo         interfaces.DataStoreRepository
	poll_service interfaces.PollingService
}

func NewMachineUsecase(repo interfaces.Repository, poll_service interfaces.PollingService) interfaces.MachineUsecase {
	return &MachineUsecase{
		repo:         repo,
		poll_service: poll_service,
	}
}

// GetMachineData - основная бизнес-логика для получения данных станка
func (u *MachineUsecase) GetMachineData(machineId string) (entities.MachineData, error) {
	data, found := u.repo.Get(machineId)
	if !found {
		return entities.MachineData{}, fmt.Errorf("данные для станка '%s' не найдены", machineId)
	}
	return data, nil
}

// StartPolling запускает опрос для указанного станка
func (u *MachineUsecase) StartPolling(machineId string, interval time.Duration) error {
	return u.poll_service.StartPollingForMachine(machineId, interval)
}

// StopPolling останавливает опрос для указанного станка
func (u *MachineUsecase) StopPolling(machineId string) error {
	return u.poll_service.StopPollingForMachine(machineId)
}

// CheckMachineConnection проверяет доступность станка
func (u *MachineUsecase) CheckMachineConnection(machineId string) error {
	return u.poll_service.CheckConnection(machineId)
}
