// package startier

// import (
// 	"log"
// 	"time"

// 	"github.com/hashicorp/go-memdb"
// )

// type Address struct {
// 	ID        string `json:"id" msgp:"id"`
// 	NodeID    string `json:"node_id" msgp:"node_id"`
// 	IPMask    string `json:"ip_mask" msgp:"ip_mask"`
// 	HostPort  string `json:"host_port" msgp:"host_port"`
// 	IsPrivate bool   `json:"is_private" msgp:"is_private"`
// }

// type Connection struct {
// 	ID       string
// 	IsTunnel bool
// }

// var _schema = &memdb.DBSchema{
// 	Tables: map[string]*memdb.TableSchema{
// 		"address": {
// 			Name: "address",
// 			Indexes: map[string]*memdb.IndexSchema{
// 				"id": {
// 					Name:    "id",
// 					Unique:  true,
// 					Indexer: &memdb.UUIDFieldIndex{Field: "ID"},
// 				},
// 				"node_id": {
// 					Name:    "node_id",
// 					Unique:  false,
// 					Indexer: &memdb.StringFieldIndex{Field: "NodeID"},
// 				},
// 				"ip_mask": {
// 					Name:    "ip_mask",
// 					Unique:  false,
// 					Indexer: &memdb.StringFieldIndex{Field: "IPMask"},
// 				},
// 				"host_port": {
// 					Name:    "host_port",
// 					Unique:  false,
// 					Indexer: &memdb.StringFieldIndex{Field: "HostPort"},
// 				},
// 				"is_private": {
// 					Name:    "is_private",
// 					Unique:  false,
// 					Indexer: &memdb.BoolFieldIndex{Field: "IsPrivate"},
// 				},
// 			},
// 		},
// 		"connection": {
// 			Name: "connection",
// 			Indexes: map[string]*memdb.IndexSchema{
// 				"id": {
// 					Name:    "id",
// 					Unique:  true,
// 					Indexer: &memdb.UUIDFieldIndex{Field: "ID"},
// 				},
// 				"is_tunnel": {
// 					Name:    "is_tunnel",
// 					Unique:  false,
// 					Indexer: &memdb.BoolFieldIndex{Field: "IsTunnel"},
// 				},
// 			},
// 		},
// 	},
// }

// var _db *memdb.MemDB

// func GetDatabase() *memdb.MemDB {
// 	return _db
// }

// func init() {
// 	db, err := memdb.NewMemDB(_schema)
// 	if err != nil {
// 		log.Panicln("new db : ", err)
// 	}
// 	_db = db
// }

// func MonitorDatabase() {
// 	for {
// 		time.Sleep(time.Second * 5)
// 		// addresses, _ := GetAllEntry[*Address]("address", "id")
// 		// log.Println("---")
// 		// for _, address := range addresses {
// 		// 	log.Printf("%+v", address)
// 		// }
// 	}
// }

// func CreateEntry(table string, obj interface{}) error {
// 	db := GetReady(GetDatabase)
// 	txn := db.Txn(true)
// 	defer txn.Abort()
// 	err := txn.Insert(table, obj)
// 	if err != nil {
// 		return err
// 	} else {
// 		txn.Commit()
// 		return nil
// 	}
// }

// func DeleteEntry(table string, obj interface{}) error {
// 	db := GetReady(GetDatabase)
// 	txn := db.Txn(true)
// 	defer txn.Abort()
// 	err := txn.Delete(table, obj)
// 	if err != nil {
// 		return err
// 	} else {
// 		txn.Commit()
// 		return nil
// 	}
// }

// func UpdateEntry(table string, old interface{}, new interface{}) error {
// 	db := GetReady(GetDatabase)
// 	txn := db.Txn(true)
// 	defer txn.Abort()
// 	err := txn.Delete(table, old)
// 	if err != nil {
// 		return err
// 	}
// 	err = txn.Insert(table, new)
// 	if err != nil {
// 		return err
// 	} else {
// 		txn.Commit()
// 		return nil
// 	}
// }

// func GetAllEntry[T any](table string, index string, args ...interface{}) ([]T, error) {
// 	db := GetReady(GetDatabase)
// 	txn := db.Txn(false)
// 	defer txn.Abort()
// 	all := []T{}
// 	it, err := txn.Get(table, index, args...)
// 	if err != nil {
// 		return all, err
// 	}
// 	for obj := it.Next(); obj != nil; obj = it.Next() {
// 		all = append(all, obj.(T))
// 	}
// 	return all, err
// }

// func GetEntry[T any](table string, index string, args ...interface{}) (T, error) {
// 	db := GetReady(GetDatabase)
// 	txn := db.Txn(false)
// 	defer txn.Abort()
// 	raw, err := txn.First(table, index, args...)
// 	var zero T
// 	if err != nil {
// 		return zero, err
// 	}
// 	if raw == nil {
// 		return zero, nil
// 	}
// 	return raw.(T), nil
// }

package startier

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Address struct {
	ID        string `gorm:"type:uuid;primaryKey" msgp:"id"`
	NodeID    string `gorm:"index:unique_address_idx,unique" msgp:"node_id"`
	IPMask    string `gorm:"index:unique_address_idx,unique" msgp:"ip_mask"`
	HostPort  string `gorm:"index:unique_address_idx,unique" msgp:"host_port"`
	IsPrivate bool   `gorm:"index:unique_address_idx,unique" msgp:"is_private"`
}

var _db *gorm.DB

func init() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	_db = db
	_db.AutoMigrate(&Address{})
}

func GetDatabase() *gorm.DB {
	return _db
}
