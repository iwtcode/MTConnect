package interfaces

import (
	"MTConnect/internal/domain/entities"
	"time"
)

// Usecases - это агрегирующий интерфейс для всех use cases
type Usecases interface {
	MachineUsecase
}

// MachineUsecase определяет контракт для бизнес-логики, связанной со станками
type MachineUsecase interface {
	GetMachineData(machineId string) (entities.MachineData, error)
	StartPolling(machineId string, interval time.Duration) error
	StopPolling(machineId string) error
	CheckMachineConnection(machineId string) error
}
