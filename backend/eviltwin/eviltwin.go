package eviltwin

import ("fmt";"log";"os";"os/exec";"sync")

type Config struct { SSID string `json:"ssid"`; Channel string `json:"channel"`; Interface string `json:"interface"`; Active bool `json:"active"` }
type Engine struct { mu sync.RWMutex; config Config; h, d *exec.Cmd }

func New() *Engine { return &Engine{config: Config{Channel: "6", Interface: "wlan0"}} }
func (e *Engine) GetConfig() Config { e.mu.RLock(); defer e.mu.RUnlock(); return e.config }
func (e *Engine) Start(ssid, ch, iface string) error {
	e.mu.Lock(); defer e.mu.Unlock()
	if e.config.Active { return fmt.Errorf("running") }
	log.Printf("[*] Twin: %s CH%s %s", ssid, ch, iface)
	os.WriteFile("/tmp/cerberus_hostapd.conf", []byte(fmt.Sprintf("interface=%s\ndriver=nl80211\nssid=%s\nchannel=%s\nhw_mode=g\n", iface, ssid, ch)), 0644)
	os.WriteFile("/tmp/cerberus_dnsmasq.conf", []byte(fmt.Sprintf("interface=%s\ndhcp-range=192.168.4.2,192.168.4.254,255.255.255.0,12h\n", iface)), 0644)
	exec.Command("ifconfig", iface, "192.168.4.1", "netmask", "255.255.255.0", "up").Run()
	exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-o", "eth0", "-j", "MASQUERADE").Run()
	os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1"), 0644)
	e.h = exec.Command("hostapd", "/tmp/cerberus_hostapd.conf"); e.h.Start()
	e.d = exec.Command("dnsmasq", "-C", "/tmp/cerberus_dnsmasq.conf", "--no-daemon"); e.d.Start()
	e.config = Config{ssid, ch, iface, true}; return nil
}
func (e *Engine) Stop() {
	e.mu.Lock(); defer e.mu.Unlock()
	if !e.config.Active { return }
	if e.h != nil { e.h.Process.Kill() }; if e.d != nil { e.d.Process.Kill() }
	e.config.Active = false; log.Println("[-] Twin stopped")
}
