package startier

import (
	"fmt"
	"sync"
	"time"
)

type Node struct {
	ID        string   `json:"id"`
	Address   string   `json:"address"`
	Port      int      `json:"port"`
	Addresses []string `json:"addresses"`
}

type Remote struct {
	NodeID  string `json:"node_id"`
	Node    *Node  `json:"node"`
	Address string `json:"address"`
}

type Address struct {
	NodeID  string `json:"node_id"`
	Node    *Node  `json:"node"`
	Address string `json:"address"`
}

type DB struct {
	nodes_mu     *sync.RWMutex
	nodes        map[string]Node
	remotes_mu   *sync.RWMutex
	remotes      map[string]Remote
	addresses_mu *sync.RWMutex
	addresses    map[string]Address
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
		nodes_mu:     new(sync.RWMutex),
		nodes:        make(map[string]Node),
		remotes_mu:   new(sync.RWMutex),
		remotes:      make(map[string]Remote),
		addresses_mu: new(sync.RWMutex),
		addresses:    make(map[string]Address),
	}
}

func MonitorDatabase() {
	db := GetReady(GetDatabase)
	for {
		nodes := db.GetAllNodes()
		for _, node := range nodes {
			fmt.Printf("[%v] node : [%v][%v][%v][%v]\n", len(nodes), node.ID, node.Address, node.Port, node.Addresses)
		}
		time.Sleep(time.Second * 3)
	}
}
