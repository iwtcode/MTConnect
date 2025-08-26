package config

import (
	"os"

	"github.com/joho/godotenv"
)

// AppConfig содержит конфигурацию приложения
type AppConfig struct {
	ServerPort  string
	KafkaBroker string // Изменено с KafkaBrokers []string
	KafkaTopic  string
	GinMode     string
}

// LoadConfiguration загружает конфигурацию из .env файла или переменных окружения
func LoadConfiguration() (*AppConfig, error) {
	// Загружаем .env файл. В случае ошибки (например, файл не найден),
	// мы все равно продолжаем, так как переменные могут быть установлены в окружении системы.
	_ = godotenv.Load()

	// KafkaBrokers теперь одна строка, разделенная запятыми
	config := &AppConfig{
		ServerPort:  getEnv("APP_PORT", "8080"),
		KafkaBroker: getEnv("KAFKA_BROKER", "localhost:9092"),
		KafkaTopic:  getEnv("KAFKA_TOPIC", "opc-data"),
		GinMode:     getEnv("GIN_MODE", "debug"),
	}

	return config, nil
}

// getEnv - вспомогательная функция для получения переменной окружения с fallback-значением
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
