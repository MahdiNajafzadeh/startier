package startier

import (
	"encoding/json"
	"os"
)

type Config struct {
	NodeID string   `json:"node_id"`
	IP     string   `json:"ip"`
	Port   string   `json:"port"`
	Peers  []string `json:"peers"`
}

var _config *Config

func GetConfig() *Config {
	return _config
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}
	_config = &config
	return _config, nil
}
