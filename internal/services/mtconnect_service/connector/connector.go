package connector

import (
	"MTConnect/internal/domain/entities"
	"MTConnect/internal/domain/models"
	"MTConnect/internal/interfaces"
	"MTConnect/internal/services/mtconnect_service/client"
	"encoding/xml"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PollingStarter interface {
	StartPollingForNewConnectionIfNeeded(conn *models.ConnectionInfo) error
	StopPollingForMachine(sessionID string) error
	LoadMetadataForEndpoint(endpointURL string) error
	CheckMachineConnection(endpointURL string) error
}

type ConnectionManager struct {
	mu         sync.RWMutex
	pool       map[string]*models.ConnectionInfo
	pollingMgr PollingStarter
	dbRepo     interfaces.CncMachineRepository
}

func NewConnectionManager(pollingMgr PollingStarter, dbRepo interfaces.CncMachineRepository) *ConnectionManager {
	return &ConnectionManager{
		pool:       make(map[string]*models.ConnectionInfo),
		pollingMgr: pollingMgr,
		dbRepo:     dbRepo,
	}
}

func (s *ConnectionManager) CreateConnection(req models.ConnectionRequest) (*models.ConnectionInfo, error) {
	s.mu.RLock()
	for _, conn := range s.pool {
		if conn.Config.EndpointURL == req.EndpointURL && conn.Config.Model == req.Model {
			s.mu.RUnlock()
			log.Printf("Запрос на подключение к уже активной сессии '%s'. Возвращаем существующие данные.", conn.SessionID)
			return conn, nil
		}
	}
	s.mu.RUnlock()

	probeURL := strings.TrimSuffix(req.EndpointURL, "/") + "/probe"
	xmlData, err := client.FetchXML(probeURL)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить /probe с %s: %w", probeURL, err)
	}

	var devices models.MTConnectDevices
	if err := xml.Unmarshal(xmlData, &devices); err != nil {
		return nil, fmt.Errorf("не удалось распарсить /probe XML с %s: %w", probeURL, err)
	}

	if len(devices.Devices) == 0 {
		return nil, fmt.Errorf("устройства не найдены в /probe ответе от %s", probeURL)
	}

	var targetDevice *models.Device
	for i := range devices.Devices {
		device := devices.Devices[i]
		if device.Description == nil {
			continue
		}
		replacer := strings.NewReplacer("\n", " ", "\t", " ", "\r", " ")
		cleanedDescription := replacer.Replace(device.Description.Value)
		normalizedDescription := strings.Join(strings.Fields(cleanedDescription), " ")
		if strings.Contains(normalizedDescription, req.Model) {
			targetDevice = &device
			break
		}
	}
	if targetDevice == nil {
		return nil, fmt.Errorf("устройство с моделью '%s' не найдено на эндпоинте %s", req.Model, req.EndpointURL)
	}

	if req.Manufacturer != "" && !strings.EqualFold(targetDevice.Description.Manufacturer, req.Manufacturer) {
		return nil, fmt.Errorf("производитель '%s' не совпадает с указанным в /probe для найденной модели: '%s'", req.Manufacturer, targetDevice.Description.Manufacturer)
	}

	if err := s.pollingMgr.LoadMetadataForEndpoint(req.EndpointURL); err != nil {
		return nil, fmt.Errorf("ошибка при загрузке метаданных для %s: %w", req.EndpointURL, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existingMachine, err := s.dbRepo.GetByEndpointAndModel(req.EndpointURL, req.Model)
	var sessionID string
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("ошибка при проверке станка в БД: %w", err)
	}

	if existingMachine != nil {
		sessionID = existingMachine.SessionID
		log.Printf("Станок '%s' (%s) уже существует в БД. Используем SessionID: %s", req.Model, req.EndpointURL, sessionID)
		if existingMachine.Status != entities.StatusConnected {
			if err := s.dbRepo.UpdateStatus(sessionID, entities.StatusConnected); err != nil {
				log.Printf("ПРЕДУПРЕЖДЕНИЕ: не удалось обновить статус для сессии %s в БД: %v", sessionID, err)
			}
		}
	} else {
		sessionID = uuid.New().String()
		log.Printf("Создается новое подключение для станка '%s' (%s). Новый SessionID: %s", req.Model, req.EndpointURL, sessionID)
		machineToSave := &entities.CncMachine{
			SessionID:    sessionID,
			EndpointURL:  req.EndpointURL,
			Model:        req.Model,
			Manufacturer: targetDevice.Description.Manufacturer,
			Status:       entities.StatusConnected,
		}
		if err := s.dbRepo.Create(machineToSave); err != nil {
			return nil, fmt.Errorf("не удалось сохранить новое подключение %s в БД: %w", sessionID, err)
		}
	}

	connInfo := &models.ConnectionInfo{
		SessionID: sessionID,
		MachineID: targetDevice.Name,
		Config: models.ConnectionConfig{
			EndpointURL:  req.EndpointURL,
			Model:        req.Model,
			Manufacturer: targetDevice.Description.Manufacturer,
		},
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
		UseCount:  1,
		IsHealthy: true,
	}

	s.pool[sessionID] = connInfo

	if err := s.pollingMgr.StartPollingForNewConnectionIfNeeded(connInfo); err != nil {
		log.Printf("ПРЕДУПРЕЖДЕНИЕ: не удалось автоматически запустить опрос для сессии %s: %v", connInfo.SessionID, err)
	}

	return connInfo, nil
}

func (s *ConnectionManager) GetConnection(sessionID string) (*models.ConnectionInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	conn, found := s.pool[sessionID]
	return conn, found
}

func (s *ConnectionManager) GetAllConnections() []*models.ConnectionInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	conns := make([]*models.ConnectionInfo, 0, len(s.pool))
	for _, conn := range s.pool {
		conns = append(conns, conn)
	}
	return conns
}

func (s *ConnectionManager) DeleteConnection(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.pool[sessionID]; !exists {
		if err := s.dbRepo.UpdateStatus(sessionID, entities.StatusDisconnected); err != nil && err != gorm.ErrRecordNotFound {
			log.Printf("ПРЕДУПРЕЖДЕНИЕ: не удалось обновить статус сессии %s в БД: %v", sessionID, err)
		}
		return fmt.Errorf("сессия '%s' не найдена в активном пуле", sessionID)
	}

	if err := s.dbRepo.UpdateStatus(sessionID, entities.StatusDisconnected); err != nil {
		log.Printf("ПРЕДУПРЕЖДЕНИЕ: не удалось обновить статус сессии %s в БД: %v", sessionID, err)
	}

	_ = s.pollingMgr.StopPollingForMachine(sessionID)
	delete(s.pool, sessionID)
	return nil
}

func (s *ConnectionManager) CheckConnection(sessionID string) (*models.ConnectionInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	conn, exists := s.pool[sessionID]
	if !exists {
		return nil, fmt.Errorf("сессия '%s' не найдена", sessionID)
	}

	err := s.pollingMgr.CheckMachineConnection(conn.Config.EndpointURL)
	conn.IsHealthy = (err == nil)
	conn.LastUsed = time.Now()
	conn.UseCount++

	return conn, err
}
