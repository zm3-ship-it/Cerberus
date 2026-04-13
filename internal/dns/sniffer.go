package dns

import (
	"log"
	"net"
	"sync"

	"cerberus/internal/devices"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/afpacket"
	"github.com/gopacket/gopacket/layers"
)

type Sniffer struct {
	iface   string
	handle  *afpacket.TPacket
	logger  *Logger
	tracker *devices.Tracker
	stop    chan struct{}
	wg      sync.WaitGroup
}

func NewSniffer(iface string, logger *Logger, tracker *devices.Tracker) (*Sniffer, error) {
	handle, err := afpacket.NewTPacket(
		afpacket.OptInterface(iface),
		afpacket.OptFrameSize(65536),
		afpacket.OptBlockSize(65536*128),
		afpacket.OptNumBlocks(8),
	)
	if err != nil {
		return nil, err
	}

	return &Sniffer{
		iface:   iface,
		handle:  handle,
		logger:  logger,
		tracker: tracker,
		stop:    make(chan struct{}),
	}, nil
}

func (s *Sniffer) Run() {
	s.wg.Add(1)
	defer s.wg.Done()

	src := gopacket.NewPacketSource(s.handle, layers.LinkTypeEthernet)
	src.NoCopy = true
	packets := src.Packets()

	for {
		select {
		case <-s.stop:
			return
		case pkt, ok := <-packets:
			if !ok {
				return
			}
			// Filter for UDP port 53 in Go (no BPF needed)
			udpLayer := pkt.Layer(layers.LayerTypeUDP)
			if udpLayer == nil {
				continue
			}
			udp := udpLayer.(*layers.UDP)
			if udp.DstPort != 53 && udp.SrcPort != 53 {
				continue
			}
			s.processPacket(pkt)
		}
	}
}

func (s *Sniffer) Stop() {
	close(s.stop)
	s.handle.Close()
	s.wg.Wait()
}

func (s *Sniffer) processPacket(pkt gopacket.Packet) {
	dnsLayer := pkt.Layer(layers.LayerTypeDNS)
	if dnsLayer == nil {
		return
	}

	dnsMsg, ok := dnsLayer.(*layers.DNS)
	if !ok {
		return
	}

	var srcIP, srcMAC string

	if ipLayer := pkt.Layer(layers.LayerTypeIPv4); ipLayer != nil {
		ip := ipLayer.(*layers.IPv4)
		srcIP = ip.SrcIP.String()
	} else if ipLayer := pkt.Layer(layers.LayerTypeIPv6); ipLayer != nil {
		ip := ipLayer.(*layers.IPv6)
		srcIP = ip.SrcIP.String()
	}

	if ethLayer := pkt.Layer(layers.LayerTypeEthernet); ethLayer != nil {
		eth := ethLayer.(*layers.Ethernet)
		srcMAC = eth.SrcMAC.String()
	}

	if isRouterIP(srcIP) {
		return
	}

	// QR=false means query, not response
	if !dnsMsg.QR {
		for _, q := range dnsMsg.Questions {
			domain := string(q.Name)
			qtype := q.Type.String()
			deviceName := s.tracker.Resolve(srcIP, srcMAC)

			rec := QueryRecord{
				Domain:    domain,
				Type:      qtype,
				ClientIP:  srcIP,
				ClientMAC: srcMAC,
				Device:    deviceName,
			}

			if err := s.logger.Log(rec); err != nil {
				log.Printf("log error: %v", err)
			}
		}
	}
}

func isRouterIP(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	routerIPs := []string{"192.168.1.1", "192.168.0.1", "10.0.0.1"}
	for _, r := range routerIPs {
		if parsed.Equal(net.ParseIP(r)) {
			return true
		}
	}
	return false
}
