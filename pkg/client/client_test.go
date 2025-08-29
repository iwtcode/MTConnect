package client

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/iwtcode/MTConnect/pkg/models"
)

// Для запуска теста должен быть запущен MTConnect сервис на localhost:8080
func TestFullClientWorkflow(t *testing.T) {
	// Инициализация клиента
	api := NewClient("http://localhost:8080")
	ctx := context.Background()

	// 1. Создание подключения
	log.Println("Шаг 1: Создание подключения...")
	createReq := models.ConnectionRequest{
		EndpointURL: "https://smstestbed.nist.gov/vds",
		Model:       "Hurco VMX 24 #1",
	}
	connResp, _, err := api.CreateConnection(ctx, createReq)
	if err != nil {
		t.Fatalf("Ошибка создания подключения: %v", err)
	}
	if connResp.Status != "ok" || connResp.ConnectionInfo.SessionID == "" {
		t.Fatalf("Некорректный ответ при создании подключения: %+v", connResp)
	}
	sessionID := connResp.ConnectionInfo.SessionID
	log.Printf("Подключение создано успешно. SessionID: %s\n", sessionID)

	// 2. Получение списка подключений
	log.Println("Шаг 2: Получение списка всех подключений...")
	listResp, _, err := api.GetConnections(ctx)
	if err != nil {
		t.Fatalf("Ошибка получения списка подключений: %v", err)
	}
	if listResp.Status != "ok" || listResp.PoolSize == 0 {
		t.Fatalf("Некорректный ответ при получении списка: %+v", listResp)
	}
	log.Printf("Получено %d активных подключений.\n", listResp.PoolSize)

	// 3. Проверка состояния
	log.Println("Шаг 3: Проверка состояния подключения...")
	checkResp, _, err := api.CheckConnection(ctx, sessionID)
	if err != nil {
		t.Fatalf("Ошибка проверки подключения: %v", err)
	}
	if checkResp.Status != "healthy" {
		t.Fatalf("Проверка показала нездоровое состояние: %s", checkResp.Status)
	}
	log.Printf("Состояние подключения: %s\n", checkResp.Status)

	// 4. Запуск опроса
	log.Println("Шаг 4: Запуск опроса данных...")
	startPollReq := models.PollingRequest{
		SessionID: sessionID,
		Interval:  1000,
	}
	startMsg, _, err := api.StartPolling(ctx, startPollReq)
	if err != nil {
		t.Fatalf("Ошибка запуска опроса: %v", err)
	}
	log.Printf("Ответ сервера: %s\n", startMsg.Message)

	// Даем опросу поработать
	log.Println("Ожидание 3 секунды, пока идет опрос...")
	time.Sleep(3 * time.Second)

	// 5. Остановка опроса
	log.Println("Шаг 5: Остановка опроса данных...")
	stopMsg, _, err := api.StopPolling(ctx, sessionID)
	if err != nil {
		t.Fatalf("Ошибка остановки опроса: %v", err)
	}
	log.Printf("Ответ сервера: %s\n", stopMsg.Message)

	// 6. Удаление подключения
	log.Println("Шаг 6: Удаление подключения...")
	deleteMsg, _, err := api.DeleteConnection(ctx, sessionID)
	if err != nil {
		t.Fatalf("Ошибка удаления подключения: %v", err)
	}
	log.Printf("Ответ сервера: %s\n", deleteMsg.Message)

	log.Println("Тест успешно завершен!")
}
