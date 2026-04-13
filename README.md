# CERBERUS

**Three-headed network guardian.** Custom OpenWrt router firmware with a full offensive security dashboard. WiFi Pineapple competitor built on real router hardware.

*Welcome to the Meme-Detector.*

---

## Features

### Core Network Monitoring
- **Real-time DNS query logging** — every domain, every device, timestamped, stored in SQLite
- **Device fingerprinting** — auto-discovers via DHCP leases + ARP table with hostname, IP, MAC, vendor, signal, bandwidth
- **Device aliasing** — rename devices from the dashboard
- **DNS-over-HTTPS blocking** — one-toggle kills all major DoH resolvers (Cloudflare, Google, Quad9, OpenDNS, AdGuard, NextDNS, Control D, Mullvad) on ports 443 + 853
- **DNS log filtering** — All/Passed/Blocked filters, domain search, 5000-line buffer

### Wireless Recon
- **Passive AP scanning** — SSID, BSSID, channel, encryption, signal, client count
- **Client enumeration** — all clients on any AP with device type, signal, traffic
- **Probe request harvesting** — see what networks devices previously connected to
- **Monitor mode management** — enable/disable from the dashboard

### WPA Handshake Capture
- **Targeted capture** — select WPA/WPA2/WPA3 AP and capture 4-way handshake
- **Auto-deauth** — forces client reconnection to trigger handshake
- **Handshake verification** — aircrack-ng validates the .cap in background
- **Download .cap files** — for offline cracking with hashcat or aircrack-ng

### Deauth Attacks
- **Targeted or broadcast** — specific client MAC or all clients
- **Continuous or burst** — set packet count or run until stopped
- **Multi-target** — deauth multiple devices simultaneously

### Evil Twin AP
- **AP cloning** — clone any SSID with hostapd
- **Built-in DHCP** — dnsmasq hands out IPs automatically
- **NAT forwarding** — victims get real internet through you
- **Captive portal redirect** — auto-redirects HTTP + DNS

### MITM (Man-in-the-Middle)
- **ARP spoofing** — bidirectional via arpspoof
- **SSL stripping** — optional sslstrip for HTTP downgrade
- **Traffic logging** — tcpdump captures HTTP traffic per target
- **Multi-target** — MITM multiple devices simultaneously

### Captive Portal / Credential Harvesting
- **Phishing templates** — Google Sign-In, Facebook Login, Hotel WiFi
- **Custom HTML** — upload your own portal page
- **Auto OS detection** — triggers native captive popup on Android/iOS/Windows
- **Live credential feed** — real-time capture display on dashboard

### Dashboard UI
- **Cinematic login screen** — animated particles, scan line, glass morphism, shield logo
- **Slogan: "Welcome to the Meme-Detector"**
- **Credentials:** BackupFaun / Creator
- **8-page navigation** — Overview, Recon, Targets, MITM, Evil Twin, Captive, Logging, Settings
- **Target AP selector** — custom dropdown in top bar
- **Live device table** — per-device MITM/Deauth toggles
- **Adapter role management** — assign scan/attack/upstream roles
- **JetBrains Mono** hacker aesthetic, dark theme, cyan/red accents

### Infrastructure
- **Go daemon** (~5MB) — cross-compiles to ARM64/MIPS/x86
- **SQLite** persistent storage
- **React** single-page frontend
- **GitHub Actions CI** — push tag → firmware images auto-built
- **Pre-installed tools** — aircrack-ng, hostapd, dnsmasq, tcpdump, arpspoof, sslstrip

## Hardware

| Router | Price |
|---|---|
| GL.iNet GL-MT3000 (recommended) | ~$70 |
| GL.iNet GL-MT1300 | ~$40 |
| Raspberry Pi 4 | ~$50 |

Plus 2x Alfa AWUS036ACH adapters (~$35 each)

## Quick Start

```bash
git clone https://github.com/YOUR_USER/cerberus.git
cd cerberus && git tag v0.1.0 && git push origin main --tags
```

Download firmware from Actions tab → flash → open `http://192.168.1.1:8443`

## Docs

- [BUILD.md](docs/BUILD.md) — Build from source, flash, configure
- [UNBRICK.md](docs/UNBRICK.md) — Recover a bricked router

## License

MIT
