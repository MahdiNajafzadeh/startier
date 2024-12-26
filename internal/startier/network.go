package startier

import (
	"io"
	"log"
	"net"
)

type Network struct {
	listener net.Listener
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
	config := GetConfig()
	_network.listener, err = net.Listen("tcp", ":"+config.Port)
	if err != nil {
		ch <- err
		return
	}
	n := _network.listener.Addr().Network()
	a := _network.listener.Addr().String()
	log.Printf("NETWORK : RUNNING ON %s://%s", n, a)
	go LoopAccept()
	go LoadRemotes()
}

func NewNetwork() (*Network, error) {
	return &Network{}, nil
}

func LoopAccept() {
	n := GetNetwork()
	for {
		conn, err := n.listener.Accept()
		if err != nil {
			continue
		}
		go LoopHandle(conn)
	}
}

func LoopHandle(conn net.Conn) {
	for {
		buf, err := io.ReadAll(conn)
		if err != nil {
			continue
		}
		log.Printf("%+v", buf)
	}
}

func LoadRemotes() {
	
}