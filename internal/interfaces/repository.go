package interfaces

import "MTConnect/internal/domain/entities"

// Repository - это агрегирующий интерфейс для всех репозиториев
type Repository interface {
	DataStoreRepository
}

// DataStoreRepository определяет контракт для хранилища данных станков
type DataStoreRepository interface {
	Set(machineId string, data entities.MachineData)
	Get(machineId string) (entities.MachineData, bool)
}
