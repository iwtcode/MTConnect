package mtconnect

import "encoding/xml"

// Структура для детального описания одной тревоги
type AlarmDetail struct {
	Status        string `json:"status"`        // FAULT, WARNING
	Type          string `json:"type"`          // Тип тревоги (SYSTEM, POSITION, и т.д.)
	ComponentName string `json:"componentName"` // Имя компонента, вызвавшего тревогу
	Message       string `json:"message"`       // Текст сообщения о тревоге
	NativeCode    string `json:"nativeCode"`    // Оригинальный код ошибки от станка
}

// Конечная модель данных для одного станка
type MachineData struct {
	MachineId           string      `json:"MachineId"`           // Идентификатор станка
	Id                  string      `json:"Id"`                  // Идентификатор станка
	Timestamp           string      `json:"Timestamp"`           // Временная метка
	IsEnabled           interface{} `json:"IsEnabled"`           // Станок включен (bool или "UNAVAILABLE")
	IsInEmergency       interface{} `json:"IsInEmergency"`       // Аварийный статус (bool или "UNAVAILABLE")
	MachineState        string      `json:"MachineState"`        // Статус выполнения программы
	ProgramMode         string      `json:"ProgramMode"`         // Режим работы (MDI / MEM / EDIT)
	TmMode              string      `json:"TmMode"`              // ЗАГЛУШКА: Режим T/M
	HandleRetraceStatus string      `json:"HandleRetraceStatus"` // ЗАГЛУШКА: Статус ручного перемещения
	AxisMovementStatus  interface{} `json:"AxisMovementStatus"`  // Статус движения осей
	MstbStatus          string      `json:"MstbStatus"`          // ЗАГЛУШКА: Статус M/S/T/B
	EmergencyStatus     string      `json:"EmergencyStatus"`     // Статус аварийного стопа
	AlarmStatus         string      `json:"AlarmStatus"`         // Общий статус тревоги (UNAVAILABLE, NORMAL, WARNING, FAULT)
	Alarms              interface{} `json:"Alarms"`              // Список тревог (слайс []AlarmDetail) или "UNAVAILABLE"
	HasAlarms           interface{} `json:"hasAlarms"`           // Флаг наличия активных тревог (bool или "UNAVAILABLE")
}

// Хранит метаданные из /probe для каждого DataItem
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
	DataItems     []DataItem     `xml:"DataItems>DataItem"`
	ComponentList *ComponentList `xml:"Components"`
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

// Вспомогательные структуры для правильного парсинга
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
