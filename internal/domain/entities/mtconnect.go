package entities

import "encoding/xml"

// AxisInfo содержит актуальную информацию о состоянии одной оси станка
type AxisInfo struct {
	ID   string                 `json:"id"`
	Name string                 `json:"name"`
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// SpindleInfo содержит актуальную информацию о состоянии одного шпинделя станка
type SpindleInfo struct {
	ID   string                 `json:"id"`
	Name string                 `json:"name"`
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// CurrentProgramInfo содержит информацию о текущей выполняемой программе
type CurrentProgramInfo struct {
	Block          string `json:"block,omitempty"`
	Program        string `json:"program,omitempty"`
	ProgramComment string `json:"program_comment,omitempty"`
	ProgramHeader  string `json:"program_header,omitempty"`
	Line           string `json:"line,omitempty"`
	LineNumber     string `json:"line_number,omitempty"`
	LineLabel      string `json:"line_label,omitempty"`
}

// MachineData - конечная модель данных для одного станка
type MachineData struct {
	MachineId           string                   `json:"machine_id"`
	Id                  string                   `json:"id"`
	Timestamp           string                   `json:"timestamp"`
	IsEnabled           interface{}              `json:"is_enabled"`
	IsInEmergency       interface{}              `json:"is_in_emergency"`
	MachineState        string                   `json:"machine_state"`
	ProgramMode         string                   `json:"program_mode"`
	TmMode              string                   `json:"tm_mode"`
	HandleRetraceStatus interface{}              `json:"handle_retrace_status"`
	AxisMovementStatus  interface{}              `json:"axis_movement_status"`
	MstbStatus          string                   `json:"mstb_status"`
	EmergencyStatus     string                   `json:"emergency_status"`
	AlarmStatus         string                   `json:"alarm_status"`
	EditStatus          string                   `json:"edit_status"`
	ManualMode          interface{}              `json:"manual_mode"`
	WriteStatus         string                   `json:"write_status"`
	LabelSkipStatus     interface{}              `json:"label_skip_status"`
	WarningStatus       string                   `json:"warning_status"`
	BatteryStatus       interface{}              `json:"battery_status"`
	ActiveToolNumber    string                   `json:"active_tool_number"`
	ToolOffsetNumber    string                   `json:"tool_offset_number"`
	AxisInfos           []AxisInfo               `json:"axis_infos"`
	FeedRate            map[string]string        `json:"feed_rate"`
	FeedOverride        map[string]string        `json:"feed_override"`
	Alarms              []map[string]interface{} `json:"alarms"`
	HasAlarms           interface{}              `json:"has_alarms"`
	PartsCount          map[string]string        `json:"parts_count"`
	AccumulatedTime     map[string]string        `json:"accumulated_time"`
	CurrentProgram      *CurrentProgramInfo      `json:"current_program,omitempty"`
	SpindleInfos        []SpindleInfo            `json:"spindle_infos"`
	ContourFeedRate     interface{}              `json:"contour_feed_rate"`
	JogOverride         interface{}              `json:"jog_override"`
}

// DataItemMetadata хранит метаданные из /probe для каждого DataItem
type DataItemMetadata struct {
	ID            string
	Name          string
	ComponentId   string
	ComponentName string
	ComponentType string
	Category      string
	Type          string
	SubType       string
}

// AxisDataItemLink - структура для связи DataItem'а с конкретной осью
type AxisDataItemLink struct {
	DeviceID        string
	AxisComponentID string
	AxisName        string
	AxisType        string
	DataKey         string
}

// SpindleDataItemLink - структура для связи DataItem'а с конкретным шпинделем
type SpindleDataItemLink struct {
	DeviceID           string
	SpindleComponentID string
	SpindleName        string
	SpindleType        string
	DataKey            string
}

// --- Структуры для парсинга /probe ---

type MTConnectDevices struct {
	XMLName xml.Name `xml:"MTConnectDevices"`
	Devices []Device `xml:"Devices>Device"`
}

type Device struct {
	XMLName       xml.Name       `xml:"Device"`
	Name          string         `xml:"name,attr"`
	UUID          string         `xml:"uuid,attr"`
	ID            string         `xml:"id,attr"`
	Description   *Description   `xml:"Description"` // Изменено для парсинга атрибутов и значения
	DataItems     []DataItem     `xml:"DataItems>DataItem"`
	ComponentList *ComponentList `xml:"Components"`
}

// Description содержит метаданные из тега Description в /probe
type Description struct {
	Manufacturer string `xml:"manufacturer,attr"`
	Model        string `xml:"model,attr"`
	SerialNumber string `xml:"serialNumber,attr"`
	Value        string `xml:",chardata"`
}

type ProbeComponent struct {
	XMLName       xml.Name
	ID            string         `xml:"id,attr"`
	Name          string         `xml:"name,attr"`
	DataItems     []DataItem     `xml:"DataItems>DataItem"`
	ComponentList *ComponentList `xml:"Components"`
}

type ComponentList struct {
	Components []ProbeComponent `xml:",any"`
}

type DataItem struct {
	ID       string `xml:"id,attr"`
	Name     string `xml:"name,attr"`
	Category string `xml:"category,attr"`
	Type     string `xml:"type,attr"`
	SubType  string `xml:"subType,attr"`
}

// --- Структуры для парсинга /current ---

type MTConnectStreams struct {
	XMLName xml.Name       `xml:"MTConnectStreams"`
	Streams []DeviceStream `xml:"Streams>DeviceStream"`
}

type DeviceStream struct {
	Name             string            `xml:"name,attr"`
	UUID             string            `xml:"uuid,attr"`
	ComponentStreams []ComponentStream `xml:"ComponentStream"`
}

type ComponentStream struct {
	Component   string      `xml:"component,attr"`
	Name        string      `xml:"name,attr"`
	ComponentId string      `xml:"componentId,attr"`
	Samples     *Samples    `xml:"Samples"`
	Events      *Events     `xml:"Events"`
	Condition   *Conditions `xml:"Condition"`
}

type Samples struct {
	Items []SampleValue `xml:",any"`
}
type Events struct {
	Items []EventValue `xml:",any"`
}
type Conditions struct {
	Items []ConditionValue `xml:",any"`
}

type SampleValue struct {
	XMLName    xml.Name
	DataItemId string `xml:"dataItemId,attr"`
	Timestamp  string `xml:"timestamp,attr"`
	Name       string `xml:"name,attr"`
	SubType    string `xml:"subType,attr"`
	Value      string `xml:",chardata"`
}

type EventValue struct {
	XMLName    xml.Name
	DataItemId string `xml:"dataItemId,attr"`
	Timestamp  string `xml:"timestamp,attr"`
	Name       string `xml:"name,attr"`
	Value      string `xml:",chardata"`
}

type ConditionValue struct {
	XMLName    xml.Name
	DataItemId string `xml:"dataItemId,attr"`
	Timestamp  string `xml:"timestamp,attr"`
	Name       string `xml:"name,attr"`
	Type       string `xml:"type,attr"`
	NativeCode string `xml:"nativeCode,attr"`
	Value      string `xml:",chardata"`
}
