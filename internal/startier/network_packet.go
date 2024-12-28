package startier

import "net"

func PacketHandler(conn net.Conn, buf []byte) error {
	n := GetNetwork()
	tun := GetTun()
	var msg PacketMessage
	_, err := n.packer.Unpack(buf, &msg)
	if err != nil {
		return err
	}
	_, err = tun.Write(msg.Payload)
	if err != nil {
		return err
	}
	return nil
}
