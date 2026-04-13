package dns

import (
	"fmt"
	"log"
	"os/exec"
	"sync"
)

// Known DoH resolver IPs — comprehensive list
var defaultDoHResolvers = []string{
	// Cloudflare
	"1.1.1.1", "1.0.0.1",
	"2606:4700:4700::1111", "2606:4700:4700::1001",
	// Google
	"8.8.8.8", "8.8.4.4",
	"2001:4860:4860::8888", "2001:4860:4860::8844",
	// Quad9
	"9.9.9.9", "149.112.112.112",
	"2620:fe::fe", "2620:fe::9",
	// OpenDNS
	"208.67.222.222", "208.67.220.220",
	// AdGuard
	"94.140.14.14", "94.140.15.15",
	// NextDNS
	"45.90.28.0", "45.90.30.0",
	// Control D
	"76.76.2.0", "76.76.10.0",
	// Mullvad
	"194.242.2.2",
}

const iptablesChain = "CERBERUS_DOH"

type DoHBlocker struct {
	enabled bool
	mu      sync.Mutex
}

func NewDoHBlocker(enabled bool) *DoHBlocker {
	return &DoHBlocker{enabled: enabled}
}

func (b *DoHBlocker) IsEnabled() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.enabled
}

func (b *DoHBlocker) Enable() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.enabled {
		return nil
	}

	if err := b.createRules(); err != nil {
		return err
	}
	b.enabled = true
	log.Println("DoH blocker enabled")
	return nil
}

func (b *DoHBlocker) Disable() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.enabled {
		return nil
	}

	if err := b.removeRules(); err != nil {
		return err
	}
	b.enabled = false
	log.Println("DoH blocker disabled")
	return nil
}

func (b *DoHBlocker) Toggle() (bool, error) {
	if b.IsEnabled() {
		return false, b.Disable()
	}
	return true, b.Enable()
}

func (b *DoHBlocker) createRules() error {
	// Create custom chain
	run("iptables", "-N", iptablesChain)
	run("ip6tables", "-N", iptablesChain)

	// Flush in case of stale rules
	run("iptables", "-F", iptablesChain)
	run("ip6tables", "-F", iptablesChain)

	// Block each DoH resolver IP on ports 443 and 853 (DoT)
	for _, ip := range defaultDoHResolvers {
		cmd := "iptables"
		if isIPv6(ip) {
			cmd = "ip6tables"
		}

		// Block HTTPS (443) to DoH resolvers
		if err := run(cmd, "-A", iptablesChain, "-d", ip, "-p", "tcp", "--dport", "443", "-j", "REJECT"); err != nil {
			log.Printf("warning: %s block %s:443 failed: %v", cmd, ip, err)
		}

		// Block DoT (853)
		if err := run(cmd, "-A", iptablesChain, "-d", ip, "-p", "tcp", "--dport", "853", "-j", "REJECT"); err != nil {
			log.Printf("warning: %s block %s:853 failed: %v", cmd, ip, err)
		}
	}

	// Insert chain into FORWARD
	run("iptables", "-I", "FORWARD", "-j", iptablesChain)
	run("ip6tables", "-I", "FORWARD", "-j", iptablesChain)

	return nil
}

func (b *DoHBlocker) removeRules() error {
	// Remove from FORWARD
	run("iptables", "-D", "FORWARD", "-j", iptablesChain)
	run("ip6tables", "-D", "FORWARD", "-j", iptablesChain)

	// Flush and delete chain
	run("iptables", "-F", iptablesChain)
	run("ip6tables", "-F", iptablesChain)
	run("iptables", "-X", iptablesChain)
	run("ip6tables", "-X", iptablesChain)

	return nil
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v: %s: %w", name, args, string(out), err)
	}
	return nil
}

func isIPv6(ip string) bool {
	for _, c := range ip {
		if c == ':' {
			return true
		}
	}
	return false
}
