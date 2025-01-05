package easynode

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
		go func() {
			c := s.AllocateContext()
			c.SetResponse(ID_WHO, &WhoMessage{Token: s.ID().(string), Sender: _config.NodeID})
			s.Send(c)
		}()
	}
	_server.OnSessionClose = func(s Session) {
		go func() {
			conn := Connection{SessionID: s.ID().(string)}
			_db.Where("session_id = ?", conn.SessionID).Delete(&conn)
		}()
	}
	_server.AddRoute(ID_WHO, func(c Context) {
		var msg WhoMessage
		if err := c.Bind(&msg); err != nil {
			return
		}
		sessionID := c.Session().ID().(string)
		if msg.Token == sessionID {
			if msg.Receiver != "" {
				if msg.Sender != "" {
					if msg.Receiver != _config.NodeID {
						_db.Model(&Node{}).Create(&Node{ID: msg.Receiver})
						_db.Model(&Edge{}).Create(&Edge{From: _config.NodeID, To: msg.Receiver})
						_db.Model(&Connection{}).Create(&Connection{SessionID: sessionID, NodeID: msg.Receiver})
						c.Session().Store().Set("node_id", msg.Receiver)
						_tun_conn_cache.Set(msg.Receiver, c.Session())
						return
					} else {
						c.SetResponse(ID_CONFLICT, 0)
						c.Send()
					}
				}
			}
		} else {
			if msg.Sender != "" {
				if msg.Receiver == "" {
					if msg.Sender != _config.NodeID {
						msg.Receiver = _config.NodeID
						c.SetResponse(ID_WHO, &msg)
						c.Send()
						_db.Model(&Node{}).Create(&Node{ID: msg.Sender})
						_db.Model(&Edge{}).Create(&Edge{From: msg.Sender, To: _config.NodeID})
						_db.Model(&Connection{}).Create(&Connection{SessionID: sessionID, NodeID: msg.Sender})
						c.Session().Store().Set("node_id", msg.Sender)
						_tun_conn_cache.Set(msg.Sender, c.Session())
						return
					} else {
						_log.Fatal("NET CONF")
					}
				}
			}
		}
		c.Session().Close()
	})
	_server.AddRoute(ID_PACKET, func(c Context) {
		var msg PacketMessage
		if err := c.Bind(&msg); err != nil {
			_log.Error(err)
			return
		}
		if msg.ToNode == _config.NodeID {
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
		s, ok := _tun_conn_cache.Get(msg.ToNode)
		if ok {
			c := s.AllocateContext()
			c.SetResponse(ID_PACKET, &msg)
			if c.Send() {
				return
			}
			_tun_conn_cache.Del(s.ID())
			s.Close()
		}
		var connections []Connection
		err := _db.
			Model(&Connection{}).
			Where("node_id = ?", msg.ToNode).
			Find(&connections).
			Error
		if err != nil {
			_log.Error(err)
			return
		}
		for _, conn := range connections {
			s, ok := _server.Sessions().Get(conn.SessionID)
			if ok {
				c := s.AllocateContext()
				c.SetResponse(ID_PACKET, &msg)
				if s.Send(c) {
					_tun_conn_cache.Set(msg.ToNode, s)
					break
				} else {
					s.Close()
				}
			}
		}
	})
	_server.AddRoute(ID_CONFLICT, func(c Context) {
		_log.Error(c.Session().Conn().RemoteAddr())
		_log.Fatal("NET CONFLIC")
	})
	_server.AddRoute(ID_JOIN, func(c Context) {
		var msg InfoMessage
		if err := c.Bind(&msg); err != nil {
			return
		}
		if len(msg.Node.Create) == 0 {
			return
		}
		node := msg.Node.Create[0]
		if node.ID == _config.NodeID {
			c.SetResponse(ID_CONFLICT, 0)
			c.Send()
			return
		}
		bmsg := msg
		bmsg.Node = Entity[Node]{Create: msg.Node.Create}
		bmsg.Edge = Entity[Edge]{Create: []Edge{{From: node.ID, To: _config.NodeID}}}
		bmsg.Address = Entity[Address]{Create: []Address{}}
		for _, v := range msg.Address.Create {
			if v.IsPrivate && v.IPMask == _config.Local {
				c.SetResponse(ID_CONFLICT, 0)
				c.Send()
				return
			}
			var count int64
			_db.Model(&Address{}).Where(&v).Count(&count)
			if count == 0 {
				bmsg.Address.Create = append(bmsg.Address.Create, v)
			}
		}
		_db.Create(&bmsg.Address.Create)
		_db.Create(&bmsg.Edge.Create)
		_server.BroadCast(ID_INFO, bmsg)
	})
	_server.AddRoute(ID_INFO, func(c Context) {
		var msg InfoMessage
		if err := c.Bind(&msg); err != nil {
			return
		}
		_db.Model(&Address{}).Create(&msg.Address.Create)
		_db.Model(&Node{}).Create(&msg.Node.Create)
		_db.Model(&Edge{}).Create(&msg.Edge.Create)
		_db.Model(&Address{}).Delete(&msg.Address.Delete)
		_db.Model(&Node{}).Delete(&msg.Node.Delete)
		_db.Model(&Edge{}).Delete(&msg.Edge.Delete)
	})
	err := _server.Run(_config.Listen)
	return err
}
