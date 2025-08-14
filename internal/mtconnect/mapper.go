package mtconnect

import (
	"strings"
	"time"
)

// Целевая структура для JSON
type MachineData struct {
	MachineId      string `json:"MachineId"`
	Id             string `json:"Id"`
	Timestamp      string `json:"Timestamp"`
	Execution      string `json:"MachineState"`
	ControllerMode string `json:"ProgramMode"`
	PartCount      string `json:"PartsCount"`
}

// Преобразует MTConnect XML в целевую структуру
func MapToMachineData(streams *MTConnectStreams) []MachineData {
	var results []MachineData

	for _, deviceStream := range streams.Streams.DeviceStreams {
		// Карта для хранения данных, где ключ - это имя тега в нижнем регистре
		dataItemsMap := make(map[string]string)

		// Функция для обработки любого среза DataItem
		processItems := func(items []DataItem) {
			for _, item := range items {
				// Используем имя тега как ключ. Приводим к нижнему регистру для универсальности.
				key := strings.ToLower(item.XMLName.Local)
				if key != "" {
					dataItemsMap[key] = item.Value
				}
			}
		}

		// Проходим по всем компонентам и собираем данные
		for _, compStream := range deviceStream.ComponentStream {
			if compStream.Events != nil {
				processItems(compStream.Events.Items)
			}
			if compStream.Samples != nil {
				processItems(compStream.Samples.Items)
			}
			if compStream.Condition != nil {
				processItems(compStream.Condition.Items)
			}
		}

		// Заполняем целевую структуру, используя стандартизированные имена тегов
		result := MachineData{
			MachineId: deviceStream.Name,
			Id:        deviceStream.UUID,
			Timestamp: time.Now().UTC().Format(time.RFC3339),

			Execution:      dataItemsMap["execution"],      // из <Execution>
			ControllerMode: dataItemsMap["controllermode"], // из <ControllerMode>
			PartCount:      dataItemsMap["partcount"],      // из <PartCount>
		}

		results = append(results, result)
	}

	return results
}
