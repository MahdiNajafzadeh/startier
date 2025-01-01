package easynode

var _store *Store

type Store struct {
	session SessionStore
	tunnel  TunnelStore
}
type SessionStore struct {
	node_to_session map[string]Session
	session_to_node map[string]Session
}
type TunnelStore struct {
	node_to_tunnel    map[string]Tunnel
	session_to_tunnel map[string]Tunnel
}

func init() {
	_store = &Store{
		session: SessionStore{
			node_to_session: map[string]Session{},
			session_to_node: map[string]Session{},
		},
		tunnel: TunnelStore{
			node_to_tunnel:    map[string]Tunnel{},
			session_to_tunnel: map[string]Tunnel{},
		},
	}
}
