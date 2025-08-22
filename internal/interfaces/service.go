package interfaces

import (
	"time"
)

// PollingService определяет контракт для сервиса, опрашивающего эндпоинты
type PollingService interface {
	StartPollingForMachine(machineId string, interval time.Duration) error
	StopPollingForMachine(machineId string) error
	CheckConnection(machineId string) error
	StopAllPolling()
}
