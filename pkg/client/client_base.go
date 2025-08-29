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

// ClientService предоставляет базовый функционал для работы с HTTP API.
type ClientService struct {
	HTTPClient *http.Client
	Host       string
}

// NewClientService создает новый экземпляр ClientService.
func NewClientService(host string) *ClientService {
	host = strings.TrimSuffix(host, "/")

	return &ClientService{
		HTTPClient: &http.Client{},
		Host:       host,
	}
}

// createRequestJSONWithContext создает HTTP-запрос с заголовком application/json.
func (s *ClientService) createRequestJSONWithContext(ctx context.Context, httpMethod, urlPath string, queryParams map[string]string, reqBody io.Reader) (*http.Request, error) {
	urlPath = strings.TrimPrefix(urlPath, "/")
	fullURL := fmt.Sprintf("%s/%s", s.Host, urlPath)

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
	return req, nil
}

// doRequest выполняет HTTP-запрос и возвращает тело ответа.
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

	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return bodyBytes, resp, fmt.Errorf("ошибка ответа сервера: статус %d, тело: %s", resp.StatusCode, string(bodyBytes))
	}

	return bodyBytes, resp, nil
}
