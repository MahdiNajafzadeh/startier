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
			AsyncRouter:      false,
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
		go RoutePacket(&msg)
	})
	_server.AddRoute(ID_CONFLICT, func(c Context) {
		_log.Error(c.Session().Conn().RemoteAddr())
		_log.Fatal("NET CONFLIC")
	})
	_server.AddRoute(ID_JOIN, func(c Context) {
		var msg JoinMessage
		if err := c.Bind(&msg); err != nil {
			return
		}
		if msg.ID == "" {
			return
		}
		if msg.ID == _config.NodeID {
			c.SetResponse(ID_CONFLICT, 0)
			return
		}
		if len(msg.Addresses) == 0 {
			return
		}
		for _, v := range msg.Addresses {
			var paddr Address
			_db.Model(&paddr).Where("node_id != ? AND ip_mask = ? AND is_private = ?", v.NodeID, v.IPMask, true).First(&paddr)
			if paddr.ID != "" {
				c.SetResponse(ID_CONFLICT, 0)
				return
			}
		}
		bmsg := NewInfoMessage()
		bmsg.Address.Create = msg.Addresses
		bmsg.Node.Create = append(bmsg.Node.Create, Node{ID: msg.ID})
		bmsg.Edge.Create = append(bmsg.Edge.Create, Edge{From: msg.ID, To: _config.NodeID})
		for _, v := range bmsg.Node.Create {
			_db.Create(&v)
		}
		for _, v := range bmsg.Address.Create {
			_db.Create(&v)
		}
		for _, v := range bmsg.Edge.Create {
			_db.Create(&v)
		}
		_db.Model(&Node{}).Find(&bmsg.Node.Create)
		_db.Model(&Address{}).Find(&bmsg.Address.Create)
		_db.Model(&Edge{}).Find(&bmsg.Edge.Create)
		_server.BroadCast(ID_INFO, &bmsg)
		c.SetResponse(ID_INFO, &bmsg)
	})
	_server.AddRoute(ID_INFO, func(c Context) {
		var msg InfoMessage
		if err := c.Bind(&msg); err != nil {
			return
		}
		for _, v := range msg.Node.Create {
			_db.Create(&v)
		}
		for _, v := range msg.Address.Create {
			_db.Create(&v)
		}
		for _, v := range msg.Edge.Create {
			_db.Create(&v)
		}
	})
	if _config.TLS.Enable {
		tlsConfig, err := loadTLSConfig(_config.TLS.Public, _config.TLS.Private, _config.TLS.Private)
		if err != nil {
			return err
		}
		if err = _server.RunTLS(_config.Listen, tlsConfig); err != nil {
			return err
		}
	} else {
		if err := _server.Run(_config.Listen); err != nil {
			return err
		}
	}
	return nil
}
