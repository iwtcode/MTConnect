package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	// Важно: импортируем модели из нового публичного пакета
	"github.com/iwtcode/MTConnect/pkg/models"
)

// ClientAPI определяет интерфейс для взаимодействия с MTConnect Service API.
type ClientAPI interface {
	// Управление подключениями
	CreateConnection(ctx context.Context, req models.ConnectionRequest) (*models.CreateConnectionResponse, *http.Response, error)
	GetConnections(ctx context.Context) (*models.GetConnectionsResponse, *http.Response, error)
	DeleteConnection(ctx context.Context, sessionID string) (*models.MessageResponse, *http.Response, error)
	CheckConnection(ctx context.Context, sessionID string) (*models.CheckConnectionResponse, *http.Response, error)

	// Управление опросом
	StartPolling(ctx context.Context, req models.PollingRequest) (*models.MessageResponse, *http.Response, error)
	StopPolling(ctx context.Context, sessionID string) (*models.MessageResponse, *http.Response, error)
}

// Client реализует интерфейс ClientAPI.
type Client struct {
	service *ClientService
}

// NewClient создает новый клиент для MTConnect Service API.
func NewClient(host string) ClientAPI {
	return &Client{
		service: NewClientService(host),
	}
}

// CreateConnection создает новое подключение к MTConnect эндпоинту.
func (c *Client) CreateConnection(ctx context.Context, req models.ConnectionRequest) (*models.CreateConnectionResponse, *http.Response, error) {
	const endpoint = "/api/v1/connect"

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка сериализации запроса: %w", err)
	}

	httpReq, err := c.service.createRequestJSONWithContext(ctx, http.MethodPost, endpoint, nil, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, nil, err
	}

	body, httpResp, err := c.service.doRequest(httpReq)
	if err != nil {
		return nil, httpResp, err
	}

	var resp models.CreateConnectionResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, httpResp, fmt.Errorf("ошибка десериализации ответа: %w", err)
	}

	return &resp, httpResp, nil
}

// GetConnections возвращает список всех активных подключений.
func (c *Client) GetConnections(ctx context.Context) (*models.GetConnectionsResponse, *http.Response, error) {
	const endpoint = "/api/v1/connect"

	httpReq, err := c.service.createRequestJSONWithContext(ctx, http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	body, httpResp, err := c.service.doRequest(httpReq)
	if err != nil {
		return nil, httpResp, err
	}

	var resp models.GetConnectionsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, httpResp, fmt.Errorf("ошибка десериализации ответа: %w", err)
	}

	return &resp, httpResp, nil
}

// DeleteConnection удаляет подключение по SessionID.
func (c *Client) DeleteConnection(ctx context.Context, sessionID string) (*models.MessageResponse, *http.Response, error) {
	const endpoint = "/api/v1/connect"

	reqBody, err := json.Marshal(models.SessionRequest{SessionID: sessionID})
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка сериализации запроса: %w", err)
	}

	httpReq, err := c.service.createRequestJSONWithContext(ctx, http.MethodDelete, endpoint, nil, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, nil, err
	}

	body, httpResp, err := c.service.doRequest(httpReq)
	if err != nil {
		return nil, httpResp, err
	}

	var resp models.MessageResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, httpResp, fmt.Errorf("ошибка десериализации ответа: %w", err)
	}

	return &resp, httpResp, nil
}

// CheckConnection проверяет состояние подключения по SessionID.
func (c *Client) CheckConnection(ctx context.Context, sessionID string) (*models.CheckConnectionResponse, *http.Response, error) {
	const endpoint = "/api/v1/connect/check"

	reqBody, err := json.Marshal(models.SessionRequest{SessionID: sessionID})
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка сериализации запроса: %w", err)
	}

	httpReq, err := c.service.createRequestJSONWithContext(ctx, http.MethodPost, endpoint, nil, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, nil, err
	}

	body, httpResp, err := c.service.doRequest(httpReq)
	if err != nil {
		return nil, httpResp, err
	}

	var resp models.CheckConnectionResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, httpResp, fmt.Errorf("ошибка десериализации ответа: %w", err)
	}

	return &resp, httpResp, nil
}

// StartPolling запускает опрос данных для указанного подключения.
func (c *Client) StartPolling(ctx context.Context, req models.PollingRequest) (*models.MessageResponse, *http.Response, error) {
	const endpoint = "/api/v1/polling/start"

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка сериализации запроса: %w", err)
	}

	httpReq, err := c.service.createRequestJSONWithContext(ctx, http.MethodPost, endpoint, nil, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, nil, err
	}

	body, httpResp, err := c.service.doRequest(httpReq)
	if err != nil {
		return nil, httpResp, err
	}

	var resp models.MessageResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, httpResp, fmt.Errorf("ошибка десериализации ответа: %w", err)
	}

	return &resp, httpResp, nil
}

// StopPolling останавливает опрос данных для указанного подключения.
func (c *Client) StopPolling(ctx context.Context, sessionID string) (*models.MessageResponse, *http.Response, error) {
	const endpoint = "/api/v1/polling/stop"

	reqBody, err := json.Marshal(models.SessionRequest{SessionID: sessionID})
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка сериализации запроса: %w", err)
	}

	httpReq, err := c.service.createRequestJSONWithContext(ctx, http.MethodPost, endpoint, nil, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, nil, err
	}

	body, httpResp, err := c.service.doRequest(httpReq)
	if err != nil {
		return nil, httpResp, err
	}

	var resp models.MessageResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, httpResp, fmt.Errorf("ошибка десериализации ответа: %w", err)
	}

	return &resp, httpResp, nil
}
