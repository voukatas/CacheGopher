package config

import (
	"os"
	"testing"
)

func TestLoadConfigSuccess(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	configJSON := `

	{
"clientConf": {
	"connectionTimeout": 300,
	"keepAliveInterval": 15,
	"unHealthyInterval": 30
},
	
		"common": {
			"production": false,
			"max_size": 10000,
			"eviction_policy": "LRU"
		},
		"servers": [
	        {
			"id": "server_A1",
			"address": "localhost:31337",
			"role": "primary",
	                "secondaries": ["server_A3", "server_A2"]
		},

               {
                 "id": "server_A3",
                 "address": "localhost:31339",
                 "role": "secondary",
                 "primary": "server_A1"
               }
	       ],
		"logging": {
			"level": "debug",
			"file": "cacheGopherServer.log"
		}
	      }`
	if _, err := tmpfile.Write([]byte(configJSON)); err != nil {
		t.Fatal(err)
	}

	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	config, err := LoadConfig(tmpfile.Name())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if config.ClientConfig.ConnectionTimeout != 300 {
		t.Errorf("Expected ConnectionTimeout to be 300")
	}
	if config.ClientConfig.KeepAliveInterval != 15 {
		t.Errorf("Expected KeepAliveInterval to be 15")
	}
	if config.ClientConfig.UnHealthyInterval != 30 {
		t.Errorf("Expected UnHealthyInterval to be 30")
	}
	if config.Common.Production != false {
		t.Errorf("Expected Production to be false")
	}
	if config.Servers[0].ID != "server_A1" {
		t.Errorf("Expected server ID to be 'server_A1'")
	}

	if config.Servers[0].Role != "primary" {
		t.Errorf("Expected server Role to be 'primary'")
	}

	if config.Servers[0].Secondaries[0] != "server_A3" {
		t.Errorf("Expected server Secondary to be 'server_A3'")
	}

	if config.Servers[0].Secondaries[1] != "server_A2" {
		t.Errorf("Expected server Secondary to be 'server_A2'")
	}

	if config.Servers[1].ID != "server_A3" {
		t.Errorf("Expected server ID to be 'server_A3'")
	}

	if config.Servers[1].Role != "secondary" {
		t.Errorf("Expected server Role to be 'secondary'")
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := LoadConfig("nonexistent.json")
	if err == nil {
		t.Errorf("Expected error, got none")
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(`{ "common": "no closing brace"`)); err != nil {
		t.Fatal(err)
	}

	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	_, err = LoadConfig(tmpfile.Name())
	if err == nil {
		t.Errorf("Expected JSON unmarshal error, got none")
	}
}
