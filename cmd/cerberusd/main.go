package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cerberus/internal/api"
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

var version = "0.1.0"

func main() {
	cfgPath := flag.String("config", "/etc/cerberus/cerberus.json", "path to config file")
	listenAddr := flag.String("listen", ":8443", "HTTP API listen address")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("cerberusd v%s\n", version)
		os.Exit(0)
	}

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// ── Core modules ─────────────────────────────────────────────
	dnsLog, err := dns.NewLogger(cfg.DBPath)
	if err != nil {
		log.Fatalf("dns logger: %v", err)
	}
	defer dnsLog.Close()

	devTracker := devices.NewTracker(cfg.LeasesFile)

	sniffer, err := dns.NewSniffer(cfg.Interface, dnsLog, devTracker)
	if err != nil {
		log.Fatalf("dns sniffer: %v", err)
	}
	go sniffer.Run()

	dohBlocker := dns.NewDoHBlocker(cfg.DoHBlockEnabled)
	if cfg.DoHBlockEnabled {
		if err := dohBlocker.Enable(); err != nil {
			log.Printf("warning: doh blocker: %v", err)
		}
	}

	go devTracker.ScanLoop(30 * time.Second)

	// ── Attack modules ───────────────────────────────────────────
	scanner := recon.NewScanner(cfg.MonitorIface)
	deauthMgr := deauth.NewManager()
	twinMgr := eviltwin.NewManager()
	shakeMgr := handshake.NewManager("/tmp/cerberus-captures")
	mitmMgr := mitm.NewManager()
	captiveMgr := captive.NewManager()

	// ── HTTP API ─────────────────────────────────────────────────
	modules := &api.Modules{
		DNSLog:     dnsLog,
		Devices:    devTracker,
		DoHBlocker: dohBlocker,
		Config:     cfg,
		Scanner:    scanner,
		Deauth:     deauthMgr,
		EvilTwin:   twinMgr,
		Handshake:  shakeMgr,
		MITM:       mitmMgr,
		Captive:    captiveMgr,
	}

	router := api.NewRouter(modules)

	srv := &http.Server{
		Addr:         *listenAddr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ── Graceful shutdown ────────────────────────────────────────
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("cerberusd v%s listening on %s", version, *listenAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Stop all attack modules cleanly
	sniffer.Stop()
	dohBlocker.Disable()
	deauthMgr.StopAll()
	twinMgr.Stop()
	shakeMgr.StopAll()
	mitmMgr.StopAll()
	captiveMgr.Stop()

	srv.Shutdown(shutCtx)
	log.Println("cerberusd stopped")
}
