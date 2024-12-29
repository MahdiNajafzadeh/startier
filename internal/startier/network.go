package startier

import (
	"log"
	"net"
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
	c := GetConfig()
	tun := GetTun()
	n := &Network{
		Sessions: make(map[string]easytcp.Session),
	}
	s := easytcp.NewServer(&easytcp.ServerOption{
		Packer:           easytcp.NewDefaultPacker(),
		Codec:            &easytcp.MsgpackCodec{},
		AsyncRouter:      true,
		DoNotPrintRoutes: true,
	})
	s.OnSessionCreate = func(sess easytcp.Session) {
		log.Printf("SESSION::CREATE::[%v]:[%s]=>[%s]", sess.ID(), sess.Conn().RemoteAddr(), sess.Conn().LocalAddr())
		n.Sessions[sess.ID().(string)] = sess
	}
	s.OnSessionClose = func(sess easytcp.Session) {
		log.Printf("SESSION::CLOSE::[%v]:[%s]=>[%s]", sess.ID(), sess.Conn().RemoteAddr(), sess.Conn().LocalAddr())
		delete(n.Sessions, sess.ID().(string))
	}
	s.Use(func(next easytcp.HandlerFunc) easytcp.HandlerFunc {
		return func(ctx easytcp.Context) {
			log.Printf("HANDLE::USE::[%v]:[%s]=>[%s]", ctx.Session().ID(), ctx.Session().Conn().RemoteAddr(), ctx.Session().Conn().LocalAddr())
			next(ctx)
		}
	})
	s.AddRoute(ID_CONFLICT, func(ctx easytcp.Context) {
		log.Fatalln("config file is make conflict in network : from ", ctx.Session().Conn().RemoteAddr())
	})
	s.AddRoute(ID_INFO, func(ctx easytcp.Context) {
		var req InfoMessage
		if err := ctx.Bind(&req); err != nil {
			return
		}
		for _, v := range req.Addresses {
			db.FirstOrCreate(&v)
		}
	})
	s.AddRoute(ID_JOIN, func(ctx easytcp.Context) {
		var req JoinMessage
		if err := ctx.Bind(&req); err != nil {
			return
		}
		if len(req.Addresses) == 0 {
			return
		}
		var nodeID string
		addrs := []Address{}
		for _, v := range req.Addresses {
			if v.NodeID == c.NodeID {
				ctx.SetResponse(ID_CONFLICT, 0)
				return
			}
			nodeID = v.NodeID
			v.ReID()
			var count int64
			db.Model(&Address{}).Where(&v).Count(&count)
			if count == 0 {
				addrs = append(addrs, v)
				db.Create(&v)
			}
		}
		if len(addrs) != 0 {
			targetHost, _, _ := net.SplitHostPort(ctx.Session().Conn().RemoteAddr().String())
			for _, sess := range n.Sessions {
				sessHost, _, _ := net.SplitHostPort(sess.Conn().RemoteAddr().String())
				if targetHost != sessHost {
					log.Println(sessHost)
					err := sess.AllocateContext().SetResponse(ID_INFO, &addrs)
					if err != nil {
						log.Println(err)
					}
				}
			}
		}
		addrs = []Address{}
		db.Where("node_id != ?", nodeID).Find(&addrs)
		ctx.SetResponse(ID_INFO, &InfoMessage{Addresses: addrs})
	})
	s.AddRoute(ID_PACKET, func(ctx easytcp.Context) {
		var req PacketMessage
		if err := ctx.Bind(&req); err != nil {
			return
		}
		_, err := tun.Write(req.Payload)
		if err != nil {
			log.Printf("TUN::PACKET::WRITE::ERR::[%s]", err.Error())
		}
	})
	s.NotFoundHandler(func(ctx easytcp.Context) {
		log.Printf("HANDLE::NOT_FOUND::[%v]:[%s]=>[%s]", ctx.Session().ID(), ctx.Session().Conn().RemoteAddr(), ctx.Session().Conn().LocalAddr())
	})
	n.Server = s
	return n, nil
}
