package models

// ErrorResponse представляет стандартный ответ с ошибкой.
type ErrorResponse struct {
	Status string `json:"status" example:"error"`
	Error  struct {
		Code    int    `json:"code" example:"404"`
		Message string `json:"message" example:"Подключение не найдено"`
	} `json:"error"`
}

// MessageResponse представляет стандартный успешный ответ с сообщением.
type MessageResponse struct {
	Status  string `json:"status" example:"ok"`
	Message string `json:"message" example:"Polling started successfully"`
}

// CreateConnectionResponse представляет ответ при успешном создании подключения.
type CreateConnectionResponse struct {
	Status         string          `json:"status" example:"ok"`
	ConnectionInfo *ConnectionInfo `json:"connection_info"`
}

// GetConnectionsResponse представляет ответ со списком всех подключений.
type GetConnectionsResponse struct {
	Status      string            `json:"status" example:"ok"`
	PoolSize    int               `json:"pool_size" example:"2"`
	Connections []*ConnectionInfo `json:"connections"`
}

// CheckConnectionResponse представляет ответ при успешной проверке подключения.
type CheckConnectionResponse struct {
	Status         string          `json:"status" example:"healthy"`
	ConnectionInfo *ConnectionInfo `json:"connection_info"`
}
