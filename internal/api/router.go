package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"cerberus/internal/captive"
	"cerberus/internal/config"
	"cerberus/internal/deauth"
	"cerberus/internal/devices"
	"cerberus/internal/dns"
	"cerberus/internal/eviltwin"
	"cerberus/internal/handshake"
	"cerberus/internal/mitm"
	"cerberus/internal/recon"
)

type Modules struct {
	DNSLog     *dns.Logger
	Devices    *devices.Tracker
	DoHBlocker *dns.DoHBlocker
	VPNBlocker *dns.VPNBlocker
	Config     *config.Config
	Scanner    *recon.Scanner
	Deauth     *deauth.Manager
	EvilTwin   *eviltwin.Manager
	Handshake  *handshake.Manager
	MITM       *mitm.Manager
	Captive    *captive.Manager
}

func NewRouter(m *Modules) http.Handler {
	mux := http.NewServeMux()

	// ── Core: DNS monitoring ─────────────────────────────────────
	mux.HandleFunc("/api/stats", handleStats(m.DNSLog))
	mux.HandleFunc("/api/queries", handleQueries(m.DNSLog))
	mux.HandleFunc("/api/top-domains", handleTopDomains(m.DNSLog))
	mux.HandleFunc("/api/devices", handleDevices(m.Devices))
	mux.HandleFunc("/api/devices/alias", handleSetAlias(m.Devices))
	mux.HandleFunc("/api/doh/status", handleDoHStatus(m.DoHBlocker))
	mux.HandleFunc("/api/doh/toggle", handleDoHToggle(m.DoHBlocker))
	mux.HandleFunc("/api/config", handleConfig(m.Config))
	mux.HandleFunc("/api/purge", handlePurge(m.DNSLog))

	// ── Recon: wireless scanning ─────────────────────────────────
	mux.HandleFunc("/api/recon/monitor/enable", handleReconMonitorEnable(m.Scanner))
	mux.HandleFunc("/api/recon/monitor/disable", handleReconMonitorDisable(m.Scanner))
	mux.HandleFunc("/api/recon/scan/start", handleReconScanStart(m.Scanner))
	mux.HandleFunc("/api/recon/scan/stop", handleReconScanStop(m.Scanner))
	mux.HandleFunc("/api/recon/aps", handleReconAPs(m.Scanner))
	mux.HandleFunc("/api/recon/clients", handleReconClients(m.Scanner))
	mux.HandleFunc("/api/recon/arpscan", handleARPScan())

	// ── Deauth ───────────────────────────────────────────────────
	mux.HandleFunc("/api/deauth/start", handleDeauthStart(m.Deauth))
	mux.HandleFunc("/api/deauth/stop", handleDeauthStop(m.Deauth))
	mux.HandleFunc("/api/deauth/list", handleDeauthList(m.Deauth))
	mux.HandleFunc("/api/deauth/stopall", handleDeauthStopAll(m.Deauth))

	// ── Evil Twin ────────────────────────────────────────────────
	mux.HandleFunc("/api/eviltwin/start", handleEvilTwinStart(m.EvilTwin))
	mux.HandleFunc("/api/eviltwin/stop", handleEvilTwinStop(m.EvilTwin))
	mux.HandleFunc("/api/eviltwin/status", handleEvilTwinStatus(m.EvilTwin))

	// ── Handshake capture ────────────────────────────────────────
	mux.HandleFunc("/api/handshake/start", handleHandshakeStart(m.Handshake))
	mux.HandleFunc("/api/handshake/stop", handleHandshakeStop(m.Handshake))
	mux.HandleFunc("/api/handshake/list", handleHandshakeList(m.Handshake))
	mux.HandleFunc("/api/handshake/download", handleHandshakeDownload(m.Handshake))

	// ── MITM ─────────────────────────────────────────────────────
	mux.HandleFunc("/api/mitm/start", handleMITMStart(m.MITM))
	mux.HandleFunc("/api/mitm/stop", handleMITMStop(m.MITM))
	mux.HandleFunc("/api/mitm/list", handleMITMList(m.MITM))
	mux.HandleFunc("/api/mitm/log", handleMITMLog(m.MITM))

	// ── Captive Portal ───────────────────────────────────────────
	mux.HandleFunc("/api/captive/start", handleCaptiveStart(m.Captive))
	mux.HandleFunc("/api/captive/stop", handleCaptiveStop(m.Captive))
	mux.HandleFunc("/api/captive/status", handleCaptiveStatus(m.Captive))
	mux.HandleFunc("/api/captive/creds", handleCaptiveCreds(m.Captive))

	// ── VPN Blocking ──────────────────────────────────────────
	mux.HandleFunc("/api/vpn/status", handleVPNStatus(m.VPNBlocker))
	mux.HandleFunc("/api/vpn/dns/enable", handleVPNDNSEnable(m.VPNBlocker))
	mux.HandleFunc("/api/vpn/dns/disable", handleVPNDNSDisable(m.VPNBlocker))
	mux.HandleFunc("/api/vpn/ports/enable", handleVPNPortsEnable(m.VPNBlocker))
	mux.HandleFunc("/api/vpn/ports/disable", handleVPNPortsDisable(m.VPNBlocker))
	mux.HandleFunc("/api/vpn/all/enable", handleVPNEnableAll(m.VPNBlocker))
	mux.HandleFunc("/api/vpn/all/disable", handleVPNDisableAll(m.VPNBlocker))

	// ── Static frontend ──────────────────────────────────────────
	fs := http.FileServer(http.Dir("/www/cerberus"))
	mux.Handle("/", fs)

	return withCORS(withLogging(mux))
}

// ═══════════════════════════════════════════════════════════════════
// CORE HANDLERS (DNS, devices, DoH)
// ═══════════════════════════════════════════════════════════════════

func handleStats(dnsLog *dns.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		since := time.Now().Add(-24 * time.Hour).Unix()
		if v := r.URL.Query().Get("since"); v != "" {
			if n, err := strconv.ParseInt(v, 10, 64); err == nil {
				since = n
			}
		}
		stats, err := dnsLog.Stats(since)
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, stats)
	}
}

func handleQueries(dnsLog *dns.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		since := time.Now().Add(-1 * time.Hour).Unix()
		until := time.Now().Unix()
		limit := 500

		if v := q.Get("since"); v != "" {
			if n, err := strconv.ParseInt(v, 10, 64); err == nil { since = n }
		}
		if v := q.Get("until"); v != "" {
			if n, err := strconv.ParseInt(v, 10, 64); err == nil { until = n }
		}
		if v := q.Get("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil { limit = n }
		}

		records, err := dnsLog.Query(since, until, q.Get("client"), q.Get("domain"), limit)
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		if records == nil { records = []dns.QueryRecord{} }
		writeJSON(w, 200, records)
	}
}

func handleTopDomains(dnsLog *dns.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		since := time.Now().Add(-24 * time.Hour).Unix()
		if v := r.URL.Query().Get("since"); v != "" {
			if n, err := strconv.ParseInt(v, 10, 64); err == nil { since = n }
		}
		limit := 20
		if v := r.URL.Query().Get("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil { limit = n }
		}
		results, err := dnsLog.TopDomains(since, r.URL.Query().Get("client"), limit)
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		if results == nil { results = []dns.DomainCount{} }
		writeJSON(w, 200, results)
	}
}

func handleDevices(tracker *devices.Tracker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, tracker.List()) }
}

func handleSetAlias(tracker *devices.Tracker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { writeError(w, 405, "POST required"); return }
		var req struct {
			MAC   string `json:"mac"`
			Alias string `json:"alias"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil { writeError(w, 400, "invalid JSON"); return }
		if req.MAC == "" { writeError(w, 400, "mac required"); return }
		tracker.SetAlias(req.MAC, req.Alias)
		writeJSON(w, 200, map[string]string{"status": "ok"})
	}
}

func handleDoHStatus(blocker *dns.DoHBlocker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]bool{"enabled": blocker.IsEnabled()})
	}
}

func handleDoHToggle(blocker *dns.DoHBlocker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { writeError(w, 405, "POST required"); return }
		enabled, err := blocker.Toggle()
		if err != nil { writeError(w, 500, err.Error()); return }
		writeJSON(w, 200, map[string]bool{"enabled": enabled})
	}
}

func handleConfig(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, 200, cfg)
		case http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			var newCfg config.Config
			if err := json.Unmarshal(body, &newCfg); err != nil { writeError(w, 400, "invalid JSON"); return }
			cfg.AlertDomains = newCfg.AlertDomains
			cfg.BlockedDomains = newCfg.BlockedDomains
			config.Save("/etc/cerberus/cerberus.json", cfg)
			writeJSON(w, 200, cfg)
		default:
			writeError(w, 405, "GET or PUT")
		}
	}
}

func handlePurge(dnsLog *dns.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { writeError(w, 405, "POST required"); return }
		days := 30
		if v := r.URL.Query().Get("days"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 { days = n }
		}
		deleted, err := dnsLog.Purge(time.Now().AddDate(0, 0, -days).Unix())
		if err != nil { writeError(w, 500, err.Error()); return }
		writeJSON(w, 200, map[string]int64{"deleted": deleted})
	}
}

// ═══════════════════════════════════════════════════════════════════
// RECON HANDLERS
// ═══════════════════════════════════════════════════════════════════

func handleReconMonitorEnable(s *recon.Scanner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { writeError(w, 405, "POST required"); return }
		if err := s.EnableMonitor(); err != nil { writeError(w, 500, err.Error()); return }
		writeJSON(w, 200, map[string]string{"status": "monitor enabled"})
	}
}

func handleReconMonitorDisable(s *recon.Scanner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { writeError(w, 405, "POST required"); return }
		if err := s.DisableMonitor(); err != nil { writeError(w, 500, err.Error()); return }
		writeJSON(w, 200, map[string]string{"status": "monitor disabled"})
	}
}

func handleReconScanStart(s *recon.Scanner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { writeError(w, 405, "POST required"); return }
		if err := s.StartScan(); err != nil { writeError(w, 500, err.Error()); return }
		writeJSON(w, 200, map[string]string{"status": "scanning"})
	}
}

func handleReconScanStop(s *recon.Scanner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { writeError(w, 405, "POST required"); return }
		s.StopScan()
		writeJSON(w, 200, map[string]string{"status": "stopped"})
	}
}

func handleReconAPs(s *recon.Scanner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, s.GetAPs()) }
}

func handleReconClients(s *recon.Scanner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bssid := r.URL.Query().Get("bssid")
		if bssid != "" {
			writeJSON(w, 200, s.GetClientsForAP(bssid))
		} else {
			writeJSON(w, 200, s.GetClients())
		}
	}
}

func handleARPScan() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hosts, err := recon.ARPScan(r.URL.Query().Get("cidr"))
		if err != nil { writeError(w, 500, err.Error()); return }
		writeJSON(w, 200, hosts)
	}
}

// ═══════════════════════════════════════════════════════════════════
// DEAUTH HANDLERS
// ═══════════════════════════════════════════════════════════════════

func handleDeauthStart(mgr *deauth.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { writeError(w, 405, "POST required"); return }
		var req struct {
			ID        string `json:"id"`
			TargetMAC string `json:"target_mac"`
			BSSID     string `json:"bssid"`
			Iface     string `json:"iface"`
			Count     int    `json:"count"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil { writeError(w, 400, "invalid JSON"); return }
		if req.ID == "" { req.ID = fmt.Sprintf("deauth-%d", time.Now().UnixMilli()) }
		if req.Iface == "" { req.Iface = "wlan1mon" }
		if err := mgr.Start(req.ID, req.TargetMAC, req.BSSID, req.Iface, req.Count); err != nil {
			writeError(w, 500, err.Error()); return
		}
		writeJSON(w, 200, map[string]string{"id": req.ID, "status": "started"})
	}
}

func handleDeauthStop(mgr *deauth.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { writeError(w, 405, "POST required"); return }
		id := r.URL.Query().Get("id")
		if id == "" { writeError(w, 400, "id required"); return }
		if err := mgr.Stop(id); err != nil { writeError(w, 500, err.Error()); return }
		writeJSON(w, 200, map[string]string{"status": "stopped"})
	}
}

func handleDeauthList(mgr *deauth.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, mgr.List()) }
}

func handleDeauthStopAll(mgr *deauth.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { writeError(w, 405, "POST required"); return }
		mgr.StopAll()
		writeJSON(w, 200, map[string]string{"status": "all stopped"})
	}
}

// ═══════════════════════════════════════════════════════════════════
// EVIL TWIN HANDLERS
// ═══════════════════════════════════════════════════════════════════

func handleEvilTwinStart(mgr *eviltwin.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { writeError(w, 405, "POST required"); return }
		var cfg eviltwin.Config
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil { writeError(w, 400, "invalid JSON"); return }
		if cfg.Iface == "" { cfg.Iface = "wlan1" }
		if cfg.OutIface == "" { cfg.OutIface = "wan" }
		if cfg.Channel == 0 { cfg.Channel = 6 }
		if err := mgr.Start(cfg); err != nil { writeError(w, 500, err.Error()); return }
		writeJSON(w, 200, map[string]string{"status": "started"})
	}
}

func handleEvilTwinStop(mgr *eviltwin.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { writeError(w, 405, "POST required"); return }
		if err := mgr.Stop(); err != nil { writeError(w, 500, err.Error()); return }
		writeJSON(w, 200, map[string]string{"status": "stopped"})
	}
}

func handleEvilTwinStatus(mgr *eviltwin.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, mgr.Status()) }
}

// ═══════════════════════════════════════════════════════════════════
// HANDSHAKE HANDLERS
// ═══════════════════════════════════════════════════════════════════

func handleHandshakeStart(mgr *handshake.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { writeError(w, 405, "POST required"); return }
		var req struct {
			ID      string `json:"id"`
			BSSID   string `json:"bssid"`
			SSID    string `json:"ssid"`
			Channel int    `json:"channel"`
			Iface   string `json:"iface"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil { writeError(w, 400, "invalid JSON"); return }
		if req.ID == "" { req.ID = fmt.Sprintf("cap-%d", time.Now().UnixMilli()) }
		if req.Iface == "" { req.Iface = "wlan1mon" }
		if err := mgr.Start(req.ID, req.BSSID, req.SSID, req.Iface, req.Channel); err != nil {
			writeError(w, 500, err.Error()); return
		}
		writeJSON(w, 200, map[string]string{"id": req.ID, "status": "capturing"})
	}
}

func handleHandshakeStop(mgr *handshake.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { writeError(w, 405, "POST required"); return }
		id := r.URL.Query().Get("id")
		if id == "" { writeError(w, 400, "id required"); return }
		if err := mgr.Stop(id); err != nil { writeError(w, 500, err.Error()); return }
		writeJSON(w, 200, map[string]string{"status": "stopped"})
	}
}

func handleHandshakeList(mgr *handshake.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, mgr.List()) }
}

func handleHandshakeDownload(mgr *handshake.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		path, err := mgr.GetCapFile(id)
		if err != nil { writeError(w, 404, err.Error()); return }
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.cap", id))
		http.ServeFile(w, r, path)
	}
}

// ═══════════════════════════════════════════════════════════════════
// MITM HANDLERS
// ═══════════════════════════════════════════════════════════════════

func handleMITMStart(mgr *mitm.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { writeError(w, 405, "POST required"); return }
		var req struct {
			ID        string `json:"id"`
			TargetIP  string `json:"target_ip"`
			GatewayIP string `json:"gateway_ip"`
			Iface     string `json:"iface"`
			SSLStrip  bool   `json:"sslstrip"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil { writeError(w, 400, "invalid JSON"); return }
		if req.ID == "" { req.ID = fmt.Sprintf("mitm-%d", time.Now().UnixMilli()) }
		if req.Iface == "" { req.Iface = "br-lan" }
		if req.GatewayIP == "" { req.GatewayIP = "192.168.1.1" }
		if err := mgr.Start(req.ID, req.TargetIP, req.GatewayIP, req.Iface, req.SSLStrip); err != nil {
			writeError(w, 500, err.Error()); return
		}
		writeJSON(w, 200, map[string]string{"id": req.ID, "status": "started"})
	}
}

func handleMITMStop(mgr *mitm.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { writeError(w, 405, "POST required"); return }
		id := r.URL.Query().Get("id")
		if id == "" { writeError(w, 400, "id required"); return }
		if err := mgr.Stop(id); err != nil { writeError(w, 500, err.Error()); return }
		writeJSON(w, 200, map[string]string{"status": "stopped"})
	}
}

func handleMITMList(mgr *mitm.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, mgr.List()) }
}

func handleMITMLog(mgr *mitm.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		data, err := mgr.GetLog(id)
		if err != nil { writeError(w, 404, err.Error()); return }
		w.Header().Set("Content-Type", "text/plain")
		w.Write(data)
	}
}

// ═══════════════════════════════════════════════════════════════════
// CAPTIVE PORTAL HANDLERS
// ═══════════════════════════════════════════════════════════════════

func handleCaptiveStart(mgr *captive.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { writeError(w, 405, "POST required"); return }
		var cfg captive.PortalConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil { writeError(w, 400, "invalid JSON"); return }
		if err := mgr.Start(cfg); err != nil { writeError(w, 500, err.Error()); return }
		writeJSON(w, 200, map[string]string{"status": "started"})
	}
}

func handleCaptiveStop(mgr *captive.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { writeError(w, 405, "POST required"); return }
		if err := mgr.Stop(); err != nil { writeError(w, 500, err.Error()); return }
		writeJSON(w, 200, map[string]string{"status": "stopped"})
	}
}

func handleCaptiveStatus(mgr *captive.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]interface{}{
			"running":  mgr.IsRunning(),
			"config":   mgr.GetConfig(),
		})
	}
}

func handleCaptiveCreds(mgr *captive.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			mgr.ClearCreds()
			writeJSON(w, 200, map[string]string{"status": "cleared"})
			return
		}
		writeJSON(w, 200, mgr.GetCreds())
	}
}

// ═══════════════════════════════════════════════════════════════════
// VPN BLOCKER HANDLERS
// ═══════════════════════════════════════════════════════════════════

func handleVPNStatus(v *dns.VPNBlocker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, v.Status()) }
}

func handleVPNDNSEnable(v *dns.VPNBlocker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { writeError(w, 405, "POST required"); return }
		if err := v.EnableDNS(); err != nil { writeError(w, 500, err.Error()); return }
		writeJSON(w, 200, map[string]string{"status": "dns blocking enabled"})
	}
}

func handleVPNDNSDisable(v *dns.VPNBlocker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { writeError(w, 405, "POST required"); return }
		if err := v.DisableDNS(); err != nil { writeError(w, 500, err.Error()); return }
		writeJSON(w, 200, map[string]string{"status": "dns blocking disabled"})
	}
}

func handleVPNPortsEnable(v *dns.VPNBlocker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { writeError(w, 405, "POST required"); return }
		if err := v.EnablePorts(); err != nil { writeError(w, 500, err.Error()); return }
		writeJSON(w, 200, map[string]string{"status": "port blocking enabled"})
	}
}

func handleVPNPortsDisable(v *dns.VPNBlocker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { writeError(w, 405, "POST required"); return }
		if err := v.DisablePorts(); err != nil { writeError(w, 500, err.Error()); return }
		writeJSON(w, 200, map[string]string{"status": "port blocking disabled"})
	}
}

func handleVPNEnableAll(v *dns.VPNBlocker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { writeError(w, 405, "POST required"); return }
		if err := v.EnableAll(); err != nil { writeError(w, 500, err.Error()); return }
		writeJSON(w, 200, map[string]string{"status": "all vpn blocking enabled"})
	}
}

func handleVPNDisableAll(v *dns.VPNBlocker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { writeError(w, 405, "POST required"); return }
		v.DisableAll()
		writeJSON(w, 200, map[string]string{"status": "all vpn blocking disabled"})
	}
}

// ═══════════════════════════════════════════════════════════════════
// MIDDLEWARE
// ═══════════════════════════════════════════════════════════════════

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" { w.WriteHeader(204); return }
		next.ServeHTTP(w, r)
	})
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
