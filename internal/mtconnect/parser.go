package mtconnect

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// formatAccumulatedTime преобразует строку с секундами в формат "ЧЧ:ММ:СС".
func formatAccumulatedTime(secondsStr string) string {
	secondsFloat, err := strconv.ParseFloat(secondsStr, 64)
	if err != nil {
		// Если не удается распарсить, возвращаем исходное значение
		return secondsStr
	}

	// Округляем до ближайшей целой секунды
	totalSeconds := int(math.Round(secondsFloat))

	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

// Вспомогательная функция для обработки DataItem'ов, связанных с осями
func processAxisDataItem(machineID string, dataItemId, value string, axisLinks map[string]AxisDataItemLink, axisInfoMap map[string]map[string]*AxisInfo) bool {
	lowerId := strings.ToLower(dataItemId)
	link, ok := axisLinks[lowerId]
	if !ok || link.DeviceID != machineID {
		return false // Это не DataItem оси для текущего станка
	}

	if _, ok := axisInfoMap[machineID]; !ok {
		axisInfoMap[machineID] = make(map[string]*AxisInfo)
	}
	if _, ok := axisInfoMap[machineID][link.AxisComponentID]; !ok {
		axisInfoMap[machineID][link.AxisComponentID] = &AxisInfo{
			ID:   link.AxisComponentID,
			Name: link.AxisName,
			Type: link.AxisType,
			Data: make(map[string]interface{}),
		}
	}

	axis := axisInfoMap[machineID][link.AxisComponentID]
	axis.Data[link.DataKey] = value

	return true
}

// Вспомогательная функция для обработки DataItem'ов, связанных со шпинделями
func processSpindleDataItem(machineID string, dataItemId, value string, spindleLinks map[string]SpindleDataItemLink, spindleInfoMap map[string]map[string]*SpindleInfo) bool {
	lowerId := strings.ToLower(dataItemId)
	link, ok := spindleLinks[lowerId]
	if !ok || link.DeviceID != machineID {
		return false // Это не DataItem шпинделя для текущего станка
	}

	if _, ok := spindleInfoMap[machineID]; !ok {
		spindleInfoMap[machineID] = make(map[string]*SpindleInfo)
	}
	if _, ok := spindleInfoMap[machineID][link.SpindleComponentID]; !ok {
		spindleInfoMap[machineID][link.SpindleComponentID] = &SpindleInfo{
			ID:   link.SpindleComponentID,
			Name: link.SpindleName,
			Type: link.SpindleType,
			Data: make(map[string]interface{}),
		}
	}

	spindle := spindleInfoMap[machineID][link.SpindleComponentID]
	spindle.Data[link.DataKey] = value

	return true
}

// Преобразует необработанные данные MTConnectStreams в срез MachineData.
func MapToMachineData(streams *MTConnectStreams, metadata map[string]DataItemMetadata, axisLinks map[string]AxisDataItemLink, spindleLinks map[string]SpindleDataItemLink) []MachineData {
	machineDataMap := make(map[string]*MachineData)
	axisInfoMap := make(map[string]map[string]*AxisInfo)
	spindleInfoMap := make(map[string]map[string]*SpindleInfo)

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
				HandleRetraceStatus: "UNAVAILABLE",
				AxisMovementStatus:  "UNAVAILABLE",
				MstbStatus:          "UNAVAILABLE",
				EmergencyStatus:     "UNAVAILABLE",
				AlarmStatus:         "UNAVAILABLE",
				WarningStatus:       "UNAVAILABLE",
				EditStatus:          "UNAVAILABLE",
				ManualMode:          "UNAVAILABLE",
				WriteStatus:         "UNAVAILABLE",
				LabelSkipStatus:     "UNAVAILABLE",
				BatteryStatus:       "UNAVAILABLE",
				ActiveToolNumber:    "UNAVAILABLE",
				ToolOffsetNumber:    "UNAVAILABLE",
				AxisInfos:           make([]AxisInfo, 0),
				FeedRate:            make(map[string]string),
				FeedOverride:        make(map[string]string),
				Alarms:              make([]map[string]interface{}, 0),
				HasAlarms:           "UNAVAILABLE",
				PartsCount:          make(map[string]string),
				AccumulatedTime:     make(map[string]string),
				SpindleInfos:        make([]SpindleInfo, 0),
				ContourFeedRate:     "UNAVAILABLE",
				JogOverride:         "UNAVAILABLE",
			}
		}
		machine := machineDataMap[machineID]

		conditionsProcessedThisCycle := false

		for _, compStream := range deviceStream.ComponentStreams {
			// Обработка Samples
			if compStream.Samples != nil {
				for _, sample := range compStream.Samples.Items {
					isAxis := processAxisDataItem(machine.MachineId, sample.DataItemId, sample.Value, axisLinks, axisInfoMap)
					isSpindle := processSpindleDataItem(machine.MachineId, sample.DataItemId, sample.Value, spindleLinks, spindleInfoMap)
					if !isAxis && !isSpindle {
						processDataItem(machine, sample.DataItemId, sample.Value, sample.Timestamp, metadata)
					}
				}
			}

			// Обработка Events
			if compStream.Events != nil {
				for _, event := range compStream.Events.Items {
					isAxis := processAxisDataItem(machine.MachineId, event.DataItemId, event.Value, axisLinks, axisInfoMap)
					isSpindle := processSpindleDataItem(machine.MachineId, event.DataItemId, event.Value, spindleLinks, spindleInfoMap)
					if !isAxis && !isSpindle {
						processDataItem(machine, event.DataItemId, event.Value, event.Timestamp, metadata)
					}
				}
			}

			// Обработка Conditions для получения статуса тревог и предупреждений
			if compStream.Condition != nil && len(compStream.Condition.Items) > 0 {
				if !conditionsProcessedThisCycle {
					machine.AlarmStatus = "NORMAL"
					machine.WarningStatus = "NORMAL"
					machine.Alarms = make([]map[string]interface{}, 0)
					conditionsProcessedThisCycle = true
				}

				for _, condition := range compStream.Condition.Items {
					status := strings.ToUpper(condition.XMLName.Local)

					// Мы собираем только FAULT и WARNING, но не NORMAL
					if status == "FAULT" || status == "WARNING" {
						alarm := make(map[string]interface{})

						alarm["level"] = status // FAULT или WARNING

						// Используем метаданные для получения более точного имени компонента
						meta, ok := metadata[strings.ToLower(condition.DataItemId)]
						if ok {
							alarm["componentName"] = meta.ComponentName
							alarm["componentId"] = meta.ComponentId
						} else {
							// Откатываемся до информации из ComponentStream, если метаданные не найдены
							alarm["componentName"] = compStream.Name
							alarm["componentId"] = compStream.ComponentId
						}

						// Динамически добавляем все непустые поля из XML
						if condition.Type != "" {
							alarm["type"] = condition.Type
						}
						if condition.NativeCode != "" {
							alarm["nativeCode"] = condition.NativeCode
						}
						if condition.Value != "" {
							alarm["message"] = strings.TrimSpace(condition.Value)
						}
						if condition.DataItemId != "" {
							alarm["dataItemId"] = condition.DataItemId
						}
						if condition.Timestamp != "" {
							alarm["timestamp"] = condition.Timestamp
						}

						machine.Alarms = append(machine.Alarms, alarm)
					}

					// Обновляем общие статусы
					if status == "FAULT" {
						machine.AlarmStatus = "FAULT"
					}
					if status == "WARNING" {
						machine.WarningStatus = "WARNING"
					}
				}
			}
		}
	}

	var machineDataSlice []MachineData
	for machineID, data := range machineDataMap {
		// Добавляем отсортированные данные по осям
		if machineAxes, ok := axisInfoMap[machineID]; ok {
			keys := make([]string, 0, len(machineAxes))
			for k := range machineAxes {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, key := range keys {
				data.AxisInfos = append(data.AxisInfos, *machineAxes[key])
			}
		}

		// Добавляем отсортированные данные по шпинделям
		if machineSpindles, ok := spindleInfoMap[machineID]; ok {
			keys := make([]string, 0, len(machineSpindles))
			for k := range machineSpindles {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, key := range keys {
				data.SpindleInfos = append(data.SpindleInfos, *machineSpindles[key])
			}
		}

		data.HasAlarms = (data.AlarmStatus == "FAULT" || data.WarningStatus == "WARNING")

		if data.EditStatus == "UNAVAILABLE" && data.ProgramMode != "UNAVAILABLE" {
			if data.ProgramMode == "EDIT" {
				data.EditStatus = "READY"
			} else {
				data.EditStatus = "NOT_READY"
			}
		}

		if data.WriteStatus == "UNAVAILABLE" && data.ProgramMode != "UNAVAILABLE" {
			if data.ProgramMode == "EDIT" {
				data.WriteStatus = "READY"
			} else {
				data.WriteStatus = "NOT_READY"
			}
		}

		machineDataSlice = append(machineDataSlice, *data)
	}

	return machineDataSlice
}

// Логика сопоставления для всех остальных (не осевых и не шпиндельных) данных.
func processDataItem(machine *MachineData, dataItemId, value, timestamp string, metadata map[string]DataItemMetadata) {
	meta, ok := metadata[strings.ToLower(dataItemId)]
	if !ok {
		return
	}

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
		machine.ManualMode = (value == "MANUAL" || value == "MANUAL_DATA_INPUT")
	case "AXIS_STATE":
		if _, isString := machine.AxisMovementStatus.(string); isString {
			machine.AxisMovementStatus = make(map[string]string)
		}
		if statusMap, ok := machine.AxisMovementStatus.(map[string]string); ok {
			if meta.ComponentName != "" {
				statusMap[meta.ComponentName] = value
			}
		}
	case "PROGRAM_EDIT":
		machine.EditStatus = value
		machine.WriteStatus = value
	case "POWER_STATE":
		machine.BatteryStatus = value
	case "TOOL_NUMBER":
		machine.ActiveToolNumber = value
	case "TOOL_OFFSET":
		machine.ToolOffsetNumber = value
	case "PATH_FEEDRATE":
		key := meta.SubType
		if key == "" {
			key = "VALUE"
		}
		machine.FeedRate[key] = value
	case "PATH_FEEDRATE_OVERRIDE":
		key := meta.SubType
		if key == "" {
			key = "VALUE"
		}
		machine.FeedOverride[key] = value
	case "PART_COUNT":
		key := meta.SubType
		if key == "" {
			key = "ALL"
		}
		machine.PartsCount[key] = value
	case "ACCUMULATED_TIME":
		key := meta.SubType
		if key == "" {
			key = "VALUE"
		}
		machine.AccumulatedTime[key] = formatAccumulatedTime(value)
	case "BLOCK":
		if machine.CurrentProgram == nil {
			machine.CurrentProgram = &CurrentProgramInfo{}
		}
		machine.CurrentProgram.Block = value
	case "PROGRAM":
		if machine.CurrentProgram == nil {
			machine.CurrentProgram = &CurrentProgramInfo{}
		}
		machine.CurrentProgram.Program = value
	case "PROGRAM_COMMENT":
		if machine.CurrentProgram == nil {
			machine.CurrentProgram = &CurrentProgramInfo{}
		}
		machine.CurrentProgram.ProgramComment = value
	case "PROGRAM_HEADER":
		if machine.CurrentProgram == nil {
			machine.CurrentProgram = &CurrentProgramInfo{}
		}
		machine.CurrentProgram.ProgramHeader = value
	case "LINE":
		if machine.CurrentProgram == nil {
			machine.CurrentProgram = &CurrentProgramInfo{}
		}
		machine.CurrentProgram.Line = value
	case "LINE_NUMBER":
		if machine.CurrentProgram == nil {
			machine.CurrentProgram = &CurrentProgramInfo{}
		}
		machine.CurrentProgram.LineNumber = value
	case "LINE_LABEL":
		if machine.CurrentProgram == nil {
			machine.CurrentProgram = &CurrentProgramInfo{}
		}
		machine.CurrentProgram.LineLabel = value
	}
}
