package cnc_machine

import (
	"github.com/iwtcode/MTConnect/internal/domain/entities"

	"gorm.io/gorm"
)

func (r *CncMachineRepositoryImpl) Create(machine *entities.CncMachine) error {
	return r.db.Create(machine).Error
}

func (r *CncMachineRepositoryImpl) GetByEndpointAndModel(endpointURL, model string) (*entities.CncMachine, error) {
	var machine entities.CncMachine
	err := r.db.Where("endpoint_url = ? AND model = ?", endpointURL, model).First(&machine).Error
	if err != nil {
		return nil, err
	}
	return &machine, nil
}

func (r *CncMachineRepositoryImpl) UpdateStatus(sessionID, status string) error {
	result := r.db.Model(&entities.CncMachine{}).Where("session_id = ?", sessionID).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdatePollingState обновляет статус и интервал опроса
func (r *CncMachineRepositoryImpl) UpdatePollingState(sessionID, status string, interval int) error {
	updates := map[string]interface{}{
		"status":   status,
		"interval": interval,
	}
	result := r.db.Model(&entities.CncMachine{}).Where("session_id = ?", sessionID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *CncMachineRepositoryImpl) Delete(sessionID string) error {
	result := r.db.Delete(&entities.CncMachine{}, "session_id = ?", sessionID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *CncMachineRepositoryImpl) GetBySessionID(sessionID string) (*entities.CncMachine, error) {
	var machine entities.CncMachine
	err := r.db.First(&machine, "session_id = ?", sessionID).Error
	if err != nil {
		return nil, err
	}
	return &machine, nil
}

func (r *CncMachineRepositoryImpl) GetAllByStatus(status string) ([]entities.CncMachine, error) {
	var machines []entities.CncMachine
	if err := r.db.Where("status = ?", status).Find(&machines).Error; err != nil {
		return nil, err
	}
	return machines, nil
}

// GetAll возвращает все сохраненные станки
func (r *CncMachineRepositoryImpl) GetAll() ([]entities.CncMachine, error) {
	var machines []entities.CncMachine
	if err := r.db.Find(&machines).Error; err != nil {
		return nil, err
	}
	return machines, nil
}
