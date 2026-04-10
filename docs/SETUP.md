# Cerberus Setup Guide — GL.iNet GL-MT3000 (Beryl AX)

---

## Hardware Shopping List

- **GL.iNet GL-MT3000 (Beryl AX)** — ~$70 on Amazon. This is your router running Cerberus firmware. MediaTek MT7981B dual-core, 512MB RAM, USB 3.0 port. Runs OpenWrt natively.
- **Alfa AWUS036ACH** x2 — ~$35 each on Amazon. Dual-band, Realtek RTL8812AU chipset, supports monitor mode + packet injection out of the box on Linux. One for scanning, one for attacking.
- **USB 3.0 Hub** — ~$10 if you need extra ports. The GL-MT3000 has one USB 3.0 port, so you'll need a hub for both Alfas.
- **Ethernet cable** — to connect the router's WAN port to your ISP modem/upstream router.
- **MicroSD card (optional)** — for expanded storage if you want to store large .cap files or logs.

---

## Path A: Quick Setup (Stock OpenWrt + Install Cerberus)

This is the fastest way. Flash stock OpenWrt, then install Cerberus on top.

### Step 1: Download OpenWrt for GL-MT3000

- Go to `https://openwrt.org/toh/gl.inet/gl-mt3000`
- Download the latest **sysupgrade** image for the GL-MT3000
- File will be named something like `openwrt-mediatek-filogic-glinet_gl-mt3000-squashfs-sysupgrade.bin`

### Step 2: Flash via GL.iNet Admin Panel

- Plug your computer into the router via ethernet (LAN port)
- Power on the router
- Open browser, go to `http://192.168.8.1`
- Log into the GL.iNet admin panel (default password is on the bottom sticker, or set one up)
- Navigate to **SYSTEM → Upgrade → Local Upgrade**
- Uncheck "Keep Settings"
- Upload the OpenWrt sysupgrade .bin file
- Click **Install** and wait 2-3 minutes
- Router will reboot into stock OpenWrt

### Step 3: Initial OpenWrt Setup

- After reboot, connect to the router via ethernet
- Go to `http://192.168.1.1` (OpenWrt default)
- Set a root password when prompted
- Go to **Network → Wireless** and set up your WiFi SSID/password
- Go to **Network → Interfaces → WAN** and configure your upstream connection (usually DHCP from your ISP modem)

### Step 4: SSH In and Install Dependencies

- Open terminal on your computer
- SSH into the router:

```
ssh root@192.168.1.1
```

- Update package lists:

```
opkg update
```

- Install required packages:

```
opkg install kmod-usb-core kmod-usb3 kmod-usb-net
opkg install aircrack-ng tcpdump arpspoof dsniff nmap
opkg install hostapd-common dnsmasq
opkg install kmod-mac80211 kmod-rtl8812au-ct
opkg install nano htop screen
```

- The `kmod-rtl8812au-ct` package is the driver for your Alfa adapters. If it's not available in the default repos, you may need to compile it from source or use the `rtl8812au` package from community feeds.

### Step 5: Verify Alfa Adapters

- Plug both Alfa adapters into the USB hub, then into the router's USB 3.0 port
- Check they're detected:

```
ip link show
iwconfig
```

- You should see `wlan1` and `wlan2` (the router's built-in WiFi is `wlan0`)
- Test monitor mode:

```
airmon-ng start wlan1
iwconfig wlan1mon
```

- If you see `Mode:Monitor`, you're good. Run `airmon-ng stop wlan1mon` to restore.

### Step 6: Install Cerberus Backend

- On your computer, cross-compile the Go backend for the router:

```
GOOS=linux GOARCH=arm64 go build -o cerberus ./cerberus/
```

- SCP it to the router:

```
scp cerberus root@192.168.1.1:/usr/bin/
scp -r cerberus-dashboard/ root@192.168.1.1:/www/cerberus/
```

- Make it executable:

```
ssh root@192.168.1.1 "chmod +x /usr/bin/cerberus"
```

- Test run:

```
ssh root@192.168.1.1 "/usr/bin/cerberus"
```

- You should see the Cerberus banner and "API listening on :1471"

### Step 7: Auto-Start on Boot

- Create an init script:

```
ssh root@192.168.1.1
cat > /etc/init.d/cerberus << 'EOF'
#!/bin/sh /etc/rc.common
START=99
STOP=10

start() {
    echo "Starting Cerberus..."
    /usr/bin/cerberus > /var/log/cerberus.log 2>&1 &
}

stop() {
    echo "Stopping Cerberus..."
    killall cerberus
}
EOF

chmod +x /etc/init.d/cerberus
/etc/init.d/cerberus enable
/etc/init.d/cerberus start
```

### Step 8: Access the Dashboard

- Open browser on any device connected to the router
- Go to `http://192.168.1.1:1471`
- You'll see the Cerberus setup screen — create your username and password
- You're in

---

## Path B: Custom Firmware Build (Advanced)

Build a complete OpenWrt image with Cerberus baked in. This is cleaner but takes longer.

### Step 1: Set Up Build Environment

- You need a Linux machine (Ubuntu 22.04+ recommended) with at least 20GB free disk space
- Install build dependencies:

```
sudo apt update
sudo apt install build-essential clang flex bison g++ gawk \
  gcc-multilib g++-multilib gettext git libncurses-dev libssl-dev \
  python3-distutils rsync unzip zlib1g-dev file wget
```

### Step 2: Clone OpenWrt Buildroot

```
git clone https://git.openwrt.org/openwrt/openwrt.git
cd openwrt
git checkout v23.05.3  # or latest stable
```

### Step 3: Update Feeds

```
./scripts/feeds update -a
./scripts/feeds install -a
```

### Step 4: Configure for GL-MT3000

```
make menuconfig
```

- Navigate to **Target System** → **MediaTek Ralink ARM**
- **Subtarget** → **Filogic 8x0 (MT798x)**
- **Target Profile** → **GL.iNet GL-MT3000**
- Under **Network**, enable: `aircrack-ng`, `tcpdump`, `dsniff`, `nmap`, `hostapd`
- Under **Kernel modules → USB Support**, enable: `kmod-usb3`, `kmod-usb-net`
- Under **Kernel modules → Wireless Drivers**, enable: `kmod-rtl8812au-ct`
- Save and exit

### Step 5: Create Cerberus Package

- Create the package directory:

```
mkdir -p package/cerberus
```

- Create the Makefile:

```
cat > package/cerberus/Makefile << 'EOF'
include $(TOPDIR)/rules.mk

PKG_NAME:=cerberus
PKG_VERSION:=0.1.0
PKG_RELEASE:=1

include $(INCLUDE_DIR)/package.mk

define Package/cerberus
  SECTION:=net
  CATEGORY:=Network
  TITLE:=Cerberus Network Control System
  DEPENDS:=+aircrack-ng +tcpdump +dsniff +hostapd-common
endef

define Package/cerberus/install
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/cerberus $(1)/usr/bin/
	$(INSTALL_DIR) $(1)/www/cerberus
	$(CP) $(PKG_BUILD_DIR)/dashboard/* $(1)/www/cerberus/
	$(INSTALL_DIR) $(1)/etc/init.d
	$(INSTALL_BIN) ./files/cerberus.init $(1)/etc/init.d/cerberus
endef

$(eval $(call BuildPackage,cerberus))
EOF
```

- Cross-compile the Go backend and place it in the build directory
- Run `make menuconfig` again, find Cerberus under **Network**, enable it

### Step 6: Build the Image

```
make -j$(nproc) V=s
```

- This takes 30-90 minutes depending on your machine
- Output image will be in `bin/targets/mediatek/filogic/`
- Look for `openwrt-mediatek-filogic-glinet_gl-mt3000-squashfs-sysupgrade.bin`

### Step 7: Flash the Custom Image

- Same process as Path A Step 2 — use the GL.iNet admin panel to flash
- Or use U-Boot recovery (see below)

---

## U-Boot Recovery (If Things Go Wrong)

If you brick the router or a bad flash leaves it unresponsive:

- Power off the router
- Hold the **Reset button** on the side
- While holding Reset, power on the router
- Keep holding Reset for 10 seconds, then release
- The router enters U-Boot recovery mode
- Connect your computer via ethernet to the LAN port
- Set your computer's IP to `192.168.1.2` with subnet `255.255.255.0`
- Open browser, go to `http://192.168.1.1`
- You'll see the U-Boot web recovery page
- Upload your firmware .bin file and flash
- Wait 3-5 minutes for it to complete
- Router reboots with fresh firmware

---

## Post-Install: Adapter Configuration

Once Cerberus is running and you're in the dashboard:

- Go to **Settings → Adapter Roles**
- Assign your adapters:
  - **wlan0** (built-in MT7981B) → **Upstream** — this is your normal WiFi, keeps the router connected to your network
  - **wlan1** (Alfa #1) → **Scanning** — passive recon, airodump, handshake capture
  - **wlan2** (Alfa #2) → **Attacking** — deauth, evil twin, packet injection
- Enable **DoH Blocking** for all major resolvers (Cloudflare, Google, Quad9)
- Enable **Port 853 blocking** to kill DNS-over-TLS

---

## Workflow Once Set Up

1. Open Cerberus dashboard at `http://192.168.1.1:1471`
2. Hit **RECON** — scanning adapter discovers nearby APs
3. Select your **target AP** from the dropdown
4. Hit **SCAN** — discovers all clients on that AP
5. Select individual devices from the **Targets** page
6. Toggle **MITM** on specific devices to intercept DNS
7. Toggle **Deauth** to kick devices off
8. Use **Handshake Capture** in Recon to grab WPA handshakes for offline cracking
9. Launch **Evil Twin** to clone an AP and lure clients
10. Deploy **Captive Portal** on the evil twin for credential harvesting
11. All DNS queries logged in **Logging** with filters

---

## Closing Notes

The GL-MT3000 is the brains, the Alfas are the muscle. Path A gets you running in under an hour. Path B gives you a clean single-image firmware you can reflash anytime. Either way, once Cerberus is up, the dashboard handles everything — no command line needed for day-to-day ops.

Back up your working firmware image after setup so you can always restore to a known-good state. Store it somewhere safe.
