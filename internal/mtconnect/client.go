package mtconnect

import (
	"fmt"
	"io"
	"net/http"
)

// GET-запрос к указанному URL, запрашивая XML
func FetchXML(url string) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса к %s: %w", url, err)
	}

	req.Header.Set("Accept", "application/xml")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса к %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("сервер %s ответил со статусом %s", url, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа от %s: %w", url, err)
	}

	return body, nil
}
