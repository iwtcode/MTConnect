package services

import (
	"MTConnect/internal/domain/entities"
	"MTConnect/internal/interfaces"
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

type PollingService struct {
	repo                 interfaces.DataStoreRepository
	producer             interfaces.DataProducer
	activePolls          map[string]*activePoll
	pollsMutex           sync.Mutex
	deviceMetadataStore  map[string]entities.DataItemMetadata
	axisDataItemLinks    map[string]entities.AxisDataItemLink
	spindleDataItemLinks map[string]entities.SpindleDataItemLink
	metadataMutex        sync.RWMutex
	axisLinksMutex       sync.RWMutex
	spindleLinksMutex    sync.RWMutex

	// --- НОВЫЕ ПОЛЯ ДЛЯ ХРАНЕНИЯ СОСТОЯНИЯ ---
	isPollingActive bool
	pollingInterval time.Duration
}

func NewPollingService(repo interfaces.DataStoreRepository, producer interfaces.DataProducer) interfaces.PollingService {
	ps := &PollingService{
		repo:                 repo,
		producer:             producer,
		activePolls:          make(map[string]*activePoll),
		deviceMetadataStore:  make(map[string]entities.DataItemMetadata),
		axisDataItemLinks:    make(map[string]entities.AxisDataItemLink),
		spindleDataItemLinks: make(map[string]entities.SpindleDataItemLink),
		isPollingActive:      false, // Изначально опрос выключен
	}
	return ps
}

// StartPollingForNewConnectionIfNeeded проверяет, активен ли опрос, и если да, запускает его для нового подключения.
func (s *PollingService) StartPollingForNewConnectionIfNeeded(conn *entities.ConnectionInfo) error {
	s.pollsMutex.Lock()
	defer s.pollsMutex.Unlock()

	if s.isPollingActive {
		log.Printf("Глобальный опрос активен. Запускаем polling для новой сессии: %s", conn.SessionID)
		// Используем уже сохраненный интервал
		return s.startPollingForMachineUnsafe(conn, s.pollingInterval)
	}
	return nil
}

// startPollingForMachineUnsafe - внутренняя версия без блокировки мьютекса
func (s *PollingService) startPollingForMachineUnsafe(conn *entities.ConnectionInfo, interval time.Duration) error {
	if _, exists := s.activePolls[conn.SessionID]; exists {
		return fmt.Errorf("опрос для сессии '%s' уже запущен", conn.SessionID)
	}

	ticker := time.NewTicker(interval)
	done := make(chan bool)

	s.activePolls[conn.SessionID] = &activePoll{
		ticker: ticker,
		done:   done,
	}

	go func() {
		log.Printf("Запуск опроса для сессии '%s' (станок: %s) с интервалом %v", conn.SessionID, conn.MachineID, interval)
		currentURL := strings.TrimSuffix(conn.Config.EndpointURL, "/") + "/current"
		for {
			select {
			case <-done:
				log.Printf("Остановлен опрос для сессии '%s'", conn.SessionID)
				return
			case <-ticker.C:
				s.processSingleEndpoint(currentURL, conn.MachineID)
			}
		}
	}()
	return nil
}

func (s *PollingService) StartPollingForMachine(conn *entities.ConnectionInfo, interval time.Duration) error {
	s.pollsMutex.Lock()
	defer s.pollsMutex.Unlock()
	return s.startPollingForMachineUnsafe(conn, interval)
}

func (s *PollingService) StopPollingForMachine(sessionID string) error {
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

func (s *PollingService) StartAllPolling(connections []*entities.ConnectionInfo, interval time.Duration) error {
	s.pollsMutex.Lock()
	defer s.pollsMutex.Unlock()

	log.Println("Запуск опроса для всех активных подключений...")
	// Сохраняем состояние
	s.isPollingActive = true
	s.pollingInterval = interval

	var errs []string
	for _, conn := range connections {
		if conn.IsHealthy {
			if err := s.startPollingForMachineUnsafe(conn, interval); err != nil {
				errs = append(errs, err.Error())
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("возникли ошибки при запуске опроса: %s", strings.Join(errs, "; "))
	}
	return nil
}

func (s *PollingService) StopAllPolling() {
	s.pollsMutex.Lock()
	defer s.pollsMutex.Unlock()

	log.Println("Остановка всех процессов опроса...")
	// Сбрасываем состояние
	s.isPollingActive = false

	for sessionID, poll := range s.activePolls {
		poll.ticker.Stop()
		poll.done <- true
		close(poll.done)
		delete(s.activePolls, sessionID)
	}
	log.Println("Все процессы опроса остановлены.")
}

// ... Остальные функции (CheckMachineConnection, LoadMetadataForEndpoint, processSingleEndpoint, и т.д.) остаются без изменений ...
// (Код остальных функций для краткости опущен, так как он не менялся)
func (s *PollingService) CheckMachineConnection(endpointURL string) error {
	probeURL := strings.TrimSuffix(endpointURL, "/") + "/probe"
	_, err := FetchXML(probeURL)
	if err != nil {
		return fmt.Errorf("проверка соединения с эндпоинтом '%s' провалена: %w", endpointURL, err)
	}
	return nil
}

func (s *PollingService) LoadMetadataForEndpoint(endpointURL string) error {
	if err := s.fetchAndParseProbe(endpointURL); err != nil {
		log.Printf("ПРЕДУПРЕЖДЕНИЕ: %v. Некоторые данные могут быть не распознаны.", err)
		return err
	}
	s.metadataMutex.RLock()
	defer s.metadataMutex.RUnlock()
	s.axisLinksMutex.RLock()
	defer s.axisLinksMutex.RUnlock()
	s.spindleLinksMutex.RLock()
	defer s.spindleLinksMutex.RUnlock()

	log.Printf("Загружено %d уникальных DataItem'ов.", len(s.deviceMetadataStore))
	log.Printf("Загружено %d ссылок на DataItem'ы осей.", len(s.axisDataItemLinks))
	log.Printf("Загружено %d ссылок на DataItem'ы шпинделей.", len(s.spindleDataItemLinks))
	return nil
}

func (s *PollingService) processSingleEndpoint(endpointURL string, targetMachineID string) {
	xmlData, err := FetchXML(endpointURL)
	if err != nil {
		log.Printf("ОШИБКА при получении XML с %s: %v\n", endpointURL, err)
		return
	}

	var streams entities.MTConnectStreams
	if err := xml.Unmarshal(xmlData, &streams); err != nil {
		log.Printf("ОШИБКА при парсинге XML с %s: %v\n", endpointURL, err)
		return
	}

	s.metadataMutex.RLock()
	s.axisLinksMutex.RLock()
	s.spindleLinksMutex.RLock()
	machineDataSlice := MapToMachineData(&streams, s.deviceMetadataStore, s.axisDataItemLinks, s.spindleDataItemLinks)
	s.spindleLinksMutex.RUnlock()
	s.axisLinksMutex.RUnlock()
	s.metadataMutex.RUnlock()

	for _, machineData := range machineDataSlice {
		if machineData.MachineId == targetMachineID {
			s.repo.Set(machineData.MachineId, machineData)

			jsonData, err := json.Marshal(machineData)
			if err != nil {
				log.Printf("ОШИБКА: не удалось сериализовать MachineData для Kafka: %v", err)
				continue
			}
			err = s.producer.Produce(context.Background(), []byte(machineData.MachineId), jsonData)
			if err != nil {
				log.Printf("ОШИБКА: не удалось отправить данные в Kafka для станка %s: %v", machineData.MachineId, err)
			}
			break
		}
	}
}

func (s *PollingService) fetchAndParseProbe(endpointURL string) error {
	probeURL := strings.TrimSuffix(endpointURL, "/") + "/probe"
	log.Printf("Загрузка метаданных с %s", probeURL)

	xmlData, err := FetchXML(probeURL)
	if err != nil {
		return fmt.Errorf("не удалось получить /probe с %s: %w", probeURL, err)
	}

	var devices entities.MTConnectDevices
	if err := xml.Unmarshal(xmlData, &devices); err != nil {
		return fmt.Errorf("не удалось распарсить /probe XML с %s: %w", probeURL, err)
	}

	for _, device := range devices.Devices {
		deviceId := device.Name
		if deviceId == "" {
			deviceId = device.UUID
		}
		for _, item := range device.DataItems {
			s.metadataMutex.Lock()
			s.deviceMetadataStore[strings.ToLower(item.ID)] = entities.DataItemMetadata{
				ID: item.ID, Name: item.Name, ComponentId: device.ID, ComponentName: device.Name,
				ComponentType: "Device", Category: item.Category, Type: item.Type, SubType: item.SubType,
			}
			s.metadataMutex.Unlock()
		}
		if device.ComponentList != nil {
			s.extractComponentMetadata(device.ComponentList.Components, deviceId)
		}
	}
	return nil
}

func (s *PollingService) extractComponentMetadata(components []entities.ProbeComponent, deviceId string) {
	for _, comp := range components {
		componentType := strings.ToUpper(comp.XMLName.Local)
		isAxisOrSpindle := componentType == "LINEAR" || componentType == "ROTARY"

		for _, item := range comp.DataItems {
			lowerId := strings.ToLower(item.ID)
			s.metadataMutex.Lock()
			s.deviceMetadataStore[lowerId] = entities.DataItemMetadata{
				ID: item.ID, Name: item.Name, ComponentId: comp.ID, ComponentName: comp.Name,
				ComponentType: strings.ToLower(comp.XMLName.Local), Category: item.Category, Type: item.Type, SubType: item.SubType,
			}
			s.metadataMutex.Unlock()

			if isAxisOrSpindle && item.Type != "" && item.Type != "AXIS_STATE" {
				dataKey := strings.ToLower(item.Type)
				switch componentType {
				case "LINEAR":
					s.axisLinksMutex.Lock()
					s.axisDataItemLinks[lowerId] = entities.AxisDataItemLink{
						DeviceID: deviceId, AxisComponentID: comp.ID, AxisName: comp.Name, AxisType: componentType, DataKey: dataKey,
					}
					s.axisLinksMutex.Unlock()
				case "ROTARY":
					s.spindleLinksMutex.Lock()
					s.spindleDataItemLinks[lowerId] = entities.SpindleDataItemLink{
						DeviceID: deviceId, SpindleComponentID: comp.ID, SpindleName: comp.Name, SpindleType: componentType, DataKey: dataKey,
					}
					s.spindleLinksMutex.Unlock()
				}
			}
		}
		if comp.ComponentList != nil {
			s.extractComponentMetadata(comp.ComponentList.Components, deviceId)
		}
	}
}
