package easynode

import (
	"fmt"
	"net"

	"github.com/songgao/water"
	"github.com/songgao/water/waterutil"
	"github.com/vishvananda/netlink"
)

var _tun *water.Interface

// var _join_msg *JoinMessage
var _tun_conn_cache Store[any, Session]
var _tun_addr_cache Store[any, Address]

func init() {
	_tun_conn_cache = newStore[any, Session]()
	tun, err := water.New(water.Config{
		DeviceType: water.TUN,
		PlatformSpecificParams: water.PlatformSpecificParams{
			Name: "easynode",
		},
	})
	if err != nil {
		_log.Error(err)
		panic(err)
	}
	_tun = tun
}

func initTun() error {
	link, err := netlink.LinkByName(_tun.Name())
	if err != nil {
		return err
	}
	ip, ipnet, err := net.ParseCIDR(_config.Local)
	if err != nil {
		return err
	}
	if ip.To4() == nil {
		return fmt.Errorf("not support IPv6 for local address : %s", ip.String())
	}
	ipnet.IP = ip
	err = netlink.AddrAdd(link, &netlink.Addr{IPNet: ipnet})
	if err != nil {
		return err
	}
	err = netlink.LinkSetUp(link)
	if err != nil {
		return err
	}
	go packetLoop()
	return nil
}

func packetLoop() {
	Load(_db)
	ip, ipnet, _ := net.ParseCIDR(_config.Local)
	ipnet.IP = ip
	buf := make([]byte, 1500)
	for {
		n, err := _tun.Read(buf)
		if err != nil {
			_log.Error(err)
			continue
		}
		if !waterutil.IsIPv4(buf[:n]) {
			continue
		}
		dst := waterutil.IPv4Destination(buf[:n])
		if !ipnet.Contains(dst) {
			continue
		}
		var address Address
		err = _db.
			Model(&Address{}).
			Where("node_id != ?", _config.NodeID).
			Where("ip_mask LIKE ?", dst.String()+"%").
			Where("is_private = ?", true).
			First(&address).
			Error
		if err != nil {
			continue
		}
		msg := &PacketMessage{FromNode: _config.NodeID, ToNode: address.NodeID, TTL: 10, Payload: buf[:n]}
		go RoutePacket(msg)
	}
}

func RoutePacket(msg *PacketMessage) {
	s, ok := _tun_conn_cache.Get(msg.ToNode)
	if ok {
		c := s.AllocateContext()
		c.SetResponse(ID_PACKET, msg)
		if s.Send(c) {
			return
		}
		_tun_conn_cache.Del(msg.ToNode)
		s.Close()
		_log.Infof("(-) TUN | CACHE SESSION : %s -> %s : %s", msg.FromNode, msg.ToNode, s.ID())
	}
	var connections []Connection
	_db.Model(&Connection{}).Where("node_id = ?", msg.ToNode).Find(&connections)
	_log.Infof("(?) TUN | DIRECT SESSION : %s -> %s : %d", msg.FromNode, msg.ToNode, len(connections))
	for _, conn := range connections {
		s, ok := _server.Sessions().Get(conn.SessionID)
		if ok {
			c := s.AllocateContext()
			c.SetResponse(ID_PACKET, msg)
			if s.Send(c) {
				_log.Infof("(+) TUN | CACHE SESSION : %s -> %s : %s", msg.FromNode, msg.ToNode, s.ID())
				_tun_conn_cache.Set(msg.ToNode, s)
				return
			}
			s.Close()
		}
	}
	_log.Infof("(?) TUN | PROXY SESSION : %s -> %s", msg.FromNode, msg.ToNode)
	sp := _graph.ShortestPath(msg.FromNode, msg.ToNode)
	_log.Infof("(?) TUN | PROXY SESSION : %+v", sp)
	if len(sp) == 0 {
		return
	}
	for _, v := range sp {
		if v == msg.FromNode || v == msg.ToNode {
			continue
		}
		var connections []Connection
		_db.Model(&Connection{}).Where("node_id = ?", v).Find(&connections)
		_log.Infof("(?) TUN | PROXY SESSION : %+v : %s : %d", sp, v, len(connections))
		for _, conn := range connections {
			s, ok := _server.Sessions().Get(conn.SessionID)
			if ok {
				c := s.AllocateContext()
				c.SetResponse(ID_PACKET, msg)
				if s.Send(c) {
					_log.Infof("(+) TUN | CACHE PROXY SESSION : %s : %s -> %s : %s", v, msg.FromNode, msg.ToNode, s.ID())
					_tun_conn_cache.Set(msg.ToNode, s)
					return
				}
				s.Close()
			}
		}
	}
}

// func RoutePacket(msg *PacketMessage) {
// 	s, ok := _tun_conn_cache.Get(msg.ToNode)
// 	if ok {
// 		c := s.AllocateContext()
// 		c.SetResponse(ID_PACKET, msg)
// 		if s.Send(c) {
// 			return
// 		}
// 		_tun_conn_cache.Del(msg.ToNode)
// 		s.Close()
// 		_log.Infof("(-) TUN | CACHE SESSION : %s -> %s : %s", msg.FromNode, msg.ToNode, s.ID())
// 	}
// 	_log.Infof("(?) TUN | DIRECT SESSION : %s -> %s", msg.FromNode, msg.ToNode)
// 	for _, s := range _server.Sessions().All() {
// 		v, ok := s.Store().Get("node_id")
// 		if !ok {
// 			continue
// 		}
// 		nodeID, ok := v.(string)
// 		if !ok {
// 			continue
// 		}
// 		if nodeID != msg.ToNode {
// 			continue
// 		}
// 		c := s.AllocateContext()
// 		c.SetResponse(ID_PACKET, msg)
// 		if s.Send(c) {
// 			_log.Infof("(+) TUN | CACHE SESSION : %s -> %s : %s", msg.FromNode, msg.ToNode, s.ID())
// 			_tun_conn_cache.Set(msg.ToNode, s)
// 			return
// 		}
// 		s.Close()
// 	}
// 	_log.Infof("(?) TUN | PROXY SESSION : %s -> %s", msg.FromNode, msg.ToNode)
// 	sp := _graph.ShortestPath(msg.FromNode, msg.ToNode)
// 	_log.Infof("(?) TUN | PROXY SESSION : %+v", sp)
// 	for _, v := range sp {
// 		if v == msg.FromNode || v == msg.ToNode {
// 			continue
// 		}
// 		_log.Infof("(?) TUN | PROXY SESSION : %+v : %s ", sp, v)
// 		for _, s := range _server.Sessions().All() {
// 			v, ok := s.Store().Get("node_id")
// 			if !ok {
// 				continue
// 			}
// 			nodeID, ok := v.(string)
// 			if !ok {
// 				continue
// 			}
// 			if nodeID != msg.ToNode {
// 				continue
// 			}
// 			c := s.AllocateContext()
// 			c.SetResponse(ID_PACKET, msg)
// 			if s.Send(c) {
// 				_log.Infof("(+) TUN | CACHE SESSION : %s -> %s : %s", msg.FromNode, msg.ToNode, s.ID())
// 				_tun_conn_cache.Set(msg.ToNode, s)
// 				return
// 			}
// 			s.Close()
// 		}
// 	}
// }
