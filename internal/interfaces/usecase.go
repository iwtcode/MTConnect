package interfaces

import "MTConnect/internal/domain/entities"

// Usecases - это агрегирующий интерфейс для всех use cases
type Usecases interface {
	MachineUsecase
}

// MachineUsecase определяет контракт для бизнес-логики, связанной со станками
type MachineUsecase interface {
	GetMachineData(machineId string) (entities.MachineData, error)
}
