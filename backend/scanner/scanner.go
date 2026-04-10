package scanner

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type Client struct {
	IP       string `json:"ip"`
	MAC      string `json:"mac"`
	Hostname string `json:"hostname"`
	Vendor   string `json:"vendor"`
	Signal   int    `json:"signal"`
	DevType  string `json:"dev_type"`
	State    string `json:"state"`
}

type Network struct {
	SSID    string `json:"ssid"`
	BSSID   string `json:"bssid"`
	Channel int    `json:"channel"`
	Enc     string `json:"enc"`
	Signal  int    `json:"signal"`
}

type Probe struct {
	MAC  string `json:"mac"`
	SSID string `json:"ssid"`
	Time string `json:"time"`
}

type Scanner struct {
	mu       sync.RWMutex
	clients  []Client
	networks []Network
	probes   []Probe
	scanning bool
}

func New() *Scanner { return &Scanner{} }

func (s *Scanner) IsScanning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.scanning
}

func (s *Scanner) GetClients() []Client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o := make([]Client, len(s.clients))
	copy(o, s.clients)
	return o
}

func (s *Scanner) GetNetworks() []Network {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o := make([]Network, len(s.networks))
	copy(o, s.networks)
	return o
}

func (s *Scanner) GetProbes() []Probe {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o := make([]Probe, len(s.probes))
	copy(o, s.probes)
	return o
}

func (s *Scanner) Scan() error {
	s.mu.Lock()
	if s.scanning {
		s.mu.Unlock()
		return fmt.Errorf("scan in progress")
	}
	s.scanning = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.scanning = false
		s.mu.Unlock()
	}()

	log.Println("[*] Starting scan...")

	// Find wireless interfaces
	iface := findWirelessInterface()
	log.Printf("[*] Using interface: %s", iface)

	// Phase 1: Scan APs using iwinfo (built into OpenWrt)
	networks := scanNetworksIwinfo(iface)
	if len(networks) == 0 {
		// Fallback to iw
		networks = scanNetworksIw(iface)
	}

	s.mu.Lock()
	s.networks = networks
	s.mu.Unlock()
	log.Printf("[*] Found %d networks", len(networks))

	// Phase 2: Find clients via ARP + DHCP leases
	clients := scanClients()

	s.mu.Lock()
	s.clients = clients
	s.mu.Unlock()
	log.Printf("[*] Found %d clients", len(clients))

	return nil
}

// ═══════════════════════════════════
// NETWORK SCANNING
// ═══════════════════════════════════

func findWirelessInterface() string {
	// Try common interface names
	for _, iface := range []string{"wlan0", "wlan1", "phy0-ap0", "phy1-ap0", "radio0", "radio1"} {
		cmd := exec.Command("iwinfo", iface, "info")
		if err := cmd.Run(); err == nil {
			return iface
		}
	}
	// Fallback: parse iw dev
	out, err := exec.Command("iw", "dev").Output()
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "Interface ") {
				return strings.TrimPrefix(line, "Interface ")
			}
		}
	}
	return "wlan0"
}

func scanNetworksIwinfo(iface string) []Network {
	var networks []Network

	out, err := exec.Command("iwinfo", iface, "scan").Output()
	if err != nil {
		log.Printf("[!] iwinfo scan failed: %v (trying all interfaces)", err)
		// Try scanning on all interfaces
		devOut, _ := exec.Command("iw", "dev").Output()
		for _, line := range strings.Split(string(devOut), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "Interface ") {
				altIface := strings.TrimPrefix(line, "Interface ")
				out2, err2 := exec.Command("iwinfo", altIface, "scan").Output()
				if err2 == nil && len(out2) > 0 {
					out = out2
					err = nil
					break
				}
			}
		}
		if err != nil {
			return networks
		}
	}

	lines := strings.Split(string(out), "\n")
	var current *Network

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// New cell
		if strings.Contains(trimmed, "Address:") {
			if current != nil {
				networks = append(networks, *current)
			}
			current = &Network{Signal: -100}
			// Extract BSSID
			re := regexp.MustCompile(`([0-9A-Fa-f:]{17})`)
			m := re.FindString(trimmed)
			if m != "" {
				current.BSSID = strings.ToUpper(m)
			}
		}
		if current == nil {
			continue
		}

		// ESSID
		if strings.Contains(trimmed, "ESSID:") {
			re := regexp.MustCompile(`ESSID:\s*"([^"]*)"`)
			m := re.FindStringSubmatch(trimmed)
			if len(m) > 1 {
				current.SSID = m[1]
			}
		}

		// Channel
		if strings.Contains(trimmed, "Channel:") {
			re := regexp.MustCompile(`Channel:\s*(\d+)`)
			m := re.FindStringSubmatch(trimmed)
			if len(m) > 1 {
				current.Channel, _ = strconv.Atoi(m[1])
			}
		}

		// Signal
		if strings.Contains(trimmed, "Signal:") {
			re := regexp.MustCompile(`Signal:\s*(-?\d+)`)
			m := re.FindStringSubmatch(trimmed)
			if len(m) > 1 {
				current.Signal, _ = strconv.Atoi(m[1])
			}
		}

		// Encryption
		if strings.Contains(trimmed, "Encryption:") {
			enc := strings.TrimSpace(strings.SplitN(trimmed, ":", 2)[1])
			if strings.Contains(enc, "none") || strings.Contains(enc, "None") {
				current.Enc = "Open"
			} else if strings.Contains(enc, "WPA3") || strings.Contains(enc, "SAE") {
				current.Enc = "WPA3"
			} else if strings.Contains(enc, "WPA2") {
				current.Enc = "WPA2"
			} else if strings.Contains(enc, "WPA") {
				current.Enc = "WPA"
			} else if strings.Contains(enc, "WEP") {
				current.Enc = "WEP"
			} else {
				current.Enc = enc
			}
		}
	}

	if current != nil {
		networks = append(networks, *current)
	}

	return networks
}

func scanNetworksIw(iface string) []Network {
	var networks []Network

	out, err := exec.Command("iw", "dev", iface, "scan").Output()
	if err != nil {
		log.Printf("[!] iw scan failed: %v", err)
		return networks
	}

	lines := strings.Split(string(out), "\n")
	var current *Network

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// New BSS
		if strings.HasPrefix(trimmed, "BSS ") {
			if current != nil {
				networks = append(networks, *current)
			}
			current = &Network{Signal: -100}
			re := regexp.MustCompile(`BSS ([0-9a-fA-F:]{17})`)
			m := re.FindStringSubmatch(trimmed)
			if len(m) > 1 {
				current.BSSID = strings.ToUpper(m[1])
			}
		}
		if current == nil {
			continue
		}

		if strings.HasPrefix(trimmed, "SSID:") {
			current.SSID = strings.TrimSpace(strings.TrimPrefix(trimmed, "SSID:"))
		}
		if strings.HasPrefix(trimmed, "signal:") {
			re := regexp.MustCompile(`(-?\d+)`)
			m := re.FindString(trimmed)
			current.Signal, _ = strconv.Atoi(m)
		}
		if strings.HasPrefix(trimmed, "freq:") {
			freq, _ := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(trimmed, "freq:")))
			current.Channel = freqToChannel(freq)
		}
		if strings.Contains(trimmed, "RSN:") || strings.Contains(trimmed, "WPA:") {
			if current.Enc == "" {
				if strings.Contains(trimmed, "RSN:") {
					current.Enc = "WPA2"
				} else {
					current.Enc = "WPA"
				}
			}
		}
	}

	if current != nil {
		networks = append(networks, *current)
	}

	// Mark networks with no encryption as Open
	for i := range networks {
		if networks[i].Enc == "" {
			networks[i].Enc = "Open"
		}
	}

	return networks
}

// ═══════════════════════════════════
// CLIENT SCANNING
// ═══════════════════════════════════

func scanClients() []Client {
	var clients []Client
	seen := make(map[string]bool)

	// Source 1: arp-scan (best — actively probes the whole subnet)
	arpClients := parseArpScan()
	for _, c := range arpClients {
		if !seen[c.MAC] {
			seen[c.MAC] = true
			clients = append(clients, c)
		}
	}

	// Source 2: DHCP leases (best for hostnames)
	leases := parseDHCPLeases()
	for _, l := range leases {
		if !seen[l.MAC] {
			seen[l.MAC] = true
			clients = append(clients, l)
		} else {
			// Update hostname from DHCP if we found it via arp-scan
			for i := range clients {
				if clients[i].MAC == l.MAC && l.Hostname != "" && clients[i].Hostname == "" {
					clients[i].Hostname = l.Hostname
				}
			}
		}
	}

	// Source 3: ARP/neighbor table (passive cache)
	neighbors := parseNeighbors()
	for _, n := range neighbors {
		if !seen[n.MAC] {
			seen[n.MAC] = true
			clients = append(clients, n)
		}
	}

	// Source 4: Associated wireless clients (for signal strength)
	wireless := parseWirelessClients()
	for _, w := range wireless {
		if !seen[w.MAC] {
			seen[w.MAC] = true
			clients = append(clients, w)
		} else {
			for i := range clients {
				if clients[i].MAC == w.MAC && w.Signal != 0 {
					clients[i].Signal = w.Signal
				}
			}
		}
	}

	return clients
}

func parseArpScan() []Client {
	var clients []Client
	out, err := exec.Command("arp-scan", "--localnet", "--quiet").Output()
	if err != nil {
		log.Printf("[!] arp-scan failed: %v", err)
		return clients
	}

	re := regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+)\s+([0-9a-fA-F:]{17})\s+(.*)`)
	for _, line := range strings.Split(string(out), "\n") {
		m := re.FindStringSubmatch(line)
		if m != nil {
			clients = append(clients, Client{
				IP:     m[1],
				MAC:    strings.ToUpper(m[2]),
				Vendor: strings.TrimSpace(m[3]),
				State:  "ARP",
			})
		}
	}
	return clients
}

func parseDHCPLeases() []Client {
	var clients []Client
	out, err := exec.Command("cat", "/tmp/dhcp.leases").Output()
	if err != nil {
		return clients
	}

	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			hostname := fields[3]
			if hostname == "*" {
				hostname = ""
			}
			clients = append(clients, Client{
				IP:       fields[2],
				MAC:      strings.ToUpper(fields[1]),
				Hostname: hostname,
				State:    "DHCP",
			})
		}
	}
	return clients
}

func parseNeighbors() []Client {
	var clients []Client

	out, err := exec.Command("ip", "neigh", "show").Output()
	if err != nil {
		return clients
	}

	for _, line := range strings.Split(string(out), "\n") {
		// Format: 192.168.1.12 dev br-lan lladdr aa:bb:cc:dd:ee:ff REACHABLE
		fields := strings.Fields(line)
		if len(fields) >= 5 && strings.Contains(line, "lladdr") {
			ip := fields[0]
			mac := ""
			state := ""
			for i, f := range fields {
				if f == "lladdr" && i+1 < len(fields) {
					mac = strings.ToUpper(fields[i+1])
				}
			}
			// Last field is state
			state = fields[len(fields)-1]

			if mac != "" && ip != "" {
				clients = append(clients, Client{
					IP:    ip,
					MAC:   mac,
					State: state,
				})
			}
		}
	}
	return clients
}

func parseWirelessClients() []Client {
	var clients []Client

	// Try iwinfo first
	ifaces := getAllWirelessInterfaces()
	for _, iface := range ifaces {
		out, err := exec.Command("iwinfo", iface, "assoclist").Output()
		if err != nil {
			continue
		}

		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			// Format: AA:BB:CC:DD:EE:FF  -45 dBm / -95 dBm (SNR 50)  120 ms ago
			re := regexp.MustCompile(`([0-9A-Fa-f:]{17})\s+(-?\d+)\s*dBm`)
			m := re.FindStringSubmatch(line)
			if len(m) > 2 {
				sig, _ := strconv.Atoi(m[2])
				clients = append(clients, Client{
					MAC:    strings.ToUpper(m[1]),
					Signal: sig,
					State:  "Associated",
				})
			}
		}
	}

	return clients
}

func getAllWirelessInterfaces() []string {
	var ifaces []string
	out, err := exec.Command("iw", "dev").Output()
	if err != nil {
		return []string{"wlan0"}
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Interface ") {
			ifaces = append(ifaces, strings.TrimPrefix(line, "Interface "))
		}
	}
	if len(ifaces) == 0 {
		return []string{"wlan0"}
	}
	return ifaces
}

func freqToChannel(freq int) int {
	switch {
	case freq == 2484:
		return 14
	case freq >= 2412 && freq <= 2472:
		return (freq - 2407) / 5
	case freq >= 5170 && freq <= 5825:
		return (freq - 5000) / 5
	default:
		return 0
	}
}
