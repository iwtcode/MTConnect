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
	Block          string `json:"BLOCK,omitempty"`
	Program        string `json:"PROGRAM,omitempty"`
	ProgramComment string `json:"PROGRAM_COMMENT,omitempty"`
	ProgramHeader  string `json:"PROGRAM_HEADER,omitempty"`
	Line           string `json:"LINE,omitempty"`
	LineNumber     string `json:"LINE_NUMBER,omitempty"`
	LineLabel      string `json:"LINE_LABEL,omitempty"`
}

// MachineData - конечная модель данных для одного станка
type MachineData struct {
	MachineId           string                   `json:"MachineId"`
	Id                  string                   `json:"Id"`
	Timestamp           string                   `json:"Timestamp"`
	IsEnabled           interface{}              `json:"IsEnabled"`
	IsInEmergency       interface{}              `json:"IsInEmergency"`
	MachineState        string                   `json:"MachineState"`
	ProgramMode         string                   `json:"ProgramMode"`
	TmMode              string                   `json:"TmMode"`
	HandleRetraceStatus interface{}              `json:"HandleRetraceStatus"`
	AxisMovementStatus  interface{}              `json:"AxisMovementStatus"`
	MstbStatus          string                   `json:"MstbStatus"`
	EmergencyStatus     string                   `json:"EmergencyStatus"`
	AlarmStatus         string                   `json:"AlarmStatus"`
	EditStatus          string                   `json:"EditStatus"`
	ManualMode          interface{}              `json:"ManualMode"`
	WriteStatus         string                   `json:"WriteStatus"`
	LabelSkipStatus     interface{}              `json:"LabelSkipStatus"`
	WarningStatus       string                   `json:"WarningStatus"`
	BatteryStatus       interface{}              `json:"BatteryStatus"`
	ActiveToolNumber    string                   `json:"activeToolNumber"`
	ToolOffsetNumber    string                   `json:"toolOffsetNumber"`
	AxisInfos           []AxisInfo               `json:"AxisInfos"`
	FeedRate            map[string]string        `json:"FeedRate"`
	FeedOverride        map[string]string        `json:"FeedOverride"`
	Alarms              []map[string]interface{} `json:"Alarms"`
	HasAlarms           interface{}              `json:"hasAlarms"`
	PartsCount          map[string]string        `json:"PartsCount"`
	AccumulatedTime     map[string]string        `json:"AccumulatedTime"`
	CurrentProgram      *CurrentProgramInfo      `json:"CurrentProgram,omitempty"`
	SpindleInfos        []SpindleInfo            `json:"SpindleInfos"`
	ContourFeedRate     interface{}              `json:"ContourFeedRate"`
	JogOverride         interface{}              `json:"JogOverride"`
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
