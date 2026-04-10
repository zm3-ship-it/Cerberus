ROUTER ?= 192.168.1.1
USER   ?= root
DIST   := dist

.PHONY: all build build-arm build-x86 frontend package install clean

all: build frontend package

# Build ARM64 binary for GL-MT3000
build-arm:
	@echo "[*] Building for ARM64 (GL-MT3000)..."
	@mkdir -p $(DIST)
	cd backend && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ../$(DIST)/cerberus .
	@echo "✓ $(DIST)/cerberus"

# Build x86 binary for local testing
build-x86:
	@echo "[*] Building for x86..."
	@mkdir -p $(DIST)
	cd backend && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ../$(DIST)/cerberus-x86 .
	@echo "✓ $(DIST)/cerberus-x86"

build: build-arm build-x86

# Package frontend
frontend:
	@echo "[*] Packaging frontend..."
	@mkdir -p $(DIST)/www
	cp frontend/cerberus-dashboard.jsx $(DIST)/www/
	cp frontend/cerberus-api.js $(DIST)/www/
	cp frontend/index.html $(DIST)/www/
	@echo "✓ Frontend ready"

# Create release tarball
package: build frontend
	@echo "[*] Packaging release..."
	@mkdir -p $(DIST)/etc
	@cp scripts/cerberus.init $(DIST)/etc/
	@cp scripts/install.sh $(DIST)/
	@chmod +x $(DIST)/install.sh $(DIST)/etc/cerberus.init
	cd $(DIST) && tar -czf ../cerberus-release.tar.gz .
	@echo "✓ cerberus-release.tar.gz"

# Deploy to router
install: build-arm frontend
	@echo "[*] Deploying to $(USER)@$(ROUTER)..."
	@ssh $(USER)@$(ROUTER) "killall cerberus 2>/dev/null || true"
	@ssh $(USER)@$(ROUTER) "opkg update >/dev/null 2>&1; opkg install aircrack-ng tcpdump dsniff nmap hostapd-common dnsmasq kmod-usb3 kmod-mac80211 2>/dev/null; mkdir -p /var/cerberus/captures /www/cerberus"
	scp $(DIST)/cerberus $(USER)@$(ROUTER):/usr/bin/cerberus
	@ssh $(USER)@$(ROUTER) "chmod +x /usr/bin/cerberus"
	scp -r $(DIST)/www/* $(USER)@$(ROUTER):/www/cerberus/
	scp scripts/cerberus.init $(USER)@$(ROUTER):/etc/init.d/cerberus
	@ssh $(USER)@$(ROUTER) "chmod +x /etc/init.d/cerberus && /etc/init.d/cerberus enable && /etc/init.d/cerberus restart"
	@sleep 2
	@echo ""
	@echo "✓ Cerberus deployed → http://$(ROUTER):1471"

clean:
	rm -rf $(DIST) cerberus-release.tar.gz
