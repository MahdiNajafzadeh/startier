package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type TLSConfig struct {
	Enable bool `mapstructure:"enable"`
	Files  struct {
		Private string `mapstructure:"private"`
		Public  string `mapstructure:"public"`
	} `mapstructure:"files"`
}

type ServerConfig struct {
	Listen string    `mapstructure:"listen"`
	Port   int       `mapstructure:"port"`
	TLS    TLSConfig `mapstructure:"tls"`
}

type WebConfig struct {
	Enable bool   `mapstructure:"enable"`
	Listen string `mapstructure:"listen"`
	Port   int    `mapstructure:"port"`
}

type Config struct {
	Hostname  string       `mapstructure:"hostname"`
	Domain    string       `mapstructure:"domain"`
	Addresses []string     `mapstructure:"addresses"`
	Remotes   []string     `mapstructure:"remotes"`
	Server    ServerConfig `mapstructure:"server"`
	Web       WebConfig    `mapstructure:"web"`
}

func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	v.SetDefault("hostname", "main")
	v.SetDefault("domain", "startier.local")
	v.SetDefault("addresses", []string{})
	v.SetDefault("remotes", []string{})
	v.SetDefault("server.listen", "0.0.0.0")
	v.SetDefault("server.port", 5555)
	v.SetDefault("server.tls.enable", false)
	v.SetDefault("server.tls.files.private", "")
	v.SetDefault("server.tls.files.public", "")
	v.SetDefault("web.enable", false)
	v.SetDefault("web.listen", "0.0.0.0")
	v.SetDefault("web.port", 3000)

	v.AddConfigPath(configPath)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file, using defaults: %v", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error in unmarshaling config file: %v", err)
	}

	return &config, nil
}
