package easynode

import (
	"time"

	"gorm.io/gorm"
)

var _server *Server

func initServer() error {
	Load(_config)
	_server = NewServer(
		&ServerOption{
			NodeID:           _config.NodeID,
			Packer:           NewDefaultPacker(),
			Codec:            &MsgpackCodec{},
			DoNotPrintRoutes: true,
			AsyncRouter:      true,
		},
	)
	_server.OnSessionCreate = func(s Session) {
		_log.Debugf("+SESSION  %v %v", s.Conn().RemoteAddr(), s.Conn().LocalAddr())
		c := s.AllocateContext()
		c.SetResponse(ID_WHO, &WhoMessage{Token: s.ID().(string), Me: _config.NodeID, You: ""})
		s.Send(c)
	}
	_server.OnSessionClose = func(s Session) {
		_log.Debugf("-SESSION  %v %v %v", s.Conn().RemoteAddr(), s.Conn().LocalAddr(), s.Get("node_id"))
	}
	_server.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) {
			_log.Debug(c.Request().ID())
			go next(c)
			if c.Session().Get("node_id") != nil {
				return
			}
			var msg BaseNodeID
			if err := c.Bind(&msg); err != nil || msg.NodeID == "" {
				return
			}
			c.Session().Set("node_id", msg.NodeID)
			_db.Create(&Connection{NodeID: msg.NodeID, SessionID: c.Session().ID().(string), Status: "OK"})
		}
	})
	_server.AddRoute(ID_WHO, func(c Context) {
		var msg WhoMessage
		if err := c.Bind(&msg); err != nil {
			return
		}
	})
	_server.AddRoute(ID_PACKET, func(c Context) {
		var msg PacketMessage
		if err := c.Bind(&msg); err != nil {
			_log.Error(err)
			return
		}
		if msg.Target == _config.NodeID {
			_, err := _tun.Write(msg.Payload)
			if err != nil {
				_log.Error(err)
			}
			return
		}
		if msg.TTL == 0 {
			return
		}
		msg.TTL -= 1
		s, ok := _store.session.node_to_session[msg.Target]
		if ok {
			c := s.AllocateContext()
			err := c.SetResponse(ID_PACKET, msg)
			if err != nil {
				_log.Error(err)
				return
			}
			if s.Send(c) {
				return
			}
		}
		var tNode Node
		err := _db.Model(Node{}).Where("id = ?", msg.Target).First(&tNode).Error
		if err != nil {
			_log.Error(err)
			return
		}
		for _, nodeIDs := range _graph.FindAllPaths(_config.NodeID, tNode.ID) {
			for _, nodeID := range nodeIDs {
				if nodeID == _config.NodeID || nodeID == tNode.ID {
					continue
				}
				s, ok := _store.session.node_to_session[nodeID]
				if ok {
					c := s.AllocateContext()
					err := c.SetResponse(ID_PACKET, msg)
					if err != nil {
						_log.Error(err)
						continue
					}
					if s.Send(c) {
						return
					}
				}
			}
		}
	})
	_server.AddRoute(ID_CONFLICT, func(c Context) {
		_log.Fatal("NET CONFLIC")
	})
	_server.AddRoute(ID_JOIN, func(c Context) {
		var msg JoinMessage
		if err := c.Bind(&msg); err != nil {
			return
		}
		if len(msg.Addresses) == 0 {
			return
		}
		addrs := []Address{}
		err := _db.Transaction(func(tx *gorm.DB) error {
			for _, v := range msg.Addresses {
				if msg.NodeID == _config.NodeID {
					err := c.SetResponse(ID_CONFLICT, 0)
					if err != nil {
						return err
					}
					c.Send()
				}
				var count int64
				err := tx.Model(&Address{}).Where(&v).Count(&count).Error
				if err != nil {
					return err
				}
				if count == 0 {
					addrs = append(addrs, v)
					err := tx.Create(&v).Error
					if err != nil {
						return err
					}
				}
				time.Sleep(time.Millisecond * 100)
			}
			return nil
		})
		if err != nil {
			_log.Error(err)
			return
		}
		for _, s := range _store.session.node_to_session {
			if s.ID() != c.Session().ID() {
				c := s.AllocateContext()
				err := c.SetResponse(ID_INFO, &InfoMessage{NodeID: _config.NodeID, Addresses: addrs})
				if err != nil {
					_log.Error(err)
					continue
				}
				s.Send(c)
			}
		}
		addrs = []Address{}
		err = _db.Where("node_id != ?", msg.NodeID).Find(&addrs).Error
		if err != nil {
			_log.Error(err)
		}
		err = c.SetResponse(ID_INFO, InfoMessage{NodeID: _config.NodeID, Addresses: addrs})
		if err != nil {
			_log.Error(err)
		}
		c.Send()
	})
	_server.AddRoute(ID_INFO, func(c Context) {
		var msg InfoMessage
		if err := c.Bind(&msg); err != nil {
			return
		}
		if len(msg.Addresses) == 0 {
			return
		}
		for _, v := range msg.Addresses {
			time.Sleep(time.Millisecond * 100)
			_db.FirstOrCreate(&v)
		}
	})
	err := _server.Run(_config.Listen)
	return err
}
