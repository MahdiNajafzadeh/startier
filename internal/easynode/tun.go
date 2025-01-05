package easynode

import (
	"fmt"
	"net"

	"github.com/songgao/water"
	"github.com/songgao/water/waterutil"
	"github.com/vishvananda/netlink"
	"gorm.io/gorm"
)

var _tun *water.Interface

// var _join_msg *JoinMessage
var _tun_conn_cache Store[any, Session]

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
TunLoop:
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
		var addr Address
		err = _db.
			Model(&Address{}).
			Where("node_id != ?", _config.NodeID).
			Where("ip_mask LIKE ?", dst.String()+"%").
			Where("is_private = ?", true).
			First(&addr).
			Error
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				_log.Error(err)
			}
			continue
		}
		msg := &PacketMessage{FromNode: _config.NodeID, ToNode: addr.NodeID, TTL: 10, Payload: buf[:n]}
		//-UseCachedConnectionForNode
		s, ok := _tun_conn_cache.Get(addr.NodeID)
		if ok {
			c := s.AllocateContext()
			c.SetResponse(ID_PACKET, msg)
			if s.Send(c) {
				continue
			}
			_tun_conn_cache.Del(addr.NodeID)
			s.Close()
		}
		//-
		var connections []Connection
		err = _db.
			Model(&Connection{}).
			Where("node_id = ?", addr.NodeID).
			Find(&connections).
			Error
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				_log.Error(err)
			}
			continue
		}
		//-UseAllConnectionToTasgetNode
		for _, conn := range connections {
			s, ok := _server.Sessions().Get(conn.SessionID)
			if ok {
				c := s.AllocateContext()
				c.SetResponse(ID_PACKET, msg)
				if s.Send(c) {
					_tun_conn_cache.Set(addr.NodeID, s)
					continue TunLoop
				} else {
					s.Close()
				}
			}
		}
		//-UseGraphConnection
	}
}

// func routePacket(msg *PacketMessage) error {
// }
