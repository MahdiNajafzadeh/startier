package tun

import (
	"errors"
	"net"
	"startier/config"
	"startier/internal/node/client"
	cnet "startier/internal/node/common/net"
	"startier/internal/node/database"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/songgao/water"
)

type Tun struct {
	tun      *water.Interface
	config   *config.Config
	database *database.Database
	client   *client.Client
}

type Packet struct {
	SrcIP   net.IP
	DstIP   net.IP
	Payload []byte
}

func New(conf *config.Config, db *database.Database, cl *client.Client) (*Tun, error) {
	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
		PlatformSpecificParams: water.PlatformSpecificParams{
			Name: "startier",
		},
	})
	if err != nil {
		return nil, err
	}
	return &Tun{
		tun:      ifce,
		config:   conf,
		database: db,
		client:   cl,
	}, nil
}

func (t *Tun) Run(ch chan error) {
	err := t.PerRun()
	if err != nil {
		ch <- err
	}
	t.RunReader()
}

func (t *Tun) PerRun() error {
	var err error
	for _, address := range t.config.Addresses {
		err = cnet.AddAddress(t.tun.Name(), address)
		if err != nil {
			return err
		}
	}
	err = cnet.UpInterface(t.tun.Name())
	if err != nil {
		return err
	}
	return nil
}

func (t *Tun) RunReader() {
	buf := make([]byte, 1500)
	for {
		n, err := t.tun.Read(buf)
		if err != nil {
			continue
		}
		packet, err := t.ReadPacket(buf[:n])
		if err != nil {
			continue
		}
		err = t.RoutePacket(packet)
		if err != nil {
			continue
		}
	}
}

func (t *Tun) ReadPacket(buf []byte) (Packet, error) {
	packet := gopacket.NewPacket(buf, layers.LayerTypeIPv4, gopacket.Default)
	ipv4Layer := packet.Layer(layers.LayerTypeIPv4)
	if ipv4Layer != nil {
		ipv4, _ := ipv4Layer.(*layers.IPv4)
		if len(ipv4.Payload) == 0 {
			return Packet{}, errors.New("zero payload")
		}
		return Packet{
			SrcIP:   ipv4.SrcIP,
			DstIP:   ipv4.DstIP,
			Payload: ipv4.Payload,
		}, nil
	}
	return Packet{}, errors.New("invalid packet: no IPv4 layer found")
}

func (t *Tun) RoutePacket(p Packet) error {
	// fmt.Printf(
	// 	"TUN::ROUTE[SRC(%s)DST(%s)PLN(%d)]\n",
	// 	p.SrcIP.String(),
	// 	p.DstIP.String(),
	// 	len(p.Payload),
	// )
	return nil
}
