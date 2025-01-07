package easynode

import (
	"context"
	"net/http"
)

type DataNode struct {
	ID        string
	Local     Address
	Addresses []Address
	Edges     []Edge
}

var webHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	msg := NewInfoMessage()
	_db.Model(&Node{}).Find(&msg.Node.Create)
	_db.Model(&Address{}).Find(&msg.Address.Create)
	_db.Model(&Edge{}).Find(&msg.Edge.Create)
	data := []DataNode{}
	for _, node := range msg.Node.Create {
		n := DataNode{ID: node.ID, Addresses: []Address{}, Edges: []Edge{}}
		for _, addr := range msg.Address.Create {
			if addr.NodeID != node.ID {
				continue
			}
			if addr.IsPrivate {
				n.Local = addr
				continue
			}
			n.Addresses = append(n.Addresses, addr)
		}
		for _, edge := range msg.Edge.Create {
			if edge.From == node.ID {
				n.Edges = append(n.Edges, edge)
			}
		}
		data = append(data, n)
	}
	Index(data).Render(context.Background(), w)
}

func runWeb(ch chan error) {
	var err error
	var server http.Server
	if _config.Web.Enable {
		if _config.Web.TLS {
			tlsConfig, err := loadTLSConfig(_config.TLS.Public, _config.TLS.Private, _config.TLS.CA)
			if err != nil {
				ch <- err
			}
			server = http.Server{
				Addr:      _config.Web.Listen,
				Handler:   webHandler,
				TLSConfig: tlsConfig,
			}
			err = server.ListenAndServeTLS("", "")
			if err != nil {
				ch <- err
			}
		} else {
			server = http.Server{
				Addr:    _config.Web.Listen,
				Handler: webHandler,
			}
			err = server.ListenAndServe()
			if err != nil {
				ch <- err
			}
		}
	}
}
