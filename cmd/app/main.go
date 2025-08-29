// @title MTConnect Service API
// @version 1.0.0
// @description API для работы с протоколом MTConnect и отправки данных в Kafka.
// @host localhost:8080
// @BasePath /api/v1
package main

import "github.com/iwtcode/MTConnect/internal/app"

func main() {
	// Создаем и запускаем новый экземпляр приложения fx
	app.New().Run()
}
