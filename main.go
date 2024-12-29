package main

import (
	"flag"
	"log"
	"startier/internal/startier"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()
}

func main() {
	log.Fatal(startier.Run(configPath))
}

// package main

// import (
// 	"github.com/nedscode/memdb"
// )

// // type Node struct {
// // 	ID string
// // }

// type Address struct {
// 	ID        string
// 	NodeID    string
// 	Host      string
// 	Port      int
// 	IsPrivate bool
// }

// // type Connection struct {
// // 	ID       string
// // 	IsTunnel bool
// // }

// func main() {
// 	// nodedb := memdb.NewStore().
// 	// 	PrimaryKey("ID")
// 	addrdb := memdb.NewStore().
// 		PrimaryKey("ID").
// 		CreateIndex("Host").
// 		CreateIndex("NodeID").
// 		CreateIndex("Port").
// 		CreateIndex("IsPrivate")
// 	// conndb := memdb.NewStore().
// 	// 	PrimaryKey("ID")
// 	addrdb.Put(&Address{})
// }
