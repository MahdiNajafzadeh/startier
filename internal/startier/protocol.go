package startier

import "github.com/vmihailenco/msgpack/v5"

type ID byte

const (
	ID_RES_CANCEL ID = iota
	ID_RES_ACCEPT
	ID_JOIN
	ID_INFO
	ID_PACKET
	ID_TUNNEL
)

type NodeData struct {
	ID      string   `msgp:"id"`
	Address string   `msgp:"address"`
	Port    int      `msgp:"port"`
	Remotes []string `msgp:"remotes"`
}

// -

type JoinMessage struct {
	Node NodeData `msgp:"nodes"`
}
type InfoMessage struct {
	Nodes []NodeData `msgp:"nodes"`
}
type PacketMessage struct {
	Payload []byte `msgp:"payload"`
}
type TunnelMessage struct {
	Address string `msgp:"address"`
}

//-

type Packer struct{}

func (p *Packer) Pack(id ID, v interface{}) ([]byte, error) {
	buf, err := msgpack.Marshal(v)
	buf = append([]byte{byte(id)}, buf...)
	return buf, err
}
func (p *Packer) Unpack(buf []byte, v interface{}) (ID, error) {
	id := ID(buf[0])
	err := msgpack.Unmarshal(buf[1:], v)
	return id, err
}
