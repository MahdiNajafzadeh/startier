package easynode

const (
	ID_CONFLICT int = iota
	ID_WHO
	ID_PACKET
	ID_JOIN
	ID_INFO
	ID_TUNNEL
)

type BaseNodeID struct {
	NodeID string `msgp:"node_id" json:"node_id"`
}
type WhoMessage struct {
	Token string `msgp:"token" json:"token"`
	Me    string `msgp:"me" json:"me"`
	You   string `msgp:"you" json:"you"`
}
type JoinMessage struct {
	NodeID    string    `msgp:"node_id" json:"node_id"`
	Addresses []Address `msgp:"addresses" json:"addresses"`
}
type InfoMessage struct {
	NodeID    string    `msgp:"node_id" json:"node_id"`
	Addresses []Address `msgp:"addresses" json:"addresses"`
}
type PacketMessage struct {
	NodeID  string `msgp:"node_id" json:"node_id"`
	Target  string `msgp:"target" json:"target"`
	TTL     int    `msgp:"ttl" json:"ttl"`
	Payload []byte `msgp:"payload" json:"payload"`
}

// type TunnelMessage struct {
// 	NodeID string `msgp:"node_id" json:"node_id"`
// 	From   string `msgp:"from" json:"from"`
// 	To     string `msgp:"to" json:"to"`
// 	ID     string `msgp:"id" json:"id"`
// }
