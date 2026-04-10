package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"cerberus/api"
	"cerberus/auth"
	"cerberus/scanner"
	"cerberus/mitm"
	"cerberus/deauth"
	"cerberus/eviltwin"
	"cerberus/captive"
	"cerberus/handshake"
	"cerberus/adapters"
)

const VERSION = "0.1.0"
const DATA_DIR = "/etc/cerberus"

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

	if os.Getuid() != 0 {
		log.Fatal("[!] Cerberus must run as root. Try: sudo ./cerberus")
	}

	os.MkdirAll(DATA_DIR, 0755)
	os.MkdirAll(DATA_DIR+"/captures", 0755)

	// Initialize all modules
	authMgr := auth.New(DATA_DIR)
	adapterMgr := adapters.New(DATA_DIR)
	scan := scanner.New()
	mitmEngine := mitm.New()
	deauthEngine := deauth.New()
	etEngine := eviltwin.New()
	capEngine := captive.New()
	hsEngine := handshake.New(DATA_DIR + "/captures")

	// Set adapter roles on engines
	roles := adapterMgr.GetRoles()
	if roles["scan"] != "" {
		hsEngine.SetInterfaces(roles["scan"], roles["attack"])
	}

	// Enable IP forwarding
	os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1"), 0644)

	// Build API
	router := api.NewRouter(authMgr, scan, mitmEngine, deauthEngine, etEngine, capEngine, hsEngine, adapterMgr)

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
		os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("0"), 0644)
		os.Exit(0)
	}()

	addr := ":1471"
	log.Printf("[*] Cerberus API on %s", addr)

	if !authMgr.IsSetup() {
		log.Println("[*] First run — create account at http://localhost:1471")
	}

	log.Fatal(http.ListenAndServe(addr, router))
}
