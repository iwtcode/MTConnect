package poller

import (
	"MTConnect/internal/domain/entities"
	"MTConnect/internal/domain/models"
	"MTConnect/internal/interfaces"
	"MTConnect/internal/services/mtconnect_service/client"
	"MTConnect/internal/services/mtconnect_service/parser"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

type activePoll struct {
	ticker *time.Ticker
	done   chan bool
}

// PollingManager - упрощенная структура, сфокусированная на опросе и отправке данных.
type PollingManager struct {
	dbRepo               interfaces.CncMachineRepository
	producer             interfaces.KafkaService
	activePolls          map[string]*activePoll
	pollsMutex           sync.Mutex
	deviceMetadataStore  map[string]models.DataItemMetadata
	axisDataItemLinks    map[string]models.AxisDataItemLink
	spindleDataItemLinks map[string]models.SpindleDataItemLink
	metadataMutex        sync.RWMutex // Один мьютекс для защиты всех метаданных.
}

// NewPollingManager - обновленный конструктор без in-memory репозитория.
func NewPollingManager(dbRepo interfaces.CncMachineRepository, producer interfaces.KafkaService) *PollingManager {
	return &PollingManager{
		dbRepo:               dbRepo,
		producer:             producer,
		activePolls:          make(map[string]*activePoll),
		deviceMetadataStore:  make(map[string]models.DataItemMetadata),
		axisDataItemLinks:    make(map[string]models.AxisDataItemLink),
		spindleDataItemLinks: make(map[string]models.SpindleDataItemLink),
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
		log.Printf("Не удалось обновить статус для сессии %s в БД при остановке опроса: %v", sessionID, err)
	}

	poll, exists := s.activePolls[sessionID]
	if !exists {
		log.Printf("Попытка остановить опрос для сессии '%s', который не был активен.", sessionID)
		return nil
	}

	poll.ticker.Stop()
	poll.done <- true
	close(poll.done)
	delete(s.activePolls, sessionID)

	return nil
}

func (s *PollingManager) StopPollingForMachine(sessionID string) error {
	s.pollsMutex.Lock()
	defer s.pollsMutex.Unlock()
	poll, exists := s.activePolls[sessionID]
	if !exists {
		return nil
	}
	poll.ticker.Stop()
	poll.done <- true
	close(poll.done)
	delete(s.activePolls, sessionID)
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

func (s *PollingManager) LoadMetadataForEndpoint(endpointURL string) error {
	if err := s.fetchAndParseProbe(endpointURL); err != nil {
		log.Printf("ПРЕДУПРЕЖДЕНИЕ: %v. Некоторые данные могут быть не распознаны.", err)
		return err
	}
	log.Printf("Загружено %d уникальных DataItem'ов.", len(s.deviceMetadataStore))
	log.Printf("Загружено %d ссылок на DataItem'ы осей.", len(s.axisDataItemLinks))
	log.Printf("Загружено %d ссылок на DataItem'ы шпинделей.", len(s.spindleDataItemLinks))
	return nil
}

// processSingleEndpoint получает данные, парсит их и отправляет в Kafka, не сохраняя в памяти.
func (s *PollingManager) processSingleEndpoint(endpointURL string, targetMachineID string) {
	xmlData, err := client.FetchXML(endpointURL)
	if err != nil {
		log.Printf("ОШИБКА при получении XML с %s: %v\n", endpointURL, err)
		return
	}

	var streams models.MTConnectStreams
	if err := xml.Unmarshal(xmlData, &streams); err != nil {
		log.Printf("ОШИБКА при парсинге XML с %s: %v\n", endpointURL, err)
		return
	}

	s.metadataMutex.RLock()
	machineDataSlice := parser.MapToMachineData(&streams, s.deviceMetadataStore, s.axisDataItemLinks, s.spindleDataItemLinks)
	s.metadataMutex.RUnlock()

	for _, machineData := range machineDataSlice {
		if machineData.MachineId == targetMachineID {
			// Данные больше не сохраняются в локальном хранилище (s.repo.Set).

			jsonData, err := json.Marshal(machineData)
			if err != nil {
				log.Printf("ОШИБКА: не удалось сериализовать MachineData для Kafka: %v", err)
				continue
			}
			err = s.producer.Produce(context.Background(), []byte(machineData.MachineId), jsonData)
			if err != nil {
				log.Printf("ОШИБКА: не удалось отправить данные в Kafka для станка %s: %v", machineData.MachineId, err)
			}
			break // Нашли нужный станок, выходим из цикла
		}
	}
}

func (s *PollingManager) fetchAndParseProbe(endpointURL string) error {
	probeURL := strings.TrimSuffix(endpointURL, "/") + "/probe"
	log.Printf("Загрузка метаданных с %s", probeURL)

	xmlData, err := client.FetchXML(probeURL)
	if err != nil {
		return fmt.Errorf("не удалось получить /probe с %s: %w", probeURL, err)
	}

	var devices models.MTConnectDevices
	if err := xml.Unmarshal(xmlData, &devices); err != nil {
		return fmt.Errorf("не удалось распарсить /probe XML с %s: %w", probeURL, err)
	}

	for _, device := range devices.Devices {
		deviceId := device.Name
		if deviceId == "" {
			deviceId = device.UUID
		}
		s.metadataMutex.Lock()
		for _, item := range device.DataItems {
			s.deviceMetadataStore[strings.ToLower(item.ID)] = models.DataItemMetadata{
				ID: item.ID, Name: item.Name, ComponentId: device.ID, ComponentName: device.Name,
				ComponentType: "Device", Category: item.Category, Type: item.Type, SubType: item.SubType,
			}
		}
		s.metadataMutex.Unlock()
		if device.ComponentList != nil {
			s.extractComponentMetadata(device.ComponentList.Components, deviceId)
		}
	}
	return nil
}

func (s *PollingManager) extractComponentMetadata(components []models.ProbeComponent, deviceId string) {
	for _, comp := range components {
		componentType := strings.ToUpper(comp.XMLName.Local)
		isAxisOrSpindle := componentType == "LINEAR" || componentType == "ROTARY"

		s.metadataMutex.Lock()
		for _, item := range comp.DataItems {
			lowerId := strings.ToLower(item.ID)
			s.deviceMetadataStore[lowerId] = models.DataItemMetadata{
				ID: item.ID, Name: item.Name, ComponentId: comp.ID, ComponentName: comp.Name,
				ComponentType: strings.ToLower(comp.XMLName.Local), Category: item.Category, Type: item.Type, SubType: item.SubType,
			}

			if isAxisOrSpindle && item.Type != "" && item.Type != "AXIS_STATE" {
				dataKey := strings.ToLower(item.Type)
				switch componentType {
				case "LINEAR":
					s.axisDataItemLinks[lowerId] = models.AxisDataItemLink{
						DeviceID: deviceId, AxisComponentID: comp.ID, AxisName: comp.Name, AxisType: componentType, DataKey: dataKey,
					}
				case "ROTARY":
					s.spindleDataItemLinks[lowerId] = models.SpindleDataItemLink{
						DeviceID: deviceId, SpindleComponentID: comp.ID, SpindleName: comp.Name, SpindleType: componentType, DataKey: dataKey,
					}
				}
			}
		}
		s.metadataMutex.Unlock()

		if comp.ComponentList != nil {
			s.extractComponentMetadata(comp.ComponentList.Components, deviceId)
		}
	}
}

func (s *PollingManager) startPollingForMachineUnsafe(sessionID, endpointURL, machineID string, interval time.Duration) {
	if _, exists := s.activePolls[sessionID]; exists {
		log.Printf("Опрос для сессии '%s' уже запущен, пропускаем.", sessionID)
		return
	}

	ticker := time.NewTicker(interval)
	done := make(chan bool)

	s.activePolls[sessionID] = &activePoll{
		ticker: ticker,
		done:   done,
	}

	go func() {
		log.Printf("Запуск опроса для сессии '%s' (станок: %s) с интервалом %v", sessionID, machineID, interval)
		currentURL := strings.TrimSuffix(endpointURL, "/") + "/current"
		for {
			select {
			case <-done:
				log.Printf("Остановлен опрос для сессии '%s'", sessionID)
				return
			case <-ticker.C:
				s.processSingleEndpoint(currentURL, machineID)
			}
		}
	}()
}
