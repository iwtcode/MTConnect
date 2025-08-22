package datastore

import (
	"MTConnect/internal/domain/entities"
	"MTConnect/internal/interfaces"
	"sync"
)

// DataStore - потокобезопасное in-memory хранилище данных
type DataStore struct {
	mu   sync.RWMutex
	data map[string]entities.MachineData
}

// NewDataStore создает новый экземпляр DataStore
func NewDataStore() interfaces.DataStoreRepository {
	return &DataStore{
		data: make(map[string]entities.MachineData),
	}
}

// Set сохраняет данные для указанного станка
func (ds *DataStore) Set(machineId string, data entities.MachineData) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.data[machineId] = data
}

// Get извлекает данные для указанного станка
func (ds *DataStore) Get(machineId string) (entities.MachineData, bool) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	machineData, found := ds.data[machineId]
	return machineData, found
}
