package dns

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
)

var vpnDomains = []string{
	"nordvpn.com", "nord-apps.com", "nordcdn.com",
	"expressvpn.com", "expressapisv2.net", "xvpn.io",
	"surfshark.com", "surfsharkdns.com",
	"protonvpn.com", "protonvpn.net", "proton.me",
	"privateinternetaccess.com", "piavpn.com",
	"cyberghostvpn.com", "cg-dialup.net",
	"mullvad.net",
	"ipvanish.com",
	"tunnelbear.com",
	"windscribe.com", "windscribe.net",
	"hotspotshield.com", "hsselite.com",
	"hide.me",
	"vyprvpn.com", "goldenfrog.com",
	"atlasvpn.com",
	"mask.icloud.com", "mask-h2.icloud.com",
	"cloudflarewarp.com",
}

type VPNPort struct {
	Port     string `json:"port"`
	Protocol string `json:"protocol"`
	Name     string `json:"name"`
}

var vpnPorts = []VPNPort{
	{"1194", "udp", "OpenVPN-UDP"},
	{"1194", "tcp", "OpenVPN-TCP"},
	{"443", "udp", "OpenVPN-HTTPS/QUIC"},
	{"51820", "udp", "WireGuard"},
	{"500", "udp", "IPSec-IKE"},
	{"4500", "udp", "IPSec-NAT-T"},
	{"1701", "udp", "L2TP"},
	{"1723", "tcp", "PPTP"},
}

const vpnChain = "CERBERUS_VPN"

type VPNBlocker struct {
	dnsEnabled   bool
	portsEnabled bool
	mu           sync.Mutex
}

type VPNStatus struct {
	DNSBlocking  bool      `json:"dns_blocking"`
	PortBlocking bool      `json:"port_blocking"`
	Domains      []string  `json:"domains"`
	Ports        []VPNPort `json:"ports"`
}

func NewVPNBlocker() *VPNBlocker {
	return &VPNBlocker{}
}

func (v *VPNBlocker) Status() VPNStatus {
	v.mu.Lock()
	defer v.mu.Unlock()
	return VPNStatus{
		DNSBlocking:  v.dnsEnabled,
		PortBlocking: v.portsEnabled,
		Domains:      vpnDomains,
		Ports:        vpnPorts,
	}
}

func (v *VPNBlocker) EnableDNS() error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dnsEnabled {
		return nil
	}
	var lines string
	for _, domain := range vpnDomains {
		lines += fmt.Sprintf("address=/%s/\n", domain)
	}
	if err := os.WriteFile("/tmp/cerberus-vpn-blocklist.conf", []byte(lines), 0644); err != nil {
		return fmt.Errorf("write vpn blocklist: %w", err)
	}
	exec.Command("sh", "-c", "kill -HUP $(pidof dnsmasq)").Run()
	v.dnsEnabled = true
	log.Println("vpn blocker: DNS blocking enabled")
	return nil
}

func (v *VPNBlocker) DisableDNS() error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if !v.dnsEnabled {
		return nil
	}
	os.Remove("/tmp/cerberus-vpn-blocklist.conf")
	exec.Command("sh", "-c", "kill -HUP $(pidof dnsmasq)").Run()
	v.dnsEnabled = false
	log.Println("vpn blocker: DNS blocking disabled")
	return nil
}

func (v *VPNBlocker) EnablePorts() error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.portsEnabled {
		return nil
	}
	run("iptables", "-N", vpnChain)
	run("iptables", "-F", vpnChain)
	for _, p := range vpnPorts {
		if p.Port == "443" && p.Protocol == "tcp" {
			continue // skip — would break all HTTPS
		}
		if err := run("iptables", "-A", vpnChain,
			"-p", p.Protocol, "--dport", p.Port,
			"-j", "REJECT"); err != nil {
			log.Printf("vpn blocker: block %s/%s (%s) failed: %v", p.Port, p.Protocol, p.Name, err)
		}
	}
	run("iptables", "-I", "FORWARD", "-j", vpnChain)
	v.portsEnabled = true
	log.Println("vpn blocker: port blocking enabled")
	return nil
}

func (v *VPNBlocker) DisablePorts() error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if !v.portsEnabled {
		return nil
	}
	run("iptables", "-D", "FORWARD", "-j", vpnChain)
	run("iptables", "-F", vpnChain)
	run("iptables", "-X", vpnChain)
	v.portsEnabled = false
	log.Println("vpn blocker: port blocking disabled")
	return nil
}

func (v *VPNBlocker) EnableAll() error {
	if err := v.EnableDNS(); err != nil {
		return err
	}
	return v.EnablePorts()
}

func (v *VPNBlocker) DisableAll() {
	v.DisableDNS()
	v.DisablePorts()
}

func (v *VPNBlocker) GetDomains() []string {
	return vpnDomains
}

func (v *VPNBlocker) AddDomain(domain string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	for _, d := range vpnDomains {
		if d == domain {
			return
		}
	}
	vpnDomains = append(vpnDomains, domain)
}

func (v *VPNBlocker) RemoveDomain(domain string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	for i, d := range vpnDomains {
		if d == domain {
			vpnDomains = append(vpnDomains[:i], vpnDomains[i+1:]...)
			return
		}
	}
}
