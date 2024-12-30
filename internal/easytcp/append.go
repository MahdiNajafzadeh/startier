/*

MAHDI NAJAFZADEH

*/

package easytcp

import (
	"io"
	"net"
)

func (s *Server) AppendConn(conn net.Conn) {
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
	s.AppendConn(conn)
	return nil
}

func (s *Server) Tunnel(from, to Session) {
	io.Copy(to.Conn(), from.Conn())
}
