package config

import (
	"encoding/json"
	"fmt"
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

func GetPrimaryServerAddress(cfg *Configuration, primaryId string) (string, error) {
	//var primaryId string

	// find the current server and its primary serverId

	// for _, server := range cfg.Servers {
	// 	if server.ID == serverId {
	// 		if strings.ToUpper(server.Role) == "PRIMARY" {
	// 			return "", fmt.Errorf("server is primary")
	//
	// 		}
	// 		primaryId = server.Primary
	// 		break
	// 	}
	// }

	for _, server := range cfg.Servers {
		if primaryId == server.ID {
			return server.Address, nil
		}
	}

	return "", fmt.Errorf("primary server not found")
}
