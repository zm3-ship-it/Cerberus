package scanner

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

type Client struct {
	IP string `json:"ip"`; MAC string `json:"mac"`; Hostname string `json:"hostname"`
	Vendor string `json:"vendor"`; Signal int `json:"signal"`; LastSeen string `json:"last_seen"`
}
type Network struct {
	SSID string `json:"ssid"`; BSSID string `json:"bssid"`; Channel int `json:"channel"`
	Enc string `json:"enc"`; Signal int `json:"signal"`
}
type Probe struct { MAC string `json:"mac"`; SSID string `json:"ssid"`; Time string `json:"time"` }

type Scanner struct { mu sync.RWMutex; clients []Client; networks []Network; probes []Probe; scanning bool }

func New() *Scanner { return &Scanner{} }
func (s *Scanner) IsScanning() bool { s.mu.RLock(); defer s.mu.RUnlock(); return s.scanning }
func (s *Scanner) GetClients() []Client { s.mu.RLock(); defer s.mu.RUnlock(); o := make([]Client, len(s.clients)); copy(o, s.clients); return o }
func (s *Scanner) GetNetworks() []Network { s.mu.RLock(); defer s.mu.RUnlock(); o := make([]Network, len(s.networks)); copy(o, s.networks); return o }
func (s *Scanner) GetProbes() []Probe { s.mu.RLock(); defer s.mu.RUnlock(); o := make([]Probe, len(s.probes)); copy(o, s.probes); return o }

func (s *Scanner) Scan() error {
	s.mu.Lock()
	if s.scanning { s.mu.Unlock(); return fmt.Errorf("scan in progress") }
	s.scanning = true; s.clients = nil; s.mu.Unlock()
	defer func() { s.mu.Lock(); s.scanning = false; s.mu.Unlock() }()
	log.Println("[*] Scanning...")
	cmd := exec.Command("arp-scan", "--localnet", "--quiet")
	out, _ := cmd.Output()
	re := regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+)\s+([0-9a-fA-F:]{17})\s+(.*)`)
	for _, line := range strings.Split(string(out), "\n") {
		m := re.FindStringSubmatch(line)
		if m != nil {
			host := m[1]
			names, err := net.LookupAddr(m[1])
			if err == nil && len(names) > 0 { host = strings.TrimSuffix(names[0], ".") }
			s.mu.Lock()
			s.clients = append(s.clients, Client{IP: m[1], MAC: strings.ToUpper(m[2]), Vendor: strings.TrimSpace(m[3]), Hostname: host, LastSeen: "now", Signal: -50})
			s.mu.Unlock()
		}
	}
	log.Printf("[*] Found %d clients", len(s.clients)); return nil
}
