package config

import (
	"os"

	"github.com/joho/godotenv"
)

// AppConfig содержит конфигурацию приложения
type AppConfig struct {
	ServerPort  string
	KafkaBroker string
	KafkaTopic  string
	GinMode     string
	Database    DatabaseConfig // Добавлено
}

// DatabaseConfig содержит конфигурацию для подключения к базе данных
type DatabaseConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
}

// LoadConfiguration загружает конфигурацию из .env файла или переменных окружения
func LoadConfiguration() (*AppConfig, error) {
	_ = godotenv.Load()

	config := &AppConfig{
		ServerPort:  getEnv("APP_PORT", "8080"),
		KafkaBroker: getEnv("KAFKA_BROKER", "localhost:9092"),
		KafkaTopic:  getEnv("KAFKA_TOPIC", "mtconnect_data"),
		GinMode:     getEnv("GIN_MODE", "debug"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Username: getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "root"),
			DBName:   getEnv("DB_NAME", "mtconnect_db"),
		},
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
