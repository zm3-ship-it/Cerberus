# SPDX-License-Identifier: MIT
#
# Cerberus — custom network monitoring daemon for OpenWrt
#

include $(TOPDIR)/rules.mk

PKG_NAME:=cerberus
PKG_VERSION:=0.1.0
PKG_RELEASE:=1
PKG_LICENSE:=MIT

PKG_BUILD_DIR:=$(BUILD_DIR)/$(PKG_NAME)-$(PKG_VERSION)

include $(INCLUDE_DIR)/package.mk

define Package/cerberus
  SECTION:=net
  CATEGORY:=Network
  TITLE:=Cerberus Network Monitor
  DEPENDS:=+libpcap +iptables
  URL:=https://github.com/YOUR_USER/cerberus
endef

define Package/cerberus/description
  Cerberus is a network monitoring dashboard for OpenWrt routers.
  It provides real-time DNS query logging, device tracking, and
  DNS-over-HTTPS blocking with a modern web interface.
endef

define Build/Prepare
	mkdir -p $(PKG_BUILD_DIR)
	$(CP) ./src/* $(PKG_BUILD_DIR)/
endef

# We cross-compile the Go binary externally via GitHub Actions
# and include the pre-built binary. This Makefile packages it.
define Build/Compile
	@echo "Using pre-built cerberusd binary"
endef

define Package/cerberus/install
	# Binary
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/cerberusd $(1)/usr/bin/cerberusd

	# Init script
	$(INSTALL_DIR) $(1)/etc/init.d
	$(INSTALL_BIN) ./files/etc/init.d/cerberus $(1)/etc/init.d/cerberus

	# Default config
	$(INSTALL_DIR) $(1)/etc/cerberus
	$(INSTALL_DATA) ./files/etc/cerberus/cerberus.json $(1)/etc/cerberus/cerberus.json

	# Web frontend
	$(INSTALL_DIR) $(1)/www/cerberus
	$(CP) ./files/www/cerberus/* $(1)/www/cerberus/
endef

define Package/cerberus/postinst
#!/bin/sh
[ -n "$${IPKG_INSTROOT}" ] || {
	/etc/init.d/cerberus enable
	/etc/init.d/cerberus start
}
endef

define Package/cerberus/prerm
#!/bin/sh
[ -n "$${IPKG_INSTROOT}" ] || {
	/etc/init.d/cerberus stop
	/etc/init.d/cerberus disable
}
endef

$(eval $(call BuildPackage,cerberus))
