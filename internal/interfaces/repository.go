package interfaces

import (
	"MTConnect/internal/domain/entities"
	"MTConnect/internal/domain/models"
)

// Repository - это агрегирующий интерфейс для всех репозиториев
type Repository interface {
	DataStoreRepository
	CncMachineRepository
}

// CncMachineRepository определяет контракт для работы с сохраненными станками в БД
type CncMachineRepository interface {
	Create(machine *entities.CncMachine) error
	GetByEndpointAndModel(endpointURL, model string) (*entities.CncMachine, error)
	UpdateStatus(sessionID, status string) error
	UpdatePollingState(sessionID, status string, interval int) error // Новый метод
	Delete(sessionID string) error
	GetBySessionID(sessionID string) (*entities.CncMachine, error)
	GetAllByStatus(status string) ([]entities.CncMachine, error)
	GetAll() ([]entities.CncMachine, error) // Новый метод для восстановления
}

// DataStoreRepository определяет контракт для хранилища данных станков
type DataStoreRepository interface {
	Set(machineId string, data models.MachineData)
	Get(machineId string) (models.MachineData, bool)
}
