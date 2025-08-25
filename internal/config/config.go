package config

import (
	"encoding/json"
	"os"
)

// AppConfig содержит конфигурацию приложения
type AppConfig struct {
	ServerPort   string   `json:"server_port"`
	KafkaBrokers []string `json:"kafka_brokers"`
	KafkaTopic   string   `json:"kafka_topic"`
}

// LoadConfiguration загружает конфигурацию из файла
func LoadConfiguration() (*AppConfig, error) {
	var config AppConfig

	configFile, err := os.ReadFile("config.json")
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(configFile, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
