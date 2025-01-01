package easynode

import (
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
		_log.Debugf("CREATE + SESSION  %v %v", s.Conn().RemoteAddr(), s.Conn().LocalAddr())
	}
	_server.OnSessionClose = func(s Session) {
		_log.Debugf("CLOSE - SESSION  %v %v %v", s.Conn().RemoteAddr(), s.Conn().LocalAddr(), s.Get("node_id"))
	}
	_server.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) {
			next(c)
			var msg BaseNodeID
			if err := c.Bind(&msg); err != nil {
				return
			}
			if msg.NodeID == "" {
				return
			}
			if c.Session().Get("node_id") != nil {
				return
			}
			_log.Debugf("NET HANDLE USE %v %v", c.Request().ID(), msg.NodeID)
			c.Session().Set("node_id", msg.NodeID)
			_store.session.node_to_session[msg.NodeID] = c.Session()
		}
	})
	_server.AddRoute(ID_PACKET, func(c Context) {
		var msg PacketMessage
		if err := c.Bind(&msg); err != nil {
			_log.Error(err)
			return
		}
		_, err := _tun.Write(msg.Payload)
		if err != nil {
			_log.Error(err)
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
			}
			return nil
		})
		if err != nil {
			_log.Error(err)
			return
		}
		for _, sess := range _server.Sessions() {
			if sess.ID() != c.Session().ID() {
				err := sess.
					AllocateContext().
					SetResponse(ID_INFO, &InfoMessage{NodeID: _config.NodeID, Addresses: addrs})
				if err != nil {
					_log.Error(err)
				}
			}
		}
		addrs = []Address{}
		_db.Where("node_id != ?", msg.NodeID).Find(&addrs)
		c.SetResponse(ID_INFO, InfoMessage{NodeID: _config.NodeID, Addresses: addrs})
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
			err := _db.FirstOrCreate(&v).Error
			if err != nil {
				_log.Error(err)
			}
		}
	})
	err := _server.Run(_config.Listen)
	return err
}
