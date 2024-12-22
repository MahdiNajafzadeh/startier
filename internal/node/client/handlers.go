package client

import (
	"fmt"
	"net"
	p "startier/internal/node/common/protocol"
	"startier/internal/node/database/models"

	"github.com/DarthPestilane/easytcp"
)

func (c *Client) decode(msg *easytcp.Message, v interface{}) error {
	if err := c.codec.Decode(msg.Data(), v); err != nil {
		return err
	}
	return nil
}

func (c *Client) handle_info(key string, conn net.Conn, msg *easytcp.Message) error {
	res := &p.InfoRes{}
	if err := c.decode(msg, res); err != nil {
		return err
	}
	for _, n := range res.Nodes {
		node := models.Node{
			Hostname: n.GetHostname(),
			Domain:   n.GetDomain(),
		}
		if c.conns[fmt.Sprintf("%s.%s", n.Hostname, n.Domain)] == nil {
			c.database.RegisterNode(&node)
		}
		for _, r := range n.Remotes {
			c.database.RegisterRemote(&n, &models.Remote{
				Host: r.GetHost(),
				Port: int(r.GetPort()),
				TLS : r.GetTLS(),
			})
		}
		for _, a := range n.Addresses {
			c.database.RegisterAddress(&n, &models.Address{
				Address: a.GetAddress(),
				Mask: int(a.GetMask()),
			})
		}
	}
	return nil
}
func (c *Client) handle_packet(key string, conn net.Conn, msg *easytcp.Message) error {
	res := &p.PacketRes{}
	if err := c.decode(msg, res); err != nil {
		return err
	}
	
	return nil
}
func (c *Client) handle_tunnel(key string, conn net.Conn, msg *easytcp.Message) error {
	res := &p.TunnelRes{}
	if err := c.decode(msg, res); err != nil {
		return err
	}
	return nil
}
func (c *Client) handle_test(key string, conn net.Conn, msg *easytcp.Message) error {
	res := &p.TestRes{}
	if err := c.decode(msg, res); err != nil {
		return err
	}
	return nil
}
