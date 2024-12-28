package startier

import (
	"net"

	"github.com/google/uuid"
)

func JoinHandler(conn net.Conn, buf []byte) error {
	n := GetNetwork()
	var msg JoinMessage
	_, err := n.packer.Unpack(buf, &msg)
	if err != nil {
		return err
	}
	node, err := GetEntry[*Node]("node", "id", msg.Node.ID)
	if err != nil {
		return err
	}
	if node != nil {
		addresses, err := GetAllEntry[*Address]("address", "node_id", msg.Node.ID)
		if err != nil {
			return err
		}
		for _, v := range addresses {
			err = DeleteEntry("address", v)
			if err != nil {
				return err
			}
		}
	}
	err = CreateEntry("node", &Node{ID: msg.Node.ID})
	if err != nil {
		return err
	}
	for _, v := range msg.Address {
		v.ID = uuid.NewString()
		err = CreateEntry("address", &v)
		if err != nil {
			return err
		}
	}
	infoMsg := InfoMessage{Nodes: []Node{}, Address: []Address{}}
	nodes, err := GetAllEntry[*Node]("node", "id")
	if err != nil {
		return err
	}
	for _, v := range nodes {
		if v.ID == msg.Node.ID {
			return err
		}
		infoMsg.Nodes = append(infoMsg.Nodes, *v)
	}
	addresses, err := GetAllEntry[*Address]("address", "host")
	if err != nil {
		return err
	}
	for _, v := range addresses {
		if v.NodeID == msg.Node.ID {
			return err
		}
		infoMsg.Address = append(infoMsg.Address, *v)
	}
	err = n.Request(conn, ID_INFO, &infoMsg)
	if err != nil {
		return err
	}
	return nil
}
