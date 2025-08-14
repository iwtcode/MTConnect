package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"MTConnect/config"
	"MTConnect/internal/mtconnect"

	"github.com/gin-gonic/gin"
)

// Потокобезопасное хранилище для данных станков
// Ключ - MachineId, значение - последние полученные данные
type DataStore struct {
	mu   sync.RWMutex
	data map[string]mtconnect.MachineData
}

// Сохраняет данные для одного станка
func (ds *DataStore) set(machineId string, data mtconnect.MachineData) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.data[machineId] = data
}

// Получает данные для одного станка
func (ds *DataStore) get(machineId string) (mtconnect.MachineData, bool) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	machineData, found := ds.data[machineId]
	return machineData, found
}

func main() {
	// 1. Загружаем конфигурацию
	cfg, err := config.LoadConfiguration("config/config.json")
	if err != nil {
		log.Fatalf("Не удалось загрузить конфигурацию: %v", err)
	}

	// 2. Инициализируем хранилище данных
	store := &DataStore{
		data: make(map[string]mtconnect.MachineData),
	}

	// 3. Запускаем фоновый процесс для постоянного опроса эндпоинтов
	go pollEndpoints(cfg, store)

	// 4. Настраиваем и запускаем Gin веб-сервер
	router := gin.Default()

	// Обработчик для получения данных по конкретному станку
	router.GET("/api/:machineId/current", func(c *gin.Context) {
		machineId := c.Param("machineId")
		machineData, found := store.get(machineId)

		if !found {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Данные для станка '%s' не найдены.", machineId)})
			return
		}

		c.JSON(http.StatusOK, machineData)
	})

	// Запускаем сервер на порту из конфигурации
	serverAddr := ":" + cfg.ServerPort
	log.Printf("Сервер запущен на http://localhost%s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}

// Функция, которая бесконечно опрашивает эндпоинты с интервалом в 1 секунду.
func pollEndpoints(cfg *config.AppConfig, store *DataStore) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		for _, endpointURL := range cfg.Endpoints {
			// Передаем cfg в processSingleEndpoint
			processSingleEndpoint(endpointURL, store, cfg)
		}
	}
}

// Обрабатывает один эндпоинт: получает, парсит и сохраняет данные.
func processSingleEndpoint(endpointURL string, store *DataStore, cfg *config.AppConfig) {
	// 1. Получаем данные в XML формате
	xmlData, err := mtconnect.FetchXML(endpointURL)
	if err != nil {
		log.Printf("ОШИБКА при получении XML с %s: %v\n", endpointURL, err)
		return
	}

	// 2. Парсим XML в Go структуры
	var streams mtconnect.MTConnectStreams
	if err := xml.Unmarshal(xmlData, &streams); err != nil {
		log.Printf("ОШИБКА при парсинге XML с %s: %v\n", endpointURL, err)
		return
	}

	// 3. Преобразуем данные в целевую структуру
	machineDataSlice := mtconnect.MapToMachineData(&streams)

	// 4. Сохраняем результат в наше хранилище
	for _, machineData := range machineDataSlice {
		if machineData.MachineId != "" {
			store.set(machineData.MachineId, machineData)
			apiURL := fmt.Sprintf("http://localhost:%s/api/%s/current", cfg.ServerPort, machineData.MachineId)
			log.Printf("Данные '%s' обновлены: %s", machineData.MachineId, apiURL)
		}
	}
}
