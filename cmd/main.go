package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"

	"MTConnect/config"
	"MTConnect/internal/mtconnect"
)

func main() {
	// 1. Загружаем конфигурацию
	cfg, err := config.LoadConfiguration("config/config.json")
	if err != nil {
		log.Fatalf("Не удалось загрузить конфигурацию: %v", err)
	}

	fmt.Printf("Найдено %d эндпоинтов для опроса.\n", len(cfg.Endpoints))
	fmt.Println("-------------------------------------------------")

	// 2. Проходим по всем эндпоинтам из конфига
	for _, endpointURL := range cfg.Endpoints {
		fmt.Printf("Обработка эндпоинта: %s\n", endpointURL)

		// 3. Получаем данные в XML формате
		xmlData, err := mtconnect.FetchXML(endpointURL)
		if err != nil {
			log.Printf("ОШИБКА при получении XML с %s: %v\n", endpointURL, err)
			fmt.Println("-------------------------------------------------")
			continue
		}

		// 4. Парсим XML в Go структуры
		var streams mtconnect.MTConnectStreams
		if err := xml.Unmarshal(xmlData, &streams); err != nil {
			log.Printf("ОШИБКА при парсинге XML с %s: %v\n", endpointURL, err)
			fmt.Println("-------------------------------------------------")
			continue
		}

		// 5. Преобразуем данные в целевую структуру
		machineDataSlice := mtconnect.MapToMachineData(&streams)

		// 6. Выводим результат в формате JSON
		for _, machineData := range machineDataSlice {
			jsonData, err := json.MarshalIndent(machineData, "", "  ")
			if err != nil {
				log.Printf("ОШИБКА при конвертации в JSON для %s: %v\n", machineData.MachineId, err)
				continue
			}

			fmt.Printf("Итоговый JSON для станка '%s':\n", machineData.MachineId)
			fmt.Println(string(jsonData))
			fmt.Println("-------------------------------------------------")
		}
	}
}
