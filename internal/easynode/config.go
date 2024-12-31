package easynode

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"

	"github.com/google/uuid"
)

type Config struct {
	NodeID string   `json:"node_id" msgp:"node_id"`
	Local  string   `json:"local"   msgp:"local"`
	Listen string   `json:"listen"  msgp:"listen"`
	Peers  []string `json:"peers"   msgp:"peers"`
}

var _config *Config

func init() {
	_config = &Config{
		NodeID: uuid.NewString(),
		Local:  fmt.Sprintf("192.168.100.%d/24", rand.Intn(253)+1),
		Listen: ":5555",
		Peers:  []string{},
	}
}

func LoadConfig(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := json.NewDecoder(file).Decode(_config); err != nil {
		return err
	}
	err = _config.LoadDatabase()
	return err
}

func (c *Config) LoadDatabase() error {
	var err error
	err = c.LoadPrivateAddress()
	if err != nil {
		return err
	}
	err = c.LoadPublicAddress()
	if err != nil {
		return err
	}
	return err
}

func (c *Config) LoadPrivateAddress() error {
	_, port, err := net.SplitHostPort(c.Listen)
	if err != nil {
		return err
	}
	ip, _, err := net.ParseCIDR(c.Local)
	if err != nil {
		return err
	}
	if ip.To4() == nil {
		return fmt.Errorf("not support IPv6 for local address : %s", ip.String())
	}
	Load(_db)
	err = _db.Create(&Address{
		NodeID:    c.NodeID,
		IPMask:    ip.String(),
		HostPort:  net.JoinHostPort(ip.String(), port),
		IsPrivate: true,
	}).Error
	return err
}

func (c *Config) LoadPublicAddress() error {
	_, port, _ := net.SplitHostPort(c.Listen)
	lip, _, _ := net.ParseCIDR(c.Local)
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return err
	}
	for _, v := range addrs {
		ip, _, err := net.ParseCIDR(v.String())
		if err != nil {
			return err
		}
		if ip.To4() == nil || ip.IsLoopback() || lip.Equal(ip) {
			continue
		}
		Load(_db)
		err = _db.Create(&Address{
			NodeID:   _config.NodeID,
			IPMask:   ip.To4().String(),
			HostPort: net.JoinHostPort(ip.To4().String(), port),
		}).Error
		if err != nil {
			return err
		}
	}
	return nil
}
