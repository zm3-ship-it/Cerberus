package recon

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

type ScanResult struct {
	BSSID     string `json:"bssid"`
	SSID      string `json:"ssid"`
	Channel   int    `json:"channel"`
	Signal    int    `json:"signal"`
	Encrypt   string `json:"encryption"`
	Clients   int    `json:"clients"`
	FirstSeen int64  `json:"first_seen"`
	LastSeen  int64  `json:"last_seen"`
}

type ClientResult struct {
	MAC       string `json:"mac"`
	BSSID     string `json:"bssid"`
	Signal    int    `json:"signal"`
	Probes    string `json:"probes"`
	FirstSeen int64  `json:"first_seen"`
	LastSeen  int64  `json:"last_seen"`
}

type Scanner struct {
	iface    string
	monIface string
	running  bool
	mu       sync.RWMutex
	aps      map[string]*ScanResult
	clients  map[string]*ClientResult
	stop     chan struct{}
}

func NewScanner(iface string) *Scanner {
	return &Scanner{
		iface:   iface,
		aps:     make(map[string]*ScanResult),
		clients: make(map[string]*ClientResult),
		stop:    make(chan struct{}),
	}
}

// EnableMonitor puts the wireless interface into monitor mode
func (s *Scanner) EnableMonitor() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already in monitor mode
	out, _ := exec.Command("iw", "dev").Output()
	if strings.Contains(string(out), "type monitor") {
		s.monIface = s.iface + "mon"
		return nil
	}

	// Try airmon-ng first
	cmd := exec.Command("airmon-ng", "start", s.iface)
	if out, err := cmd.CombinedOutput(); err == nil {
		// Parse the monitor interface name from output
		re := regexp.MustCompile(`\(monitor mode.*enabled on (\w+)\)`)
		if matches := re.FindSubmatch(out); len(matches) > 1 {
			s.monIface = string(matches[1])
		} else {
			s.monIface = s.iface + "mon"
		}
		return nil
	}

	// Fallback: manual monitor mode
	cmds := [][]string{
		{"ip", "link", "set", s.iface, "down"},
		{"iw", "dev", s.iface, "set", "type", "monitor"},
		{"ip", "link", "set", s.iface, "up"},
	}
	for _, c := range cmds {
		if err := exec.Command(c[0], c[1:]...).Run(); err != nil {
			return fmt.Errorf("monitor mode setup failed at '%s': %w", strings.Join(c, " "), err)
		}
	}
	s.monIface = s.iface
	return nil
}

// DisableMonitor restores the interface to managed mode
func (s *Scanner) DisableMonitor() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Try airmon-ng first
	if err := exec.Command("airmon-ng", "stop", s.monIface).Run(); err == nil {
		return nil
	}

	// Fallback
	cmds := [][]string{
		{"ip", "link", "set", s.monIface, "down"},
		{"iw", "dev", s.monIface, "set", "type", "managed"},
		{"ip", "link", "set", s.monIface, "up"},
	}
	for _, c := range cmds {
		exec.Command(c[0], c[1:]...).Run()
	}
	return nil
}

// StartScan begins continuous airodump-ng scanning
func (s *Scanner) StartScan() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("scan already running")
	}
	if s.monIface == "" {
		s.mu.Unlock()
		return fmt.Errorf("monitor mode not enabled — call EnableMonitor first")
	}
	s.running = true
	s.stop = make(chan struct{})
	s.mu.Unlock()

	go s.scanLoop()
	return nil
}

func (s *Scanner) StopScan() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}
	close(s.stop)
	s.running = false
}

func (s *Scanner) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *Scanner) scanLoop() {
	csvFile := "/tmp/cerberus-scan"

	for {
		select {
		case <-s.stop:
			return
		default:
		}

		// Run airodump-ng for 10 seconds, write CSV
		cmd := exec.Command("airodump-ng",
			"--write", csvFile,
			"--write-interval", "5",
			"--output-format", "csv",
			"--band", "abg",
			s.monIface,
		)
		cmd.Start()

		select {
		case <-s.stop:
			cmd.Process.Kill()
			return
		case <-time.After(10 * time.Second):
			cmd.Process.Kill()
		}

		cmd.Wait()

		// Parse the CSV output
		s.parseCSV(csvFile + "-01.csv")

		// Clean up temp files
		exec.Command("rm", "-f",
			csvFile+"-01.csv",
			csvFile+"-01.cap",
			csvFile+"-01.kismet.csv",
			csvFile+"-01.kismet.netxml",
		).Run()
	}
}

func (s *Scanner) parseCSV(path string) {
	cmd := exec.Command("cat", path)
	out, err := cmd.Output()
	if err != nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().Unix()
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	inClients := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "Station MAC") {
			inClients = true
			continue
		}
		if strings.HasPrefix(line, "BSSID") {
			inClients = false
			continue
		}
		if line == "" {
			continue
		}

		fields := strings.Split(line, ",")
		for i := range fields {
			fields[i] = strings.TrimSpace(fields[i])
		}

		if !inClients && len(fields) >= 14 {
			bssid := fields[0]
			if !isMAC(bssid) {
				continue
			}

			ap := &ScanResult{
				BSSID:    bssid,
				SSID:     strings.TrimSpace(fields[13]),
				Encrypt:  fields[5],
				LastSeen: now,
			}

			fmt.Sscanf(fields[3], "%d", &ap.Channel)
			fmt.Sscanf(fields[8], "%d", &ap.Signal)

			if existing, ok := s.aps[bssid]; ok {
				existing.LastSeen = now
				existing.Signal = ap.Signal
				existing.Clients = ap.Clients
			} else {
				ap.FirstSeen = now
				s.aps[bssid] = ap
			}
		}

		if inClients && len(fields) >= 6 {
			mac := fields[0]
			if !isMAC(mac) {
				continue
			}

			client := &ClientResult{
				MAC:      mac,
				BSSID:    fields[5],
				Probes:   strings.Join(fields[6:], ","),
				LastSeen: now,
			}
			fmt.Sscanf(fields[3], "%d", &client.Signal)

			if existing, ok := s.clients[mac]; ok {
				existing.LastSeen = now
				existing.Signal = client.Signal
				if client.BSSID != "(not associated)" {
					existing.BSSID = client.BSSID
				}
			} else {
				client.FirstSeen = now
				s.clients[mac] = client
			}
		}
	}

	// Count clients per AP
	for _, ap := range s.aps {
		ap.Clients = 0
	}
	for _, client := range s.clients {
		if ap, ok := s.aps[client.BSSID]; ok {
			ap.Clients++
		}
	}
}

// GetAPs returns all discovered access points
func (s *Scanner) GetAPs() []ScanResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]ScanResult, 0, len(s.aps))
	for _, ap := range s.aps {
		result = append(result, *ap)
	}
	return result
}

// GetClients returns all discovered clients
func (s *Scanner) GetClients() []ClientResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]ClientResult, 0, len(s.clients))
	for _, c := range s.clients {
		result = append(result, *c)
	}
	return result
}

// GetClientsForAP returns clients connected to a specific AP
func (s *Scanner) GetClientsForAP(bssid string) []ClientResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []ClientResult
	for _, c := range s.clients {
		if strings.EqualFold(c.BSSID, bssid) {
			result = append(result, *c)
		}
	}
	return result
}

func (s *Scanner) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.aps = make(map[string]*ScanResult)
	s.clients = make(map[string]*ClientResult)
}

// ARPScan does a quick LAN ARP scan for live hosts
func ARPScan(cidr string) ([]ARPHost, error) {
	if cidr == "" {
		cidr = "192.168.1.0/24"
	}

	cmd := exec.Command("arp-scan", "--localnet", "--interface=br-lan")
	out, err := cmd.Output()
	if err != nil {
		// Fallback to ping sweep + arp
		return pingARPFallback(cidr)
	}

	var hosts []ARPHost
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	re := regexp.MustCompile(`^(\d+\.\d+\.\d+\.\d+)\s+([\da-fA-F:]+)\s+(.*)$`)

	for scanner.Scan() {
		matches := re.FindStringSubmatch(scanner.Text())
		if len(matches) == 4 {
			hosts = append(hosts, ARPHost{
				IP:     matches[1],
				MAC:    matches[2],
				Vendor: strings.TrimSpace(matches[3]),
			})
		}
	}
	return hosts, nil
}

func pingARPFallback(cidr string) ([]ARPHost, error) {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var hosts []ARPHost

	for ip := ip.Mask(ipNet.Mask); ipNet.Contains(ip); incIP(ip) {
		wg.Add(1)
		target := ip.String()
		go func() {
			defer wg.Done()
			cmd := exec.Command("ping", "-c", "1", "-W", "1", target)
			if cmd.Run() == nil {
				mu.Lock()
				hosts = append(hosts, ARPHost{IP: target})
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	// Now read the ARP table
	out, err := exec.Command("ip", "neigh", "show").Output()
	if err != nil {
		return hosts, nil
	}

	arpMap := make(map[string]string)
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 5 && fields[len(fields)-1] != "FAILED" {
			arpMap[fields[0]] = fields[4]
		}
	}

	for i, h := range hosts {
		if mac, ok := arpMap[h.IP]; ok {
			hosts[i].MAC = mac
		}
	}

	return hosts, nil
}

type ARPHost struct {
	IP     string `json:"ip"`
	MAC    string `json:"mac"`
	Vendor string `json:"vendor"`
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func isMAC(s string) bool {
	_, err := net.ParseMAC(s)
	return err == nil
}

func init() {
	// Suppress airmon-ng "found N processes" warnings
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
