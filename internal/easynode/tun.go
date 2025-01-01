package easynode

import (
	"fmt"
	"net"

	"github.com/songgao/water"
	"github.com/songgao/water/waterutil"
	"github.com/vishvananda/netlink"
)

var _tun *water.Interface
var _join_msg *JoinMessage

func init() {
	tun, err := water.New(water.Config{
		DeviceType: water.TUN,
		PlatformSpecificParams: water.PlatformSpecificParams{
			Name: "startier",
			// Persist: true,
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
	// go tunnelLoop()
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
		var addr Address
		err = _db.
			Model(&Address{}).
			Where("node_id != ?", _config.NodeID).
			Where("ip_mask LIKE ?", dst.String()+"%").
			Where("is_private = ?", true).
			First(&addr).
			Error
		if err != nil {
			_log.Error(err)
			continue
		}
		s, ok := _store.session.node_to_session[addr.NodeID]
		if ok {
			c := s.AllocateContext()
			err := c.SetResponse(ID_PACKET, &PacketMessage{NodeID: _config.NodeID, Payload: buf[:n]})
			s.Send(c)
			if err != nil {
				_log.Error(err)
			}
		}
		
	}
}

// func tunnelLoop() {
// 	Load(_db)
// 	Load(_join_msg)
// 	for {
// 		var nodes []string
// 		err := _db.
// 			Model(&Address{}).
// 			Where("node_id != ?", _config.NodeID).
// 			Distinct("node_id").
// 			Pluck("node_id", &nodes).
// 			Error
// 		if err != nil {
// 			_log.Error(err)
// 		}
// 		for _, n := range nodes {
// 			_, ok := _store.session.node_to_session[n]
// 			if !ok {
// 				var addrs []Address
// 				err := _db.Model(&Address{}).Where("node_id = ? AND is_private = ?", n, false).Find(&addrs).Error
// 				if err != nil {
// 					_log.Error(err)
// 					continue
// 				}
// 				for _, addr := range addrs {
// 					err := _server.Request(addr.HostPort, ID_JOIN, _join_msg)
// 					if err != nil {
// 						_log.Debug(err)
// 					}
// 				}
// 			}
// 		}
// 		time.Sleep(time.Second * 5)
// 	}
// }
