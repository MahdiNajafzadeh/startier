package easynode

import (
	"os"
	"sync"
	"time"
)

func Run(configPath string) error {
	Load(_log)
	_log.Infof("(+) APP PID  : %d", os.Getpid())
	_log.Infof("(+) APP PPID : %d", os.Getppid())
	err := LoadConfig(configPath)
	if err != nil {
		_log.Infof("(x) APP CONFIG ERROR : %s", err)
		return err
	}
	_log.Info("(+) APP CONFIG")
	_log.Infof("(+) CONFIG %s", _config.JSON())
	ch := make(chan error)
	go runServer(ch)
	go runTun(ch)
	go runWeb(ch)
	go runPostRun()
	return <-ch
}

func runServer(ch chan<- error) {
	Load(_config)
	err := initServer()
	if err != nil {
		ch <- err
	}
	_log.Info("(+) APP SERVER")
}

func runTun(ch chan<- error) {
	Load(_config)
	err := initTun()
	if err != nil {
		ch <- err
		return
	}
	_log.Info("(+) APP TUN")
}

func runPostRun() {
	Load(_server)
	msg := JoinMessage{ID: _config.NodeID, Addresses: []Address{}}
	_db.Model(&Address{}).Where("node_id = ?", _config.NodeID).Find(&msg.Addresses)
	wg := sync.WaitGroup{}
	for _, peer := range _config.Peers {
		go func(addr string, msg JoinMessage) {
			count := 0
			for count < 10 {
				err := _server.Request(addr, ID_JOIN, &msg)
				if err == nil {
					_log.Infof("(+) PEER CONNECT : %s", addr)
					break
				}
				_log.Warnf("(x) PEER CONNECT : %s : %s", addr, err.Error())
				time.Sleep(time.Second * 5)
				count++
			}
			wg.Done()
		}(peer, msg)
		wg.Add(1)
	}
	wg.Wait()
	_log.Info("(+) APP POST-RUN")
}
