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
	tun := GetReady(GetTun)
	buf := make([]byte, 1500)
	for {
		n, err := tun.Read(buf)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		if !waterutil.IsIPv4(buf[:n]) {
			continue
		}
		// dst := waterutil.IPv4Destination(buf[:n])
		// log.Println(dst.To4().String())
	}
}
