package client

import (
	"net"
	"startier/config"
	"startier/internal/node/database"

	"github.com/DarthPestilane/easytcp"
)

type Client struct {
	config   *config.Config
	database *database.Database
	conns    map[string]net.Conn
	packer   easytcp.Packer
	codec    easytcp.Codec
}

func New(conf *config.Config, db *database.Database) (*Client, error) {
	cl := &Client{
		config:   conf,
		database: db,
		packer:   easytcp.NewDefaultPacker(),
		codec:    &easytcp.ProtobufCodec{},
		conns:    make(map[string]net.Conn),
	}
	return cl, nil
}

func (c *Client) Run(ch chan error) {
	err := c.PreRun()
	if err != nil {
		ch <- err
	}
	go c.runReader()
	go c.runTester()
}
