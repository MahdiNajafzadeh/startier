package net

import (
	"fmt"
	"net"
	"strings"

	"github.com/vishvananda/netlink"
)

func AddAddress(dev string, address string) error {
	var ipNet *net.IPNet
	var ip net.IP
	var err error

	if strings.Contains(address, "/") {
		ip, ipNet, err = net.ParseCIDR(address)
		if err != nil {
			return fmt.Errorf("Error parsing CIDR '%s': %v", address, err)
		}
		ipNet.IP = ip
	} else {
		ip = net.ParseIP(address)
		if ip == nil {
			return fmt.Errorf("Invalid IP address: '%s'", address)
		}
		if ip.To4() != nil {
			ipNet = &net.IPNet{IP: ip, Mask: net.CIDRMask(24, 32)}
		} else {
			ipNet = &net.IPNet{IP: ip, Mask: net.CIDRMask(64, 128)}
		}
	}

	iface, err := netlink.LinkByName(dev)
	if err != nil {
		return fmt.Errorf("Error getting interface '%s': %v", dev, err)
	}

	addrs, err := netlink.AddrList(iface, netlink.FAMILY_ALL)
	if err != nil {
		return fmt.Errorf("Error listing addresses for interface '%s': %v", dev, err)
	}
	for _, a := range addrs {
		if a.IP.Equal(ip) {
			return nil
		}
	}

	err = netlink.AddrAdd(iface, &netlink.Addr{IPNet: ipNet})
	if err != nil {
		return fmt.Errorf("Failed to add address '%s' to interface '%s': %v", address, dev, err)
	}
	return nil
}

func UpInterface(dev string) error {
	iface, err := netlink.LinkByName(dev)
	if err != nil {
		return fmt.Errorf("Error in getting name of interface : %s : %+v", dev, err)
	}
	err = netlink.LinkSetUp(iface)
	if err != nil {
		return fmt.Errorf("Error in set up interface : %s : %+v", dev, err)
	}
	return nil
}
