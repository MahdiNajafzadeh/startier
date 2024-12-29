package startier

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/google/uuid"
)

type Config struct {
	NodeID string   `json:"node_id"`
	Local  string   `json:"local"`
	Listen string   `json:"listen"`
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
	err = _config.ToDatabase()
	return _config, err
}

func (c *Config) ToJSON() string {
	b, _ := json.Marshal(c)
	return string(b)
}

func (c *Config) ToDatabase() error {
	db := GetDatabase()
	_, port, err := net.SplitHostPort(c.Listen)
	if err != nil {
		return err
	}
	iaddrs, err := net.InterfaceAddrs()
	if err != nil {
		return err
	}
	lip, lipnet, err := net.ParseCIDR(c.Local)
	if err != nil {
		return err
	}
	if lip.To4() == nil || lip.IsLoopback() {
		return fmt.Errorf("unsupported this local ip address : %v", lip.String())
	}
	lipnet.IP = lip
	address := Address{
		ID:        uuid.NewString(),
		NodeID:    c.NodeID,
		IPMask:    lipnet.String(),
		HostPort:  net.JoinHostPort(lip.To4().String(), port),
		IsPrivate: true,
	}
	if err := db.Create(&address).Error; err != nil {
		return err
	}

	for _, v := range iaddrs {
		ip, ipnet, err := net.ParseCIDR(v.String())
		if err != nil {
			return err
		}
		if ip.To4() == nil || ip.IsLoopback() || ip.Equal(lip) {
			continue
		}
		ipnet.IP = ip
		address := Address{
			ID:        uuid.NewString(),
			NodeID:    c.NodeID,
			IPMask:    ipnet.String(),
			HostPort:  net.JoinHostPort(ip.To4().String(), port),
			IsPrivate: false,
		}
		if err := db.Create(&address).Error; err != nil {
			return err
		}
	}
	return nil
}
