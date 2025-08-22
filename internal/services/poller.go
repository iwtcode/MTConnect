package services

import (
	"MTConnect/internal/config"
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

// activePoll хранит информацию о запущенном процессе опроса для одного станка
type activePoll struct {
	ticker *time.Ticker
	done   chan bool
}

type PollingService struct {
	cfg                  *config.AppConfig
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
}

func NewPollingService(cfg *config.AppConfig, repo interfaces.DataStoreRepository, producer interfaces.DataProducer) interfaces.PollingService {
	ps := &PollingService{
		cfg:                  cfg,
		repo:                 repo,
		producer:             producer,
		activePolls:          make(map[string]*activePoll),
		deviceMetadataStore:  make(map[string]entities.DataItemMetadata),
		axisDataItemLinks:    make(map[string]entities.AxisDataItemLink),
		spindleDataItemLinks: make(map[string]entities.SpindleDataItemLink),
	}
	ps.loadInitialMetadata() // Загружаем метаданные при создании сервиса
	return ps
}

// findEndpointByMachineId ищет URL эндпоинта по ID станка
func (s *PollingService) findEndpointByMachineId(machineId string) (string, error) {
	for _, endpoint := range s.cfg.Endpoints {
		if strings.HasSuffix(strings.ToLower(endpoint), strings.ToLower(machineId)) {
			return endpoint, nil
		}
	}
	return "", fmt.Errorf("эндпоинт для станка '%s' не найден в конфигурации", machineId)
}

func (s *PollingService) StartPollingForMachine(machineId string, interval time.Duration) error {
	s.pollsMutex.Lock()
	defer s.pollsMutex.Unlock()

	if _, exists := s.activePolls[machineId]; exists {
		return fmt.Errorf("опрос для станка '%s' уже запущен", machineId)
	}

	endpointURL, err := s.findEndpointByMachineId(machineId)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	done := make(chan bool)

	s.activePolls[machineId] = &activePoll{
		ticker: ticker,
		done:   done,
	}

	go func() {
		log.Printf("Запуск опроса для '%s' с интервалом %v", machineId, interval)
		currentURL := strings.TrimSuffix(endpointURL, "/") + "/current"
		for {
			select {
			case <-done:
				log.Printf("Остановлен опрос для '%s'", machineId)
				return
			case <-ticker.C:
				s.processSingleEndpoint(currentURL)
			}
		}
	}()

	return nil
}

func (s *PollingService) StopPollingForMachine(machineId string) error {
	s.pollsMutex.Lock()
	defer s.pollsMutex.Unlock()

	poll, exists := s.activePolls[machineId]
	if !exists {
		return fmt.Errorf("опрос для станка '%s' не был запущен", machineId)
	}

	poll.ticker.Stop()
	poll.done <- true
	close(poll.done)
	delete(s.activePolls, machineId)

	return nil
}

func (s *PollingService) StopAllPolling() {
	s.pollsMutex.Lock()
	defer s.pollsMutex.Unlock()
	log.Println("Остановка всех процессов опроса...")
	for machineId, poll := range s.activePolls {
		poll.ticker.Stop()
		poll.done <- true
		close(poll.done)
		delete(s.activePolls, machineId)
	}
	log.Println("Все процессы опроса остановлены.")
}

func (s *PollingService) CheckConnection(machineId string) error {
	endpointURL, err := s.findEndpointByMachineId(machineId)
	if err != nil {
		return err
	}
	probeURL := strings.TrimSuffix(endpointURL, "/") + "/probe"
	_, err = FetchXML(probeURL)
	if err != nil {
		return fmt.Errorf("проверка соединения со станком '%s' провалена: %w", machineId, err)
	}
	return nil
}

func (s *PollingService) processSingleEndpoint(endpointURL string) {
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
		if machineData.MachineId != "" {
			// 1. Сохраняем в локальное хранилище
			s.repo.Set(machineData.MachineId, machineData)

			// 2. Отправляем в Kafka
			jsonData, err := json.Marshal(machineData)
			if err != nil {
				log.Printf("ОШИБКА: не удалось сериализовать MachineData для Kafka: %v", err)
				continue
			}
			err = s.producer.Produce(context.Background(), []byte(machineData.MachineId), jsonData)
			if err != nil {
				log.Printf("ОШИБКА: не удалось отправить данные в Kafka для станка %s: %v", machineData.MachineId, err)
			}
		}
	}
}

// ... (остальные функции loadInitialMetadata, fetchAndParseProbe, extractComponentMetadata остаются без изменений) ...
func (s *PollingService) loadInitialMetadata() {
	for _, endpoint := range s.cfg.Endpoints {
		if err := s.fetchAndParseProbe(endpoint); err != nil {
			log.Printf("ПРЕДУПРЕЖДЕНИЕ: %v. Некоторые данные могут быть не распознаны.", err)
		}
	}
	s.metadataMutex.RLock()
	s.axisLinksMutex.RLock()
	s.spindleLinksMutex.RLock()
	log.Printf("Загружено %d уникальных DataItem'ов.", len(s.deviceMetadataStore))
	log.Printf("Загружено %d ссылок на DataItem'ы осей.", len(s.axisDataItemLinks))
	log.Printf("Загружено %d ссылок на DataItem'ы шпинделей.", len(s.spindleDataItemLinks))
	s.spindleLinksMutex.RUnlock()
	s.axisLinksMutex.RUnlock()
	s.metadataMutex.RUnlock()
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
