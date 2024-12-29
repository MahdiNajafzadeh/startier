package startier

const (
	ID_CONFLICT int = iota
	ID_JOIN
	ID_JOIN_CAST
	ID_INFO
	ID_PACKET
	ID_TUNNEL
)

type JoinMessage struct {
	Addresses []Address `msgp:"addresses"`
}
type InfoMessage struct {
	Addresses []Address `msgp:"addresses"`
}
type JoinCastMessage struct {
	Addresses []Address `msgp:"addresses"`
}
type PacketMessage struct {
	Payload []byte `msgp:"payload"`
}
type TunnelMessage struct {
	Address Address `msgp:"address"`
}
