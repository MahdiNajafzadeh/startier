package startier

const (
	ID_RES_CANCEL int = iota
	ID_RES_ACCEPT
	ID_RES_NOT_FOUND
	ID_JOIN
	ID_JOIN_CAST
	ID_INFO
	ID_PACKET
	ID_TUNNEL
)

// -

type JoinMessage struct {
	Address []Address `msgp:"addresses"`
}
type InfoMessage struct {
	Address []Address `msgp:"addresses"`
}
type PacketMessage struct {
	Payload []byte `msgp:"payload"`
}
type TunnelMessage struct {
	Address Address `msgp:"address"`
}
