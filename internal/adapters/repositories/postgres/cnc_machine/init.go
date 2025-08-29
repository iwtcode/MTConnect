package cnc_machine

import (
	"github.com/iwtcode/MTConnect/internal/interfaces"

	"gorm.io/gorm"
)

type CncMachineRepositoryImpl struct {
	db *gorm.DB
}

func NewCncMachineRepository(db *gorm.DB) interfaces.CncMachineRepository {
	return &CncMachineRepositoryImpl{db: db}
}
