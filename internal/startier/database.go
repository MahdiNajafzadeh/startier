package startier

import (
	"fmt"
	"log"
	"time"

	"github.com/mitchellh/hashstructure/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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

func MonitorDatabase() {
	db := GetDatabase()
	for {
		time.Sleep(time.Second * 5)
		var addrs []Address
		db.Find(&addrs)
		fmt.Println("===")
		for _, v := range addrs {
			fmt.Printf("%+v\n", v)
		}
		fmt.Println("===")
	}
}

type Address struct {
	ID        string `gorm:"primaryKey" msgp:"id" json:"-" hash:"-"`
	NodeID    string `gorm:"index:unique_address_idx,unique" msgp:"node_id" json:"node_id"`
	IPMask    string `gorm:"index:unique_address_idx,unique" msgp:"ip_mask" json:"ip_mask"`
	HostPort  string `gorm:"index:unique_address_idx,unique" msgp:"host_port" json:"host_port"`
	IsPrivate bool   `gorm:"index:unique_address_idx,unique" msgp:"is_private" json:"is_private"`
}

func (a *Address) ReID() {
	hash, _ := hashstructure.Hash(a, hashstructure.FormatV2, nil)
	a.ID = fmt.Sprintf("%d", hash)
}

func (a *Address) BeforeSave(tx *gorm.DB) (err error) {
	a.ReID()
	log.Printf("%+v", a)
	return nil
}
