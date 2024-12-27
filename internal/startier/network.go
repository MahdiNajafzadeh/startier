package startier

import (
	"fmt"
	"io"
	"log"
	"net"
)

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
	n.listener, err = net.Listen("tcp", c.Listen)
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
				if node, ok := db.GetNode(msg.Node.ID); ok {
					log.Println("node exsit")
					db.UpdateNode(node.ID, Node{
						ID:        node.ID,
						Port:      msg.Node.Port,
						Address:   msg.Node.Address,
						Addresses: msg.Node.Remotes,
					})
				} else {
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
	switch target := target.(type) {
	case net.Conn:
		{
			data, err := n.packer.Pack(id, v)
			if err != nil {
				return err
			}
			_, err = target.Write(data)
			if err != nil {
				return err
			}
		}
	case string:
		{
			host, _, err := net.SplitHostPort(target)
			if err != nil {
				return err
			}
			r, err := net.LookupHost(host)
			log.Println(r, err)
			conn, err := net.Dial("tcp", target)
			if err != nil {
				return err
			}
			data, err := n.packer.Pack(id, v)
			if err != nil {
				return err
			}
			_, err = conn.Write(data)
			if err != nil {
				return err
			}
		}
	default:
		{
			return fmt.Errorf("unknown type for target : %T", target)
		}
	}
	return nil
}

// func (n *Network) MakeConnection(address string) (net.Conn, error) {
// }

func (n *Network) GetAddress() ([]string, error) {
	result := []string{}
	_, port, err := net.SplitHostPort(n.listener.Addr().String())
	if err != nil {
		return result, err
	}
	iaddrs, err := net.InterfaceAddrs()
	if err != nil {
		return result, err
	}
	for _, v := range iaddrs {
		ip, _, err := net.ParseCIDR(v.String())
		if err != nil || ip.IsLoopback() || ip.To4() == nil {
			continue
		}
		result = append(result, fmt.Sprintf("%s:%s", ip.To4().String(), port))
	}
	return result, err
}
