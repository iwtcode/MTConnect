package entities

import "time"

// ConnectionRequest определяет структуру для нового запроса на подключение.
type ConnectionRequest struct {
	EndpointURL  string `json:"EndpointURL" binding:"required"`
	Model        string `json:"Model" binding:"required"`
	Manufacturer string `json:"Manufacturer,omitempty"`
}

// SessionRequest определяет структуру для запросов, использующих SessionID.
type SessionRequest struct {
	SessionID string `json:"SessionID" binding:"required"`
}

// ConnectionConfig содержит проверенную конфигурацию подключения.
type ConnectionConfig struct {
	EndpointURL  string `json:"EndpointURL"`
	Model        string `json:"Model"`
	Manufacturer string `json:"Manufacturer,omitempty"`
}

// ConnectionInfo представляет активное подключение в пуле.
type ConnectionInfo struct {
	SessionID string           `json:"SessionID"`
	MachineID string           `json:"-"` // Внутренний идентификатор станка из probe
	Config    ConnectionConfig `json:"Config"`
	CreatedAt time.Time        `json:"CreatedAt"`
	LastUsed  time.Time        `json:"LastUsed"`
	UseCount  int64            `json:"UseCount"`
	IsHealthy bool             `json:"IsHealthy"`
}
