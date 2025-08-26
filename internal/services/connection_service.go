package services

import (
	"MTConnect/internal/domain/entities"
	"MTConnect/internal/interfaces"
	"encoding/xml"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ConnectionService struct {
	mu         sync.RWMutex
	pool       map[string]*entities.ConnectionInfo
	pollingSvc interfaces.PollingService
	dbRepo     interfaces.CncMachineRepository // Добавлено
}

func NewConnectionService(pollingSvc interfaces.PollingService, dbRepo interfaces.CncMachineRepository) interfaces.ConnectionService { // Добавлен dbRepo
	return &ConnectionService{
		pool:       make(map[string]*entities.ConnectionInfo),
		pollingSvc: pollingSvc,
		dbRepo:     dbRepo, // Добавлено
	}
}

// CreateConnection проверяет новый запрос на подключение и добавляет его в пул.
// Функция стала идемпотентной: если станок уже есть в БД, она использует существующий
// SessionID и обновляет его статус. Если нет - создает новую запись.
func (s *ConnectionService) CreateConnection(req entities.ConnectionRequest) (*entities.ConnectionInfo, error) {
	// 1. Проверка на дубликаты в текущем пуле активных подключений (в памяти)
	s.mu.RLock()
	for _, conn := range s.pool {
		if conn.Config.EndpointURL == req.EndpointURL && conn.Config.Model == req.Model {
			s.mu.RUnlock()
			log.Printf("Запрос на подключение к уже активной сессии '%s'. Возвращаем существующие данные.", conn.SessionID)
			return conn, nil // Возвращаем существующее активное подключение
		}
	}
	s.mu.RUnlock()

	// 2. Получение и парсинг /probe для валидации эндпоинта и модели
	probeURL := strings.TrimSuffix(req.EndpointURL, "/") + "/probe"

	xmlData, err := FetchXML(probeURL)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить /probe с %s: %w", probeURL, err)
	}

	var devices entities.MTConnectDevices
	if err := xml.Unmarshal(xmlData, &devices); err != nil {
		return nil, fmt.Errorf("не удалось распарсить /probe XML с %s: %w", probeURL, err)
	}

	if len(devices.Devices) == 0 {
		return nil, fmt.Errorf("устройства не найдены в /probe ответе от %s", probeURL)
	}

	var targetDevice *entities.Device
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

	// Загружаем метаданные до блокировки, чтобы не задерживать другие операции
	if err := s.pollingSvc.LoadMetadataForEndpoint(req.EndpointURL); err != nil {
		return nil, fmt.Errorf("ошибка при загрузке метаданных для %s: %w", req.EndpointURL, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 3. Проверяем, существует ли станок в базе данных
	existingMachine, err := s.dbRepo.GetByEndpointAndModel(req.EndpointURL, req.Model)

	var sessionID string
	if err != nil && err != gorm.ErrRecordNotFound {
		// Ошибка при обращении к БД, не связанная с отсутствием записи
		return nil, fmt.Errorf("ошибка при проверке станка в БД: %w", err)
	}

	if existingMachine != nil {
		// 4. СТАНОК УЖЕ СУЩЕСТВУЕТ В БД - используем его данные
		sessionID = existingMachine.SessionID
		log.Printf("Станок '%s' (%s) уже существует в БД. Используем SessionID: %s", req.Model, req.EndpointURL, sessionID)

		// Обновляем статус на 'connected', если он был другим
		if existingMachine.Status != entities.StatusConnected {
			if err := s.dbRepo.UpdateStatus(sessionID, entities.StatusConnected); err != nil {
				log.Printf("ПРЕДУПРЕЖДЕНИЕ: не удалось обновить статус для сессии %s в БД: %v", sessionID, err)
			}
		}
	} else {
		// 5. СТАНОК НОВЫЙ - создаем запись в БД
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

	// 6. Создаем или обновляем информацию о подключении в памяти
	connInfo := &entities.ConnectionInfo{
		SessionID: sessionID,
		MachineID: targetDevice.Name,
		Config: entities.ConnectionConfig{
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

	// 7. Запускаем опрос, если он уже активен глобально
	if err := s.pollingSvc.StartPollingForNewConnectionIfNeeded(connInfo); err != nil {
		log.Printf("ПРЕДУПРЕЖДЕНИЕ: не удалось автоматически запустить опрос для сессии %s: %v", connInfo.SessionID, err)
	}

	return connInfo, nil
}

// ... (GetConnection и GetAllConnections без изменений) ...
func (s *ConnectionService) GetConnection(sessionID string) (*entities.ConnectionInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	conn, found := s.pool[sessionID]
	return conn, found
}

func (s *ConnectionService) GetAllConnections() []*entities.ConnectionInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	conns := make([]*entities.ConnectionInfo, 0, len(s.pool))
	for _, conn := range s.pool {
		conns = append(conns, conn)
	}
	return conns
}

func (s *ConnectionService) DeleteConnection(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.pool[sessionID]; !exists {
		// Если сессии нет в памяти, все равно пытаемся обновить статус в БД
		if err := s.dbRepo.UpdateStatus(sessionID, entities.StatusDisconnected); err != nil && err != gorm.ErrRecordNotFound {
			log.Printf("ПРЕДУПРЕЖДЕНИЕ: не удалось обновить статус сессии %s в БД: %v", sessionID, err)
		}
		return fmt.Errorf("сессия '%s' не найдена в активном пуле", sessionID)
	}

	if err := s.dbRepo.UpdateStatus(sessionID, entities.StatusDisconnected); err != nil {
		log.Printf("ПРЕДУПРЕЖДЕНИЕ: не удалось обновить статус сессии %s в БД: %v", sessionID, err)
	}

	_ = s.pollingSvc.StopPollingForMachine(sessionID)

	delete(s.pool, sessionID)
	return nil
}

// ... (CheckConnection без изменений) ...
func (s *ConnectionService) CheckConnection(sessionID string) (*entities.ConnectionInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	conn, exists := s.pool[sessionID]
	if !exists {
		return nil, fmt.Errorf("сессия '%s' не найдена", sessionID)
	}

	err := s.pollingSvc.CheckMachineConnection(conn.Config.EndpointURL)
	conn.IsHealthy = (err == nil)
	conn.LastUsed = time.Now()
	conn.UseCount++

	return conn, err
}
