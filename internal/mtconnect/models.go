package mtconnect

import "encoding/xml"

// Корневой элемент XML-документа
type MTConnectStreams struct {
	XMLName xml.Name `xml:"MTConnectStreams"`
	Header  Header   `xml:"Header"`
	Streams Streams  `xml:"Streams"`
}

// Содержит мета-информацию
type Header struct {
	CreationTime string `xml:"creationTime,attr"`
	Sender       string `xml:"sender,attr"`
	InstanceID   string `xml:"instanceId,attr"`
	BufferSize   int    `xml:"bufferSize,attr"`
	NextSequence int64  `xml:"nextSequence,attr"`
}

// Содержит потоки данных от одного или нескольких устройств
type Streams struct {
	DeviceStreams []DeviceStream `xml:"DeviceStream"`
}

// Содержит данные для одного конкретного устройства.
type DeviceStream struct {
	Name            string            `xml:"name,attr"`
	UUID            string            `xml:"uuid,attr"`
	ComponentStream []ComponentStream `xml:"ComponentStream"`
}

// Поток данных от компонента (оси, шпинделя, контроллера)
type ComponentStream struct {
	Component   string     `xml:"component,attr"`
	Name        string     `xml:"name,attr"`
	ComponentID string     `xml:"componentId,attr"`
	Samples     *Samples   `xml:"Samples"`
	Events      *Events    `xml:"Events"`
	Condition   *Condition `xml:"Condition"`
}

// Обёртка для тега <Samples>
type Samples struct {
	Items []DataItem `xml:",any"`
}

// Обёртка для тега <Events>
type Events struct {
	Items []DataItem `xml:",any"`
}

// Обёртка для тега <Condition>
type Condition struct {
	Items []DataItem `xml:",any"`
}

// Универсальная структура для любого элемента данных
type DataItem struct {
	XMLName    xml.Name
	DataItemID string `xml:"dataItemId,attr"`
	Type       string `xml:"type,attr"`
	SubType    string `xml:"subType,attr"`
	Timestamp  string `xml:"timestamp,attr"`
	Sequence   int64  `xml:"sequence,attr"`
	Value      string `xml:",chardata"`
}
