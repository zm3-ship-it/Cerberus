# 🐕 CERBERUS

**Network Control System** — Open-source offensive network toolkit running on a $70 travel router.

![License](https://img.shields.io/badge/license-MIT-00ffc8)
![Go](https://img.shields.io/badge/Go-1.22-00ADD8)
![Platform](https://img.shields.io/badge/platform-GL--MT3000-ff4757)

---

## ⚠️ Disclaimer

**This software is provided for educational and authorized security testing purposes only.**

The author(s) of this project are **not responsible** for any misuse, damage, or illegal activity caused by this software. By downloading, installing, or using Cerberus, **you accept full responsibility** for your actions.

It is your obligation to comply with all applicable local, state, federal, and international laws. Unauthorized interception of network traffic, unauthorized access to computer systems, and deploying rogue access points on networks you do not own or have explicit written permission to test is **illegal** in most jurisdictions.

**Use at your own risk. You have been warned.**

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
| **Router Management** | WAN/LAN/WiFi config, DHCP, firmware flash, reboot |
| **Modern Dashboard** | Dark theme, real-time updates, mobile-friendly |
| **Hashed Auth** | SHA-256 passwords, persistent sessions |

---

## Hardware

| Part | Price | Notes |
|------|-------|-------|
| GL.iNet GL-MT3000 (Beryl AX) | ~$70 | Runs OpenWrt natively, USB 3.0 |
| Alfa AWUS036ACH (or Panda PAU09) | ~$20-35 | Monitor mode + injection, dual band |
| USB Hub (optional) | ~$10 | If running 2 adapters |

---

## Quick Install

### 1. Flash OpenWrt on GL-MT3000

```bash
ssh root@192.168.8.1
cd /tmp
wget https://downloads.openwrt.org/releases/25.12.0/targets/mediatek/filogic/openwrt-25.12.0-mediatek-filogic-glinet_gl-mt3000-squashfs-sysupgrade.bin
sysupgrade -n /tmp/openwrt-25.12.0-mediatek-filogic-glinet_gl-mt3000-squashfs-sysupgrade.bin
```

### 2. Deploy Cerberus

Download the latest release from the [Releases](../../releases) tab, then:

```bash
# Upload binary
scp cerberus root@192.168.1.1:/usr/bin/cerberus

# Upload frontend
scp -r www/* root@192.168.1.1:/www/cerberus/

# SSH in and finish setup
ssh root@192.168.1.1

chmod +x /usr/bin/cerberus
mkdir -p /www/cerberus /var/cerberus/captures
opkg update && opkg install aircrack-ng tcpdump dsniff nmap hostapd-common dnsmasq

cat > /etc/init.d/cerberus << 'EOF'
#!/bin/sh /etc/rc.common
START=99
STOP=10
start() { /usr/bin/cerberus > /var/log/cerberus.log 2>&1 & }
stop() { killall cerberus 2>/dev/null; }
restart() { stop; sleep 1; start; }
EOF

chmod +x /etc/init.d/cerberus
/etc/init.d/cerberus enable
/etc/init.d/cerberus start
```

### 3. Open Dashboard

`http://192.168.1.1:1471`

First visit — create your account. Plug in your Alfa/Panda adapter for offensive features.

---

## Build from Source

```bash
git clone https://github.com/YOUR_USERNAME/cerberus.git
cd cerberus
make build
```

Or push a tag to trigger GitHub Actions auto-build:

```bash
git tag v0.1.0 && git push --tags
```

---

## Project Structure

```
cerberus/
├── backend/
│   ├── main.go                 # Entry point
│   ├── scanner/scanner.go      # Network discovery
│   ├── mitm/mitm.go           # ARP spoof + DNS capture
│   ├── deauth/deauth.go       # WiFi deauthentication
│   ├── eviltwin/eviltwin.go   # Rogue access point
│   ├── captive/captive.go     # Phishing portal
│   ├── handshake/handshake.go # WPA handshake capture
│   ├── adapters/adapters.go   # Wireless adapter management
│   ├── system/system.go       # System info, reboot, firmware
│   ├── network/network.go     # WAN/LAN/WiFi/DHCP via UCI
│   └── api/routes.go          # 40+ REST endpoints
├── frontend/
│   ├── index.html
│   ├── cerberus-dashboard.jsx
│   └── cerberus-api.js
├── .github/workflows/build.yml
├── Makefile
└── README.md
```

---

## API

All endpoints return JSON. CORS enabled. Port 1471.

<details>
<summary>Full API Reference (click to expand)</summary>

### Scanner
| Method | Path | Description |
|--------|------|-------------|
| POST | /api/scan | Start network scan |
| GET | /api/clients | List discovered clients |
| GET | /api/networks | List discovered APs |
| GET | /api/probes | List probe requests |

### MITM
| Method | Path | Description |
|--------|------|-------------|
| POST | /api/mitm/start | Start MITM `{mac, ip}` |
| POST | /api/mitm/stop | Stop MITM `{mac}` |
| GET | /api/mitm/targets | Active targets |
| GET | /api/mitm/dns | DNS query log |

### Deauth
| Method | Path | Description |
|--------|------|-------------|
| POST | /api/deauth/start | Start `{mac, bssid}` |
| POST | /api/deauth/stop | Stop `{mac}` |
| GET | /api/deauth/targets | Active targets |

### Evil Twin
| Method | Path | Description |
|--------|------|-------------|
| POST | /api/eviltwin/start | Launch `{ssid, channel, iface}` |
| POST | /api/eviltwin/stop | Stop |
| GET | /api/eviltwin/status | Status |

### Captive Portal
| Method | Path | Description |
|--------|------|-------------|
| POST | /api/captive/start | Launch `{template}` |
| POST | /api/captive/stop | Stop |
| GET | /api/captive/creds | Captured credentials |

### Handshake
| Method | Path | Description |
|--------|------|-------------|
| POST | /api/handshake/start | Capture `{bssid, ssid, channel}` |
| POST | /api/handshake/stop | Stop capture |
| GET | /api/handshake/status | Capture status |
| GET | /api/handshake/captures | List .cap files |
| GET | /api/handshake/download/:file | Download .cap |

### System
| Method | Path | Description |
|--------|------|-------------|
| GET | /api/system/info | System info |
| POST | /api/system/reboot | Reboot router |
| POST | /api/system/hostname | Set hostname |
| POST | /api/system/firmware | Flash firmware (multipart) |

### Network
| Method | Path | Description |
|--------|------|-------------|
| GET/POST | /api/network/wan | WAN config |
| GET/POST | /api/network/lan | LAN config |
| GET/POST | /api/network/wifi | WiFi config |
| GET | /api/network/interfaces | Interface status |
| GET | /api/network/dhcp/leases | DHCP leases |
| GET/POST/DELETE | /api/network/dhcp/static | Static leases |

### Adapters
| Method | Path | Description |
|--------|------|-------------|
| GET | /api/adapters | List adapters |
| POST | /api/adapters/role | Assign role `{adapter, role}` |
| POST | /api/adapters/detect | Re-detect adapters |

</details>

---

## License

LGPL-2.0 — see [LICENSE](LICENSE)
