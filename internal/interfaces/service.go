package interfaces

// PollingService определяет контракт для сервиса, опрашивающего эндпоинты
type PollingService interface {
	StartPolling()
	StopPolling()
}
