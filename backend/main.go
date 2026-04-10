package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"cerberus/adapters"
	"cerberus/api"
	"cerberus/captive"
	"cerberus/deauth"
	"cerberus/eviltwin"
	"cerberus/handshake"
	"cerberus/mitm"
	"cerberus/network"
	"cerberus/scanner"
)

const VERSION = "0.2.0"
const DATA_DIR = "/etc/cerberus"
const CAP_DIR = DATA_DIR + "/captures"

func main() {
	fmt.Printf(`
   ██████╗███████╗██████╗ ██████╗ ███████╗██████╗ ██╗   ██╗███████╗
  ██╔════╝██╔════╝██╔══██╗██╔══██╗██╔════╝██╔══██╗██║   ██║██╔════╝
  ██║     █████╗  ██████╔╝██████╔╝█████╗  ██████╔╝██║   ██║███████╗
  ██║     ██╔══╝  ██╔══██╗██╔══██╗██╔══╝  ██╔══██╗██║   ██║╚════██║
  ╚██████╗███████╗██║  ██║██████╔╝███████╗██║  ██║╚██████╔╝███████║
   ╚═════╝╚══════╝╚═╝  ╚═╝╚═════╝ ╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚══════╝
  Network Control System v%s
`, VERSION)

	// Must run as root for raw sockets, arp, hostapd, etc.
	if os.Getuid() != 0 {
		log.Fatal("[!] Cerberus must run as root. Try: sudo ./cerberus")
	}

	// Create data directories
	os.MkdirAll(DATA_DIR, 0755)
	os.MkdirAll(CAP_DIR, 0755)

	// Initialize modules
	scan := scanner.New()
	mitmEngine := mitm.New()
	deauthEngine := deauth.New()
	etEngine := eviltwin.New()
	capEngine := captive.New()
	hsEngine := handshake.New(CAP_DIR)
	adapterMgr := adapters.New(DATA_DIR)
	netMgr := network.New()

	// Wire adapter roles to handshake engine
	roles := adapterMgr.GetRoles()
	if roles["scan"] != "" {
		hsEngine.SetInterfaces(roles["scan"], roles["attack"])
	}

	// Enable IP forwarding
	if err := os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1"), 0644); err != nil {
		log.Printf("[!] Warning: Could not enable IP forwarding: %v", err)
	}

	// Build API router
	router := api.NewRouter(scan, mitmEngine, deauthEngine, etEngine, capEngine, hsEngine, adapterMgr, netMgr)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n[*] Shutting down Cerberus...")
		mitmEngine.StopAll()
		deauthEngine.StopAll()
		etEngine.Stop()
		capEngine.Stop()
		hsEngine.Stop()
		os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("0"), 0644)
		os.Exit(0)
	}()

	addr := ":1471"
	log.Printf("[*] Cerberus API listening on %s", addr)
	log.Printf("[*] Dashboard: http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}
