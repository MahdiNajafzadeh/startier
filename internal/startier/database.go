package startier

import "sync"

type Node struct {
	ID        string   `json:"id"`
	IP        string   `json:"ip"`
	Port      string   `json:"port"`
	Addresses []string `json:"addresses"`
}

type Remote struct {
	NodeID  string `json:"node_id"`
	Node    *Node  `json:"node"`
	Address string `json:"address"`
}

type DB struct {
	nodes_mu   *sync.RWMutex
	nodes      map[string]Node
	remotes_mu *sync.RWMutex
	remotes    map[string]Remote
}

var _db *DB

func GetDatabase() *DB {
	return _db
}

func RunDatabase(ch chan error) {
	if GetDatabase() == nil {
		_db = NewDatabase()
	}
}

func NewDatabase() *DB {
	return &DB{
		nodes_mu: new(sync.RWMutex),
		nodes:    make(map[string]Node),
	}
}
