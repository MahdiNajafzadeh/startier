package easynode

import (
	"fmt"

	"github.com/mitchellh/hashstructure/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var _db *gorm.DB

func init() {
	db, err := gorm.Open(sqlite.Open("file::memory:?mode=memory&cache=shared&_fk=1"), &gorm.Config{})
	if err != nil {
		_log.Errorf("database error in init : %s", err)
		panic(err)
	}
	err = db.AutoMigrate(&Address{})
	if err != nil {
		_log.Errorf("database error in migration : %s", err)
		panic(err)
	}
	_db = db
}

type Address struct {
	ID        string `gorm:"primaryKey" msgp:"id" json:"-" hash:"-"`
	NodeID    string `gorm:"index:unique_address_idx,unique" msgp:"node_id" json:"node_id"`
	IPMask    string `gorm:"index:unique_address_idx,unique" msgp:"ip_mask" json:"ip_mask"`
	HostPort  string `gorm:"index:unique_address_idx,unique" msgp:"host_port" json:"host_port"`
	IsPrivate bool   `gorm:"index:unique_address_idx,unique" msgp:"is_private" json:"is_private"`
}

func (a *Address) BeforeSave(tx *gorm.DB) error {
	hash, err := hashstructure.Hash(a, hashstructure.FormatV2, nil)
	if err != nil {
		_log.Errorf("database error in calculate address id hash : %s", err)
		return err
	}
	a.ID = fmt.Sprintf("%d", hash)
	return nil
}
