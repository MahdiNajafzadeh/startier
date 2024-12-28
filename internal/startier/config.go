package startier

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"

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
	return _config, nil
}

func (c *Config) ToJSON() string {
	b, _ := json.Marshal(c)
	return string(b)
}

func (c *Config) ToDatabase() error {
	err := CreateEntry("node", &Node{ID: c.NodeID})
	if err != nil {
		return err
	}
	_, port, err := net.SplitHostPort(c.Listen)
	if err != nil {
		return err

	}
	nport, err := strconv.Atoi(port)
	if err != nil {
		return err

	}
	iaddrs, err := net.InterfaceAddrs()
	if err != nil {
		return err

	}
	localIP, ipnet, err := net.ParseCIDR(c.Local)
	if err != nil {
		return err

	}
	if localIP.To4() == nil {
		return fmt.Errorf("unsupport IPv6 for local : %v", localIP.String())
	}
	localMask, _ := ipnet.Mask.Size()
	err = CreateEntry("address",
		&Address{
			ID:        uuid.NewString(),
			NodeID:    c.NodeID,
			Host:      localIP.To4().String(),
			Mask:      localMask,
			Port:      nport,
			IsPrivate: true,
		},
	)
	if err != nil {
		return err

	}
	for _, v := range iaddrs {
		ip, ipnet, err := net.ParseCIDR(v.String())
		if err != nil {
			return err

		}
		if ip.To4() == nil || ip.IsLoopback() || ip.Equal(localIP) {
			continue
		}
		mask, _ := ipnet.Mask.Size()
		err = CreateEntry("address",
			&Address{
				ID:        uuid.NewString(),
				NodeID:    c.NodeID,
				Host:      ip.To4().String(),
				Mask:      mask,
				Port:      nport,
				IsPrivate: false,
			},
		)
		if err != nil {
			return err

		}
	}
	return nil
}
