package client

import "time"

// --- Модели запросов ---

// ConnectionRequest определяет структуру для запроса на новое подключение.
type ConnectionRequest struct {
	EndpointURL  string `json:"endpoint_url"`
	Model        string `json:"model"`
	Manufacturer string `json:"manufacturer,omitempty"`
}

// SessionRequest определяет структуру для запросов, использующих SessionID.
type SessionRequest struct {
	SessionID string `json:"session_id"`
}

// PollingRequest определяет структуру для запроса на запуск опроса.
type PollingRequest struct {
	SessionID string `json:"session_id"`
	Interval  int    `json:"interval"` // в миллисекундах
}

// --- Модели ответов ---

// ConnectionConfig содержит конфигурацию подключения.
type ConnectionConfig struct {
	EndpointURL  string `json:"endpoint_url"`
	Model        string `json:"model"`
	Manufacturer string `json:"manufacturer,omitempty"`
}

// ConnectionInfo представляет информацию об активном подключении.
type ConnectionInfo struct {
	SessionID string           `json:"session_id"`
	Config    ConnectionConfig `json:"config"`
	CreatedAt time.Time        `json:"created_at"`
	LastUsed  time.Time        `json:"last_used"`
	UseCount  int64            `json:"use_count"`
	IsHealthy bool             `json:"is_healthy"`
}

// ErrorDetail содержит код и сообщение об ошибке.
type ErrorDetail struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse представляет стандартный ответ с ошибкой.
type ErrorResponse struct {
	Status string      `json:"status"`
	Error  ErrorDetail `json:"error"`
}

// MessageResponse представляет стандартный успешный ответ с сообщением.
type MessageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// CreateConnectionResponse представляет ответ при успешном создании подключения.
type CreateConnectionResponse struct {
	Status         string          `json:"status"`
	ConnectionInfo *ConnectionInfo `json:"connection_info"`
}

// GetConnectionsResponse представляет ответ со списком всех подключений.
type GetConnectionsResponse struct {
	Status      string            `json:"status"`
	PoolSize    int               `json:"pool_size"`
	Connections []*ConnectionInfo `json:"connections"`
}

// CheckConnectionResponse представляет ответ при проверке подключения.
type CheckConnectionResponse struct {
	Status         string          `json:"status"`
	ConnectionInfo *ConnectionInfo `json:"connection_info"`
}
