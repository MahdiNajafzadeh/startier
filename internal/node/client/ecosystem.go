package client

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	p "startier/internal/node/common/protocol"
	ctls "startier/internal/node/common/tls"
	"startier/internal/node/database/models"
	"strconv"
	"sync"
	"time"

	"github.com/DarthPestilane/easytcp"
)

func (c *Client) PreRun() error {
	remotes := []models.Remote{}
	for _, v := range c.config.Remotes {
		remote, err := parseURL(v)
		if err != nil {
			return err
		}
		remotes = append(remotes, *remote)
	}
	go c.loadRemotes(remotes)
	return nil
}

func (c *Client) loadRemotes(remotes []models.Remote) {
	for _, remote := range remotes {
		addr := HostToAddr(remote.Host, remote.Port)
		if remote.TLS && c.config.Server.TLS.Enable {
			tlsConfig, err := ctls.GetTLSConfig(c.config)
			if err != nil {
				println(err.Error())
				continue
			}
			conn, err := tls.Dial("tcp", addr, tlsConfig)
			if err != nil {
				println(err.Error())
				continue
			}
			c.conns[UnRegisterForm(conn)] = conn
		} else {
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				println(err.Error())
				continue
			}
			c.conns[UnRegisterForm(conn)] = conn
		}
	}
}

func (c *Client) runReader() {
	var wg sync.WaitGroup
	for {
		for key, conn := range c.conns {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					msg, err := c.packer.Unpack(conn)
					if err != nil {
						println(err.Error())
						continue
					}
					switch msg.ID() {
					case p.MSG_ID_TUNNEL_RES:
						c.handle_tunnel(key, conn, msg)
					case p.MSG_ID_PACKET_RES:
						c.handle_packet(key, conn, msg)
					case p.MSG_ID_INFO_RES:
						c.handle_info(key, conn, msg)
					case p.MSG_ID_TEST_RES:
						c.handle_test(key, conn, msg)
					default:
						continue
					}
				}
			}()
		}
		wg.Wait()
	}
}

func (c *Client) runTester() {
	for {
		for _, conn := range c.conns {
			c.write(conn, p.MSG_ID_TEST_REQ, &p.TestReq{})
			time.Sleep(time.Second * 10)
		}
	}
}

func (c *Client) write(conn net.Conn, id, v interface{}) error {
	data, err := c.codec.Encode(v)
	if err != nil {
		return err
	}
	msg, err := c.packer.Pack(easytcp.NewMessage(id, data))
	if err != nil {
		return err
	}
	if _, err := conn.Write(msg); err != nil {
		return err
	}
	return nil
}

func parseURL(rawURL string) (*models.Remote, error) {
	parseURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("error in parse peers: %v", err)
	}
	schema := parseURL.Scheme
	if schema != "startier" {
		return nil, fmt.Errorf(
			"error in parse peers: protocol is not 'startier', protocol: '%s'",
			schema,
		)
	}
	host, portString, err := net.SplitHostPort(parseURL.Host)
	if err != nil {
		return nil, fmt.Errorf("error in parse peers: can't parse host & port: %v", err)
	}
	port, err := strconv.Atoi(portString)
	if err != nil {
		return nil, fmt.Errorf("error in parse peers: can't parse port: %v", err)
	}
	tls := parseURL.Query().Get("tls")
	return &models.Remote{Host: host, Port: port, TLS: tls == "enable"}, nil
}

func UnRegisterForm(conn net.Conn) string {
	return conn.RemoteAddr().String()
}

func HostToAddr(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}
