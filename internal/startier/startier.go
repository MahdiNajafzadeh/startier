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
	n := GetReady(GetNetwork)
	c := GetConfig()
	db := GetDatabase()
	msg := JoinMessage{Addresses: []Address{}}
	db.Find(&msg.Addresses)
	for _, addr := range msg.Addresses {
		log.Printf("%+v", addr)
	}
	for _, peer := range c.Peers {
		go func(peer string) {
			for {
				time.Sleep(time.Second)
				err := n.Server.NewRequest(peer, ID_JOIN, msg)
				if err == nil {
					break
				}
				log.Println(err)
			}
		}(peer)
	}
}
