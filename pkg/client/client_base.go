package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// ClientService обрабатывает низкоуровневое HTTP-взаимодействие.
type ClientService struct {
	HTTPClient *http.Client
	Host       string
}

// NewClientService создает новый сервис для клиента.
func NewClientService(host string) *ClientService {
	host = strings.TrimSuffix(host, "/")
	return &ClientService{
		HTTPClient: &http.Client{},
		Host:       host,
	}
}

// createRequestJSON создает HTTP-запрос с JSON-заголовками.
func (s *ClientService) createRequestJSON(ctx context.Context, httpMethod, urlPath string, queryParams map[string]string, reqBody io.Reader) (*http.Request, error) {
	processedUrlPath := strings.TrimPrefix(urlPath, "/")
	fullURL := fmt.Sprintf("%s/api/v1/%s", s.Host, processedUrlPath)

	if len(queryParams) > 0 {
		params := url.Values{}
		for param, value := range queryParams {
			params.Add(param, value)
		}
		fullURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, httpMethod, fullURL, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}

// doRequest отправляет запрос и возвращает тело ответа, полную структуру ответа или ошибку.
func (s *ClientService) doRequest(req *http.Request) ([]byte, *http.Response, error) {
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, resp, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}

	// Создаем новый reader для тела ответа, чтобы его можно было прочитать снова при необходимости.
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return bodyBytes, resp, fmt.Errorf("запрос завершился с кодом состояния %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return bodyBytes, resp, nil
}
