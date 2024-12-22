package models

import (
	"gorm.io/gorm"
)

type Node struct {
	ID            int          `gorm:"primaryKey;autoIncrement"`
	Hostname      string       `gorm:"not null"`
	Domain        string       `gorm:"not null"`
	IsMe          bool         `gorm:"default:false"`
	Addresses     []Address    `gorm:"foreignKey:NodeID"`
	FromRoutes    []Route      `gorm:"foreignKey:FromID"`
	ToRoutes      []Route      `gorm:"foreignKey:ToID"`
	BetweenRoutes []*Route     `gorm:"many2many:between_routes;"`
	Connections   []Connection `gorm:"foreignKey:NodeID"`
	Remotes       []Remote     `gorm:"foreignKey:NodeID"`

	UniqueHostnameDomain struct{} `gorm:"uniqueIndex:idx_hostname_domain"`
}

type Address struct {
	ID      int    `gorm:"primaryKey;autoIncrement"`
	NodeID  int    `gorm:"not null"`
	Node    Node   `gorm:"foreignKey:NodeID;constraint:OnDelete:CASCADE"`
	Address string `gorm:"not null"`
	Mask    int    `gorm:"not null"`
	IsMe    bool   `gorm:"default:false"`

	UniqueAddressMask struct{} `gorm:"uniqueIndex:idx_address_mask"`
}

type Route struct {
	ID      int     `gorm:"primaryKey;autoIncrement"`
	FromID  int     `gorm:"not null"`
	From    Node    `gorm:"foreignKey:FromID;constraint:OnDelete:CASCADE"`
	ToID    int     `gorm:"not null"`
	To      Node    `gorm:"foreignKey:ToID;constraint:OnDelete:CASCADE"`
	Between []*Node `gorm:"many2many:between_routes;constraint:OnDelete:CASCADE"`

	UniqueFromTo struct{} `gorm:"uniqueIndex:idx_from_to"`
}

type Connection struct {
	ID       int    `gorm:"primaryKey;autoIncrement"`
	NodeID   int    `gorm:"not null"`
	Node     Node   `gorm:"foreignKey:NodeID;constraint:OnDelete:CASCADE"`
	RemoteID int    `gorm:"not null"`
	Remote   Remote `gorm:"foreignKey:RemoteID;constraint:OnDelete:CASCADE"`
	Ping     int    `gorm:"not null"`

	UniqueNodeRemote struct{} `gorm:"uniqueIndex:idx_node_remote"`
}

type Remote struct {
	ID         int          `gorm:"primaryKey;autoIncrement"`
	NodeID     int          `gorm:"not null"`
	Node       Node         `gorm:"foreignKey:NodeID;constraint:OnDelete:CASCADE"`
	Host       string       `gorm:"not null"`
	Port       int          `gorm:"not null"`
	TLS        bool         `gorm:"default:false"`
	Connection []Connection `gorm:"foreignKey:RemoteID"`

	UniqueHostPort struct{} `gorm:"uniqueIndex:idx_host_port"`
	UniqueNodeHost struct{} `gorm:"uniqueIndex:idx_node_host"`
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&Node{}, &Address{}, &Route{}, &Connection{}, &Remote{})
}
