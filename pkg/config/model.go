package config

type Configuration struct {
	Server  ServerConfig  `json:"server"`
	Logging LoggingConfig `json:"logging"`
}

type ServerConfig struct {
	Address        string `json:"address"`
	Port           string `json:"port"`
	Production     bool   `json:"production"`
	MaxSize        int    `json:"max_size"`
	EvictionPolicy string `json:"eviction_policy"`
}

type LoggingConfig struct {
	//Enabled bool   `json:"enabled"`
	Level string `json:"level"`
	File  string `json:"file"`
}
