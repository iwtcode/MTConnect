package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ClientAPI определяет интерфейс для клиента сервиса MTConnect.
type ClientAPI interface {
	// CreateConnection создает новое подключение к эндпоинту MTConnect.
	CreateConnection(ctx context.Context, endpointURL, model, manufacturer string) (*CreateConnectionResponse, error)
	// GetConnections возвращает список всех активных подключений.
	GetConnections(ctx context.Context) (*GetConnectionsResponse, error)
	// DeleteConnection удаляет подключение по его SessionID.
	DeleteConnection(ctx context.Context, sessionID string) (*MessageResponse, error)
	// CheckConnection проверяет состояние подключения по его SessionID.
	CheckConnection(ctx context.Context, sessionID string) (*CheckConnectionResponse, error)
	// StartPolling запускает сбор данных для указанного подключения.
	StartPolling(ctx context.Context, sessionID string, interval int) (*MessageResponse, error)
	// StopPolling останавливает сбор данных для указанного подключения.
	StopPolling(ctx context.Context, sessionID string) (*MessageResponse, error)
}

// Client является конкретной реализацией ClientAPI.
type Client struct {
	service *ClientService
}

// NewClient создает новый клиент MTConnect.
func NewClient(host string) ClientAPI {
	return &Client{
		service: NewClientService(host),
	}
}

// unmarshalResponse — это вспомогательная функция для декодирования JSON-ответов и обработки ошибок.
func unmarshalResponse[T any](body []byte) (*T, error) {
	// Сначала пытаемся десериализовать в обобщенный ответ об ошибке.
	var genericResp map[string]interface{}
	if err := json.Unmarshal(body, &genericResp); err == nil {
		if status, ok := genericResp["status"].(string); ok && status == "error" {
			var errResp ErrorResponse
			if err := json.Unmarshal(body, &errResp); err == nil {
				return nil, fmt.Errorf("ошибка API: %s (код: %d)", errResp.Error.Message, errResp.Error.Code)
			}
		}
	}

	// Если ошибки нет, десериализуем в целевой тип.
	var target T
	if err := json.Unmarshal(body, &target); err != nil {
		return nil, fmt.Errorf("не удалось десериализовать успешный ответ: %w", err)
	}
	return &target, nil
}

func (c *Client) CreateConnection(ctx context.Context, endpointURL, model, manufacturer string) (*CreateConnectionResponse, error) {
	const endpoint = "connect"
	reqBody, err := json.Marshal(ConnectionRequest{
		EndpointURL:  endpointURL,
		Model:        model,
		Manufacturer: manufacturer,
	})
	if err != nil {
		return nil, fmt.Errorf("не удалось сериализовать тело запроса: %w", err)
	}

	req, err := c.service.createRequestJSON(ctx, http.MethodPost, endpoint, nil, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("не удалось создать запрос: %w", err)
	}

	body, _, err := c.service.doRequest(req)
	if err != nil {
		return nil, err
	}
	return unmarshalResponse[CreateConnectionResponse](body)
}

func (c *Client) GetConnections(ctx context.Context) (*GetConnectionsResponse, error) {
	const endpoint = "connect"
	req, err := c.service.createRequestJSON(ctx, http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать запрос: %w", err)
	}

	body, _, err := c.service.doRequest(req)
	if err != nil {
		return nil, err
	}
	return unmarshalResponse[GetConnectionsResponse](body)
}

func (c *Client) DeleteConnection(ctx context.Context, sessionID string) (*MessageResponse, error) {
	const endpoint = "connect"
	reqBody, err := json.Marshal(SessionRequest{SessionID: sessionID})
	if err != nil {
		return nil, fmt.Errorf("не удалось сериализовать тело запроса: %w", err)
	}

	req, err := c.service.createRequestJSON(ctx, http.MethodDelete, endpoint, nil, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("не удалось создать запрос: %w", err)
	}

	body, _, err := c.service.doRequest(req)
	if err != nil {
		return nil, err
	}
	return unmarshalResponse[MessageResponse](body)
}

func (c *Client) CheckConnection(ctx context.Context, sessionID string) (*CheckConnectionResponse, error) {
	const endpoint = "connect/check"
	reqBody, err := json.Marshal(SessionRequest{SessionID: sessionID})
	if err != nil {
		return nil, fmt.Errorf("не удалось сериализовать тело запроса: %w", err)
	}

	req, err := c.service.createRequestJSON(ctx, http.MethodPost, endpoint, nil, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("не удалось создать запрос: %w", err)
	}

	body, _, err := c.service.doRequest(req)
	if err != nil {
		return nil, err
	}
	return unmarshalResponse[CheckConnectionResponse](body)
}

func (c *Client) StartPolling(ctx context.Context, sessionID string, interval int) (*MessageResponse, error) {
	const endpoint = "polling/start"
	reqBody, err := json.Marshal(PollingRequest{
		SessionID: sessionID,
		Interval:  interval,
	})
	if err != nil {
		return nil, fmt.Errorf("не удалось сериализовать тело запроса: %w", err)
	}

	req, err := c.service.createRequestJSON(ctx, http.MethodPost, endpoint, nil, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("не удалось создать запрос: %w", err)
	}

	body, _, err := c.service.doRequest(req)
	if err != nil {
		return nil, err
	}
	return unmarshalResponse[MessageResponse](body)
}

func (c *Client) StopPolling(ctx context.Context, sessionID string) (*MessageResponse, error) {
	const endpoint = "polling/stop"
	reqBody, err := json.Marshal(SessionRequest{SessionID: sessionID})
	if err != nil {
		return nil, fmt.Errorf("не удалось сериализовать тело запроса: %w", err)
	}

	req, err := c.service.createRequestJSON(ctx, http.MethodPost, endpoint, nil, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("не удалось создать запрос: %w", err)
	}

	body, _, err := c.service.doRequest(req)
	if err != nil {
		return nil, err
	}
	return unmarshalResponse[MessageResponse](body)
}
