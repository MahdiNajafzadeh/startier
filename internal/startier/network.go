package startier

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
)

const BASE_NETWORK_PROTOCOL = "tcp"

type Network struct {
	listener net.Listener
	packer   Packer
}

var _network *Network

func GetNetwork() *Network {
	return _network
}

func RunNetwork(ch chan error) {
	var err error
	if GetNetwork() == nil {
		_network, err = NewNetwork()
	}
	if err != nil {
		ch <- err
		return
	}
	c := GetConfig()
	n := GetNetwork()
	n.listener, err = net.Listen(BASE_NETWORK_PROTOCOL, c.Listen)
	if err != nil {
		log.Println(err)
		ch <- err
		return
	}
	log.Printf("NETWORK : RUNNING ON %s://%s", n.listener.Addr().Network(), n.listener.Addr().String())
	go n.LoopAccept()
}

func NewNetwork() (*Network, error) {
	return &Network{
		packer: Packer{},
	}, nil
}

func (n *Network) LoopAccept() {
	for {
		conn, err := n.listener.Accept()
		if err != nil {
			continue
		}
		go n.LoopHandle(conn)
	}
}

func (n *Network) LoopHandle(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 2048)
	db := GetDatabase()
	log.Println(conn.RemoteAddr().String())
	for {
		bufLen, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			continue
		}
		switch id := buf[0]; ID(id) {
		case ID_JOIN:
			{
				var msg JoinMessage
				_, err := n.packer.Unpack(buf[:bufLen], &msg)
				if err != nil {
					log.Printf("error in json message handle : %v", err)
					break
				}
				log.Printf("%#+v", msg.Node.ID)
				if _, ok := db.GetNode(msg.Node.ID); !ok {
					log.Println("node no exsit")
					db.AddNode(Node{
						ID:        msg.Node.ID,
						Port:      msg.Node.Port,
						Address:   msg.Node.Address,
						Addresses: msg.Node.Remotes,
					})
				}
			}
		default:
			{
				log.Printf("unknown id : %b", id)
			}
		}
	}
}

func (n *Network) Request(target interface{}, id ID, v interface{}) error {
	log.Printf("%T : %+v", target, target)
	db := GetDatabase()
	switch t := target.(type) {
	case net.Conn:
		data, err := n.packer.Pack(id, v)
		if err != nil {
			return err
		}
		_, err = t.Write(data)
		return err
	case string:
		if node, ok := db.GetNode(t); ok {
			return n.Request(node, id, v)
		} else if remote, ok := db.GetRemote(t); ok {
			return n.Request(remote, id, v)
		} else {
			host, _, err := net.SplitHostPort(t)
			if err != nil {
				return err
			}
			resolvedIPs, err := net.LookupHost(host)
			if err != nil || len(resolvedIPs) == 0 {
				return fmt.Errorf("unable to resolve address: %v", err)
			}
			conn, err := net.Dial(BASE_NETWORK_PROTOCOL, t)
			if err != nil {
				return err
			}
			defer conn.Close()
			return n.Request(conn, id, v)
		}
	case Node:
		address := fmt.Sprintf("%s:%d", t.Address, t.Port)
		return n.Request(address, id, v)
	case Remote:
		return n.Request(t.Address, id, v)
	default:
		return fmt.Errorf("unsupported target type: %T", t)
	}
}

func (n *Network) MakeConnection(address string) (net.Conn, error) {
	return nil, nil
}

func (n *Network) GetAddress() ([]string, int, error) {
	result := []string{}
	_, port, err := net.SplitHostPort(n.listener.Addr().String())
	if err != nil {
		return result, 0, err
	}
	iaddrs, err := net.InterfaceAddrs()
	if err != nil {
		return result, 0, err
	}
	for _, v := range iaddrs {
		ip, _, err := net.ParseCIDR(v.String())
		if err != nil || ip.IsLoopback() || ip.To4() == nil {
			continue
		}
		result = append(result, fmt.Sprintf("%s:%s", ip.To4().String(), port))
	}
	nport, err := strconv.Atoi(port)
	if err != nil {
		return result, 0, err
	}
	return result, nport, err
}
