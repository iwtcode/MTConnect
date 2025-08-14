package mtconnect

import "time"

// Целевая структура для JSON
type MachineData struct {
	MachineId      string `json:"MachineId"`
	Id             string `json:"Id"`
	Timestamp      string `json:"Timestamp"`
	Execution      string `json:"MachineState"`
	ControllerMode string `json:"ProgramMode"`
	PartCount      string `json:"PartsCount"`
	PowerOnTime    string `json:"PowerOnTime"`
	// ... добавить остальные поля
}

// Преобразует MTConnect XML в целевую структуру
func MapToMachineData(streams *MTConnectStreams) []MachineData {
	var results []MachineData

	for _, deviceStream := range streams.Streams.DeviceStreams {
		dataItemsMap := make(map[string]string)

		for _, compStream := range deviceStream.ComponentStream {
			if compStream.Events != nil {
				for _, item := range compStream.Events.Items {
					// Ключом может быть ID (предпочтительно) или имя тега (XMLName.Local)
					key := item.DataItemID
					if key == "" {
						key = item.XMLName.Local
					}
					dataItemsMap[key] = item.Value
				}
			}
			if compStream.Samples != nil {
				for _, item := range compStream.Samples.Items {
					key := item.DataItemID
					if key == "" {
						key = item.XMLName.Local
					}
					dataItemsMap[key] = item.Value
				}
			}
			if compStream.Condition != nil {
				for _, item := range compStream.Condition.Items {
					key := item.DataItemID
					if key == "" {
						key = item.XMLName.Local
					}
					dataItemsMap[key] = item.Value
				}
			}
		}

		// Заполняем целевую структуру
		result := MachineData{
			MachineId: deviceStream.Name,
			Id:        deviceStream.UUID,
			Timestamp: time.Now().UTC().Format(time.RFC3339),

			// Ищем нужные значения в нашей карте
			// Ключи - dataItemId из Devices.xml
			Execution:      dataItemsMap["execution"],            // Mazak
			ControllerMode: dataItemsMap["mode"],                 // Mazak
			PartCount:      dataItemsMap["PartCountAct"],         // Mazak
			PowerOnTime:    dataItemsMap["LpTotalOperatingTime"], // OKUMA
		}

		// Добавляем логику для OKUMA, если поля для Mazak не нашлись
		if result.Execution == "" {
			result.Execution = dataItemsMap["Lpexecution"]
		}
		if result.ControllerMode == "" {
			result.ControllerMode = dataItemsMap["Lpmode"]
		}
		if result.PartCount == "" {
			result.PartCount = dataItemsMap["Lppartcount"]
		}

		results = append(results, result)
	}

	return results
}
