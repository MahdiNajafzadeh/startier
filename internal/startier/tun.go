package startier

import (
	"log"
	"net"

	"github.com/songgao/water"
	"github.com/songgao/water/waterutil"
	"github.com/vishvananda/netlink"
)

const (
	TUN_NAME = "startier"
)

var _tun *water.Interface

func GetTun() *water.Interface {
	return _tun
}

func RunTun(ch chan error) {
	if GetTun() == nil {
		tun, err := NewTun()
		if err != nil {
			ch <- err
			return
		}
		_tun = tun
	}
	go LoopTun(ch)
}

func NewTun() (*water.Interface, error) {
	var err error
	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
		PlatformSpecificParams: water.PlatformSpecificParams{
			Name: TUN_NAME,
		},
	})
	if err != nil {
		return nil, err
	}
	link, err := netlink.LinkByName(ifce.Name())
	if err != nil {
		return nil, err
	}
	config := GetConfig()
	ip, ipnet, err := net.ParseCIDR(config.Local)
	if err != nil {
		return nil, err
	}
	ipnet.IP = ip
	err = netlink.AddrAdd(link, &netlink.Addr{IPNet: ipnet})
	if err != nil {
		return nil, err
	}
	err = netlink.LinkSetUp(link)
	if err != nil {
		return nil, err
	}
	return ifce, nil
}

func LoopTun(ch chan error) {
	c := GetReady(GetConfig)
	tun := GetReady(GetTun)
	db := GetReady(GetDatabase)
	network := GetReady(GetNetwork)
	_, ipnet, _ := net.ParseCIDR(c.Local)
	buf := make([]byte, 1500)
	for {
		n, err := tun.Read(buf)
		if err != nil {
			continue
		}
		if waterutil.IsIPv6(buf[:n]) {
			continue
		}
		ip := waterutil.IPv4Destination(buf[:n])
		if !ipnet.Contains(ip) {
			continue
		}
		var addr Address
		err = db.
			Where("ip_mask LIKE ?", ip.String()+"%").
			Where("is_private = ?", true).
			Find(&addr).
			Error
		if err != nil {
			continue
		}
		// ...
		log.Printf("TUN::PACKET::READ::[%s]:[%s]", addr.NodeID, addr.IPMask)

	}
}
