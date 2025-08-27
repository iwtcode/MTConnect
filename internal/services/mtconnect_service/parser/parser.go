package parser

import (
	"MTConnect/internal/domain/models"
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
		return secondsStr
	}
	totalSeconds := int(math.Round(secondsFloat))
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

// processAxisDataItem обрабатывает DataItem'ы, связанные с осями
func processAxisDataItem(machineID, dataItemId, value string, axisLinks map[string]models.AxisDataItemLink, axisInfoMap map[string]map[string]*models.AxisInfo) bool {
	lowerId := strings.ToLower(dataItemId)
	link, ok := axisLinks[lowerId]
	if !ok || link.DeviceID != machineID {
		return false
	}
	if _, ok := axisInfoMap[machineID]; !ok {
		axisInfoMap[machineID] = make(map[string]*models.AxisInfo)
	}
	if _, ok := axisInfoMap[machineID][link.AxisComponentID]; !ok {
		axisInfoMap[machineID][link.AxisComponentID] = &models.AxisInfo{
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

// processSpindleDataItem обрабатывает DataItem'ы, связанные со шпинделями
func processSpindleDataItem(machineID, dataItemId, value string, spindleLinks map[string]models.SpindleDataItemLink, spindleInfoMap map[string]map[string]*models.SpindleInfo) bool {
	lowerId := strings.ToLower(dataItemId)
	link, ok := spindleLinks[lowerId]
	if !ok || link.DeviceID != machineID {
		return false
	}
	if _, ok := spindleInfoMap[machineID]; !ok {
		spindleInfoMap[machineID] = make(map[string]*models.SpindleInfo)
	}
	if _, ok := spindleInfoMap[machineID][link.SpindleComponentID]; !ok {
		spindleInfoMap[machineID][link.SpindleComponentID] = &models.SpindleInfo{
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

// MapToMachineData преобразует необработанные данные MTConnectStreams в срез MachineData
func MapToMachineData(streams *models.MTConnectStreams, metadata map[string]models.DataItemMetadata, axisLinks map[string]models.AxisDataItemLink, spindleLinks map[string]models.SpindleDataItemLink) []models.MachineData {
	machineDataMap := make(map[string]*models.MachineData)
	axisInfoMap := make(map[string]map[string]*models.AxisInfo)
	spindleInfoMap := make(map[string]map[string]*models.SpindleInfo)

	for _, deviceStream := range streams.Streams {
		machineID := deviceStream.Name
		if machineID == "" {
			machineID = deviceStream.UUID
		}
		if _, ok := machineDataMap[machineID]; !ok {
			machineDataMap[machineID] = &models.MachineData{
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
				AxisInfos:           make([]models.AxisInfo, 0),
				FeedRate:            make(map[string]string),
				FeedOverride:        make(map[string]string),
				Alarms:              make([]map[string]interface{}, 0),
				HasAlarms:           "UNAVAILABLE",
				PartsCount:          make(map[string]string),
				AccumulatedTime:     make(map[string]string),
				SpindleInfos:        make([]models.SpindleInfo, 0),
				ContourFeedRate:     "UNAVAILABLE",
				JogOverride:         "UNAVAILABLE",
			}
		}
		machine := machineDataMap[machineID]
		conditionsProcessedThisCycle := false

		for _, compStream := range deviceStream.ComponentStreams {
			if compStream.Samples != nil {
				for _, sample := range compStream.Samples.Items {
					if !processAxisDataItem(machine.MachineId, sample.DataItemId, sample.Value, axisLinks, axisInfoMap) &&
						!processSpindleDataItem(machine.MachineId, sample.DataItemId, sample.Value, spindleLinks, spindleInfoMap) {
						processDataItem(machine, sample.DataItemId, sample.Value, sample.Timestamp, metadata)
					}
				}
			}
			if compStream.Events != nil {
				for _, event := range compStream.Events.Items {
					if !processAxisDataItem(machine.MachineId, event.DataItemId, event.Value, axisLinks, axisInfoMap) &&
						!processSpindleDataItem(machine.MachineId, event.DataItemId, event.Value, spindleLinks, spindleInfoMap) {
						processDataItem(machine, event.DataItemId, event.Value, event.Timestamp, metadata)
					}
				}
			}
			if compStream.Condition != nil && len(compStream.Condition.Items) > 0 {
				if !conditionsProcessedThisCycle {
					machine.AlarmStatus = "NORMAL"
					machine.WarningStatus = "NORMAL"
					machine.Alarms = make([]map[string]interface{}, 0)
					conditionsProcessedThisCycle = true
				}
				for _, condition := range compStream.Condition.Items {
					status := strings.ToUpper(condition.XMLName.Local)
					if status == "FAULT" || status == "WARNING" {
						alarm := make(map[string]interface{})
						alarm["level"] = status
						if meta, ok := metadata[strings.ToLower(condition.DataItemId)]; ok {
							alarm["componentName"], alarm["componentId"] = meta.ComponentName, meta.ComponentId
						} else {
							alarm["componentName"], alarm["componentId"] = compStream.Name, compStream.ComponentId
						}
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

	var machineDataSlice []models.MachineData
	for machineID, data := range machineDataMap {
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
			data.EditStatus = "NOT_READY"
			if data.ProgramMode == "EDIT" {
				data.EditStatus = "READY"
			}
		}
		if data.WriteStatus == "UNAVAILABLE" && data.ProgramMode != "UNAVAILABLE" {
			data.WriteStatus = "NOT_READY"
			if data.ProgramMode == "EDIT" {
				data.WriteStatus = "READY"
			}
		}
		machineDataSlice = append(machineDataSlice, *data)
	}
	return machineDataSlice
}

// processDataItem - логика сопоставления для всех остальных данных
func processDataItem(machine *models.MachineData, dataItemId, value, timestamp string, metadata map[string]models.DataItemMetadata) {
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
		if statusMap, ok := machine.AxisMovementStatus.(map[string]string); ok && meta.ComponentName != "" {
			statusMap[meta.ComponentName] = value
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
			machine.CurrentProgram = &models.CurrentProgramInfo{}
		}
		machine.CurrentProgram.Block = value
	case "PROGRAM":
		if machine.CurrentProgram == nil {
			machine.CurrentProgram = &models.CurrentProgramInfo{}
		}
		machine.CurrentProgram.Program = value
	case "PROGRAM_COMMENT":
		if machine.CurrentProgram == nil {
			machine.CurrentProgram = &models.CurrentProgramInfo{}
		}
		machine.CurrentProgram.ProgramComment = value
	case "PROGRAM_HEADER":
		if machine.CurrentProgram == nil {
			machine.CurrentProgram = &models.CurrentProgramInfo{}
		}
		machine.CurrentProgram.ProgramHeader = value
	case "LINE":
		if machine.CurrentProgram == nil {
			machine.CurrentProgram = &models.CurrentProgramInfo{}
		}
		machine.CurrentProgram.Line = value
	case "LINE_NUMBER":
		if machine.CurrentProgram == nil {
			machine.CurrentProgram = &models.CurrentProgramInfo{}
		}
		machine.CurrentProgram.LineNumber = value
	case "LINE_LABEL":
		if machine.CurrentProgram == nil {
			machine.CurrentProgram = &models.CurrentProgramInfo{}
		}
		machine.CurrentProgram.LineLabel = value
	}
}
