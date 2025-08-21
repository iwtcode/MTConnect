package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"MTConnect/config"
	"MTConnect/internal/mtconnect"

	"github.com/gin-gonic/gin"
)

// Потокобезопасное хранилище для данных станков
type DataStore struct {
	mu   sync.RWMutex
	data map[string]mtconnect.MachineData
}

func (ds *DataStore) set(machineId string, data mtconnect.MachineData) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.data[machineId] = data
}

func (ds *DataStore) get(machineId string) (mtconnect.MachineData, bool) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	machineData, found := ds.data[machineId]
	return machineData, found
}

// Глобальное хранилище метаданных: ключ - dataItemId в нижнем регистре, значение - метаданные
var deviceMetadataStore = make(map[string]mtconnect.DataItemMetadata)
var metadataMutex = &sync.RWMutex{}

// Глобальное хранилище связей: ключ - dataItemId, значение - информация о связи с осью
var axisDataItemLinks = make(map[string]mtconnect.AxisDataItemLink)
var axisLinksMutex = &sync.RWMutex{}

// Глобальное хранилище связей шпинделей: ключ - dataItemId, значение - информация о связи со шпинделем
var spindleDataItemLinks = make(map[string]mtconnect.SpindleDataItemLink)
var spindleLinksMutex = &sync.RWMutex{}

// Рекурсивная функция для извлечения метаданных из всех компонентов
func extractComponentMetadata(components []mtconnect.ProbeComponent, deviceId string) {
	for _, comp := range components {
		componentType := strings.ToUpper(comp.XMLName.Local)
		isAxisOrSpindle := componentType == "LINEAR" || componentType == "ROTARY"

		// Обрабатываем DataItem'ы внутри текущего компонента
		for _, item := range comp.DataItems {
			lowerId := strings.ToLower(item.ID)

			// 1. Сохраняем общие метаданные
			metadataMutex.Lock()
			deviceMetadataStore[lowerId] = mtconnect.DataItemMetadata{
				ID:            item.ID,
				Name:          item.Name,
				ComponentId:   comp.ID,
				ComponentName: comp.Name,
				ComponentType: strings.ToLower(comp.XMLName.Local),
				Category:      item.Category,
				Type:          item.Type,
				SubType:       item.SubType,
			}
			metadataMutex.Unlock()

			// 2. Если компонент является осью (LINEAR) или шпинделем (ROTARY), создаем соответствующую связь
			if isAxisOrSpindle && item.Type != "" && item.Type != "AXIS_STATE" {
				dataKey := strings.ToLower(item.Type)

				switch componentType {
				case "LINEAR":
					link := mtconnect.AxisDataItemLink{
						DeviceID:        deviceId,
						AxisComponentID: comp.ID,
						AxisName:        comp.Name,
						AxisType:        componentType,
						DataKey:         dataKey,
					}
					axisLinksMutex.Lock()
					axisDataItemLinks[lowerId] = link
					axisLinksMutex.Unlock()
				case "ROTARY":
					link := mtconnect.SpindleDataItemLink{
						DeviceID:           deviceId,
						SpindleComponentID: comp.ID,
						SpindleName:        comp.Name,
						SpindleType:        componentType,
						DataKey:            dataKey,
					}
					spindleLinksMutex.Lock()
					spindleDataItemLinks[lowerId] = link
					spindleLinksMutex.Unlock()
				}
			}
		}

		// Рекурсивно обходим дочерние компоненты
		if comp.ComponentList != nil {
			extractComponentMetadata(comp.ComponentList.Components, deviceId)
		}
	}
}

// Загружает и парсит /probe ответ для одного эндпоинта
func fetchAndParseProbe(endpointURL string) error {
	probeURL := strings.TrimSuffix(endpointURL, "/") + "/probe"
	log.Printf("Загрузка метаданных с %s", probeURL)

	xmlData, err := mtconnect.FetchXML(probeURL)
	if err != nil {
		return fmt.Errorf("не удалось получить /probe с %s: %w", probeURL, err)
	}

	var devices mtconnect.MTConnectDevices
	if err := xml.Unmarshal(xmlData, &devices); err != nil {
		return fmt.Errorf("не удалось распарсить /probe XML с %s: %w", probeURL, err)
	}

	for _, device := range devices.Devices {
		deviceId := device.Name
		if deviceId == "" {
			deviceId = device.UUID
		}

		// Обрабатываем DataItem'ы на верхнем уровне устройства
		for _, item := range device.DataItems {
			metadataMutex.Lock()
			deviceMetadataStore[strings.ToLower(item.ID)] = mtconnect.DataItemMetadata{
				ID:            item.ID,
				Name:          item.Name,
				ComponentId:   device.ID,
				ComponentName: device.Name,
				ComponentType: "Device",
				Category:      item.Category,
				Type:          item.Type,
				SubType:       item.SubType,
			}
			metadataMutex.Unlock()
		}
		// Запускаем рекурсивный обход для всех компонентов
		if device.ComponentList != nil {
			extractComponentMetadata(device.ComponentList.Components, deviceId)
		}
	}
	return nil
}

func main() {
	cfg, err := config.LoadConfiguration("config/config.json")
	if err != nil {
		log.Fatalf("Не удалось загрузить конфигурацию: %v", err)
	}

	// 1. ЗАГРУЗКА МЕТАДАННЫХ ПЕРЕД ЗАПУСКОМ
	for _, endpoint := range cfg.Endpoints {
		if err := fetchAndParseProbe(endpoint); err != nil {
			log.Printf("ПРЕДУПРЕЖДЕНИЕ: %v. Некоторые данные могут быть не распознаны.", err)
		}
	}
	log.Printf("Загружено %d уникальных DataItem'ов из всех /probe эндпоинтов.", len(deviceMetadataStore))
	log.Printf("Загружено %d ссылок на DataItem'ы осей.", len(axisDataItemLinks))
	log.Printf("Загружено %d ссылок на DataItem'ы шпинделей.", len(spindleDataItemLinks))

	store := &DataStore{
		data: make(map[string]mtconnect.MachineData),
	}

	go pollEndpoints(cfg, store)

	router := gin.Default()
	router.GET("/api/:machineId/current", func(c *gin.Context) {
		machineId := c.Param("machineId")
		machineData, found := store.get(machineId)
		if !found {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Данные для станка '%s' не найдены.", machineId)})
			return
		}
		c.JSON(http.StatusOK, machineData)
	})

	serverAddr := ":" + cfg.ServerPort
	log.Printf("Сервер запущен на http://localhost%s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}

func pollEndpoints(cfg *config.AppConfig, store *DataStore) {
	for _, endpointURL := range cfg.Endpoints {
		currentURL := strings.TrimSuffix(endpointURL, "/") + "/current"
		processSingleEndpoint(currentURL, store)
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		for _, endpointURL := range cfg.Endpoints {
			currentURL := strings.TrimSuffix(endpointURL, "/") + "/current"
			processSingleEndpoint(currentURL, store)
		}
	}
}

func processSingleEndpoint(endpointURL string, store *DataStore) {
	xmlData, err := mtconnect.FetchXML(endpointURL)
	if err != nil {
		log.Printf("ОШИБКА при получении XML с %s: %v\n", endpointURL, err)
		return
	}

	var streams mtconnect.MTConnectStreams
	if err := xml.Unmarshal(xmlData, &streams); err != nil {
		log.Printf("ОШИБКА при парсинге XML с %s: %v\n", endpointURL, err)
		return
	}

	// Передаем метаданные и связи осей/шпинделей в маппер
	metadataMutex.RLock()
	axisLinksMutex.RLock()
	spindleLinksMutex.RLock()
	machineDataSlice := mtconnect.MapToMachineData(&streams, deviceMetadataStore, axisDataItemLinks, spindleDataItemLinks)
	spindleLinksMutex.RUnlock()
	axisLinksMutex.RUnlock()
	metadataMutex.RUnlock()

	for _, machineData := range machineDataSlice {
		if machineData.MachineId != "" {
			store.set(machineData.MachineId, machineData)
		}
	}
}
