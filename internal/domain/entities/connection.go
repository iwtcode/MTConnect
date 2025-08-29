package entities

import "time"

// ConnectionRequest определяет структуру для нового запроса на подключение.
type ConnectionRequest struct {
	EndpointURL  string `json:"endpoint_url" binding:"required"`
	Model        string `json:"model" binding:"required"`
	Manufacturer string `json:"manufacturer,omitempty"`
}

// SessionRequest определяет структуру для запросов, использующих SessionID.
type SessionRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}

// ConnectionConfig содержит проверенную конфигурацию подключения.
type ConnectionConfig struct {
	EndpointURL  string `json:"endpoint_url"`
	Model        string `json:"model"`
	Manufacturer string `json:"manufacturer,omitempty"`
}

// ConnectionInfo представляет активное подключение в пуле.
type ConnectionInfo struct {
	SessionID string           `json:"session_id"`
	MachineID string           `json:"-"` // Внутренний идентификатор станка из probe
	Config    ConnectionConfig `json:"config"`
	CreatedAt time.Time        `json:"created_at"`
	LastUsed  time.Time        `json:"last_used"`
	UseCount  int64            `json:"use_count"`
	IsHealthy bool             `json:"is_healthy"`
}
