package config

import (
	"encoding/json"
	"os"
)

// Список конечных точек
type AppConfig struct {
	Endpoints []string `json:"endpoints"`
}

// Парсинг конфигурационного файла
func LoadConfiguration(filePath string) (*AppConfig, error) {
	var config AppConfig

	configFile, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(configFile, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
