package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"cerberus/adapters"
	"cerberus/captive"
	"cerberus/deauth"
	"cerberus/eviltwin"
	"cerberus/handshake"
	"cerberus/mitm"
	"cerberus/network"
	"cerberus/scanner"
	"cerberus/system"
)

type Router struct {
	scan  *scanner.Scanner
	mitm  *mitm.Engine
	dea   *deauth.Engine
	et    *eviltwin.Engine
	cap   *captive.Engine
	hs    *handshake.Engine
	adpt  *adapters.Manager
	net   *network.Manager
}

func NewRouter(s *scanner.Scanner, m *mitm.Engine, d *deauth.Engine, e *eviltwin.Engine, c *captive.Engine, h *handshake.Engine, a *adapters.Manager, n *network.Manager) http.Handler {
	r := &Router{scan: s, mitm: m, dea: d, et: e, cap: c, hs: h, adpt: a, net: n}
	mux := http.NewServeMux()
	w := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == "OPTIONS" { w.WriteHeader(200); return }
			h(w, r)
		}
	}

	// Scanner
	mux.HandleFunc("/api/scan", w(func(rw http.ResponseWriter, rq *http.Request) { go r.scan.Scan(); j(rw, M{"status": "scanning"}) }))
	mux.HandleFunc("/api/clients", w(func(rw http.ResponseWriter, rq *http.Request) { j(rw, r.scan.GetClients()) }))
	mux.HandleFunc("/api/networks", w(func(rw http.ResponseWriter, rq *http.Request) { j(rw, r.scan.GetNetworks()) }))
	mux.HandleFunc("/api/probes", w(func(rw http.ResponseWriter, rq *http.Request) { j(rw, r.scan.GetProbes()) }))

	// MITM
	mux.HandleFunc("/api/mitm/start", w(func(rw http.ResponseWriter, rq *http.Request) { var t T; dec(rq, &t); r.mitm.StartTarget(t.MAC, t.IP); j(rw, M{"status": "ok"}) }))
	mux.HandleFunc("/api/mitm/stop", w(func(rw http.ResponseWriter, rq *http.Request) { var t T; dec(rq, &t); r.mitm.StopTarget(t.MAC); j(rw, M{"status": "ok"}) }))
	mux.HandleFunc("/api/mitm/targets", w(func(rw http.ResponseWriter, rq *http.Request) { j(rw, r.mitm.GetActiveTargets()) }))
	mux.HandleFunc("/api/mitm/dns", w(func(rw http.ResponseWriter, rq *http.Request) { j(rw, r.mitm.GetDNSLog()) }))

	// Deauth
	mux.HandleFunc("/api/deauth/start", w(func(rw http.ResponseWriter, rq *http.Request) { var t T; dec(rq, &t); r.dea.Start(t.MAC, t.BSSID); j(rw, M{"status": "ok"}) }))
	mux.HandleFunc("/api/deauth/stop", w(func(rw http.ResponseWriter, rq *http.Request) { var t T; dec(rq, &t); r.dea.Stop(t.MAC); j(rw, M{"status": "ok"}) }))
	mux.HandleFunc("/api/deauth/targets", w(func(rw http.ResponseWriter, rq *http.Request) { j(rw, r.dea.GetActiveTargets()) }))

	// Evil Twin
	mux.HandleFunc("/api/eviltwin/start", w(func(rw http.ResponseWriter, rq *http.Request) { var e struct{ SSID, Channel, Iface string }; dec(rq, &e); r.et.Start(e.SSID, e.Channel, e.Iface); j(rw, M{"status": "ok"}) }))
	mux.HandleFunc("/api/eviltwin/stop", w(func(rw http.ResponseWriter, rq *http.Request) { r.et.Stop(); j(rw, M{"status": "ok"}) }))
	mux.HandleFunc("/api/eviltwin/status", w(func(rw http.ResponseWriter, rq *http.Request) { j(rw, r.et.GetConfig()) }))

	// Captive
	mux.HandleFunc("/api/captive/start", w(func(rw http.ResponseWriter, rq *http.Request) { var c struct{ Template string }; dec(rq, &c); r.cap.Start(c.Template); j(rw, M{"status": "ok"}) }))
	mux.HandleFunc("/api/captive/stop", w(func(rw http.ResponseWriter, rq *http.Request) { r.cap.Stop(); j(rw, M{"status": "ok"}) }))
	mux.HandleFunc("/api/captive/creds", w(func(rw http.ResponseWriter, rq *http.Request) { j(rw, r.cap.GetCredentials()) }))

	// Handshake
	mux.HandleFunc("/api/handshake/start", w(func(rw http.ResponseWriter, rq *http.Request) { var h struct{ BSSID, SSID, Channel string }; dec(rq, &h); r.hs.Start(h.BSSID, h.SSID, h.Channel); j(rw, M{"status": "ok"}) }))
	mux.HandleFunc("/api/handshake/stop", w(func(rw http.ResponseWriter, rq *http.Request) { r.hs.Stop(); j(rw, M{"status": "ok"}) }))
	mux.HandleFunc("/api/handshake/status", w(func(rw http.ResponseWriter, rq *http.Request) { j(rw, r.hs.GetStatus()) }))
	mux.HandleFunc("/api/handshake/captures", w(func(rw http.ResponseWriter, rq *http.Request) { j(rw, r.hs.ListCapFiles()) }))
	mux.HandleFunc("/api/handshake/download/", w(func(rw http.ResponseWriter, rq *http.Request) {
		fn := filepath.Base(rq.URL.Path)
		p := r.hs.GetCapFilePath(fn)
		if _, err := os.Stat(p); err != nil { je(rw, "not found", 404); return }
		rw.Header().Set("Content-Disposition", "attachment; filename="+fn)
		http.ServeFile(rw, rq, p)
	}))

	// Adapters
	mux.HandleFunc("/api/adapters", w(func(rw http.ResponseWriter, rq *http.Request) { j(rw, r.adpt.GetAdapters()) }))
	mux.HandleFunc("/api/adapters/role", w(func(rw http.ResponseWriter, rq *http.Request) { var a struct{ Adapter, Role string }; dec(rq, &a); r.adpt.SetRole(a.Role, a.Adapter); j(rw, M{"status": "ok"}) }))
	mux.HandleFunc("/api/adapters/detect", w(func(rw http.ResponseWriter, rq *http.Request) { r.adpt.Detect(); j(rw, r.adpt.GetAdapters()) }))

	// System
	mux.HandleFunc("/api/system/info", w(func(rw http.ResponseWriter, rq *http.Request) { j(rw, system.GetInfo()) }))
	mux.HandleFunc("/api/system/reboot", w(func(rw http.ResponseWriter, rq *http.Request) { j(rw, M{"status": "rebooting"}); go system.Reboot() }))
	mux.HandleFunc("/api/system/hostname", w(func(rw http.ResponseWriter, rq *http.Request) { var h struct{ Hostname string }; dec(rq, &h); system.SetHostname(h.Hostname); j(rw, M{"status": "ok"}) }))
	mux.HandleFunc("/api/system/firmware", w(system.FlashFirmware))

	// Network
	mux.HandleFunc("/api/network/wan", w(func(rw http.ResponseWriter, rq *http.Request) {
		if rq.Method == "GET" { j(rw, r.net.GetWAN()); return }
		var cfg network.WANConfig; dec(rq, &cfg); r.net.SetWAN(cfg); j(rw, M{"status": "ok"})
	}))
	mux.HandleFunc("/api/network/lan", w(func(rw http.ResponseWriter, rq *http.Request) {
		if rq.Method == "GET" { j(rw, r.net.GetLAN()); return }
		var cfg network.LANConfig; dec(rq, &cfg); r.net.SetLAN(cfg); j(rw, M{"status": "ok"})
	}))
	mux.HandleFunc("/api/network/wifi", w(func(rw http.ResponseWriter, rq *http.Request) {
		if rq.Method == "GET" { j(rw, r.net.GetWiFi()); return }
		var b struct { Index int `json:"index"`; Cfg network.WiFiConfig `json:"config"` }; dec(rq, &b); r.net.SetWiFi(b.Index, b.Cfg); j(rw, M{"status": "ok"})
	}))
	mux.HandleFunc("/api/network/interfaces", w(func(rw http.ResponseWriter, rq *http.Request) { j(rw, r.net.GetInterfaces()) }))
	mux.HandleFunc("/api/network/dhcp/leases", w(func(rw http.ResponseWriter, rq *http.Request) { j(rw, r.net.GetDHCPLeases()) }))
	mux.HandleFunc("/api/network/dhcp/static", w(func(rw http.ResponseWriter, rq *http.Request) {
		if rq.Method == "GET" { j(rw, r.net.GetStaticLeases()); return }
		if rq.Method == "POST" { var l struct{ MAC, IP, Name string }; dec(rq, &l); r.net.AddStaticLease(l.MAC, l.IP, l.Name); j(rw, M{"status": "ok"}); return }
		if rq.Method == "DELETE" { var d struct{ Index int }; dec(rq, &d); r.net.DeleteStaticLease(d.Index); j(rw, M{"status": "ok"}); return }
	}))

	// Status
	mux.HandleFunc("/api/status", w(func(rw http.ResponseWriter, rq *http.Request) {
		si := system.GetInfo()
		j(rw, M{"scanning": r.scan.IsScanning(), "clients": len(r.scan.GetClients()), "mitm": len(r.mitm.GetActiveTargets()), "deauth": len(r.dea.GetActiveTargets()), "eviltwin": r.et.GetConfig().Active, "captive": r.cap.IsActive(), "handshake": r.hs.GetStatus().State, "uptime": si.Uptime, "cpu": si.CPUUsage, "mem": si.MemPercent, "hostname": si.Hostname})
	}))

	// Frontend
	mux.Handle("/", http.FileServer(http.Dir("/www/cerberus")))
	return mux
}

type M = map[string]interface{}
type T struct { MAC, IP, BSSID string }

func j(w http.ResponseWriter, d interface{}) { w.Header().Set("Content-Type", "application/json"); json.NewEncoder(w).Encode(d) }
func je(w http.ResponseWriter, m string, c int) { w.WriteHeader(c); j(w, M{"error": m}) }
func dec(r *http.Request, v interface{}) { json.NewDecoder(r.Body).Decode(v) }
func init() { log.SetFlags(log.Ltime | log.Lshortfile) }
