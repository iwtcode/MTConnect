package poller

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/iwtcode/MTConnect/internal/domain/entities"
	"github.com/iwtcode/MTConnect/internal/domain/models"
	"github.com/iwtcode/MTConnect/internal/interfaces"
	"github.com/iwtcode/MTConnect/internal/middleware/logging"
	"github.com/iwtcode/MTConnect/internal/services/mtconnect_service/client"
	"github.com/iwtcode/MTConnect/internal/services/mtconnect_service/parser"
)

type activePoll struct {
	ticker *time.Ticker
	done   chan bool
}

// MetadataStore хранит все метаданные для одного эндпоинта.
type MetadataStore struct {
	DeviceMetadata map[string]models.DataItemMetadata
	AxisLinks      map[string]models.AxisDataItemLink
	SpindleLinks   map[string]models.SpindleDataItemLink
}

// PollingManager - упрощенная структура, сфокусированная на опросе и отправке данных.
type PollingManager struct {
	dbRepo         interfaces.CncMachineRepository
	producer       interfaces.KafkaService
	logger         *logging.Logger
	activePolls    map[string]*activePoll
	pollsMutex     sync.Mutex
	metadataStores map[string]*MetadataStore
}

// NewPollingManager - конструктор Polling
func NewPollingManager(dbRepo interfaces.CncMachineRepository, producer interfaces.KafkaService, logger *logging.Logger) *PollingManager {
	return &PollingManager{
		dbRepo:         dbRepo,
		producer:       producer,
		logger:         logger.WithPrefix("POLLER"),
		activePolls:    make(map[string]*activePoll),
		metadataStores: make(map[string]*MetadataStore),
	}
}

func (s *PollingManager) IsPollingActive(sessionID string) bool {
	s.pollsMutex.Lock()
	defer s.pollsMutex.Unlock()
	_, exists := s.activePolls[sessionID]
	return exists
}

func (s *PollingManager) StartPolling(conn *models.ConnectionInfo, interval time.Duration) error {
	s.pollsMutex.Lock()
	defer s.pollsMutex.Unlock()

	sessionID := conn.SessionID

	if _, exists := s.activePolls[sessionID]; exists {
		return fmt.Errorf("опрос для сессии '%s' уже запущен", sessionID)
	}

	// Загружаем метаданные для этой сессии при первом запуске опроса.
	// Это происходит один раз и защищено мьютексом.
	if _, exists := s.metadataStores[sessionID]; !exists {
		store, err := s.fetchAndParseProbe(conn.Config.EndpointURL)
		if err != nil {
			s.logger.Error("Failed to load metadata, polling will not start", "endpoint", conn.Config.EndpointURL, "error", err)
			return fmt.Errorf("не удалось загрузить метаданные для эндпоинта: %w", err)
		}
		s.metadataStores[sessionID] = store
		s.logger.Info("Successfully loaded metadata for session",
			"sessionID", sessionID,
			"dataItems", len(store.DeviceMetadata),
			"axisLinks", len(store.AxisLinks),
			"spindleLinks", len(store.SpindleLinks))
	}

	if err := s.dbRepo.UpdatePollingState(sessionID, entities.StatusPolled, int(interval.Milliseconds())); err != nil {
		return fmt.Errorf("не удалось обновить статус станка в БД: %w", err)
	}

	s.startPollingForMachineUnsafe(sessionID, conn.Config.EndpointURL, conn.MachineID, interval)

	return nil
}

func (s *PollingManager) StopPolling(sessionID string) error {
	s.pollsMutex.Lock()
	defer s.pollsMutex.Unlock()

	if err := s.dbRepo.UpdatePollingState(sessionID, entities.StatusConnected, 0); err != nil {
		s.logger.Error("Failed to update status in DB when stopping polling", "sessionID", sessionID, "error", err)
	}

	poll, exists := s.activePolls[sessionID]
	if !exists {
		s.logger.Warn("Attempt to stop polling for a session that was not active.", "sessionID", sessionID)
		return nil
	}

	poll.ticker.Stop()
	poll.done <- true
	close(poll.done)
	delete(s.activePolls, sessionID)
	// Удаляем метаданные, связанные с этой сессией
	delete(s.metadataStores, sessionID)

	return nil
}

func (s *PollingManager) CheckMachineConnection(endpointURL string) error {
	probeURL := strings.TrimSuffix(endpointURL, "/") + "/probe"
	_, err := client.FetchXML(probeURL)
	if err != nil {
		return fmt.Errorf("проверка соединения с эндпоинтом '%s' провалена: %w", endpointURL, err)
	}
	return nil
}

// processSingleEndpoint получает данные, парсит их и отправляет в Kafka
func (s *PollingManager) processSingleEndpoint(sessionID string, endpointURL string, targetMachineID string) {
	xmlData, err := client.FetchXML(endpointURL)
	if err != nil {
		s.logger.Error("Error fetching XML", "url", endpointURL, "error", err)
		return
	}

	var streams models.MTConnectStreams
	if err := xml.Unmarshal(xmlData, &streams); err != nil {
		s.logger.Error("Error parsing XML", "url", endpointURL, "error", err)
		return
	}

	store, exists := s.metadataStores[sessionID]

	if !exists {
		s.logger.Error("Metadata store not found for a running poll", "sessionID", sessionID)
		return
	}

	machineDataSlice := parser.MapToMachineData(&streams, store.DeviceMetadata, store.AxisLinks, store.SpindleLinks)

	for _, machineData := range machineDataSlice {
		if machineData.MachineId == targetMachineID {

			jsonData, err := json.Marshal(machineData)
			if err != nil {
				s.logger.Error("Failed to serialize MachineData for Kafka", "error", err)
				continue
			}
			err = s.producer.Produce(context.Background(), []byte(machineData.MachineId), jsonData)
			if err != nil {
				s.logger.Error("Failed to send data to Kafka", "machineId", machineData.MachineId, "error", err)
			}
			break // Нашли нужный станок, выходим из цикла
		}
	}
}

func (s *PollingManager) fetchAndParseProbe(endpointURL string) (*MetadataStore, error) {
	probeURL := strings.TrimSuffix(endpointURL, "/") + "/probe"
	s.logger.Info("Loading metadata", "url", probeURL)

	xmlData, err := client.FetchXML(probeURL)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить /probe с %s: %w", probeURL, err)
	}

	var devices models.MTConnectDevices
	if err := xml.Unmarshal(xmlData, &devices); err != nil {
		return nil, fmt.Errorf("не удалось распарсить /probe XML с %s: %w", probeURL, err)
	}

	store := &MetadataStore{
		DeviceMetadata: make(map[string]models.DataItemMetadata),
		AxisLinks:      make(map[string]models.AxisDataItemLink),
		SpindleLinks:   make(map[string]models.SpindleDataItemLink),
	}

	for _, device := range devices.Devices {
		deviceId := device.Name
		if deviceId == "" {
			deviceId = device.UUID
		}

		for _, item := range device.DataItems {
			store.DeviceMetadata[strings.ToLower(item.ID)] = models.DataItemMetadata{
				ID: item.ID, Name: item.Name, ComponentId: device.ID, ComponentName: device.Name,
				ComponentType: "Device", Category: item.Category, Type: item.Type, SubType: item.SubType,
			}
		}

		if device.ComponentList != nil {
			s.extractComponentMetadata(store, device.ComponentList.Components, deviceId)
		}
	}
	return store, nil
}

func (s *PollingManager) extractComponentMetadata(store *MetadataStore, components []models.ProbeComponent, deviceId string) {
	for _, comp := range components {
		componentType := strings.ToUpper(comp.XMLName.Local)
		isAxisOrSpindle := componentType == "LINEAR" || componentType == "ROTARY"

		for _, item := range comp.DataItems {
			lowerId := strings.ToLower(item.ID)
			store.DeviceMetadata[lowerId] = models.DataItemMetadata{
				ID: item.ID, Name: item.Name, ComponentId: comp.ID, ComponentName: comp.Name,
				ComponentType: strings.ToLower(comp.XMLName.Local), Category: item.Category, Type: item.Type, SubType: item.SubType,
			}

			if isAxisOrSpindle && item.Type != "" && item.Type != "AXIS_STATE" {
				dataKey := strings.ToLower(item.Type)
				switch componentType {
				case "LINEAR":
					store.AxisLinks[lowerId] = models.AxisDataItemLink{
						DeviceID: deviceId, AxisComponentID: comp.ID, AxisName: comp.Name, AxisType: componentType, DataKey: dataKey,
					}
				case "ROTARY":
					store.SpindleLinks[lowerId] = models.SpindleDataItemLink{
						DeviceID: deviceId, SpindleComponentID: comp.ID, SpindleName: comp.Name, SpindleType: componentType, DataKey: dataKey,
					}
				}
			}
		}

		if comp.ComponentList != nil {
			s.extractComponentMetadata(store, comp.ComponentList.Components, deviceId)
		}
	}
}

func (s *PollingManager) startPollingForMachineUnsafe(sessionID, endpointURL, machineID string, interval time.Duration) {
	if _, exists := s.activePolls[sessionID]; exists {
		s.logger.Warn("Polling for session already started, skipping.", "sessionID", sessionID)
		return
	}

	ticker := time.NewTicker(interval)
	done := make(chan bool)

	s.activePolls[sessionID] = &activePoll{
		ticker: ticker,
		done:   done,
	}

	go func() {
		s.logger.Info("Starting polling", "sessionID", sessionID, "machineID", machineID, "interval", interval)
		currentURL := strings.TrimSuffix(endpointURL, "/") + "/current"
		for {
			select {
			case <-done:
				s.logger.Info("Polling stopped", "sessionID", sessionID)
				return
			case <-ticker.C:
				s.processSingleEndpoint(sessionID, currentURL, machineID)
			}
		}
	}()
}
