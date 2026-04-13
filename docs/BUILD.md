# CERBERUS — Build Guide

Build Cerberus firmware from source, flash it to your router, and get full DNS visibility on your network.

---

## Prerequisites

**Your machine needs:**

- Go 1.22+ ([https://go.dev/dl](https://go.dev/dl))
- Node.js 20+ and npm ([https://nodejs.org](https://nodejs.org))
- Git
- 10GB free disk space (OpenWrt Image Builder is chunky)
- Linux x86_64 (native or WSL2 on Windows)

**Target router (pick one):**

| Router | Price | Arch | OpenWrt Target |
|---|---|---|---|
| GL.iNet GL-MT3000 (Beryl AX) | ~$70 | aarch64 | mediatek/filogic |
| GL.iNet GL-MT1300 | ~$40 | mipsle | ramips/mt7621 |
| Raspberry Pi 4 + USB ethernet | ~$50 | aarch64 | bcm27xx/bcm2711 |
| Any x86 mini PC | varies | amd64 | x86/64 |

The GL-MT3000 is the recommended target. It ships with OpenWrt, has good specs, USB-C power, and costs less than a nice dinner.

---

## Option A: Build Locally

### Step 1 — Clone the repo

```bash
git clone https://github.com/YOUR_USER/cerberus.git
cd cerberus
```

### Step 2 — Build the Go daemon

For GL-MT3000 (aarch64):

```bash
sudo apt-get install -y gcc-aarch64-linux-gnu libpcap-dev

GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc \
  go build -ldflags="-s -w" -o cerberusd ./cmd/cerberusd/
```

For x86_64:

```bash
sudo apt-get install -y libpcap-dev

GOOS=linux GOARCH=amd64 CGO_ENABLED=1 \
  go build -ldflags="-s -w" -o cerberusd ./cmd/cerberusd/
```

For MIPS (GL-MT1300):

```bash
sudo apt-get install -y gcc-mipsel-linux-gnu

GOOS=linux GOARCH=mipsle CGO_ENABLED=1 CC=mipsel-linux-gnu-gcc \
  go build -ldflags="-s -w" -o cerberusd ./cmd/cerberusd/
```

### Step 3 — Build the web frontend

```bash
cd web
npm ci
npm run build
cd ..
```

The built files land in `web/dist/`.

### Step 4 — Download OpenWrt Image Builder

Pick the right one for your router.

For GL-MT3000:

```bash
OPENWRT_VER=23.05.5
TARGET=mediatek
SUBTARGET=filogic

wget "https://downloads.openwrt.org/releases/${OPENWRT_VER}/targets/${TARGET}/${SUBTARGET}/openwrt-imagebuilder-${OPENWRT_VER}-${TARGET}-${SUBTARGET}.Linux-x86_64.tar.xz"
tar xf openwrt-imagebuilder-*.tar.xz
mv openwrt-imagebuilder-*/ imagebuilder/
```

### Step 5 — Prepare custom files overlay

```bash
mkdir -p custom-files/usr/bin
mkdir -p custom-files/etc/init.d
mkdir -p custom-files/etc/cerberus
mkdir -p custom-files/www/cerberus

cp cerberusd custom-files/usr/bin/cerberusd
chmod +x custom-files/usr/bin/cerberusd

cp openwrt/files/etc/init.d/cerberus custom-files/etc/init.d/cerberus
chmod +x custom-files/etc/init.d/cerberus

cp openwrt/files/etc/cerberus/cerberus.json custom-files/etc/cerberus/cerberus.json

cp -r web/dist/* custom-files/www/cerberus/
```

### Step 6 — Build the firmware image

```bash
cd imagebuilder

make image \
  PROFILE="glinet_gl-mt3000" \
  PACKAGES="libpcap iptables kmod-nf-nat luci-ssl" \
  FILES="../custom-files" \
  EXTRA_IMAGE_NAME="cerberus"
```

The output firmware lands in `imagebuilder/bin/targets/mediatek/filogic/`.

You want the file ending in `-sysupgrade.bin`.

### Step 7 — Flash it

**If your router is running stock GL.iNet firmware:**

1. Open `192.168.8.1` in your browser
2. Go to **Upgrade** → **Local Upgrade**
3. Upload the `*-sysupgrade.bin` file
4. Uncheck "Keep settings" (clean flash is safer)
5. Click Upgrade and wait 2-3 minutes

**If already running OpenWrt:**

```bash
scp openwrt-*-cerberus-*-sysupgrade.bin root@192.168.1.1:/tmp/
ssh root@192.168.1.1
sysupgrade -n /tmp/openwrt-*-cerberus-*-sysupgrade.bin
```

The `-n` flag means don't keep old config. Fresh start.

### Step 8 — Verify

After the router reboots (~2 minutes):

```bash
ssh root@192.168.1.1
# Check daemon is running
ps | grep cerberusd

# Check logs
logread | grep cerberus
```

Open `http://192.168.1.1:8443` in your browser. You should see the Cerberus dashboard.

---

## Option B: Build with GitHub Actions

Push to your fork with a tag and CI builds everything automatically.

```bash
git tag v0.1.0
git push origin main --tags
```

The workflow builds:

1. `cerberusd` binaries for all architectures
2. Vue frontend
3. Complete firmware images with Cerberus baked in

Download the firmware artifact from the Actions tab and flash per Step 7 above.

---

## Post-Flash Setup

### Access the dashboard

Navigate to `http://192.168.1.1:8443`

### Name your devices

Go to the **Devices** tab. Click **Rename** next to any device to give it a friendly name like "Jake's Laptop" instead of seeing raw IPs.

### DoH blocking

The DoH blocker is **on by default**. The red "DoH BLOCKED" button in the header toggles it. When active, it blocks outbound connections to all major DNS-over-HTTPS resolver IPs (Cloudflare, Google, Quad9, OpenDNS, AdGuard, NextDNS, Control D, Mullvad) on ports 443 and 853.

This forces all devices on your network to fall back to the router's own DNS resolver, which Cerberus sniffs and logs.

### SSH access

Default credentials after flash:

```
Host: 192.168.1.1
User: root
Password: (none — set one immediately)
```

```bash
ssh root@192.168.1.1
passwd
```

---

## Troubleshooting

**Dashboard shows "Connection error"**
- Verify cerberusd is running: `ps | grep cerberusd`
- Check port: `netstat -tlnp | grep 8443`
- Check logs: `logread | grep cerberus`

**No DNS queries showing up**
- Confirm the interface name: `ip addr` — look for `br-lan`
- If your interface differs, edit `/etc/cerberus/cerberus.json` and change `interface`
- Restart: `/etc/init.d/cerberus restart`

**Device falls back to DoH despite blocking**
- Some apps (Firefox, Chrome) use built-in DoH. Check browser settings and disable "DNS over HTTPS" or "Secure DNS" in the browser
- Firefox: `about:config` → `network.trr.mode` → set to `5` (disabled)
- Chrome: `chrome://settings/security` → disable "Use secure DNS"

**Dashboard not loading (404)**
- Verify frontend files exist: `ls /www/cerberus/`
- Should contain `index.html` and an `assets/` directory

**Daemon crashes on start**
- Usually a permissions issue with pcap
- Check: `ls -la /usr/bin/cerberusd` — needs to be executable
- Check libpcap: `opkg list-installed | grep pcap`
