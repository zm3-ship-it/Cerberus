package dns

import (
	"log"
	"net"
	"sync"

	"cerberus/internal/devices"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type Sniffer struct {
	iface   string
	handle  *pcap.Handle
	logger  *Logger
	tracker *devices.Tracker
	stop    chan struct{}
	wg      sync.WaitGroup
}

func NewSniffer(iface string, logger *Logger, tracker *devices.Tracker) (*Sniffer, error) {
	handle, err := pcap.OpenLive(iface, 1600, true, pcap.BlockForever)
	if err != nil {
		return nil, err
	}

	// Capture only DNS traffic (port 53)
	if err := handle.SetBPFFilter("udp port 53"); err != nil {
		handle.Close()
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

	src := gopacket.NewPacketSource(s.handle, s.handle.LinkType())
	packets := src.Packets()

	for {
		select {
		case <-s.stop:
			return
		case pkt, ok := <-packets:
			if !ok {
				return
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

	dns, ok := dnsLayer.(*layers.DNS)
	if !ok {
		return
	}

	// Extract source IP and MAC
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

	// Skip packets from the router itself
	if isRouterIP(srcIP) {
		return
	}

	// Process queries (QR=false means query, not response)
	if !dns.QR {
		for _, q := range dns.Questions {
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
	// Common router self-IPs — adjust if your LAN differs
	routerIPs := []string{"192.168.1.1", "192.168.0.1", "10.0.0.1"}
	for _, r := range routerIPs {
		if parsed.Equal(net.ParseIP(r)) {
			return true
		}
	}
	return false
}
