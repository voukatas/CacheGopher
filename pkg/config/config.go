package config

import (
	"encoding/json"
	"os"
)

func LoadConfig(configFile string) (*Configuration, error) {
	file, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config Configuration
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
