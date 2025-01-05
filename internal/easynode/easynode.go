package easynode

import (
	"os"
	"sync"
	"time"
)

func Run(configPath string) error {
	Load(_log)
	_log.Infof("APP PID  : %d", os.Getpid())
	_log.Infof("APP PPID : %d", os.Getppid())
	err := LoadConfig(configPath)
	if err != nil {
		return err
	}
	// _log.Info("APP LOAD CONFIG")
	// _log.Infof("CONFIG %s", _config.JSON())
	ch := make(chan error)
	go runServer(ch)
	go runTun(ch)
	go runPostRun(ch)
	err = <-ch
	_log.Errorf(err.Error())
	return err
}

func runServer(ch chan<- error) {
	Load(_config)
	err := initServer()
	if err != nil {
		ch <- err
	}
	// _log.Info("APP LOAD SERVER")
}

func runTun(ch chan<- error) {
	Load(_config)
	err := initTun()
	if err != nil {
		ch <- err
		return
	}
	// _log.Info("APP LOAD TUN")
}

func runPostRun(ch chan<- error) {
	Load(_server)
	var addrs []Address
	_db.Model(&Address{}).Where("node_id = ?", _config.NodeID).Find(&addrs)
	msg := &InfoMessage{
		Node:    Entity[Node]{Create: []Node{{ID: _config.NodeID}}},
		Address: Entity[Address]{Create: addrs},
	}
	wg := sync.WaitGroup{}
	for _, peer := range _config.Peers {
		go func(addr string) {
			for {
				err := _server.Request(addr, ID_JOIN, msg)
				if err == nil {
					// _log.Infof("CONNECT PEER SUCCESS %s", addr)
					break
				}
				// _log.Errorf("CONNECT PEER ERROR %s", err.Error())
				time.Sleep(time.Second * 5)
			}
			wg.Done()
		}(peer)
		wg.Add(1)
	}
	wg.Wait()
	// _log.Info("APP LOAD POST-RUN")
}
