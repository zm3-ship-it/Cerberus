package adapters

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

type Adapter struct {
	ID    string   `json:"id"`
	Chip  string   `json:"chip"`
	MAC   string   `json:"mac"`
	Modes []string `json:"modes"`
	Band  string   `json:"band"`
	Role  string   `json:"role"` // scan, attack, upstream
}

type Manager struct {
	mu       sync.RWMutex
	adapters []Adapter
	roles    map[string]string // role -> adapter id
	filePath string
}

func New(dataDir string) *Manager {
	m := &Manager{
		roles:    map[string]string{"scan": "", "attack": "", "upstream": ""},
		filePath: dataDir + "/adapters.json",
	}
	m.Detect()
	m.loadRoles()
	return m
}

func (m *Manager) GetAdapters() []Adapter {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Adapter, len(m.adapters))
	copy(out, m.adapters)
	for i := range out {
		for role, id := range m.roles {
			if id == out[i].ID {
				out[i].Role = role
			}
		}
	}
	return out
}

func (m *Manager) GetRoles() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]string)
	for k, v := range m.roles {
		out[k] = v
	}
	return out
}

func (m *Manager) SetRole(role, adapterID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	valid := false
	for _, a := range m.adapters {
		if a.ID == adapterID {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("adapter %s not found", adapterID)
	}

	if role != "scan" && role != "attack" && role != "upstream" {
		return fmt.Errorf("invalid role: %s", role)
	}

	// Swap: if another role has this adapter, give it the old adapter
	oldAdapter := m.roles[role]
	for r, id := range m.roles {
		if id == adapterID && r != role {
			m.roles[r] = oldAdapter
		}
	}
	m.roles[role] = adapterID

	m.saveRoles()
	log.Printf("[*] Adapter %s assigned to role: %s", adapterID, role)
	return nil
}

// Detect discovers wireless interfaces
func (m *Manager) Detect() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.adapters = nil

	// List wireless interfaces
	cmd := exec.Command("iw", "dev")
	out, err := cmd.Output()
	if err != nil {
		log.Printf("[!] Failed to detect adapters: %v", err)
		return
	}

	ifaceRe := regexp.MustCompile(`Interface\s+(\w+)`)
	macRe := regexp.MustCompile(`addr\s+([0-9a-fA-F:]+)`)

	blocks := strings.Split(string(out), "Interface")
	for _, block := range blocks[1:] {
		ifMatch := ifaceRe.FindStringSubmatch("Interface" + block)
		macMatch := macRe.FindStringSubmatch(block)

		if ifMatch == nil {
			continue
		}

		iface := ifMatch[1]
		mac := ""
		if macMatch != nil {
			mac = strings.ToUpper(macMatch[1])
		}

		adapter := Adapter{
			ID:   iface,
			MAC:  mac,
			Chip: getChipset(iface),
			Band: getBand(iface),
		}

		// Check capabilities
		adapter.Modes = getModes(iface)

		m.adapters = append(m.adapters, adapter)
	}

	// Also check for eth0 as upstream
	if ifExists("eth0") {
		m.adapters = append(m.adapters, Adapter{
			ID:    "eth0",
			MAC:   getMAC("eth0"),
			Chip:  "Ethernet",
			Band:  "Wired",
			Modes: []string{"Managed"},
		})
	}

	// Auto-assign roles if not set
	if m.roles["upstream"] == "" {
		for _, a := range m.adapters {
			if a.ID == "eth0" || strings.Contains(a.Chip, "Ethernet") {
				m.roles["upstream"] = a.ID
				break
			}
		}
	}
	if m.roles["scan"] == "" || m.roles["attack"] == "" {
		for _, a := range m.adapters {
			if a.ID == m.roles["upstream"] {
				continue
			}
			hasInjection := false
			for _, mode := range a.Modes {
				if mode == "Monitor" || mode == "Injection" {
					hasInjection = true
				}
			}
			if hasInjection {
				if m.roles["scan"] == "" {
					m.roles["scan"] = a.ID
				} else if m.roles["attack"] == "" {
					m.roles["attack"] = a.ID
				}
			}
		}
	}

	log.Printf("[*] Detected %d adapters. Roles: scan=%s attack=%s upstream=%s",
		len(m.adapters), m.roles["scan"], m.roles["attack"], m.roles["upstream"])
}

func (m *Manager) loadRoles() {
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return
	}
	json.Unmarshal(data, &m.roles)
}

func (m *Manager) saveRoles() {
	data, _ := json.MarshalIndent(m.roles, "", "  ")
	os.WriteFile(m.filePath, data, 0644)
}

func getChipset(iface string) string {
	cmd := exec.Command("ethtool", "-i", iface)
	out, err := cmd.Output()
	if err != nil {
		// Try reading from /sys
		data, err := os.ReadFile(fmt.Sprintf("/sys/class/net/%s/device/uevent", iface))
		if err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(line, "DRIVER=") {
					return strings.TrimPrefix(line, "DRIVER=")
				}
			}
		}
		return "Unknown"
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "driver:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "driver:"))
		}
	}
	return "Unknown"
}

func getBand(iface string) string {
	cmd := exec.Command("iw", "phy")
	out, _ := cmd.Output()
	if strings.Contains(string(out), "5180") {
		return "2.4/5 GHz"
	}
	return "2.4 GHz"
}

func getModes(iface string) []string {
	var modes []string
	modes = append(modes, "Managed")

	// Check monitor mode support
	cmd := exec.Command("iw", "list")
	out, _ := cmd.Output()
	if strings.Contains(string(out), "monitor") {
		modes = append(modes, "Monitor")
	}

	// Check injection support
	cmd = exec.Command("aireplay-ng", "--test", iface)
	out, err := cmd.Output()
	if err == nil && strings.Contains(string(out), "Injection is working") {
		modes = append(modes, "Injection")
	}

	return modes
}

func ifExists(iface string) bool {
	_, err := os.Stat(fmt.Sprintf("/sys/class/net/%s", iface))
	return err == nil
}

func getMAC(iface string) string {
	data, err := os.ReadFile(fmt.Sprintf("/sys/class/net/%s/address", iface))
	if err != nil {
		return ""
	}
	return strings.ToUpper(strings.TrimSpace(string(data)))
}
