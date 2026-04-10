#!/bin/bash
# ═══════════════════════════════════════════════════════════
#  CERBERUS FIRMWARE BUILDER
#  Builds a flashable OpenWrt image with Cerberus baked in
#  Target: GL.iNet GL-MT3000 (Beryl AX)
# ═══════════════════════════════════════════════════════════

set -e

OPENWRT_VERSION="25.12.0"
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
  exit 1
fi
echo -e "${G}  ✓ All dependencies found${N}"

# ─── Step 2: Check Go ───
echo -e "${C}[2/7]${N} Checking Go..."
if ! command -v go &>/dev/null; then
  echo -e "${R}Go not found. Install from https://go.dev/dl/${N}"
  exit 1
fi
echo -e "${G}  ✓ $(go version | awk '{print $3}')${N}"

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

for f in "react/18.2.0/umd/react.production.min.js" "react-dom/18.2.0/umd/react-dom.production.min.js" "babel-standalone/7.23.9/babel.min.js"; do
  fname=$(basename "$f")
  if [ ! -f "${OUTPUT_DIR}/www/${fname}" ]; then
    curl -sL "https://cdnjs.cloudflare.com/ajax/libs/${f}" -o "${OUTPUT_DIR}/www/${fname}"
  fi
done
echo -e "${G}  ✓ Frontend bundled${N}"

# ─── Step 5: Clone/update OpenWrt ───
echo -e "${C}[5/7]${N} Setting up OpenWrt ${OPENWRT_VERSION}..."
cd "${CERBERUS_DIR}/firmware"

if [ ! -d "${OPENWRT_DIR}" ]; then
  git clone --depth 1 --branch "v${OPENWRT_VERSION}" "https://github.com/openwrt/openwrt.git" "${OPENWRT_DIR}"
fi

cd "${OPENWRT_DIR}"

# Remove problematic feeds
sed -i '/telephony/d' feeds.conf.default 2>/dev/null || true
sed -i '/video/d' feeds.conf.default 2>/dev/null || true

./scripts/feeds update -a
./scripts/feeds install -a
echo -e "${G}  ✓ OpenWrt source ready${N}"

# ─── Step 6: Configure + Stage files ───
echo -e "${C}[6/7]${N} Configuring..."

# Create filesystem overlay — this is how we bake Cerberus into the image
# Everything in files/ gets merged into the root filesystem
mkdir -p files/usr/bin
mkdir -p files/www/cerberus
mkdir -p files/etc/init.d
mkdir -p files/etc/uci-defaults
mkdir -p files/etc/config
mkdir -p files/etc/cerberus
mkdir -p files/var/cerberus/captures

# Copy Cerberus binary
cp "${OUTPUT_DIR}/cerberus" files/usr/bin/cerberus
chmod +x files/usr/bin/cerberus

# Copy frontend
cp "${OUTPUT_DIR}/www/"* files/www/cerberus/

# Copy init script
cp "${CERBERUS_DIR}/firmware/package/cerberus/files/cerberus.init" files/etc/init.d/cerberus
chmod +x files/etc/init.d/cerberus

# Copy banner
if [ -f "${CERBERUS_DIR}/firmware/files/etc/banner" ]; then
  cp "${CERBERUS_DIR}/firmware/files/etc/banner" files/etc/banner
fi

# Copy system config (hostname=Cerberus)
if [ -f "${CERBERUS_DIR}/firmware/files/etc/config/system" ]; then
  cp "${CERBERUS_DIR}/firmware/files/etc/config/system" files/etc/config/system
fi

# First-boot script (sets password to toor, enables cerberus)
cat > files/etc/uci-defaults/99-cerberus << 'FIRSTBOOT'
#!/bin/sh
echo -e "toor\ntoor" | passwd root
uci set system.@system[0].hostname='Cerberus'
uci commit system
echo 1 > /proc/sys/net/ipv4/ip_forward
/etc/init.d/cerberus enable 2>/dev/null
exit 0
FIRSTBOOT
chmod +x files/etc/uci-defaults/99-cerberus

echo -e "${G}  ✓ Files overlay staged:${N}"
find files/ -type f | sort | sed 's/^/      /'

# Write OpenWrt config
cat > .config << 'DEFCONFIG'
CONFIG_TARGET_mediatek=y
CONFIG_TARGET_mediatek_filogic=y
CONFIG_TARGET_MULTI_PROFILE=y
CONFIG_TARGET_PER_DEVICE_ROOTFS=y
CONFIG_TARGET_mediatek_filogic_DEVICE_glinet_gl-mt3000=y
CONFIG_PACKAGE_aircrack-ng=y
CONFIG_PACKAGE_tcpdump=y
CONFIG_PACKAGE_nmap=y
CONFIG_PACKAGE_hostapd=y
CONFIG_PACKAGE_dnsmasq-full=y
CONFIG_PACKAGE_kmod-usb-core=y
CONFIG_PACKAGE_kmod-usb3=y
CONFIG_PACKAGE_kmod-mac80211=y
CONFIG_PACKAGE_iwinfo=y
CONFIG_PACKAGE_ethtool=y
CONFIG_PACKAGE_openssh-sftp-server=y
CONFIG_PACKAGE_arp-scan=y
DEFCONFIG

make defconfig

# Force GL-MT3000 device in case defconfig reset it
sed -i 's/# CONFIG_TARGET_mediatek_filogic_DEVICE_glinet_gl-mt3000 is not set/CONFIG_TARGET_mediatek_filogic_DEVICE_glinet_gl-mt3000=y/' .config
grep -q "CONFIG_TARGET_mediatek_filogic_DEVICE_glinet_gl-mt3000=y" .config || echo "CONFIG_TARGET_mediatek_filogic_DEVICE_glinet_gl-mt3000=y" >> .config
make defconfig

# Verify
if grep -q "CONFIG_TARGET_mediatek_filogic_DEVICE_glinet_gl-mt3000=y" .config; then
  echo -e "${G}  ✓ GL-MT3000 device selected${N}"
else
  echo -e "${R}  ✗ GL-MT3000 not selected — check .config${N}"
  exit 1
fi

# ─── Step 7: Build ───
echo -e "${C}[7/7]${N} Building firmware..."
echo "     Cores: $(nproc)"
echo ""

make -j$(nproc) V=s 2>&1 | tee "${OUTPUT_DIR}/build.log" || make -j1 V=s 2>&1 | tee -a "${OUTPUT_DIR}/build.log"

# Find output
FWFILE=$(find bin/targets/mediatek/filogic/ -name "*gl-mt3000*squashfs-sysupgrade*" | head -1)

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
  echo -e "  ${B}Flash:${N}"
  echo "  scp ${OUTPUT_DIR}/cerberus-firmware-gl-mt3000.bin root@192.168.1.1:/tmp/fw.bin"
  echo "  ssh root@192.168.1.1 'sysupgrade -n /tmp/fw.bin'"
  echo ""
  echo -e "  ${B}After flash:${N}"
  echo "  SSH: root@192.168.1.1 (password: toor)"
  echo "  Web: http://192.168.1.1:1471"
  echo ""
else
  echo -e "${R}Build failed — no sysupgrade image found${N}"
  echo "Check ${OUTPUT_DIR}/build.log"
  exit 1
fi
