package network

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

type WANConfig struct {
	Proto   string `json:"proto"`   // dhcp, static, pppoe
	IP      string `json:"ip"`
	Netmask string `json:"netmask"`
	Gateway string `json:"gateway"`
	DNS1    string `json:"dns1"`
	DNS2    string `json:"dns2"`
	// PPPoE
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type LANConfig struct {
	IP      string `json:"ip"`
	Netmask string `json:"netmask"`
	DHCPOn  bool   `json:"dhcp_enabled"`
	Start   int    `json:"dhcp_start"`
	Limit   int    `json:"dhcp_limit"`
	Lease   string `json:"dhcp_lease"`
}

type WiFiConfig struct {
	Enabled    bool   `json:"enabled"`
	SSID       string `json:"ssid"`
	Password   string `json:"password"`
	Encryption string `json:"encryption"` // psk2, psk2+ccmp, sae, sae-mixed, none
	Channel    string `json:"channel"`
	Band       string `json:"band"` // 2g, 5g
	Hidden     bool   `json:"hidden"`
	Mode       string `json:"mode"` // ap, sta, monitor
	HWMode     string `json:"hwmode"`
	HTMode     string `json:"htmode"`
	TxPower    int    `json:"txpower"`
}

type DHCPLease struct {
	Expire   string `json:"expire"`
	MAC      string `json:"mac"`
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
}

type StaticLease struct {
	MAC  string `json:"mac"`
	IP   string `json:"ip"`
	Name string `json:"name"`
}

type InterfaceStatus struct {
	Name    string `json:"name"`
	Up      bool   `json:"up"`
	IP      string `json:"ip"`
	Netmask string `json:"netmask"`
	MAC     string `json:"mac"`
	RX      int64  `json:"rx_bytes"`
	TX      int64  `json:"tx_bytes"`
	Speed   string `json:"speed"`
}

type Manager struct {
	mu sync.RWMutex
}

func New() *Manager {
	return &Manager{}
}

// ═══════════════════════════════════
// UCI HELPERS
// ═══════════════════════════════════

func uciGet(path string) string {
	out, err := exec.Command("uci", "get", path).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func uciSet(path, value string) error {
	return exec.Command("uci", "set", fmt.Sprintf("%s=%s", path, value)).Run()
}

func uciDelete(path string) error {
	return exec.Command("uci", "delete", path).Run()
}

func uciCommit(config string) error {
	return exec.Command("uci", "commit", config).Run()
}

func uciAddList(path, value string) error {
	return exec.Command("uci", "add_list", fmt.Sprintf("%s=%s", path, value)).Run()
}

func uciDelList(path, value string) error {
	return exec.Command("uci", "del_list", fmt.Sprintf("%s=%s", path, value)).Run()
}

// ═══════════════════════════════════
// WAN
// ═══════════════════════════════════

func (m *Manager) GetWAN() WANConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return WANConfig{
		Proto:   uciGet("network.wan.proto"),
		IP:      uciGet("network.wan.ipaddr"),
		Netmask: uciGet("network.wan.netmask"),
		Gateway: uciGet("network.wan.gateway"),
		DNS1:    uciGet("network.wan.dns"),
	}
}

func (m *Manager) SetWAN(cfg WANConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Printf("[*] Setting WAN: proto=%s", cfg.Proto)

	uciSet("network.wan.proto", cfg.Proto)

	switch cfg.Proto {
	case "static":
		uciSet("network.wan.ipaddr", cfg.IP)
		uciSet("network.wan.netmask", cfg.Netmask)
		uciSet("network.wan.gateway", cfg.Gateway)
		if cfg.DNS1 != "" {
			uciDelete("network.wan.dns")
			uciAddList("network.wan.dns", cfg.DNS1)
			if cfg.DNS2 != "" {
				uciAddList("network.wan.dns", cfg.DNS2)
			}
		}
	case "pppoe":
		uciSet("network.wan.username", cfg.Username)
		uciSet("network.wan.password", cfg.Password)
		if cfg.DNS1 != "" {
			uciSet("network.wan.peerdns", "0")
			uciDelete("network.wan.dns")
			uciAddList("network.wan.dns", cfg.DNS1)
			if cfg.DNS2 != "" {
				uciAddList("network.wan.dns", cfg.DNS2)
			}
		}
	case "dhcp":
		// Clean up static fields
		uciDelete("network.wan.ipaddr")
		uciDelete("network.wan.netmask")
		uciDelete("network.wan.gateway")
	}

	uciCommit("network")
	return restartNetwork()
}

// ═══════════════════════════════════
// LAN
// ═══════════════════════════════════

func (m *Manager) GetLAN() LANConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	start, _ := strconv.Atoi(uciGet("dhcp.lan.start"))
	limit, _ := strconv.Atoi(uciGet("dhcp.lan.limit"))
	ignore := uciGet("dhcp.lan.ignore")

	return LANConfig{
		IP:      uciGet("network.lan.ipaddr"),
		Netmask: uciGet("network.lan.netmask"),
		DHCPOn:  ignore != "1",
		Start:   start,
		Limit:   limit,
		Lease:   uciGet("dhcp.lan.leasetime"),
	}
}

func (m *Manager) SetLAN(cfg LANConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Printf("[*] Setting LAN: ip=%s", cfg.IP)

	uciSet("network.lan.ipaddr", cfg.IP)
	uciSet("network.lan.netmask", cfg.Netmask)

	if cfg.DHCPOn {
		uciDelete("dhcp.lan.ignore")
	} else {
		uciSet("dhcp.lan.ignore", "1")
	}
	uciSet("dhcp.lan.start", strconv.Itoa(cfg.Start))
	uciSet("dhcp.lan.limit", strconv.Itoa(cfg.Limit))
	if cfg.Lease != "" {
		uciSet("dhcp.lan.leasetime", cfg.Lease)
	}

	uciCommit("network")
	uciCommit("dhcp")
	return restartNetwork()
}

// ═══════════════════════════════════
// WIFI
// ═══════════════════════════════════

func (m *Manager) GetWiFi() []WiFiConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var configs []WiFiConfig

	// Read all wireless interfaces
	out, err := exec.Command("uci", "show", "wireless").Output()
	if err != nil {
		return configs
	}

	// Parse UCI output for wifi-iface sections
	ifaceCount := 0
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "=wifi-iface") {
			ifaceCount++
		}
	}

	for i := 0; i < ifaceCount; i++ {
		prefix := fmt.Sprintf("wireless.@wifi-iface[%d]", i)
		device := uciGet(prefix + ".device")
		disabled := uciGet(prefix + ".disabled")

		cfg := WiFiConfig{
			Enabled:    disabled != "1",
			SSID:       uciGet(prefix + ".ssid"),
			Encryption: uciGet(prefix + ".encryption"),
			Mode:       uciGet(prefix + ".mode"),
			Hidden:     uciGet(prefix+".hidden") == "1",
		}

		// Get radio settings
		if device != "" {
			cfg.Channel = uciGet(fmt.Sprintf("wireless.%s.channel", device))
			cfg.Band = uciGet(fmt.Sprintf("wireless.%s.band", device))
			cfg.HWMode = uciGet(fmt.Sprintf("wireless.%s.hwmode", device))
			cfg.HTMode = uciGet(fmt.Sprintf("wireless.%s.htmode", device))
			txp := uciGet(fmt.Sprintf("wireless.%s.txpower", device))
			cfg.TxPower, _ = strconv.Atoi(txp)
		}

		configs = append(configs, cfg)
	}

	return configs
}

func (m *Manager) SetWiFi(index int, cfg WiFiConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	prefix := fmt.Sprintf("wireless.@wifi-iface[%d]", index)
	device := uciGet(prefix + ".device")

	log.Printf("[*] Setting WiFi[%d]: ssid=%s", index, cfg.SSID)

	uciSet(prefix+".ssid", cfg.SSID)
	uciSet(prefix+".encryption", cfg.Encryption)

	if cfg.Password != "" {
		uciSet(prefix+".key", cfg.Password)
	}

	if cfg.Enabled {
		uciDelete(prefix + ".disabled")
	} else {
		uciSet(prefix+".disabled", "1")
	}

	if cfg.Hidden {
		uciSet(prefix+".hidden", "1")
	} else {
		uciDelete(prefix + ".hidden")
	}

	// Radio settings
	if device != "" && cfg.Channel != "" {
		uciSet(fmt.Sprintf("wireless.%s.channel", device), cfg.Channel)
	}

	uciCommit("wireless")
	exec.Command("wifi", "reload").Run()
	return nil
}

// ═══════════════════════════════════
// DHCP LEASES
// ═══════════════════════════════════

func (m *Manager) GetDHCPLeases() []DHCPLease {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var leases []DHCPLease

	data, err := exec.Command("cat", "/tmp/dhcp.leases").Output()
	if err != nil {
		return leases
	}

	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			leases = append(leases, DHCPLease{
				Expire:   fields[0],
				MAC:      strings.ToUpper(fields[1]),
				IP:       fields[2],
				Hostname: fields[3],
			})
		}
	}
	return leases
}

func (m *Manager) GetStaticLeases() []StaticLease {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var leases []StaticLease
	out, _ := exec.Command("uci", "show", "dhcp").Output()

	count := 0
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "=host") {
			count++
		}
	}

	for i := 0; i < count; i++ {
		prefix := fmt.Sprintf("dhcp.@host[%d]", i)
		leases = append(leases, StaticLease{
			MAC:  uciGet(prefix + ".mac"),
			IP:   uciGet(prefix + ".ip"),
			Name: uciGet(prefix + ".name"),
		})
	}
	return leases
}

func (m *Manager) AddStaticLease(mac, ip, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	exec.Command("uci", "add", "dhcp", "host").Run()

	// Find the new index
	out, _ := exec.Command("uci", "show", "dhcp").Output()
	count := 0
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "=host") {
			count++
		}
	}
	idx := count - 1

	prefix := fmt.Sprintf("dhcp.@host[%d]", idx)
	uciSet(prefix+".mac", mac)
	uciSet(prefix+".ip", ip)
	if name != "" {
		uciSet(prefix+".name", name)
	}
	uciSet(prefix+".dns", "1")

	uciCommit("dhcp")
	exec.Command("/etc/init.d/dnsmasq", "restart").Run()
	return nil
}

func (m *Manager) DeleteStaticLease(index int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	exec.Command("uci", "delete", fmt.Sprintf("dhcp.@host[%d]", index)).Run()
	uciCommit("dhcp")
	exec.Command("/etc/init.d/dnsmasq", "restart").Run()
	return nil
}

// ═══════════════════════════════════
// INTERFACE STATUS
// ═══════════════════════════════════

func (m *Manager) GetInterfaces() []InterfaceStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var ifaces []InterfaceStatus

	out, err := exec.Command("ip", "-j", "addr", "show").Output()
	if err != nil {
		// Fallback to non-JSON
		return m.getInterfacesFallback()
	}

	// Parse JSON output from ip command
	// For simplicity, use the text-based fallback
	_ = out
	return m.getInterfacesFallback()
}

func (m *Manager) getInterfacesFallback() []InterfaceStatus {
	var ifaces []InterfaceStatus

	out, _ := exec.Command("ip", "addr", "show").Output()
	lines := strings.Split(string(out), "\n")

	var current *InterfaceStatus
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// New interface line: "2: eth0: <BROADCAST,..."
		if len(line) > 0 && line[0] >= '0' && line[0] <= '9' {
			if current != nil {
				ifaces = append(ifaces, *current)
			}
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				name := strings.TrimRight(fields[1], ":")
				if name == "lo" {
					current = nil
					continue
				}
				current = &InterfaceStatus{
					Name: name,
					Up:   strings.Contains(line, "UP"),
				}
			}
		}
		if current == nil {
			continue
		}
		if strings.HasPrefix(line, "link/ether") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				current.MAC = strings.ToUpper(fields[1])
			}
		}
		if strings.HasPrefix(line, "inet ") && !strings.HasPrefix(line, "inet6") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				parts := strings.Split(fields[1], "/")
				current.IP = parts[0]
				if len(parts) > 1 {
					prefix, _ := strconv.Atoi(parts[1])
					current.Netmask = prefixToNetmask(prefix)
				}
			}
		}
	}
	if current != nil {
		ifaces = append(ifaces, *current)
	}

	// Add RX/TX stats
	for i := range ifaces {
		rxData, _ := exec.Command("cat", fmt.Sprintf("/sys/class/net/%s/statistics/rx_bytes", ifaces[i].Name)).Output()
		txData, _ := exec.Command("cat", fmt.Sprintf("/sys/class/net/%s/statistics/tx_bytes", ifaces[i].Name)).Output()
		ifaces[i].RX, _ = strconv.ParseInt(strings.TrimSpace(string(rxData)), 10, 64)
		ifaces[i].TX, _ = strconv.ParseInt(strings.TrimSpace(string(txData)), 10, 64)

		speedData, _ := exec.Command("cat", fmt.Sprintf("/sys/class/net/%s/speed", ifaces[i].Name)).Output()
		speed := strings.TrimSpace(string(speedData))
		if speed != "" && speed != "-1" {
			ifaces[i].Speed = speed + " Mbps"
		}
	}

	return ifaces
}

func prefixToNetmask(prefix int) string {
	mask := uint32(0xFFFFFFFF) << (32 - prefix)
	return fmt.Sprintf("%d.%d.%d.%d",
		(mask>>24)&0xFF, (mask>>16)&0xFF, (mask>>8)&0xFF, mask&0xFF)
}

func restartNetwork() error {
	log.Println("[*] Restarting network...")
	return exec.Command("/etc/init.d/network", "restart").Run()
}
