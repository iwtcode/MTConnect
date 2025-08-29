package interfaces

import (
	"github.com/iwtcode/MTConnect/internal/domain/entities"
)

// CncMachineRepository определяет контракт для работы с сохраненными станками в БД
type CncMachineRepository interface {
	Create(machine *entities.CncMachine) error
	GetByEndpointAndModel(endpointURL, model string) (*entities.CncMachine, error)
	UpdateStatus(sessionID, status string) error
	UpdatePollingState(sessionID, status string, interval int) error
	Delete(sessionID string) error
	GetBySessionID(sessionID string) (*entities.CncMachine, error)
	GetAllByStatus(status string) ([]entities.CncMachine, error)
	GetAll() ([]entities.CncMachine, error)
}
