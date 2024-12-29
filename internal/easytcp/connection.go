/*

MAHDI NAJAFZADEH

*/

package easytcp

import (
	"io"
	"net"
)

func (s *Server) AddConnection(conn net.Conn) {
	go s.handleConn(conn)
}

func (s *Server) NewRequest(address string, id interface{}, v interface{}) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	data, err := s.Codec.Encode(v)
	if err != nil {
		return err
	}
	packet, err := s.Packer.Pack(NewMessage(id, data))
	if err != nil {
		return err
	}
	_, err = conn.Write(packet)
	if err != nil {
		return err
	}
	s.AddConnection(conn)
	return nil
}

func (s *Server) TunnelConnection(from, to Session) {
	io.Copy(to.Conn(), from.Conn())
}
