package usecases

import "github.com/iwtcode/MTConnect/internal/interfaces"

// UseCases - агрегатор всех use case интерфейсов
type UseCases struct {
	interfaces.Usecases
}

// NewUsecases - конструктор для UseCases
func NewUsecases(
	mtconnectSvc interfaces.MTConnectService,
) interfaces.Usecases {
	return NewUsecase(mtconnectSvc)
}
