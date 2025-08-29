package client

import (
	"context"
	"log"
	"testing"
	"time"
)

// ВНИМАНИЕ: Эти тесты являются интеграционными и требуют, чтобы основной
// сервис MTConnect был запущен и доступен по адресу, указанному в 'testHost'.
const (
	testHost = "http://localhost:8080"
	// Используется публичный и стабильный эндпоинт для тестов
	testEndpointURL = "https://smstestbed.nist.gov/vds"
	testModel       = "Mazak Integrex 100-IV"
)

// TestClientWorkflow проверяет полный цикл работы с API.
func TestClientWorkflow(t *testing.T) {
	// 1. Создание нового клиента
	client := NewClient(testHost)
	ctx := context.Background()

	// 2. Создание нового подключения
	t.Log("Шаг 1: Создание нового подключения...")
	createResp, err := client.CreateConnection(ctx, testEndpointURL, testModel, "")
	if err != nil {
		t.Fatalf("Не удалось создать подключение: %v", err)
	}
	if createResp.Status != "ok" || createResp.ConnectionInfo == nil || createResp.ConnectionInfo.SessionID == "" {
		t.Fatalf("Получен некорректный ответ при создании подключения: %+v", createResp)
	}
	sessionID := createResp.ConnectionInfo.SessionID
	t.Logf("Подключение успешно создано. SessionID: %s", sessionID)

	// 3. Отложенное удаление для очистки ресурсов после теста
	defer func() {
		t.Logf("Очистка: Удаление подключения с SessionID: %s", sessionID)
		_, err := client.DeleteConnection(ctx, sessionID)
		if err != nil {
			// Используем log вместо t.Fatalf, чтобы не маскировать исходную ошибку теста
			log.Printf("ПРЕДУПРЕЖДЕНИЕ: Не удалось удалить подключение во время очистки: %v", err)
		} else {
			t.Log("Очистка ресурсов прошла успешно.")
		}
	}()

	// 4. Получение списка всех подключений для проверки
	t.Log("Шаг 2: Получение списка подключений для проверки...")
	getResp, err := client.GetConnections(ctx)
	if err != nil {
		t.Fatalf("Не удалось получить список подключений: %v", err)
	}
	var found bool
	for _, conn := range getResp.Connections {
		if conn.SessionID == sessionID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Созданное подключение с SessionID %s не найдено в общем списке", sessionID)
	}
	t.Log("Проверка подтвердила, что подключение присутствует в пуле.")

	// 5. Проверка состояния подключения
	t.Log("Шаг 3: Проверка состояния подключения...")
	checkResp, err := client.CheckConnection(ctx, sessionID)
	if err != nil {
		t.Fatalf("Не удалось проверить состояние подключения: %v", err)
	}
	if checkResp.Status != "healthy" || !checkResp.ConnectionInfo.IsHealthy {
		t.Fatalf("Проверка вернула нездоровый статус: %+v", checkResp)
	}
	t.Log("Проверка состояния подключения пройдена успешно.")

	// 6. Запуск опроса данных (polling)
	t.Log("Шаг 4: Запуск опроса данных...")
	startResp, err := client.StartPolling(ctx, sessionID, 1000)
	if err != nil {
		t.Fatalf("Не удалось запустить опрос: %v", err)
	}
	if startResp.Status != "ok" {
		t.Fatalf("Получен некорректный ответ при запуске опроса: %+v", startResp)
	}
	t.Log("Опрос данных успешно запущен.")

	// Дадим время на выполнение хотя бы одного цикла опроса
	time.Sleep(1200 * time.Millisecond)

	// 7. Остановка опроса данных
	t.Log("Шаг 5: Остановка опроса данных...")
	stopResp, err := client.StopPolling(ctx, sessionID)
	if err != nil {
		t.Fatalf("Не удалось остановить опрос: %v", err)
	}
	if stopResp.Status != "ok" {
		t.Fatalf("Получен некорректный ответ при остановке опроса: %+v", stopResp)
	}
	t.Log("Опрос данных успешно остановлен.")

	t.Log("Тестовый сценарий завершен. Запускается отложенная очистка...")
}
