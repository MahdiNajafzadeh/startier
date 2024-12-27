package startier

import (
	"encoding/json"
	"os"
)

type Config struct {
	NodeID  string   `json:"node_id"`
	Address string   `json:"address"`
	Listen  string   `json:"listen"`
	Peers   []string `json:"peers"`
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

func (c *Config) ToJSON() string {
	b, _ := json.Marshal(c)
	return string(b)
}
