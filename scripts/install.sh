#!/bin/bash
# ═══════════════════════════════════════════════════════════
#  CERBERUS INSTALLER v0.2.0
#  Deploys Cerberus Network Control System to GL-MT3000
# ═══════════════════════════════════════════════════════════

set -e

ROUTER="${1:-192.168.1.1}"
USER="root"
PORT="1471"
DIR="$(cd "$(dirname "$0")" && pwd)"

C='\033[0;36m'  # cyan
G='\033[0;32m'  # green
R='\033[0;31m'  # red
Y='\033[0;33m'  # yellow
D='\033[0;90m'  # dim
N='\033[0m'     # reset
B='\033[1m'     # bold

banner() {
  echo ""
  echo -e "${C}"
  echo "   ██████╗███████╗██████╗ ██████╗ ███████╗██████╗ ██╗   ██╗███████╗"
  echo "  ██╔════╝██╔════╝██╔══██╗██╔══██╗██╔════╝██╔══██╗██║   ██║██╔════╝"
  echo "  ██║     █████╗  ██████╔╝██████╔╝█████╗  ██████╔╝██║   ██║███████╗"
  echo "  ██║     ██╔══╝  ██╔══██╗██╔══██╗██╔══╝  ██╔══██╗██║   ██║╚════██║"
  echo "  ╚██████╗███████╗██║  ██║██████╔╝███████╗██║  ██║╚██████╔╝███████║"
  echo "   ╚═════╝╚══════╝╚═╝  ╚═╝╚═════╝ ╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚══════╝"
  echo -e "${N}"
  echo -e "  ${C}${B}Network Control System${N} ${D}v0.2.0${N}"
  echo ""
}

ok()   { echo -e "  ${G}✓${N} $1"; }
fail() { echo -e "  ${R}✗${N} $1"; }
info() { echo -e "  ${C}→${N} $1"; }
warn() { echo -e "  ${Y}!${N} $1"; }
step() { echo ""; echo -e "${C}${B}[$1/8]${N} ${B}$2${N}"; }

banner
echo -e "  Target: ${C}${USER}@${ROUTER}${N}"
echo -e "  Source: ${D}${DIR}${N}"
echo ""

# ═══════════════════════════════════════════════════════════
# STEP 1: Verify local files exist
# ═══════════════════════════════════════════════════════════
step 1 "Checking local files"

BINARY="${DIR}/cerberus"
WWW="${DIR}/www"
INIT="${DIR}/etc/cerberus.init"
MISSING=0

if [ -f "$BINARY" ]; then
  ok "Binary found: $(du -h "$BINARY" | cut -f1)"
else
  fail "Binary not found: ${BINARY}"
  MISSING=1
fi

if [ -d "$WWW" ]; then
  ok "Frontend found: $(ls "$WWW" | wc -l) files"
else
  fail "Frontend not found: ${WWW}"
  MISSING=1
fi

if [ -f "$INIT" ]; then
  ok "Init script found"
else
  warn "Init script missing — will create on router"
fi

if [ "$MISSING" = "1" ]; then
  echo ""
  fail "Missing required files. Make sure you extracted the release properly."
  echo -e "  ${D}Expected structure:${N}"
  echo -e "  ${D}  cerberus          (ARM64 binary)${N}"
  echo -e "  ${D}  www/              (frontend files)${N}"
  echo -e "  ${D}  etc/cerberus.init (service script)${N}"
  exit 1
fi

# ═══════════════════════════════════════════════════════════
# STEP 2: Test SSH connection
# ═══════════════════════════════════════════════════════════
step 2 "Testing SSH to ${ROUTER}"

if ssh -o ConnectTimeout=5 -o BatchMode=yes -o StrictHostKeyChecking=no ${USER}@${ROUTER} "echo ok" > /dev/null 2>&1; then
  ok "SSH connection OK"
else
  fail "Cannot connect via SSH"
  echo ""
  echo -e "  ${Y}Troubleshooting:${N}"
  echo -e "  ${D}1. Connect to router via ethernet${N}"
  echo -e "  ${D}2. Make sure router is running OpenWrt${N}"
  echo -e "  ${D}3. Copy your SSH key:  ssh-copy-id root@${ROUTER}${N}"
  echo -e "  ${D}4. Or set a password:  ssh root@${ROUTER}${N}"
  exit 1
fi

# Check if OpenWrt
FIRMWARE=$(ssh ${USER}@${ROUTER} "cat /etc/openwrt_release 2>/dev/null | grep DISTRIB_DESCRIPTION | cut -d\\' -f2" 2>/dev/null || echo "")
if [ -n "$FIRMWARE" ]; then
  ok "Firmware: ${FIRMWARE}"
else
  warn "Could not detect OpenWrt — proceeding anyway"
fi

# ═══════════════════════════════════════════════════════════
# STEP 3: Stop existing Cerberus
# ═══════════════════════════════════════════════════════════
step 3 "Stopping existing Cerberus (if running)"

RUNNING=$(ssh ${USER}@${ROUTER} "pgrep cerberus 2>/dev/null" || echo "")
if [ -n "$RUNNING" ]; then
  ssh ${USER}@${ROUTER} "killall cerberus 2>/dev/null; sleep 1"
  ok "Stopped running instance (PID: ${RUNNING})"
else
  ok "No existing instance"
fi

# ═══════════════════════════════════════════════════════════
# STEP 4: Install dependencies
# ═══════════════════════════════════════════════════════════
step 4 "Installing dependencies on router"

ssh ${USER}@${ROUTER} << 'DEPS'
  # Update package list (silent)
  opkg update > /dev/null 2>&1

  INSTALLED=0
  SKIPPED=0
  FAILED=0

  install_pkg() {
    if opkg list-installed 2>/dev/null | grep -q "^$1 "; then
      SKIPPED=$((SKIPPED+1))
    elif opkg install "$1" > /dev/null 2>&1; then
      echo "  + $1"
      INSTALLED=$((INSTALLED+1))
    else
      echo "  ? $1 (unavailable)"
      FAILED=$((FAILED+1))
    fi
  }

  # Core offensive tools
  install_pkg aircrack-ng
  install_pkg tcpdump
  install_pkg dsniff
  install_pkg nmap
  install_pkg hostapd-common
  install_pkg dnsmasq

  # USB + wireless
  install_pkg kmod-usb-core
  install_pkg kmod-usb3
  install_pkg kmod-usb-net
  install_pkg kmod-mac80211

  # Alfa adapter driver
  install_pkg kmod-rtl8812au-ct

  # Network management tools (for UCI / system control)
  install_pkg uci
  install_pkg luci-lib-jsonc
  install_pkg iwinfo
  install_pkg ethtool

  # Create required directories
  mkdir -p /var/cerberus/captures
  mkdir -p /www/cerberus
  mkdir -p /var/log

  echo "  --- $INSTALLED installed, $SKIPPED already had, $FAILED unavailable"
DEPS

ok "Dependencies ready"

# ═══════════════════════════════════════════════════════════
# STEP 5: Upload binary
# ═══════════════════════════════════════════════════════════
step 5 "Uploading Cerberus binary"

scp -q "$BINARY" ${USER}@${ROUTER}:/usr/bin/cerberus
ssh ${USER}@${ROUTER} "chmod +x /usr/bin/cerberus"
REMOTE_SIZE=$(ssh ${USER}@${ROUTER} "du -h /usr/bin/cerberus | cut -f1")
ok "Binary deployed: ${REMOTE_SIZE}"

# ═══════════════════════════════════════════════════════════
# STEP 6: Upload frontend
# ═══════════════════════════════════════════════════════════
step 6 "Uploading dashboard frontend"

scp -q -r "${WWW}"/* ${USER}@${ROUTER}:/www/cerberus/
FCOUNT=$(ssh ${USER}@${ROUTER} "ls /www/cerberus/ | wc -l")
ok "Frontend deployed: ${FCOUNT} files"

# ═══════════════════════════════════════════════════════════
# STEP 7: Install service
# ═══════════════════════════════════════════════════════════
step 7 "Installing service"

if [ -f "$INIT" ]; then
  scp -q "$INIT" ${USER}@${ROUTER}:/etc/init.d/cerberus
else
  # Create init script directly on router
  ssh ${USER}@${ROUTER} << 'INITSCRIPT'
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
    killall cerberus 2>/dev/null
}

restart() {
    stop
    sleep 1
    start
}
EOF
INITSCRIPT
fi

ssh ${USER}@${ROUTER} "chmod +x /etc/init.d/cerberus && /etc/init.d/cerberus enable 2>/dev/null"
ok "Service installed and enabled on boot"

# Enable IP forwarding (needed for MITM)
ssh ${USER}@${ROUTER} "echo 1 > /proc/sys/net/ipv4/ip_forward 2>/dev/null"
# Make it persistent
ssh ${USER}@${ROUTER} "uci set network.globals.ip_forward='1' 2>/dev/null; uci commit network 2>/dev/null"
ok "IP forwarding enabled"

# ═══════════════════════════════════════════════════════════
# STEP 8: Start and verify
# ═══════════════════════════════════════════════════════════
step 8 "Starting Cerberus"

ssh ${USER}@${ROUTER} "/etc/init.d/cerberus restart"
sleep 3

PID=$(ssh ${USER}@${ROUTER} "pgrep cerberus 2>/dev/null" || echo "")

if [ -n "$PID" ]; then
  ok "Cerberus is running (PID: ${PID})"

  # Detect adapters
  ADAPTERS=$(ssh ${USER}@${ROUTER} "iw dev 2>/dev/null | grep Interface | wc -l" || echo "?")
  ok "Wireless adapters detected: ${ADAPTERS}"

  # Check if Alfa is plugged in
  USB_WIFI=$(ssh ${USER}@${ROUTER} "lsusb 2>/dev/null | grep -i 'realtek\|ralink\|atheros\|mediatek' | wc -l" || echo "0")
  if [ "$USB_WIFI" -gt "0" ]; then
    ok "USB WiFi adapter(s) detected: ${USB_WIFI}"
  else
    warn "No USB WiFi adapter found — offensive features need an Alfa plugged in"
  fi

  echo ""
  echo -e "  ${G}${B}═══════════════════════════════════════════${N}"
  echo -e "  ${G}${B}  INSTALL COMPLETE${N}"
  echo -e "  ${G}${B}═══════════════════════════════════════════${N}"
  echo ""
  echo -e "  ${C}Dashboard${N}     http://${ROUTER}:${PORT}"
  echo -e "  ${C}SSH${N}           ssh root@${ROUTER}"
  echo -e "  ${C}Logs${N}          ssh root@${ROUTER} tail -f /var/log/cerberus.log"
  echo -e "  ${C}Captures${N}      /var/cerberus/captures/"
  echo -e "  ${C}Restart${N}       ssh root@${ROUTER} /etc/init.d/cerberus restart"
  echo -e "  ${C}Stop${N}          ssh root@${ROUTER} /etc/init.d/cerberus stop"
  echo ""
  echo -e "  ${D}First visit: create your account at the login screen.${N}"
  echo -e "  ${D}Plug in Alfa adapter for deauth/evil twin/handshake.${N}"
  echo ""
else
  fail "Cerberus failed to start"
  echo ""
  echo -e "  ${Y}Debug:${N}"
  echo -e "  ${D}  ssh root@${ROUTER} cat /var/log/cerberus.log${N}"
  echo -e "  ${D}  ssh root@${ROUTER} /usr/bin/cerberus${N}"
  exit 1
fi
