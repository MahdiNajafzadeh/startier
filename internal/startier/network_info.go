package startier

import (
	"net"
)

func InfoHandler(conn net.Conn, buf []byte) error {
	n := GetNetwork()
	var msg InfoMessage
	_, err := n.packer.Unpack(buf, &msg)
	if err != nil {
		return err
	}
	for _, v := range msg.Nodes {
		err = CreateEntry("node", &v)
		if err != nil {
			return err
		}
	}
	for _, v := range msg.Address {
		err = CreateEntry("address", &v)
		if err != nil {
			return err
		}
	}
	return nil
}
