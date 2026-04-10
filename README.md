# 🐕 CERBERUS

**Custom router firmware with built-in offensive network toolkit and full router management.**

One flash. Everything works. No setup needed.

![Go](https://img.shields.io/badge/Go-1.22-00ADD8)
![Platform](https://img.shields.io/badge/platform-GL--MT3000-ff4757)
![OpenWrt](https://img.shields.io/badge/OpenWrt-25.12.0-blue)

---

## ⚠️ Disclaimer

**This software is provided for educational and authorized security testing purposes only.**

The author(s) are **not responsible** for any misuse, damage, or illegal activity caused by this software. By using Cerberus, **you accept full responsibility** for your actions. Unauthorized network interception is **illegal** in most jurisdictions.

**Use at your own risk.**

---

## What Is This?

Cerberus is a custom firmware for the GL.iNet GL-MT3000 router (~$70) that combines:

- **Offensive WiFi toolkit** — network recon, MITM, deauth, handshake capture, evil twin, captive portal, DNS spoofing
- **Full router management** — WAN/LAN/WiFi config, DHCP, firmware updates, system monitoring
- **Modern dashboard** — dark themed web UI at `http://192.168.1.1:1471`

Two modes in the sidebar:
- **🐕 CERBERUS** — Offensive tools (cyan)
- **🌐 ROUTER** — Management (blue)

Default credentials: `root` / `toor`

---

## Features

### Offensive (Cerberus Mode)
| Feature | Description |
|---------|-------------|
| Network Recon | Discover nearby APs, clients, signal strength, probe requests |
| Per-Device MITM | ARP spoof specific targets, intercept DNS in real-time |
| Deauth | Kick individual devices off any network |
| Handshake Capture | Grab WPA 4-way handshakes with targeted client deauth, download .cap files |
| Evil Twin | Clone any AP with one click |
| Captive Portal | Credential harvesting (Google, Facebook, WiFi, Hotel templates) |
| DNS Spoofing | Redirect any domain to a custom IP |
| DNS Logging | 5000-line buffer with pass/block/domain filters |
| DoH/DoT Blocking | Force plaintext DNS by blocking encrypted resolvers |

### Router Management (Router Mode)
| Feature | Description |
|---------|-------------|
| WAN Config | DHCP, Static IP, PPPoE with DNS settings |
| LAN Config | IP, subnet, DHCP server toggle with range/lease |
| WiFi Config | Per-radio SSID, password, encryption, channel, enable/disable |
| Interface Status | All interfaces with IP, MAC, speed, up/down |
| DHCP Leases | Active leases with hostname, IP, MAC |
| System Info | Hostname, uptime, firmware, kernel, CPU, RAM, disk with live gauges |
| Firmware Flash | Upload .bin to flash new firmware from the dashboard |
| Reboot | One-click with confirmation |
| Adapter Roles | Assign WiFi adapters to scan/attack/upstream roles |

---

## Hardware

| Part | Price | Required? |
|------|-------|-----------|
| GL.iNet GL-MT3000 (Beryl AX) | ~$70 | Yes |
| Alfa AWUS036ACH or Panda PAU09 | $20-35 | Optional — needed for deauth, evil twin, handshake capture |
| USB Hub | ~$10 | Only if running 2 adapters |

**Without an adapter:** Recon, MITM, DNS logging, DNS spoofing, all router management features work using the built-in radio.

**With an adapter:** Adds deauth, handshake capture, evil twin, captive portal.

---

## Install

### Option 1: Flash the Firmware (Recommended)

Download `cerberus-firmware-gl-mt3000.bin` from the [Releases](../../releases) page.

**If already on OpenWrt:**
```bash
scp cerberus-firmware-gl-mt3000.bin root@192.168.1.1:/tmp/
ssh root@192.168.1.1 "sysupgrade -n /tmp/cerberus-firmware-gl-mt3000.bin"
```

**If on stock GL.iNet firmware:**
```bash
ssh root@192.168.8.1
cd /tmp
wget [URL to .bin from Releases]
sysupgrade -n /tmp/cerberus-firmware-gl-mt3000.bin
```

**From bricked/any state (U-Boot):**
1. Hold Reset button while powering on for 10 seconds
2. Connect ethernet to LAN port
3. Set your PC IP to `192.168.1.2`
4. Browse to `http://192.168.1.1`
5. Upload the .bin file

After flash, wait 3-5 minutes. Then:
- **Dashboard:** `http://192.168.1.1:1471` — login with `root` / `toor`
- **SSH:** `ssh root@192.168.1.1` — password `toor`

```
root@Cerberus:~#
```

### Option 2: Install on Existing OpenWrt

If you already have OpenWrt running and don't want to reflash:

Download `cerberus-release.tar.gz` from [Releases](../../releases). Extract it, then:

```bash
# Upload binary (from the router, not Windows PowerShell)
cd /tmp
wget [URL to cerberus binary from Releases]
chmod +x cerberus
mv cerberus /usr/bin/cerberus

# Upload frontend (from your PC)
# PowerShell breaks binaries — use this for text files only:
cat www/index.html | ssh root@192.168.1.1 "mkdir -p /www/cerberus && cat > /www/cerberus/index.html"
cat www/cerberus-dashboard.jsx | ssh root@192.168.1.1 "cat > /www/cerberus/cerberus-dashboard.jsx"
cat www/cerberus-api.js | ssh root@192.168.1.1 "cat > /www/cerberus/cerberus-api.js"
cat www/react.production.min.js | ssh root@192.168.1.1 "cat > /www/cerberus/react.production.min.js"
cat www/react-dom.production.min.js | ssh root@192.168.1.1 "cat > /www/cerberus/react-dom.production.min.js"
cat www/babel.min.js | ssh root@192.168.1.1 "cat > /www/cerberus/babel.min.js"

# SSH into router and finish
ssh root@192.168.1.1
mkdir -p /var/cerberus/captures

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

Dashboard at `http://192.168.1.1:1471`

> **Note:** On Windows, never use PowerShell's `cat` to upload binary files (like the `cerberus` executable). It corrupts them. Always download binaries directly on the router with `wget`.

---

## Usage

1. Open `http://192.168.1.1:1471`
2. Login with `root` / `toor` (change in Settings)
3. Switch between **Cerberus** (offensive) and **Router** (management) in the sidebar

### Offensive Workflow
1. Hit **RECON** — discovers nearby APs
2. Select target AP from dropdown
3. Hit **SCAN** — finds clients
4. Go to **Targets** — toggle MITM/Deauth per device
5. Check **Logging** for DNS queries
6. Use **MITM** page for DNS spoofing rules
7. Use **Recon** for handshake capture (select client to deauth)
8. Launch **Evil Twin** + **Captive Portal** for credential harvesting

### Router Management
1. Switch to **Router** mode in sidebar
2. **Network** — configure WAN, LAN, WiFi, view DHCP leases
3. **System** — monitor resources, flash firmware, reboot

---

## Building

### App Only (Go binary + frontend)

Automatically built by GitHub Actions on every tagged push:

```bash
git tag Release-vX && git push --tags
```

Download from Releases tab. Binary is cross-compiled for ARM64 (GL-MT3000).

### Full Firmware (.bin)

Builds a complete flashable firmware image with Cerberus baked in:

```bash
git tag fw-vX && git push --tags
```

Or: **Actions** → **Build Firmware** → **Run workflow**

Takes ~2 hours (compiling entire Linux OS from source). Output is a `.bin` file ready to flash.

**Local build** (needs Linux + 20GB disk):
```bash
cd firmware && chmod +x build.sh && ./build.sh
```

---

## Project Structure

```
cerberus/
├── .github/workflows/
│   ├── build.yml              # CI: app binary + frontend
│   └── firmware.yml           # CI: full firmware image
├── backend/
│   ├── main.go                # Entry point
│   ├── scanner/               # Network discovery (iwinfo + arp-scan)
│   ├── mitm/                  # ARP spoof + DNS capture
│   ├── deauth/                # WiFi deauthentication
│   ├── eviltwin/              # Rogue access point
│   ├── captive/               # Phishing portal
│   ├── handshake/             # WPA handshake capture
│   ├── adapters/              # Wireless adapter management
│   ├── system/                # System info, reboot, firmware flash
│   ├── network/               # WAN/LAN/WiFi/DHCP via UCI
│   └── api/                   # 40+ REST endpoints
├── frontend/
│   ├── index.html             # Loads React locally (no CDN)
│   ├── cerberus-dashboard.jsx # Full dashboard (~800 lines)
│   └── cerberus-api.js        # API client library
├── firmware/
│   ├── build.sh               # Local firmware build script
│   ├── files/                 # Filesystem overlay (banner, hostname, password)
│   └── package/cerberus/      # OpenWrt package definition
├── scripts/
│   ├── install.sh             # One-command deploy script
│   └── cerberus.init          # OpenWrt service script
├── docs/SETUP.md              # Detailed hardware setup guide
├── LICENSE                    # Source-Available (non-commercial)
├── Makefile
└── README.md
```

---

## API

All endpoints return JSON. Port 1471.

<details>
<summary>Full API Reference (click to expand)</summary>

### Scanner
| Method | Path | Description |
|--------|------|-------------|
| POST | /api/scan | Start network scan |
| GET | /api/clients | Discovered clients |
| GET | /api/networks | Discovered APs |
| GET | /api/probes | Probe requests |

### MITM
| Method | Path | Description |
|--------|------|-------------|
| POST | /api/mitm/start | Start `{mac, ip}` |
| POST | /api/mitm/stop | Stop `{mac}` |
| GET | /api/mitm/targets | Active targets |
| GET | /api/mitm/dns | DNS query log |

### DNS Spoofing
| Method | Path | Description |
|--------|------|-------------|
| GET | /api/dns/spoof | Current spoof rules |
| POST | /api/dns/spoof | Set rules `{rules:[{domain,ip}]}` |

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
| POST | /api/handshake/start | Capture `{bssid, ssid, channel, client}` |
| POST | /api/handshake/stop | Stop |
| GET | /api/handshake/status | Status |
| GET | /api/handshake/captures | List .cap files |
| GET | /api/handshake/download/:file | Download .cap |

### System
| Method | Path | Description |
|--------|------|-------------|
| GET | /api/system/info | System info |
| POST | /api/system/reboot | Reboot |
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
| POST | /api/adapters/role | Assign role |
| POST | /api/adapters/detect | Re-detect |

</details>

---

## License

Cerberus Source-Available License v1.0 — see [LICENSE](LICENSE)

Non-commercial use only. Attribution required. See LICENSE for full terms.
