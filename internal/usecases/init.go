package usecases

import "MTConnect/internal/interfaces"

// UseCases - агрегатор всех use case интерфейсов
type UseCases struct {
	interfaces.ConnectionUsecase
}

// NewUsecases - конструктор для UseCases
func NewUsecases(
	repo interfaces.Repository,
	pollSvc interfaces.PollingService,
	connSvc interfaces.ConnectionService,
) interfaces.Usecases {
	return &UseCases{
		ConnectionUsecase: NewConnectionUsecase(connSvc, pollSvc),
	}
}
