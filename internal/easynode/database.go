package easynode

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mitchellh/hashstructure/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	DB_FILE_PATH = "/tmp/easynode.db"
	DB_MEM_PATH  = "file::memory:?mode=memory&cache=shared"
)

var _db *gorm.DB

func init() {
	if _, err := os.Stat(DB_FILE_PATH); err == nil {
		if err := os.Remove(DB_FILE_PATH); err != nil {
			panic(err)
		}
	}
	var logLevel logger.LogLevel
	logLevel = logger.Error
	logLevel = logger.Info
	logLevel = logger.Silent
	db, err := gorm.Open(sqlite.Open(DB_FILE_PATH), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		_log.Error(err)
		panic(err)
	}
	err = db.AutoMigrate(&Address{}, &Node{}, &Connection{}, &Edge{})
	if err != nil {
		_log.Error(err)
		panic(err)
	}
	_db = db
}

type Node struct {
	ID string `gorm:"primaryKey" msgp:"id" json:"id"`
}

type Address struct {
	ID        string `gorm:"primaryKey" msgp:"id" json:"-" hash:"-"`
	NodeID    string `gorm:"index:unique_address_idx,unique" msgp:"node_id" json:"node_id"`
	IPMask    string `gorm:"index:unique_address_idx,unique" msgp:"ip_mask" json:"ip_mask"`
	HostPort  string `gorm:"index:unique_address_idx,unique" msgp:"host_port" json:"-"`
	IsPrivate bool   `gorm:"index:unique_address_idx,unique" msgp:"is_private" json:"is_private"`
}

type Edge struct {
	From string `gorm:"index:unique_edge_idx,unique" msgp:"from" json:"from"`
	To   string `gorm:"index:unique_edge_idx,unique" msgp:"to" json:"to"`
}

type Connection struct {
	SessionID string `grom:"primaryKey" msgp:"session_id" json:"session_id"`
	NodeID    string `msgp:"node_id" json:"node_id"`
}


//---

func (a *Address) JSON() string {
	b, _ := json.Marshal(a)
	return string(b)
}

func (a *Address) BeforeCreate(tx *gorm.DB) error {
	hash, _ := hashstructure.Hash(a, hashstructure.FormatV2, nil)
	a.ID = fmt.Sprintf("%d", hash)
	return nil
}

func (a *Address) AfterCreate(tx *gorm.DB) error {
	_log.Infof("(+) ADDRESS    %s", a.JSON())
	if !a.IsPrivate {
		go _server.Connect(a.HostPort)
	}
	go tx.Create(&Node{ID: a.NodeID})
	return nil
}

//---

func (n *Node) JSON() string {
	b, _ := json.Marshal(n)
	return string(b)
}

func (n *Node) AfterCreate(tx *gorm.DB) error {
	_log.Infof("(+) NODE       %s", n.JSON())
	_graph.AddNode(n.ID)
	return nil
}

//---

func (c *Connection) JSON() string {
	b, _ := json.Marshal(c)
	return string(b)
}

func (c *Connection) AfterSave(tx *gorm.DB) error {
	_log.Infof("(+) CONNECTION %s", c.JSON())
	return nil
}

func (c *Connection) AfterDelete(tx *gorm.DB) error {
	_log.Infof("(-) CONNECTION %s", c.JSON())
	return nil
}

func (g *Edge) JSON() string {
	b, _ := json.Marshal(g)
	return string(b)
}

func (e *Edge) AfterCreate(tx *gorm.DB) error {
	_log.Infof("(+) EDGE       %s", e.JSON())
	_graph.AddEdge(e.From, e.To)
	go _server.BroadCast(ID_INFO, InfoMessage{Edge: Entity[Edge]{Create: []Edge{*e}}})
	return nil
}
