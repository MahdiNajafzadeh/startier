package easynode

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
)

type Config struct {
	NodeID string   `json:"node_id" msgp:"node_id"`
	Local  string   `json:"local"   msgp:"local"`
	Listen string   `json:"listen"  msgp:"listen"`
	Peers  []string `json:"peers"   msgp:"peers"`
}

var _config *Config

func LoadConfig(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return err
	}
	_config = &config
	err = _config.LoadDatabase()
	return err
}

func (c *Config) LoadDatabase() error {
	if err := c.LoadPrivateAddress(); err != nil {
		return err
	}
	if err := c.LoadPublicAddress(); err != nil {
		return err
	}
	return nil
}

func (c *Config) LoadPrivateAddress() error {
	_, port, err := net.SplitHostPort(c.Listen)
	if err != nil {
		return err
	}
	ip, ipnet, err := net.ParseCIDR(c.Local)
	if err != nil {
		return err
	}
	if ip.To4() == nil || ip.IsLoopback() {
		return fmt.Errorf("not support IPv6 for local address : %s", ip.String())
	}
	ipnet.IP = ip
	Load(_db)
	err = _db.Create(&Address{
		NodeID:    c.NodeID,
		IPMask:    ipnet.String(),
		HostPort:  net.JoinHostPort(ip.String(), port),
		IsPrivate: true,
	}).Error
	return err
}

func (c *Config) LoadPublicAddress() error {
	_, port, _ := net.SplitHostPort(c.Listen)
	localIP, _, _ := net.ParseCIDR(c.Local)
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return err
	}
	for _, v := range addrs {
		ip, ipnet, err := net.ParseCIDR(v.String())
		if err != nil {
			return err
		}
		if ip.To4() == nil || ip.IsLoopback() || localIP.Equal(ip) {
			continue
		}
		ipnet.IP = ip 
		Load(_db)
		err = _db.Create(&Address{
			NodeID:   _config.NodeID,
			IPMask:   ipnet.String(),
			HostPort: net.JoinHostPort(ip.To4().String(), port),
		}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) ToJSON() string {
	b, _ := json.Marshal(c)
	return string(b)
}
