package startier

import (
	"log"
	"net"
	"time"

	"github.com/hashicorp/go-memdb"
)

type Node struct {
	ID string `json:"id" msgp:"id"`
}
type Address struct {
	ID string `json:"-" msgp:"-"`
	// Node      *Node  // `json:"-" msgp:"-"`
	NodeID    string `json:"node_id" msgp:"node_id"`
	Host      string `json:"host" msgp:"host"`
	Port      int    `json:"port" msgp:"port"`
	Mask      int    `json:"mask" msgp:"mask"`
	IsPrivate bool   `json:"is_private" msgp:"is_private"`
}
type Connection struct {
	Address  *Address
	Conn     *net.Conn
	IsTunnel bool
}

var _schema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		"node": {
			Name: "node",
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:    "id",
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "ID"},
				},
			},
		},
		"address": {
			Name: "address",
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:    "id",
					Unique:  true,
					Indexer: &memdb.UUIDFieldIndex{Field: "ID"},
				},
				"node_id": {
					Name:    "node_id",
					Unique:  false,
					Indexer: &memdb.StringFieldIndex{Field: "NodeID"},
				},
				"host": {
					Name:    "host",
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "Host"},
				},
				"port": {
					Name:    "port",
					Unique:  false,
					Indexer: &memdb.IntFieldIndex{Field: "Port"},
				},
				"mask": {
					Name:    "mask",
					Unique:  false,
					Indexer: &memdb.IntFieldIndex{Field: "Mask"},
				},
				"is_private": {
					Name:    "is_private",
					Unique:  false,
					Indexer: &memdb.BoolFieldIndex{Field: "IsPrivate"},
				},
			},
		},
		"connection": {
			Name: "connection",
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:    "id",
					Unique:  true,
					Indexer: &memdb.UUIDFieldIndex{Field: "ID"},
				},
				"address_host": {
					Name:    "address_host",
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "Address.Host"},
				},
				"is_tunnel": {
					Name:    "is_tunnel",
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "IsTunnel"},
				},
			},
		},
	},
}

var _db *memdb.MemDB

func GetDatabase() *memdb.MemDB {
	return _db
}

func RunDatabase() error {
	var err error
	if GetDatabase() == nil {
		_db, err = NewDatabase()
	}
	return err
}

func NewDatabase() (*memdb.MemDB, error) {
	return memdb.NewMemDB(_schema)
}

func MonitorDatabase() {
	for {
		time.Sleep(time.Second * 5)
		// nodes, _ := GetAllEntry[*Node]("node", "id")
		// for _, node := range nodes {
		// 	log.Printf("node [%v]", node.ID)
		// }
		addresses, _ := GetAllEntry[*Address]("address", "id")
		log.Println("---")
		for _, address := range addresses {
			log.Printf("address [%#+v]", address)
		}
	}
}

func CreateEntry(table string, obj interface{}) error {
	db := GetDatabase()
	txn := db.Txn(true)
	defer txn.Abort()
	err := txn.Insert(table, obj)
	if err != nil {
		return err
	} else {
		txn.Commit()
		return nil
	}
}

func DeleteEntry(table string, obj interface{}) error {
	db := GetDatabase()
	txn := db.Txn(true)
	defer txn.Abort()
	err := txn.Delete(table, obj)
	if err != nil {
		return err
	} else {
		txn.Commit()
		return nil
	}
}

func UpdateEntry(table string, id string, old interface{}, new interface{}) error {
	db := GetDatabase()
	txn := db.Txn(true)
	defer txn.Abort()
	err := txn.Delete(table, old)
	if err != nil {
		return err
	}
	err = txn.Insert(table, new)
	if err != nil {
		return err
	} else {
		txn.Commit()
		return nil
	}
}

func GetAllEntry[T any](table string, id string, args ...interface{}) ([]T, error) {
	db := GetDatabase()
	txn := db.Txn(false)
	defer txn.Abort()
	all := []T{}
	it, err := txn.Get(table, id, args...)
	if err != nil {
		return all, err
	}
	for obj := it.Next(); obj != nil; obj = it.Next() {
		all = append(all, obj.(T))
	}
	return all, err
}

func GetEntry[T any](table string, id string, args ...interface{}) (T, error) {
	db := GetDatabase()
	txn := db.Txn(false)
	defer txn.Abort()
	raw, err := txn.First(table, id, args...)
	var zero T
	if err != nil {
		return zero, err
	}
	if raw == nil {
		return zero, nil
	}
	return raw.(T), nil
}
