package mtconnect

import (
	"strings"
)

// Преобразует необработанные данные MTConnectStreams в срез MachineData.
func MapToMachineData(streams *MTConnectStreams, metadata map[string]DataItemMetadata) []MachineData {
	machineDataMap := make(map[string]*MachineData)

	for _, deviceStream := range streams.Streams {
		machineID := deviceStream.Name
		if machineID == "" {
			machineID = deviceStream.UUID
		}

		if _, ok := machineDataMap[machineID]; !ok {
			machineDataMap[machineID] = &MachineData{
				MachineId:           machineID,
				Id:                  deviceStream.UUID,
				IsEnabled:           "UNAVAILABLE",
				IsInEmergency:       "UNAVAILABLE",
				MachineState:        "UNAVAILABLE",
				ProgramMode:         "UNAVAILABLE",
				TmMode:              "UNAVAILABLE",
				AxisMovementStatus:  "UNAVAILABLE",
				MstbStatus:          "UNAVAILABLE",
				EmergencyStatus:     "UNAVAILABLE",
				AlarmStatus:         "UNAVAILABLE",
				Alarms:              "UNAVAILABLE",
				HasAlarms:           "UNAVAILABLE",
				EditStatus:          "UNAVAILABLE",
				HandleRetraceStatus: "UNAVAILABLE",
			}
		}
		machine := machineDataMap[machineID]

		conditionsProcessedThisCycle := false

		for _, compStream := range deviceStream.ComponentStreams {
			// Обработка Samples
			if compStream.Samples != nil {
				for _, sample := range compStream.Samples.Items {
					processDataItem(machine, sample.DataItemId, sample.Value, sample.Timestamp, metadata)
				}
			}

			// Обработка Events
			if compStream.Events != nil {
				for _, event := range compStream.Events.Items {
					processDataItem(machine, event.DataItemId, event.Value, event.Timestamp, metadata)
				}
			}

			// Обработка Conditions для получения статуса тревог
			if compStream.Condition != nil && len(compStream.Condition.Items) > 0 {
				if !conditionsProcessedThisCycle {
					machine.AlarmStatus = "NORMAL"
					machine.Alarms = make([]AlarmDetail, 0)
					conditionsProcessedThisCycle = true
				}

				for _, condition := range compStream.Condition.Items {
					status := strings.ToUpper(condition.XMLName.Local)

					if status == "FAULT" || status == "WARNING" {
						message := condition.Value
						if message == "" {
							message = "No details provided"
						}

						// Значения по умолчанию, если метаданные не найдены
						componentName := compStream.Name
						alarmType := condition.Type

						// Ищем метаданные по dataItemId тревоги
						meta, ok := metadata[strings.ToLower(condition.DataItemId)]
						if ok {
							componentName = meta.ComponentName
						}

						alarm := AlarmDetail{
							Status:        status,
							Type:          alarmType,
							ComponentName: componentName,
							Message:       message,
							NativeCode:    condition.NativeCode,
						}

						if alarmList, ok := machine.Alarms.([]AlarmDetail); ok {
							machine.Alarms = append(alarmList, alarm)
						}
					}

					if status == "FAULT" {
						machine.AlarmStatus = "FAULT"
					} else if status == "WARNING" && machine.AlarmStatus != "FAULT" {
						machine.AlarmStatus = "WARNING"
					}
				}
			}
		}
	}

	// Преобразуем карту в срез и вычисляем производные значения
	var machineDataSlice []MachineData
	for _, data := range machineDataMap {
		// Устанавливаем флаг HasAlarms на основе итогового статуса
		if data.AlarmStatus != "UNAVAILABLE" {
			data.HasAlarms = (data.AlarmStatus == "FAULT" || data.AlarmStatus == "WARNING")
		}

		// Логика отката для EditStatus на основе ProgramMode
		// Эта проверка выполняется только если настоящее значение EditStatus не было получено
		if data.EditStatus == "UNAVAILABLE" && data.ProgramMode != "UNAVAILABLE" {
			if data.ProgramMode == "EDIT" {
				data.EditStatus = "READY"
			} else {
				data.EditStatus = "NOT_READY"
			}
		}

		machineDataSlice = append(machineDataSlice, *data)
	}

	return machineDataSlice
}

// Логика сопоставления. Перезаписывает значения "UNAVAILABLE", если найдет реальные данные в потоке
func processDataItem(machine *MachineData, dataItemId, value, timestamp string, metadata map[string]DataItemMetadata) {
	meta, ok := metadata[strings.ToLower(dataItemId)]
	if !ok {
		return
	}

	// Обновляем глобальную временную метку станка на самую последнюю
	if machine.Timestamp < timestamp {
		machine.Timestamp = timestamp
	}

	switch meta.Type {
	case "AVAILABILITY":
		machine.IsEnabled = (value == "AVAILABLE")
	case "EMERGENCY_STOP":
		machine.IsInEmergency = (value == "TRIGGERED")
		machine.EmergencyStatus = value
	case "EXECUTION":
		machine.MachineState = value
	case "CONTROLLER_MODE":
		machine.ProgramMode = value
		machine.HandleRetraceStatus = (value == "MANUAL")
	case "T_M_MODE": // ЗАГЛУШКА
		machine.TmMode = value
	case "AXIS_STATE":
		if _, isString := machine.AxisMovementStatus.(string); isString {
			machine.AxisMovementStatus = make(map[string]string)
		}
		if statusMap, ok := machine.AxisMovementStatus.(map[string]string); ok {
			if meta.ComponentName != "" {
				statusMap[meta.ComponentName] = value
			}
		}
	case "MSTB_STATUS": // ЗАГЛУШКА
		machine.MstbStatus = value
	case "PROGRAM_EDIT":
		machine.EditStatus = value
	}
}
