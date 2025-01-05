package easynode

import (
	"encoding/json"
	"fmt"

	"github.com/mitchellh/hashstructure/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var _db *gorm.DB

func init() {
	var logLevel logger.LogLevel
	logLevel = logger.Error
	logLevel = logger.Info
	logLevel = logger.Silent
	db, err := gorm.Open(sqlite.Open("file::memory:?mode=memory&cache=shared"), &gorm.Config{
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

type Connection struct {
	SessionID string `grom:"primaryKey" msgp:"session_id" json:"session_id"`
	NodeID    string `msgp:"node_id" json:"node_id"`
}

type Edge struct {
	ID   int    `gorm:"primaryKey" msgp:"id" json:"id" hash:"-"`
	From string `gorm:"index:unique_edge_idx,unique" msgp:"from" json:"from"`
	To   string `gorm:"index:unique_edge_idx,unique" msgp:"to" json:"to"`
}

//---

func (a *Address) JSON() string {
	b, _ := json.Marshal(a)
	return string(b)
}

func (a *Address) BeforeSave(tx *gorm.DB) error {
	hash, _ := hashstructure.Hash(a, hashstructure.FormatV2, nil)
	a.ID = fmt.Sprintf("%d", hash)
	go tx.Create(&Node{ID: a.NodeID})
	return nil
}

func (a *Address) AfterSave(tx *gorm.DB) error {
	// _log.Infof("(+) ADDRESS    %s", a.JSON())
	if !a.IsPrivate {
		go _server.Connect(a.HostPort)
	}
	return nil
}

//---

func (n *Node) JSON() string {
	b, _ := json.Marshal(n)
	return string(b)
}

func (n *Node) GraphID() int64 {
	h, _ := hashstructure.Hash(n, hashstructure.FormatV2, nil)
	return int64(h)
}

func (n *Node) BeforeSave(tx *gorm.DB) error {
	return nil
}

func (n *Node) AfterSave(tx *gorm.DB) error {
	// _log.Infof("(+) NODE       %s", n.JSON())
	gn, _ := _graph.NodeWithID(n.GraphID())
	_graph.AddNode(gn)
	return nil
}

func (n *Node) AfterDelete(tx *gorm.DB) error {
	// _log.Infof("(-) NODE       %s", n.JSON())
	return nil
}

//---

func (c *Connection) JSON() string {
	b, _ := json.Marshal(c)
	return string(b)
}

func (c *Connection) AfterSave(tx *gorm.DB) error {
	// _log.Infof("(+) CONNECTION %s", c.JSON())
	return nil
}

func (c *Connection) AfterDelete(tx *gorm.DB) error {
	// _log.Infof("(-) CONNECTION %s", c.JSON())
	return nil
}

//---

func (g *Edge) JSON() string {
	b, _ := json.Marshal(g)
	return string(b)
}

func (e *Edge) BeforeSave(tx *gorm.DB) error {
	if e.From == e.To {
		return fmt.Errorf("'Edge.From' & 'Edge.To' is equal")
	}
	h, _ := hashstructure.Hash(e, hashstructure.FormatV2, nil)
	e.ID = int(uint64(h))
	return nil
}

func (g *Edge) AfterSave(tx *gorm.DB) error {
	_log.Infof("(+) EDGE       %s", g.JSON())
	_graph.SetEdge(
		_graph.NewEdge(
			_graph.Node(),
			_graph.Node(),
		),
	)
	return nil
}
