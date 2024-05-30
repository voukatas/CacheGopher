package config

type Configuration struct {
	ClientConfig ClientConfig   `json:"clientConf"`
	Common       Common         `json:"common"`
	Servers      []ServerConfig `json:"servers"`
	Logging      LoggingConfig  `json:"logging"`
}

type Common struct {
	Production     bool   `json:"production"`
	MaxSize        int    `json:"max_size"`
	EvictionPolicy string `json:"eviction_policy"`
}

type ServerConfig struct {
	ID          string   `json:"id"`
	Address     string   `json:"address"`
	Role        string   `json:"role"`
	Secondaries []string `json:"secondaries,omitempty"`
	Primary     string   `json:"primary,omitempty"`
}

type LoggingConfig struct {
	//Enabled bool   `json:"enabled"`
	Level string `json:"level"`
	File  string `json:"file"`
}

type ClientConfig struct {
	ConnectionTimeout int `json:"connectionTimeout"`
	KeepAliveInterval int `json:"keepAliveInterval"`
	UnHealthyInterval int `json:"unHealthyInterval"`
}
