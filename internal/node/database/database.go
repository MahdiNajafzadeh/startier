package database

import (
	"fmt"
	"net"
	"startier/config"
	"startier/internal/node/database/models"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Database struct {
	config *config.Config
	db     *gorm.DB
}

func New(conf *config.Config) (*Database, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &Database{config: conf, db: db}, nil
}

func (d *Database) Run(ch chan error) {
	err := models.Migrate(d.db)
	if err != nil {
		ch <- err
	}
	err = d.PostRun()
	if err != nil {
		ch <- err
	}
}

func (d *Database) PostRun() error {
	node := models.Node{
		Hostname: d.config.Hostname,
		Domain:   d.config.Domain,
		IsMe:     true,
	}
	r := d.db.Create(&node)
	if r.Error != nil {
		return r.Error
	}
	for _, v := range d.config.Addresses {
		address := models.Address{
			NodeID: node.ID,
			IsMe:   true,
		}
		if strings.Contains(v, "/") {
			ip, ipnet, err := net.ParseCIDR(v)
			if err != nil {
				return fmt.Errorf("error parsing CIDR '%s' : %v", v, err)
			}
			address.Address = ip.To4().String()
			address.Mask, _ = ipnet.Mask.Size()
		} else {
			ip := net.ParseIP(v)
			if ip == nil {
				return fmt.Errorf("invalid IP address: '%s'", v)
			}
			if ip.To4() == nil {
				return fmt.Errorf("no support for IPv6 '%s'", v)
			}
			address.Address = ip.To4().String()
			address.Mask = 24
		}
		r := d.db.Create(&address)
		if r.Error != nil {
			return r.Error
		}
	}
	return nil
}
