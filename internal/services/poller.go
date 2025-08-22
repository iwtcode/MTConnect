package services

import (
	"encoding/xml"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"MTConnect/internal/config"
	"MTConnect/internal/domain/entities"
	"MTConnect/internal/interfaces"
)

type PollingService struct {
	cfg                  *config.AppConfig
	repo                 interfaces.DataStoreRepository
	ticker               *time.Ticker
	done                 chan bool
	deviceMetadataStore  map[string]entities.DataItemMetadata
	axisDataItemLinks    map[string]entities.AxisDataItemLink
	spindleDataItemLinks map[string]entities.SpindleDataItemLink
	metadataMutex        sync.RWMutex
	axisLinksMutex       sync.RWMutex
	spindleLinksMutex    sync.RWMutex
}

func NewPollingService(cfg *config.AppConfig, repo interfaces.DataStoreRepository) interfaces.PollingService {
	return &PollingService{
		cfg:                  cfg,
		repo:                 repo,
		done:                 make(chan bool),
		deviceMetadataStore:  make(map[string]entities.DataItemMetadata),
		axisDataItemLinks:    make(map[string]entities.AxisDataItemLink),
		spindleDataItemLinks: make(map[string]entities.SpindleDataItemLink),
	}
}

func (s *PollingService) StartPolling() {
	s.loadInitialMetadata()
	s.pollAllEndpoints()

	s.ticker = time.NewTicker(1 * time.Second)
	for {
		select {
		case <-s.done:
			return
		case <-s.ticker.C:
			s.pollAllEndpoints()
		}
	}
}

func (s *PollingService) StopPolling() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.done <- true
}

func (s *PollingService) pollAllEndpoints() {
	for _, endpointURL := range s.cfg.Endpoints {
		currentURL := strings.TrimSuffix(endpointURL, "/") + "/current"
		s.processSingleEndpoint(currentURL)
	}
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
			s.repo.Set(machineData.MachineId, machineData)
		}
	}
}

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
