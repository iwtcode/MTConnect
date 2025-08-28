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
	existingMachineDB, err := s.dbRepo.GetByEndpointAndModel(req.EndpointURL, req.Model)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("ошибка при проверке станка в БД: %w", err)
	}
	if existingMachineDB != nil {
		s.mu.RLock()
		_, exists := s.pool[existingMachineDB.SessionID]
		s.mu.RUnlock()
		if exists {
			return nil, fmt.Errorf("подключение для '%s' на '%s' уже активно с SessionID: %s", req.Model, req.EndpointURL, existingMachineDB.SessionID)
		}
		return nil, fmt.Errorf("подключение для '%s' на '%s' уже существует в БД с SessionID: %s. Удалите старое подключение перед созданием нового", req.Model, req.EndpointURL, existingMachineDB.SessionID)
	}

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

	targetDevice, err := findTargetDevice(&devices, req.Model, req.Manufacturer, req.EndpointURL)
	if err != nil {
		return nil, err
	}

	if err := s.pollingMgr.LoadMetadataForEndpoint(req.EndpointURL); err != nil {
		return nil, fmt.Errorf("ошибка при загрузке метаданных для %s: %w", req.EndpointURL, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	sessionID := uuid.New().String()
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

	connInfo := createConnectionInfo(sessionID, targetDevice.Name, req, targetDevice.Description.Manufacturer)

	// Начальная проверка состояния
	errCheck := s.pollingMgr.CheckMachineConnection(connInfo.Config.EndpointURL)
	connInfo.IsHealthy = (errCheck == nil)
	if !connInfo.IsHealthy {
		log.Printf("ПРЕДУПРЕЖДЕНИЕ: Начальная проверка состояния для сессии %s провалена. Эндпоинт недоступен.", sessionID)
	}

	s.pool[sessionID] = connInfo

	return connInfo, nil
}

// RestoreConnection восстанавливает подключение из БД в пул памяти.
// Эта функция теперь всегда успешна, даже если эндпоинт недоступен,
// помечая такое соединение как IsHealthy: false.
func (s *ConnectionManager) RestoreConnection(machine entities.CncMachine) (*models.ConnectionInfo, error) {
	req := models.ConnectionRequest{
		EndpointURL:  machine.EndpointURL,
		Model:        machine.Model,
		Manufacturer: machine.Manufacturer,
	}

	// Создаем базовый объект подключения, по умолчанию нездоровый
	connInfo := createConnectionInfo(machine.SessionID, "unknown", req, machine.Manufacturer)
	connInfo.IsHealthy = false

	probeURL := strings.TrimSuffix(machine.EndpointURL, "/") + "/probe"
	xmlData, err := client.FetchXML(probeURL)
	if err != nil {
		log.Printf("ПРЕДУПРЕЖДЕНИЕ: Не удалось получить /probe для сессии %s: %v. Соединение будет восстановлено как нездоровое.", machine.SessionID, err)
	} else {
		var devices models.MTConnectDevices
		if err := xml.Unmarshal(xmlData, &devices); err != nil {
			log.Printf("ПРЕДУПРЕЖДЕНИЕ: Не удалось распарсить /probe для сессии %s: %v.", machine.SessionID, err)
		} else if len(devices.Devices) == 0 {
			log.Printf("ПРЕДУПРЕЖДЕНИЕ: Устройства не найдены в /probe для сессии %s.", machine.SessionID)
		} else {
			targetDevice, err := findTargetDevice(&devices, machine.Model, machine.Manufacturer, machine.EndpointURL)
			if err != nil {
				log.Printf("ПРЕДУПРЕЖДЕНИЕ: Целевое устройство не найдено для сессии %s: %v.", machine.SessionID, err)
			} else {
				// Все успешно, обновляем информацию и статус
				connInfo.MachineID = targetDevice.Name
				connInfo.Config.Manufacturer = targetDevice.Description.Manufacturer
				connInfo.IsHealthy = true
				if err := s.pollingMgr.LoadMetadataForEndpoint(machine.EndpointURL); err != nil {
					log.Printf("ПРЕДУПРЕЖДЕНИЕ: Не удалось загрузить метаданные для %s: %v", machine.EndpointURL, err)
				}
			}
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.pool[machine.SessionID] = connInfo

	// Всегда возвращаем информацию, ошибки логируются выше
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
		err := s.dbRepo.Delete(sessionID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("сессия '%s' не найдена ни в активном пуле, ни в БД", sessionID)
			}
			return fmt.Errorf("ошибка удаления сессии '%s' из БД: %w", sessionID, err)
		}
		log.Printf("Сессия '%s' (не в пуле) успешно удалена из БД.", sessionID)
		return nil
	}

	_ = s.pollingMgr.StopPollingForMachine(sessionID)
	delete(s.pool, sessionID)

	if err := s.dbRepo.Delete(sessionID); err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("ошибка удаления сессии '%s' из БД: %w", sessionID, err)
	}

	log.Printf("Сессия '%s' успешно удалена.", sessionID)
	return nil
}

func (s *ConnectionManager) CheckConnection(sessionID string) (*models.ConnectionInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	conn, exists := s.pool[sessionID]
	if !exists {
		return nil, fmt.Errorf("сессия '%s' не найдена", sessionID)
	}

	previousHealth := conn.IsHealthy
	err := s.pollingMgr.CheckMachineConnection(conn.Config.EndpointURL)
	conn.IsHealthy = (err == nil)
	conn.LastUsed = time.Now()
	conn.UseCount++

	if previousHealth != conn.IsHealthy {
		log.Printf("Статус здоровья сессии '%s' изменен: %v -> %v", sessionID, previousHealth, conn.IsHealthy)
	}

	return conn, err
}

// Вспомогательная функция для поиска устройства
func findTargetDevice(devices *models.MTConnectDevices, model, manufacturer, endpointURL string) (*models.Device, error) {
	var targetDevice *models.Device
	for i := range devices.Devices {
		device := &devices.Devices[i]
		if device.Description == nil {
			continue
		}
		replacer := strings.NewReplacer("\n", " ", "\t", " ", "\r", " ")
		cleanedDescription := replacer.Replace(device.Description.Value)
		normalizedDescription := strings.Join(strings.Fields(cleanedDescription), " ")
		if strings.Contains(normalizedDescription, model) {
			targetDevice = device
			break
		}
	}
	if targetDevice == nil {
		return nil, fmt.Errorf("устройство с моделью '%s' не найдено на эндпоинте %s", model, endpointURL)
	}
	if manufacturer != "" && !strings.EqualFold(targetDevice.Description.Manufacturer, manufacturer) {
		return nil, fmt.Errorf("производитель '%s' не совпадает с указанным в /probe для найденной модели: '%s'", manufacturer, targetDevice.Description.Manufacturer)
	}
	return targetDevice, nil
}

// Вспомогательная функция для создания ConnectionInfo
func createConnectionInfo(sessionID, machineID string, req models.ConnectionRequest, probeManufacturer string) *models.ConnectionInfo {
	return &models.ConnectionInfo{
		SessionID: sessionID,
		MachineID: machineID,
		Config: models.ConnectionConfig{
			EndpointURL:  req.EndpointURL,
			Model:        req.Model,
			Manufacturer: probeManufacturer,
		},
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
		UseCount:  1,
		IsHealthy: true, // По умолчанию true, но может быть немедленно перезаписано
	}
}
