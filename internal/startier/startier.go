package startier

import (
	"log"
	"os"
	"reflect"
	"time"
)

func Run(configPath string) error {
	log.Printf("PPID : %d", os.Getppid())
	log.Printf("PID  : %d", os.Getpid())
	c, err := LoadConfig(configPath)
	if err != nil {
		return err
	}
	RunDatabase()
	c.ToDatabase()
	log.Printf("CONFIG : %+v", c.ToJSON())
	ch := make(chan error)
	defer close(ch)
	go MonitorDatabase()
	go RunTun(ch)
	go RunNetwork(ch)
	go PostRun(ch)
	return <-ch
}

func GetReady[T any](getter func() T) T {
	for {
		things := getter()
		if reflect.ValueOf(things).IsNil() {
			time.Sleep(time.Millisecond)
		} else {
			return things
		}
	}
}

func PostRun(ch chan error) {
	c := GetReady(GetConfig)
	n := GetReady(GetNetwork)
	txn := GetReady(GetDatabase).Txn(false)
	for _, peer := range c.Peers {
		go func(peer string) {
			msg := JoinMessage{Node: Node{ID: c.NodeID}, Address: []Address{}}
			it, err := txn.Get("address", "node_id", c.NodeID)
			if err != nil {
				log.Println(err)
				return
			}
			for obj := it.Next(); obj != nil; obj = it.Next() {
				msg.Address = append(msg.Address, *obj.(*Address))
				log.Printf("address : %+v", *obj.(*Address))
			}
			for {
				err = n.Request(peer, ID_JOIN, &msg)
				time.Sleep(time.Second)
				if err != nil {
					log.Println(err)
					continue
				}
				break
			}
		}(peer)
	}
}
