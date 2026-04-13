package devices

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Device struct {
	IP       string `json:"ip"`
	MAC      string `json:"mac"`
	Hostname string `json:"hostname"`
	Alias    string `json:"alias"`
	Vendor   string `json:"vendor"`
	LastSeen int64  `json:"last_seen"`
	Online   bool   `json:"online"`
}

type Tracker struct {
	leasesFile string
	devices    map[string]*Device // keyed by MAC
	aliases    map[string]string  // MAC -> user-assigned alias
	mu         sync.RWMutex
}

func NewTracker(leasesFile string) *Tracker {
	return &Tracker{
		leasesFile: leasesFile,
		devices:    make(map[string]*Device),
		aliases:    make(map[string]string),
	}
}

func (t *Tracker) ScanLoop(interval time.Duration) {
	t.scan()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		t.scan()
	}
}

func (t *Tracker) scan() {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Mark all offline first
	for _, d := range t.devices {
		d.Online = false
	}

	// Parse DHCP leases: "timestamp mac ip hostname clientid"
	if f, err := os.Open(t.leasesFile); err == nil {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) < 4 {
				continue
			}
			mac := strings.ToLower(fields[1])
			ip := fields[2]
			hostname := fields[3]
			if hostname == "*" {
				hostname = ""
			}

			dev, exists := t.devices[mac]
			if !exists {
				dev = &Device{MAC: mac}
				t.devices[mac] = dev
			}
			dev.IP = ip
			dev.Hostname = hostname
			if alias, ok := t.aliases[mac]; ok {
				dev.Alias = alias
			}
		}
		f.Close()
	}

	// Parse ARP table for online status
	out, err := exec.Command("ip", "neigh", "show").Output()
	if err != nil {
		log.Printf("arp scan error: %v", err)
		return
	}

	now := time.Now().Unix()
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		ip := fields[0]
		mac := strings.ToLower(fields[4])
		state := fields[len(fields)-1]

		if state == "FAILED" {
			continue
		}

		dev, exists := t.devices[mac]
		if !exists {
			dev = &Device{MAC: mac, IP: ip}
			t.devices[mac] = dev
		}
		dev.Online = true
		dev.LastSeen = now
		dev.IP = ip

		if alias, ok := t.aliases[mac]; ok {
			dev.Alias = alias
		}
		if dev.Vendor == "" {
			dev.Vendor = lookupVendor(mac)
		}
	}
}

func (t *Tracker) List() []Device {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]Device, 0, len(t.devices))
	for _, d := range t.devices {
		result = append(result, *d)
	}
	return result
}

func (t *Tracker) SetAlias(mac, alias string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	mac = strings.ToLower(mac)
	t.aliases[mac] = alias
	if dev, ok := t.devices[mac]; ok {
		dev.Alias = alias
	}
}

func (t *Tracker) Resolve(ip, mac string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	mac = strings.ToLower(mac)
	if dev, ok := t.devices[mac]; ok {
		if dev.Alias != "" {
			return dev.Alias
		}
		if dev.Hostname != "" {
			return dev.Hostname
		}
	}
	return ip
}

// lookupVendor returns a short vendor name from the MAC OUI.
// This is a minimal built-in list; extend or use an OUI database file.
func lookupVendor(mac string) string {
	if len(mac) < 8 {
		return ""
	}
	oui := strings.ToUpper(strings.ReplaceAll(mac[:8], ":", ""))

	vendors := map[string]string{
		"AABBCC": "Apple",
		"001A2B": "Apple",
		"3C22FB": "Apple",
		"A4C3F0": "Apple",
		"DC56E7": "Apple",
		"F0D5BF": "Apple",
		"002248": "Microsoft",
		"7C5CF8": "Samsung",
		"E4FAED": "Dell",
		"B8AC6F": "Dell",
		"000C29": "VMware",
		"D8BB2C": "Google",
		"54602B": "Google",
	}

	if v, ok := vendors[oui]; ok {
		return v
	}

	// Try to read from /usr/share/oui if available
	parsed, err := net.ParseMAC(mac)
	if err != nil {
		return ""
	}
	_ = parsed
	return ""
}

func (t *Tracker) GetByIP(ip string) *Device {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, d := range t.devices {
		if d.IP == ip {
			cp := *d
			return &cp
		}
	}
	return nil
}

func (t *Tracker) GetByMAC(mac string) *Device {
	t.mu.RLock()
	defer t.mu.RUnlock()

	mac = strings.ToLower(mac)
	if d, ok := t.devices[mac]; ok {
		cp := *d
		return &cp
	}
	return nil
}

func (t *Tracker) Count() (total, online int) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	total = len(t.devices)
	for _, d := range t.devices {
		if d.Online {
			online++
		}
	}
	return
}

func (t *Tracker) String() string {
	total, online := t.Count()
	return fmt.Sprintf("%d devices (%d online)", total, online)
}
