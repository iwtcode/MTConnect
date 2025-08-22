package usecases

import (
	"MTConnect/internal/domain/entities"
	"MTConnect/internal/interfaces"
	"fmt"
)

type MachineUsecase struct {
	repo interfaces.DataStoreRepository
}

func NewMachineUsecase(repo interfaces.Repository) interfaces.MachineUsecase {
	return &MachineUsecase{repo: repo}
}

// GetMachineData - основная бизнес-логика для получения данных станка
func (u *MachineUsecase) GetMachineData(machineId string) (entities.MachineData, error) {
	data, found := u.repo.Get(machineId)
	if !found {
		return entities.MachineData{}, fmt.Errorf("данные для станка '%s' не найдены", machineId)
	}
	return data, nil
}
