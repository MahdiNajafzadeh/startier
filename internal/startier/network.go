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
	for {
		_, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Println("Connection closed by peer")
				conn.Close()
				break
			}
			log.Printf("Error reading data: %v", err)
			continue
		}
		id := ID(buf[0])
		switch id {
		case ID_INFO:
			{
				err = InfoHandler(conn, buf)
				if err != nil {
					log.Println(err)
				}
			}
		case ID_JOIN:
			{
				err = JoinHandler(conn, buf)
				if err != nil {
					log.Println(err)
				}
			}
		case ID_PACKET:
			{
				err = PacketHandler(conn, buf)
				if err != nil {
					log.Println(err)
				}
			}
		default:
			log.Printf("unknown id : %b", id)
		}
	}
	log.Printf("END OF LOOP HANDLE CONNECTION : %v", conn.RemoteAddr())
}

func (n *Network) Request(target interface{}, id ID, v interface{}) error {
	// db := GetDatabase()
	switch t := target.(type) {
	case net.Conn:
		data, err := n.packer.Pack(id, v)
		if err != nil {
			return err
		}
		_, err = t.Write(data)
		return err
	case string:
		// if node, ok := db.GetNode(t); ok {
		// 	return n.Request(node, id, v)
		// } else if remote, ok := db.GetRemote(t); ok {
		// 	return n.Request(remote, id, v)
		// } else {
		_, _, err := net.SplitHostPort(t)
		if err != nil {
			return err
		}
		conn, err := net.Dial(BASE_NETWORK_PROTOCOL, t)
		if err != nil {
			return err
		}
		go n.LoopHandle(conn)
		return n.Request(conn, id, v)
		// }
	// case Node:
	// 	address := fmt.Sprintf("%s:%d", t.Local, t.Port)
	// 	return n.Request(address, id, v)
	// case Remote:
	// 	return n.Request(t.Node.Local, id, v)
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
		result = append(result, net.JoinHostPort(ip.To4().String(), port))
	}
	nport, err := strconv.Atoi(port)
	if err != nil {
		return result, 0, err
	}
	return result, nport, err
}
