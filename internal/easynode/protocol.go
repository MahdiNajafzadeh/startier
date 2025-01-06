package easynode

const (
	ID_CONFLICT int = iota
	ID_WHO
	ID_INFO
	ID_JOIN
	ID_PACKET
	ID_BC_INFO
)

type WhoMessage struct {
	Token    string `msgp:"token" json:"token"`
	Sender   string `msgp:"sender" json:"sender"`
	Receiver string `msgp:"receiver" json:"receiver"`
}
type JoinMessage struct {
	ID        string    `msgp:"id"`
	Addresses []Address `msgp:"addresses"`
}
type Entity[T any] struct {
	Create []T `msgp:"create"`
	Update []T `msgp:"update"`
	Delete []T `msgp:"delete"`
}
type InfoMessage struct {
	Node    Entity[Node]    `msgp:"node"`
	Address Entity[Address] `msgp:"address"`
	Edge    Entity[Edge]    `msgp:"edge"`
}
type PacketMessage struct {
	FromNode string `msgp:"from_node" json:"from_node"`
	ToNode   string `msgp:"to_node" json:"to_node"`
	TTL      int    `msgp:"ttl" json:"ttl"`
	Payload  []byte `msgp:"payload" json:"payload"`
}

func NewInfoMessage() InfoMessage {
	return InfoMessage{
		Node: Entity[Node]{
			Create: []Node{},
			Update: []Node{},
			Delete: []Node{},
		},
		Address: Entity[Address]{
			Create: []Address{},
			Update: []Address{},
			Delete: []Address{},
		},
		Edge: Entity[Edge]{
			Create: []Edge{},
			Update: []Edge{},
			Delete: []Edge{},
		},
	}
}
