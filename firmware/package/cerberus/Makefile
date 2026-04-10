include $(TOPDIR)/rules.mk

PKG_NAME:=cerberus
PKG_VERSION:=0.2.0
PKG_RELEASE:=1

PKG_BUILD_DEPENDS:=golang/host

include $(INCLUDE_DIR)/package.mk

define Package/cerberus
  SECTION:=net
  CATEGORY:=Network
  TITLE:=Cerberus Network Control System
  URL:=https://github.com/BackupFaun/cerberus
  DEPENDS:=+aircrack-ng +tcpdump +nmap +hostapd +dnsmasq
  PKGARCH:=all
endef

define Package/cerberus/description
  Open-source offensive network toolkit with a modern web dashboard.
  Provides network recon, MITM, deauth, evil twin, captive portal,
  WPA handshake capture, and full router management.
endef

define Package/cerberus/install
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/cerberus $(1)/usr/bin/

	$(INSTALL_DIR) $(1)/www/cerberus
	$(CP) $(PKG_BUILD_DIR)/www/* $(1)/www/cerberus/

	$(INSTALL_DIR) $(1)/etc/init.d
	$(INSTALL_BIN) ./files/cerberus.init $(1)/etc/init.d/cerberus

	$(INSTALL_DIR) $(1)/etc/cerberus
	$(INSTALL_DIR) $(1)/var/cerberus/captures
endef

define Package/cerberus/postinst
#!/bin/sh
/etc/init.d/cerberus enable
echo "Cerberus installed. Dashboard at http://\$$(uci get network.lan.ipaddr):1471"
endef

$(eval $(call BuildPackage,cerberus))
