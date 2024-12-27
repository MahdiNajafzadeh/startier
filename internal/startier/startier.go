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
	go RunDatabase(ch)
	go MonitorDatabase()
	// go RunTun(ch)
	go RunNetwork(ch)
	GetReady[*Network](GetNetwork)
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
	c := GetConfig()
	n := GetNetwork()
	addrs, err := n.GetAddress()
	if err != nil {
		ch <- err
		return
	}
	for _, peer := range c.Peers {
		go func(peer string) {
			for {
				err = n.Request(peer, ID_JOIN, JoinMessage{
					Node: NodeData{
						ID:      c.NodeID,
						Address: c.Address,
						Remotes: addrs,
					},
				})
				time.Sleep(time.Second * 5)
				if err != nil {
					log.Println(err)
					continue
				}
				break
			}
		}(peer)
	}
}
