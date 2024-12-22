package node

import (
	"startier/config"
	"startier/internal/node/client"
	"startier/internal/node/database"
	"startier/internal/node/server"
	"startier/internal/node/tun"
	"startier/internal/node/web"
)

type Node struct {
	config   *config.Config
	database *database.Database
	server   *server.Server
	client   *client.Client
	tun      *tun.Tun
	web      *web.Server
}

func New(conf *config.Config) (*Node, error) {
	n := &Node{config: conf}
	var err error
	n.database, err = database.New(n.config)
	if err != nil {
		return nil, err
	}
	n.client, err = client.New(n.config, n.database)
	if err != nil {
		return nil, err
	}
	n.tun, err = tun.New(n.config, n.database, n.client)
	if err != nil {
		return nil, err
	}
	n.server, err = server.New(n.config, n.database, n.client, n.tun)
	if err != nil {
		return nil, err
	}
	n.web, err = web.New(n.config, n.database)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (n *Node) Run() error {
	ch := make(chan error)
	defer close(ch)
	go n.database.Run(ch)
	go n.client.Run(ch)
	go n.tun.Run(ch)
	go n.server.Run(ch)
	go n.web.Run(ch)
	return <-ch
}
