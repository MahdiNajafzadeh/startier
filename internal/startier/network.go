package startier

import (
	"log"
	"startier/internal/easytcp"
)

type Network struct {
	Server   *easytcp.Server
	Sessions map[string]easytcp.Session
}

var _network *Network

func GetNetwork() *Network {
	return _network
}

func RunNetwork(ch chan error) {
	c := GetConfig()
	var err error
	if GetNetwork() == nil {
		_network, err = NewNetwork()
		if err != nil {
			ch <- err
			return
		}
	}
	err = _network.Server.Run(c.Listen)
	if err != nil {
		ch <- err
	}
}

func NewNetwork() (*Network, error) {
	db := GetDatabase()
	n := &Network{
		Sessions: make(map[string]easytcp.Session),
	}
	s := easytcp.NewServer(&easytcp.ServerOption{
		Packer:      easytcp.NewDefaultPacker(),
		Codec:       &easytcp.MsgpackCodec{},
		AsyncRouter: false,
	})
	s.OnSessionCreate = func(sess easytcp.Session) {
		log.Printf("CREATE : %v : %s => %s", sess.ID(), sess.Conn().RemoteAddr(), sess.Conn().LocalAddr())
		n.Sessions[sess.ID().(string)] = sess
	}
	s.OnSessionClose = func(sess easytcp.Session) {
		log.Printf("CLOSE  : %v : %s => %s", sess.ID(), sess.Conn().RemoteAddr(), sess.Conn().LocalAddr())
		delete(n.Sessions, sess.ID().(string))
	}
	s.Use(func(next easytcp.HandlerFunc) easytcp.HandlerFunc {
		return func(ctx easytcp.Context) {
			log.Printf("USE    : %v : %s => %s", ctx.Session().ID(), ctx.Session().Conn().RemoteAddr().String(), ctx.Session().Conn().LocalAddr().String())
			next(ctx)
		}
	})
	s.AddRoute(ID_JOIN, func(ctx easytcp.Context) {
		var req JoinMessage
		if err := ctx.Bind(&req); err != nil {
			return
		}

		db := GetDatabase()
		var existingAddresses []Address
		db.Find(&existingAddresses)

		var matchedAddress *Address
		for _, addr := range existingAddresses {
			if addr.NodeID != req.Address[0].NodeID && addr.Host == req.Address[0].Host && addr.Port == req.Address[0].Port {
				matchedAddress = &addr
				break
			}
		}

		if matchedAddress != nil {
			go ctx.SetResponse(ID_INFO, &InfoMessage{Address: existingAddresses})
		}

		for _, sess := range n.Sessions {
			if sess.ID() != ctx.Session().ID() {
				ctx.SetResponse(ID_JOIN_CAST, req)
			}
		}
	})
	s.NotFoundHandler(func(ctx easytcp.Context) {
		log.Printf("NOT FOUND : %v", ctx.Session().Conn().RemoteAddr().String())
	})
	n.Server = s
	return n, nil
}
