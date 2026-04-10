#!/bin/bash
# ═══════════════════════════════════════════════════════════
#  CERBERUS FIRMWARE BUILDER
#  Builds a flashable OpenWrt image with Cerberus baked in
#  Target: GL.iNet GL-MT3000 (Beryl AX) — MediaTek MT7981B
# ═══════════════════════════════════════════════════════════

set -e

OPENWRT_VERSION="25.12.0"
OPENWRT_URL="https://github.com/openwrt/openwrt.git"
OPENWRT_DIR="openwrt"
CERBERUS_DIR="$(cd "$(dirname "$0")/.." && pwd)"
OUTPUT_DIR="${CERBERUS_DIR}/firmware/output"

C='\033[0;36m'
G='\033[0;32m'
R='\033[0;31m'
N='\033[0m'
B='\033[1m'

echo -e "${C}${B}"
echo "  ╔═══════════════════════════════════════╗"
echo "  ║  CERBERUS FIRMWARE BUILDER            ║"
echo "  ║  Target: GL-MT3000 (Beryl AX)         ║"
echo "  ║  Base: OpenWrt ${OPENWRT_VERSION}               ║"
echo "  ╚═══════════════════════════════════════╝"
echo -e "${N}"

# ─── Step 1: Check dependencies ───
echo -e "${C}[1/7]${N} Checking build dependencies..."
MISSING=""
for tool in git make gcc g++ python3 wget curl rsync; do
  if ! command -v $tool &>/dev/null; then
    MISSING="$MISSING $tool"
  fi
done

if [ -n "$MISSING" ]; then
  echo -e "${R}Missing:${N}$MISSING"
  echo "Install with: sudo apt install build-essential clang flex bison g++ gawk gcc-multilib g++-multilib gettext git libncurses-dev libssl-dev python3-distutils rsync unzip zlib1g-dev file wget"
  exit 1
fi
echo -e "${G}  ✓ All dependencies found${N}"

# ─── Step 2: Check Go ───
echo -e "${C}[2/7]${N} Checking Go..."
if ! command -v go &>/dev/null; then
  echo -e "${R}Go not found. Install from https://go.dev/dl/${N}"
  exit 1
fi
GO_VER=$(go version | awk '{print $3}')
echo -e "${G}  ✓ ${GO_VER}${N}"

# ─── Step 3: Cross-compile Cerberus backend ───
echo -e "${C}[3/7]${N} Cross-compiling Cerberus for ARM64..."
mkdir -p "${OUTPUT_DIR}"
cd "${CERBERUS_DIR}/backend"
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o "${OUTPUT_DIR}/cerberus" .
echo -e "${G}  ✓ Binary: $(du -h ${OUTPUT_DIR}/cerberus | cut -f1)${N}"

# ─── Step 4: Bundle frontend ───
echo -e "${C}[4/7]${N} Bundling frontend..."
mkdir -p "${OUTPUT_DIR}/www"
cp "${CERBERUS_DIR}/frontend/index.html" "${OUTPUT_DIR}/www/"
cp "${CERBERUS_DIR}/frontend/cerberus-dashboard.jsx" "${OUTPUT_DIR}/www/"
cp "${CERBERUS_DIR}/frontend/cerberus-api.js" "${OUTPUT_DIR}/www/"

# Download React/Babel if not present
for f in "react/18.2.0/umd/react.production.min.js" "react-dom/18.2.0/umd/react-dom.production.min.js" "babel-standalone/7.23.9/babel.min.js"; do
  fname=$(basename "$f")
  if [ ! -f "${OUTPUT_DIR}/www/${fname}" ]; then
    curl -sL "https://cdnjs.cloudflare.com/ajax/libs/${f}" -o "${OUTPUT_DIR}/www/${fname}"
  fi
done
echo -e "${G}  ✓ Frontend bundled: $(ls ${OUTPUT_DIR}/www/ | wc -l) files${N}"

# ─── Step 5: Clone/update OpenWrt ───
echo -e "${C}[5/7]${N} Setting up OpenWrt ${OPENWRT_VERSION}..."
cd "${CERBERUS_DIR}/firmware"

if [ ! -d "${OPENWRT_DIR}" ]; then
  git clone --depth 1 --branch "v${OPENWRT_VERSION}" "${OPENWRT_URL}" "${OPENWRT_DIR}"
else
  echo "  Using existing OpenWrt source"
fi

cd "${OPENWRT_DIR}"

# Update feeds
./scripts/feeds update -a
./scripts/feeds install -a
echo -e "${G}  ✓ OpenWrt source ready${N}"

# ─── Step 6: Configure ───
echo -e "${C}[6/7]${N} Configuring for GL-MT3000..."

# Create the Cerberus package in the OpenWrt tree
mkdir -p package/cerberus/files
cp "${CERBERUS_DIR}/firmware/package/cerberus/Makefile" package/cerberus/
cp "${CERBERUS_DIR}/firmware/package/cerberus/files/cerberus.init" package/cerberus/files/

# Create the package build dir with our pre-compiled binary and frontend
mkdir -p "build_dir/cerberus"
cp "${OUTPUT_DIR}/cerberus" "build_dir/cerberus/"
cp -r "${OUTPUT_DIR}/www" "build_dir/cerberus/"

# Generate default config for GL-MT3000
cat > .config << 'DEFCONFIG'
CONFIG_TARGET_mediatek=y
CONFIG_TARGET_mediatek_filogic=y
CONFIG_TARGET_mediatek_filogic_DEVICE_glinet_gl-mt3000=y

# Cerberus and dependencies
CONFIG_PACKAGE_cerberus=y
CONFIG_PACKAGE_aircrack-ng=y
CONFIG_PACKAGE_tcpdump=y
CONFIG_PACKAGE_nmap=y
CONFIG_PACKAGE_hostapd=y
CONFIG_PACKAGE_dnsmasq-full=y

# USB support for Alfa adapters
CONFIG_PACKAGE_kmod-usb-core=y
CONFIG_PACKAGE_kmod-usb3=y
CONFIG_PACKAGE_kmod-usb-net=y
CONFIG_PACKAGE_kmod-mac80211=y

# Alfa RTL8812AU driver
CONFIG_PACKAGE_kmod-rtl8812au-ct=y

# Network tools
CONFIG_PACKAGE_iwinfo=y
CONFIG_PACKAGE_ethtool=y
CONFIG_PACKAGE_ip-full=y

# SSH
CONFIG_PACKAGE_openssh-sftp-server=y
CONFIG_PACKAGE_dropbear=y

# Disable LuCI (Cerberus replaces it)
# CONFIG_PACKAGE_luci is not set
DEFCONFIG

make defconfig

# Copy custom filesystem overlay (banner, hostname, password, etc.)
cp -r "${CERBERUS_DIR}/firmware/files" ./files/
chmod +x ./files/etc/uci-defaults/99-cerberus

echo -e "${G}  ✓ Config generated for GL-MT3000${N}"
echo -e "${G}  ✓ Custom files overlay applied (hostname=Cerberus, password=toor)${N}"

# ─── Step 7: Build ───
echo -e "${C}[7/7]${N} Building firmware (this takes 30-90 minutes)..."
echo "     Cores: $(nproc)"
echo "     Output: bin/targets/mediatek/filogic/"
echo ""

make -j$(nproc) FILES=files/ V=s 2>&1 | tee "${OUTPUT_DIR}/build.log"

# Copy output
FWFILE=$(find bin/targets/mediatek/filogic/ -name "*gl-mt3000*sysupgrade*" -not -name "*.sha256" | head -1)

if [ -n "$FWFILE" ]; then
  cp "$FWFILE" "${OUTPUT_DIR}/cerberus-firmware-gl-mt3000.bin"
  FWSIZE=$(du -h "${OUTPUT_DIR}/cerberus-firmware-gl-mt3000.bin" | cut -f1)
  echo ""
  echo -e "${G}${B}═══════════════════════════════════════${N}"
  echo -e "${G}${B}  FIRMWARE BUILD COMPLETE${N}"
  echo -e "${G}${B}═══════════════════════════════════════${N}"
  echo ""
  echo -e "  ${C}File:${N} ${OUTPUT_DIR}/cerberus-firmware-gl-mt3000.bin"
  echo -e "  ${C}Size:${N} ${FWSIZE}"
  echo ""
  echo -e "  ${B}Flash via SSH:${N}"
  echo -e "  scp cerberus-firmware-gl-mt3000.bin root@192.168.1.1:/tmp/"
  echo -e "  ssh root@192.168.1.1 'sysupgrade -n /tmp/cerberus-firmware-gl-mt3000.bin'"
  echo ""
  echo -e "  ${B}Flash via U-Boot:${N}"
  echo -e "  Hold reset + power on for 10s, browse to http://192.168.1.1"
  echo ""
else
  echo -e "${R}Build failed. Check ${OUTPUT_DIR}/build.log${N}"
  exit 1
fi
