package interfaces

import "MTConnect/internal/domain/entities"

// Repository - это агрегирующий интерфейс для всех репозиториев
type Repository interface {
	DataStoreRepository
	CncMachineRepository // Добавлено
}

// CncMachineRepository определяет контракт для работы с сохраненными станками в БД
type CncMachineRepository interface {
	Create(machine *entities.CncMachine) error
	GetByEndpointAndModel(endpointURL, model string) (*entities.CncMachine, error)
	UpdateStatus(sessionID, status string) error
	Delete(sessionID string) error
	GetBySessionID(sessionID string) (*entities.CncMachine, error)
	GetAllByStatus(status string) ([]entities.CncMachine, error)
}

// DataStoreRepository определяет контракт для хранилища данных станков
type DataStoreRepository interface {
	Set(machineId string, data entities.MachineData)
	Get(machineId string) (entities.MachineData, bool)
}
