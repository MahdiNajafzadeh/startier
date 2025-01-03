package easynode

import (
	"encoding/json"
	"fmt"
	"strings"

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
	err = db.AutoMigrate(&Address{}, &Node{}, &Connection{})
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
	Status    string `msgp:"status" json:"status"`
}

func (a *Address) JSON() string {
	b, _ := json.Marshal(a)
	return string(b)
}

func (a *Address) BeforeSave(tx *gorm.DB) error {
	hash, err := hashstructure.Hash(a, hashstructure.FormatV2, nil)
	if err != nil {
		_log.Error(err)
		return err
	}
	a.ID = fmt.Sprintf("%d", hash)
	defer tx.Create(&Node{ID: a.NodeID})
	return nil
}

func (a *Address) AfterSave(tx *gorm.DB) error {
	_log.Infof("DB CREATE ADDRESS %s", a.JSON())
	return nil
}

func (n *Node) JSON() string {
	b, _ := json.Marshal(n)
	return string(b)
}

func (n *Node) AfterSave(tx *gorm.DB) error {
	_log.Infof("DB CREATE NODE    %s", n.JSON())
	return nil
}

func (c *Connection) JSON() string {
	b, _ := json.Marshal(c)
	return string(b)
}

func (c *Connection) AfterSave(tx *gorm.DB) error {
	_log.Infof("DB CREATE CONNECTION %s", c.JSON())
	return nil
}

func LogDBErr(err error) {
	if err != gorm.ErrRecordNotFound || strings.Contains(err.Error(), "database table is locked") {
		_log.Error(err)
	}
}
