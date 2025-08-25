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
)

type ConnectionService struct {
	mu         sync.RWMutex
	pool       map[string]*entities.ConnectionInfo
	pollingSvc interfaces.PollingService
}

func NewConnectionService(pollingSvc interfaces.PollingService) interfaces.ConnectionService {
	return &ConnectionService{
		pool:       make(map[string]*entities.ConnectionInfo),
		pollingSvc: pollingSvc,
	}
}

// CreateConnection проверяет новый запрос на подключение и добавляет его в пул.
func (s *ConnectionService) CreateConnection(req entities.ConnectionRequest) (*entities.ConnectionInfo, error) {
	s.mu.RLock()
	for _, conn := range s.pool {
		if conn.Config.EndpointURL == req.EndpointURL && conn.Config.Model == req.Model {
			s.mu.RUnlock()
			return nil, fmt.Errorf("подключение для модели '%s' на эндпоинте '%s' уже существует с SessionID: %s", req.Model, req.EndpointURL, conn.SessionID)
		}
	}
	s.mu.RUnlock()

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

	if err := s.pollingSvc.LoadMetadataForEndpoint(req.EndpointURL); err != nil {
		return nil, fmt.Errorf("ошибка при загрузке метаданных для %s: %w", req.EndpointURL, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	sessionID := uuid.New().String()
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

	// --- ИЗМЕНЕНИЕ ЗДЕСЬ ---
	// После успешного добавления подключения в пул,
	// пытаемся запустить для него опрос, если глобальный опрос активен.
	if err := s.pollingSvc.StartPollingForNewConnectionIfNeeded(connInfo); err != nil {
		// Эта ошибка не должна откатывать создание подключения,
		// но ее стоит залогировать.
		log.Printf("ПРЕДУПРЕЖДЕНИЕ: не удалось автоматически запустить опрос для новой сессии %s: %v", connInfo.SessionID, err)
	}

	return connInfo, nil
}

// ... Остальные функции (GetConnection, GetAllConnections, DeleteConnection, CheckConnection) остаются без изменений ...
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
		return fmt.Errorf("сессия '%s' не найдена", sessionID)
	}

	_ = s.pollingSvc.StopPollingForMachine(sessionID)

	delete(s.pool, sessionID)
	return nil
}

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
