package usecases

import "MTConnect/internal/interfaces"

// UseCases - агрегатор всех use case интерфейсов
type UseCases struct {
	interfaces.MachineUsecase
}

// NewUsecases - конструктор для UseCases
func NewUsecases(repo interfaces.Repository, poll_service interfaces.PollingService) interfaces.Usecases {
	return &UseCases{
		MachineUsecase: NewMachineUsecase(repo, poll_service),
	}
}
