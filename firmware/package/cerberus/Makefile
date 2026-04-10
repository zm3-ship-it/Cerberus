include $(TOPDIR)/rules.mk

PKG_NAME:=cerberus
PKG_VERSION:=0.2.0
PKG_RELEASE:=1

include $(INCLUDE_DIR)/package.mk

define Package/cerberus
  SECTION:=net
  CATEGORY:=Network
  TITLE:=Cerberus Network Control System
  URL:=https://github.com/zm3-ship-it/Cerberus
  DEPENDS:=+aircrack-ng +tcpdump +nmap +hostapd +dnsmasq-full
endef

define Package/cerberus/description
  Offensive network toolkit with web dashboard.
endef

define Build/Compile
	@echo "Using pre-built Cerberus binary"
endef

define Package/cerberus/install
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_DIR) $(1)/www/cerberus
	$(INSTALL_DIR) $(1)/etc/init.d
	$(INSTALL_DIR) $(1)/etc/cerberus
	$(INSTALL_DIR) $(1)/var/cerberus/captures

	$(INSTALL_BIN) $(TOPDIR)/cerberus-prebuilt/cerberus $(1)/usr/bin/cerberus
	$(CP) $(TOPDIR)/cerberus-prebuilt/www/* $(1)/www/cerberus/
	$(INSTALL_BIN) $(TOPDIR)/cerberus-prebuilt/cerberus.init $(1)/etc/init.d/cerberus
endef

define Package/cerberus/postinst
#!/bin/sh
[ -z "$${IPKG_INSTROOT}" ] && /etc/init.d/cerberus enable
exit 0
endef

$(eval $(call BuildPackage,cerberus))
