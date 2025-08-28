package models

// ErrorResponse представляет стандартный ответ с ошибкой.
type ErrorResponse struct {
	Status string `json:"Status" example:"error"`
	Error  struct {
		Code    int    `json:"code" example:"404"`
		Message string `json:"message" example:"Подключение не найдено"`
	} `json:"error"`
}

// MessageResponse представляет стандартный успешный ответ с сообщением.
type MessageResponse struct {
	Status  string `json:"Status" example:"ok"`
	Message string `json:"Message" example:"Polling started successfully"`
}

// CreateConnectionResponse представляет ответ при успешном создании подключения.
type CreateConnectionResponse struct {
	Status         string          `json:"Status" example:"ok"`
	ConnectionInfo *ConnectionInfo `json:"connectionInfo"`
}

// GetConnectionsResponse представляет ответ со списком всех подключений.
type GetConnectionsResponse struct {
	Status      string            `json:"Status" example:"ok"`
	PoolSize    int               `json:"PoolSize" example:"2"`
	Connections []*ConnectionInfo `json:"Connections"`
}

// CheckConnectionResponse представляет ответ при успешной проверке подключения.
type CheckConnectionResponse struct {
	Status         string          `json:"Status" example:"healthy"`
	ConnectionInfo *ConnectionInfo `json:"connectionInfo"`
}
