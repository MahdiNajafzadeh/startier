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
	var logLevel logger.LogLevel = logger.Silent
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
	ID  string `gorm:"primaryKey" msgp:"id" json:"id"`
	GID string `msgp:"-" json:"gid"`
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
	ID   string `gorm:"primaryKey" msgp:"id" json:"-" hash:"-"`
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

func (n *Node) BeforeSave(tx *gorm.DB) error {
	return nil
}

func (n *Node) AfterSave(tx *gorm.DB) error {
	gnode := _graph.NewNode()
	n.GID = fmt.Sprintf("%d", gnode.ID())
	_log.Infof("(+) NODE       %s", n.JSON())
	_graph.AddNode(gnode)
	err := tx.UpdateColumn("gid", n.GID).Error
	_log.Error(err)
	return nil
}

func (n *Node) AfterDelete(tx *gorm.DB) error {
	_log.Infof("(-) NODE       %s", n.JSON())
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

func (g *Edge) BeforeSave(tx *gorm.DB) error {
	if g.From == g.To {
		return fmt.Errorf("'Edge.From' & 'Edge.To' is equal")
	}
	hash, _ := hashstructure.Hash(g, hashstructure.FormatV2, nil)
	g.ID = fmt.Sprintf("%d", hash)
	return nil
}

func (g *Edge) AfterSave(tx *gorm.DB) error {
	_log.Infof("(+) EDGE       %s", g.JSON())
	nodeFrom := Node{ID: g.From}
	nodeTo := Node{ID: g.To}
	_db.Table("edgs").Find(&nodeFrom)
	_db.Table("edgs").Find(&nodeTo)
	_log.Infof("AFTER-ADD-EDGE : %s", g.JSON())
	// _graph.SetEdge(
	// 	_graph.NewEdge(
	// 		_graph.Node(nodeFrom.GID),
	// 		_graph.Node(nodeTo.GID),
	// 	),
	// )
	return nil
}
