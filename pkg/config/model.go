package config

type Configuration struct {
	Server  ServerConfig  `json:"server"`
	Logging LoggingConfig `json:"logging"`
}

type ServerConfig struct {
	Address    string `json:"address"`
	Port       string `json:"port"`
	Production bool   `json:"production"`
}

type LoggingConfig struct {
	//Enabled bool   `json:"enabled"`
	Level string `json:"level"`
	File  string `json:"file"`
}
