#!/bin/sh
# ═══════════════════════════════════════════════════════════════
# CERBERUS INSTALLER — for existing OpenWrt installations
# Usage: scp this folder to router, then run: sh install.sh
# ═══════════════════════════════════════════════════════════════

set -e

GREEN='\033[0;32m'
CYAN='\033[0;36m'
RED='\033[0;31m'
BOLD='\033[1m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

print_step() {
    echo "${CYAN}[CERBERUS]${NC} $1"
}

print_ok() {
    echo "${GREEN}[  OK  ]${NC} $1"
}

print_fail() {
    echo "${RED}[ FAIL ]${NC} $1"
}

# ─── Pre-flight checks ───────────────────────────────────────
echo ""
echo "${CYAN}${BOLD}"
echo "   ██████╗███████╗██████╗ ██████╗ ███████╗██████╗ ██╗   ██╗███████╗"
echo "  ██╔════╝██╔════╝██╔══██╗██╔══██╗██╔════╝██╔══██╗██║   ██║██╔════╝"
echo "  ██║     █████╗  ██████╔╝██████╔╝█████╗  ██████╔╝██║   ██║███████╗"
echo "  ██║     ██╔══╝  ██╔══██╗██╔══██╗██╔══╝  ██╔══██╗██║   ██║╚════██║"
echo "  ╚██████╗███████╗██║  ██║██████╔╝███████╗██║  ██║╚██████╔╝███████║"
echo "   ╚═════╝╚══════╝╚═╝  ╚═╝╚═════╝ ╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚══════╝"
echo "${NC}"
echo "  ${CYAN}Welcome to the Meme-Detector${NC}"
echo "  ${BOLD}Installer for OpenWrt${NC}"
echo ""

# Check we're on OpenWrt
if [ ! -f /etc/openwrt_release ]; then
    print_fail "This doesn't look like OpenWrt. Aborting."
    exit 1
fi
print_ok "OpenWrt detected"

# Check for cerberusd binary
if [ ! -f "$SCRIPT_DIR/cerberusd" ]; then
    print_fail "cerberusd binary not found in $SCRIPT_DIR"
    print_fail "Download the correct binary for your architecture from GitHub releases"
    echo ""
    echo "  Your arch: $(uname -m)"
    echo ""
    exit 1
fi
print_ok "cerberusd binary found ($(uname -m))"

# Check for frontend
if [ ! -d "$SCRIPT_DIR/www" ]; then
    print_fail "www/ frontend directory not found in $SCRIPT_DIR"
    exit 1
fi
print_ok "Frontend files found"

# ─── Install packages ────────────────────────────────────────
print_step "Updating package lists..."
opkg update > /tmp/cerberus-install.log 2>&1 || true

print_step "Installing dependencies..."
for pkg in tcpdump aircrack-ng wireless-tools iw ip-full iptables; do
    opkg install "$pkg" >> /tmp/cerberus-install.log 2>&1 && \
        print_ok "Installed $pkg" || \
        print_fail "Failed to install $pkg (may already be present)"
done

# Optional packages — best effort
for pkg in dsniff python3 python3-pip; do
    opkg install "$pkg" >> /tmp/cerberus-install.log 2>&1 && \
        print_ok "Installed $pkg" || true
done

# sslstrip via pip (best effort)
if command -v pip3 > /dev/null 2>&1; then
    pip3 install sslstrip >> /tmp/cerberus-install.log 2>&1 && \
        print_ok "Installed sslstrip" || true
fi

# ─── Install cerberusd ───────────────────────────────────────
print_step "Installing cerberusd..."

cp "$SCRIPT_DIR/cerberusd" /usr/bin/cerberusd
chmod +x /usr/bin/cerberusd
print_ok "Binary installed to /usr/bin/cerberusd"

# ─── Config ──────────────────────────────────────────────────
print_step "Setting up configuration..."

mkdir -p /etc/cerberus
mkdir -p /tmp/cerberus
mkdir -p /tmp/cerberus-captures

cat > /etc/cerberus/cerberus.json << 'CFGEOF'
{
  "db_path": "/tmp/cerberus/cerberus.db",
  "interface": "br-lan",
  "monitor_iface": "wlan1",
  "wan_iface": "wan",
  "leases_file": "/tmp/dhcp.leases",
  "doh_block_enabled": false,
  "doh_resolvers": [
    "1.1.1.1", "1.0.0.1",
    "8.8.8.8", "8.8.4.4",
    "9.9.9.9", "149.112.112.112",
    "208.67.222.222", "208.67.220.220",
    "94.140.14.14", "94.140.15.15",
    "76.76.2.0", "76.76.10.0"
  ],
  "alert_domains": [],
  "blocked_domains": []
}
CFGEOF
print_ok "Config written to /etc/cerberus/cerberus.json"

# ─── Frontend ────────────────────────────────────────────────
print_step "Installing dashboard..."

mkdir -p /www/cerberus
cp -r "$SCRIPT_DIR/www/"* /www/cerberus/
print_ok "Dashboard installed to /www/cerberus/"

# ─── Init script ─────────────────────────────────────────────
print_step "Installing init script..."

cat > /etc/init.d/cerberus << 'INITEOF'
#!/bin/sh /etc/rc.common

START=99
STOP=10

USE_PROCD=1
PROG=/usr/bin/cerberusd

start_service() {
    mkdir -p /tmp/cerberus
    mkdir -p /tmp/cerberus-captures

    [ -x "$PROG" ] || return 1

    sleep 3

    procd_open_instance
    procd_set_param command "$PROG" \
        -config /etc/cerberus/cerberus.json \
        -listen :8443
    procd_set_param respawn 3600 30 3
    procd_set_param stdout 1
    procd_set_param stderr 1
    procd_set_param pidfile /var/run/cerberusd.pid
    procd_close_instance
}

stop_service() {
    killall cerberusd 2>/dev/null
}

reload_service() {
    stop
    start
}
INITEOF
chmod +x /etc/init.d/cerberus
print_ok "Init script installed"

# ─── Enable service ──────────────────────────────────────────
print_step "Enabling cerberus service..."
/etc/init.d/cerberus enable
print_ok "Cerberus will start on boot"

# ─── Change root password ────────────────────────────────────
print_step "Setting root password..."
echo "root:toor" | chpasswd
print_ok "Root password changed to: toor"

# ─── Set banner ──────────────────────────────────────────────
print_step "Setting login banner..."

cat > /etc/banner << 'BANNEREOF'

   ██████╗███████╗██████╗ ██████╗ ███████╗██████╗ ██╗   ██╗███████╗
  ██╔════╝██╔════╝██╔══██╗██╔══██╗██╔════╝██╔══██╗██║   ██║██╔════╝
  ██║     █████╗  ██████╔╝██████╔╝█████╗  ██████╔╝██║   ██║███████╗
  ██║     ██╔══╝  ██╔══██╗██╔══██╗██╔══╝  ██╔══██╗██║   ██║╚════██║
  ╚██████╗███████╗██║  ██║██████╔╝███████╗██║  ██║╚██████╔╝███████║
   ╚═════╝╚══════╝╚═╝  ╚═╝╚═════╝ ╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚══════╝

   Welcome to the Meme-Detector.
   Dashboard: http://192.168.1.1:8443
   Password:  toor

BANNEREOF
print_ok "Banner set"

# ─── Start cerberus ──────────────────────────────────────────
print_step "Starting cerberus daemon..."
/etc/init.d/cerberus start
sleep 2

if pgrep cerberusd > /dev/null 2>&1; then
    print_ok "cerberusd is running (PID: $(pidof cerberusd))"
else
    print_fail "cerberusd failed to start — check: logread | grep cerberus"
fi

# ─── Done ─────────────────────────────────────────────────────
echo ""
echo "${GREEN}${BOLD}═══════════════════════════════════════════════════${NC}"
echo "${GREEN}${BOLD}  CERBERUS INSTALLED SUCCESSFULLY${NC}"
echo "${GREEN}${BOLD}═══════════════════════════════════════════════════${NC}"
echo ""
echo "  Dashboard:  ${CYAN}http://192.168.1.1:8443${NC}"
echo "  OpenWrt:    ${CYAN}http://192.168.1.1${NC}"
echo "  SSH:        ${CYAN}ssh root@192.168.1.1${NC}"
echo "  Password:   ${BOLD}toor${NC}"
echo ""
echo "  Install log: /tmp/cerberus-install.log"
echo ""
