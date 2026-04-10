# 🐕 CERBERUS

**Network Control System** — WiFi Pineapple-grade offensive network toolkit running on a $70 router.

![License](https://img.shields.io/badge/license-MIT-00ffc8)
![Go](https://img.shields.io/badge/Go-1.22-00ADD8)
![Platform](https://img.shields.io/badge/platform-GL--MT3000-ff4757)

---

## Features

| Feature | Description |
|---------|-------------|
| **Network Recon** | Discover APs, clients, signal strength, probe requests |
| **Per-Device MITM** | ARP spoof specific targets, intercept DNS in real-time |
| **Deauth** | Kick individual devices off any network |
| **Handshake Capture** | Grab WPA 4-way handshakes, download .cap files |
| **Evil Twin** | Clone any AP with one click |
| **Captive Portal** | Credential harvesting (Google, Facebook, WiFi templates) |
| **DNS Logging** | 5000-line buffer with pass/block filters and domain search |
| **DoH/DoT Blocking** | Force plaintext DNS by blocking encrypted resolvers |
| **Modern Dashboard** | Dark theme, real-time updates, mobile-friendly |
| **Hashed Auth** | SHA-256 passwords, persistent sessions |

---

## Hardware

| Part | Price | Notes |
|------|-------|-------|
| GL.iNet GL-MT3000 (Beryl AX) | ~$70 | Runs OpenWrt natively, USB 3.0 |
| Alfa AWUS036ACH | ~$35 | Monitor mode + injection, dual band |
| USB Hub (optional) | ~$10 | If running 2 adapters |

**Minimum:** Router only ($70) — gets you MITM, DNS logging, DoH blocking.
**Recommended:** Router + 1 Alfa ($105) — adds deauth, handshake capture, evil twin.

---

## Install (One Command)

### Option 1: Download Release

Go to [Releases](../../releases), download the latest `cerberus-release.tar.gz`.

```bash
tar -xzf cerberus-release.tar.gz
./install.sh 192.168.1.1
```

That's it. Dashboard at `http://192.168.1.1:1471`.

### Option 2: Build from Source

```bash
git clone https://github.com/YOUR_USERNAME/cerberus.git
cd cerberus
make install ROUTER=192.168.1.1
```

### Option 3: Fork & Auto-Build

1. Fork this repo
2. Push a tag: `git tag v0.1.0 && git push --tags`
3. GitHub Actions builds everything automatically
4. Download from Releases tab
5. `./install.sh 192.168.1.1`

---

## Prerequisites

Your GL-MT3000 must be running **OpenWrt** (not stock GL.iNet firmware).

### Flash OpenWrt

1. Download OpenWrt for GL-MT3000 from [openwrt.org](https://openwrt.org/toh/gl.inet/gl-mt3000)
2. Connect to router via ethernet, go to `http://192.168.8.1`
3. **System → Upgrade → Local Upgrade**
4. Uncheck "Keep Settings", upload the .bin
5. Wait 2-3 minutes, router reboots into OpenWrt at `http://192.168.1.1`

If you brick it — hold Reset while powering on for 10 seconds. Go to `http://192.168.1.1` to access U-Boot recovery.

---

## Usage

```
1. Open http://192.168.1.1:1471
2. Create account (first visit) or login
3. Hit RECON — discovers nearby APs
4. Select target AP from dropdown
5. Hit SCAN — finds clients on that AP
6. Toggle MITM / Deauth per device
7. View DNS logs with filters
8. Capture WPA handshakes
9. Launch Evil Twin + Captive Portal
```

---

## Project Structure

```
cerberus/
├── .github/workflows/build.yml    # CI/CD
├── backend/
│   ├── main.go                    # Entry point
│   ├── go.mod
│   ├── scanner/scanner.go         # ARP scan, airodump, probes
│   ├── mitm/mitm.go              # Per-target ARP spoof + DNS
│   ├── deauth/deauth.go          # Per-target deauth
│   ├── eviltwin/eviltwin.go      # Rogue AP via hostapd
│   ├── captive/captive.go        # Phishing portal
│   ├── handshake/handshake.go    # WPA capture + download
│   ├── adapters/adapters.go      # Interface detection + roles
│   └── api/routes.go             # REST API
├── frontend/
│   ├── index.html
│   ├── cerberus-dashboard.jsx    # Full React dashboard
│   └── cerberus-api.js           # API client
├── Makefile
├── docs/
│   └── SETUP.md                  # Detailed setup guide
└── README.md
```

---

## API

All endpoints return JSON. CORS enabled.

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/scan` | Start network scan |
| GET | `/api/clients` | List discovered clients |
| GET | `/api/networks` | List discovered APs |
| GET | `/api/probes` | List probe requests |
| POST | `/api/mitm/start` | Start MITM on target |
| POST | `/api/mitm/stop` | Stop MITM on target |
| GET | `/api/mitm/dns` | Get DNS query log |
| POST | `/api/deauth/start` | Start deauth on target |
| POST | `/api/deauth/stop` | Stop deauth on target |
| POST | `/api/eviltwin/start` | Launch evil twin |
| POST | `/api/eviltwin/stop` | Stop evil twin |
| POST | `/api/captive/start` | Start captive portal |
| GET | `/api/captive/creds` | Get captured credentials |
| POST | `/api/handshake/start` | Start handshake capture |
| GET | `/api/handshake/status` | Capture status |
| GET | `/api/handshake/download/:file` | Download .cap file |
| GET | `/api/adapters` | List wireless adapters |
| POST | `/api/adapters/role` | Assign adapter role |
| GET | `/api/status` | System status |

---

## Disclaimer

For authorized testing and parental monitoring only. You are responsible for complying with local laws.

---

## License

MIT
