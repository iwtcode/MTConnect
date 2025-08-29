package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// AppConfig содержит конфигурацию приложения
type AppConfig struct {
	ServerPort  string
	KafkaBroker string
	KafkaTopic  string
	GinMode     string
	Database    DatabaseConfig
	Logging     LoggerConfig
}

// LoggerConfig содержит настройки логгера
type LoggerConfig struct {
	Enable     bool
	LogsDir    string
	Level      string
	SavingDays int
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
		Logging: LoggerConfig{ // Добавлено
			Enable:     getEnvAsBool("LOGGER_ENABLE", true),
			LogsDir:    getEnv("LOGGER_LOGS_DIR", "./logs"),
			Level:      getEnv("LOGGER_LOG_LEVEL", "DEBUG"),
			SavingDays: getEnvAsInt("LOGGER_SAVING_DAYS", 7),
		},
	}

	return config, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(name string, defaultValue int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	val, _ := strconv.ParseBool(value)
	return val
}
