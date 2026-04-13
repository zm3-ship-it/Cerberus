# Unbricking Your Router

Your router won't boot, the lights are wrong, and you're staring at a $70 paperweight. Don't panic. Almost every "bricked" router can be recovered. This guide covers the most common methods, from easiest to most involved.

---

## First: Is It Actually Bricked?

Before you start, verify the router is genuinely dead and not just slow.

**Not bricked (just wait):**
- Firmware upgrade in progress — can take up to 5 minutes. Do NOT unplug it.
- LED is slowly blinking — it's booting. Give it 3 full minutes.
- You changed the IP and can't find it — try `192.168.1.1`, `192.168.8.1`, `10.0.0.1`, or run `arp -a` after plugging in via ethernet.

**Actually bricked:**
- Power light is on but nothing else happens after 5 minutes
- LEDs are in an unusual repeating pattern (fast blink loop)
- No response to ping on any IP
- Boot loops (LEDs cycle the same pattern endlessly)

If it's genuinely dead, keep reading.

---

## Method 1: Failsafe Mode (Easiest)

Almost all OpenWrt-compatible routers have a built-in failsafe mode that survives bad flashes. This is your first attempt.

### Steps

1. **Connect** your computer to the router's LAN port with an ethernet cable. Disable Wi-Fi on your computer.

2. **Set a static IP** on your computer:
   - IP: `192.168.1.2`
   - Subnet: `255.255.255.0`
   - Gateway: `192.168.1.1`

3. **Power cycle** the router (unplug, wait 5 seconds, plug back in).

4. **Watch the LEDs.** When you see the LED start blinking rapidly (usually ~5 seconds after power on), **press and hold the reset button** for 5-10 seconds. Timing varies by model.

   Common triggers by model:
   - **GL.iNet GL-MT3000**: Hold reset button while powering on, release after 10 seconds
   - **GL.iNet GL-MT1300**: Same as above
   - **TP-Link**: Press reset exactly when the gear/system LED starts blinking
   - **Netgear**: Hold reset during power on until power LED blinks amber

5. **Wait 60 seconds** after releasing the button.

6. **Try to connect:**
   ```bash
   ping 192.168.1.1
   ```
   If you get responses, it's in failsafe.

7. **SSH in and reflash:**
   ```bash
   ssh root@192.168.1.1

   # Once connected, mount the filesystem read-write
   mount_root

   # Now you can fix things or flash a fresh firmware
   # First, get your firmware file onto the router
   # From your computer (in another terminal):
   scp openwrt-*-sysupgrade.bin root@192.168.1.1:/tmp/

   # Back in the router SSH session:
   sysupgrade -n /tmp/openwrt-*-sysupgrade.bin
   ```

8. Wait for the reboot (2-3 minutes). Router should come back clean.

---

## Method 2: TFTP Recovery

If failsafe doesn't work, most routers have a hardware bootloader (U-Boot) that can load firmware via TFTP. The bootloader lives in a protected flash region that's almost impossible to corrupt with a bad firmware flash.

### Setup TFTP Server

**Linux:**
```bash
sudo apt-get install -y tftpd-hpa

# Place your firmware file in the TFTP directory
sudo cp openwrt-*-sysupgrade.bin /srv/tftp/
# Some routers expect a specific filename:
# GL-MT3000: openwrt-mediatek-filogic-glinet_gl-mt3000-squashfs-sysupgrade.bin
# Check your router's wiki page for the expected filename

sudo systemctl restart tftpd-hpa
```

**macOS:**
```bash
# Built-in TFTP
sudo launchctl load -F /System/Library/LaunchDaemons/tftp.plist
cp openwrt-*-sysupgrade.bin /private/tftpboot/
```

**Windows:**
Download Tftpd64 from [https://pjo2.github.io/tftpd64/](https://pjo2.github.io/tftpd64/). Point it at the folder containing your firmware file.

### Flash via TFTP

1. **Set your computer's static IP** to `192.168.1.2` (or whatever the bootloader expects — see model-specific notes below).

2. **Connect** ethernet from your computer to the router's LAN port.

3. **Start a continuous ping** so you can see when the bootloader is alive:
   ```bash
   ping -t 192.168.1.1    # Windows
   ping 192.168.1.1        # Linux/macOS
   ```

4. **Power on the router while holding the reset button.** Hold for 10-30 seconds depending on model.

5. **Watch for ping responses.** When you see replies, the bootloader is up and requesting firmware via TFTP. Your TFTP server should show transfer activity.

6. **Wait.** The transfer and flash takes 2-5 minutes. Do NOT touch anything. The router will reboot automatically when done.

### Model-Specific TFTP Notes

**GL.iNet GL-MT3000 / MT1300:**
- Bootloader IP: `192.168.1.1`
- Your IP: `192.168.1.2`
- Hold reset button during power on
- U-Boot looks for firmware automatically via TFTP

**TP-Link routers:**
- Bootloader IP: `192.168.0.1` (note: .0, not .1)
- Your IP: `192.168.0.2`
- Expected filename varies — check the OpenWrt wiki for your exact model
- Some models need the file renamed to a specific name like `tp_recovery.bin`

**Netgear routers:**
- Use Netgear's "nmrpflash" utility instead of raw TFTP
- Download: [https://github.com/jclehner/nmrpflash](https://github.com/jclehner/nmrpflash)
- Usage: `nmrpflash -i eth0 -f openwrt-*-factory.img`

---

## Method 3: Serial Console (Last Resort)

If both failsafe and TFTP fail, you need to talk directly to the bootloader via serial UART. This requires opening the router case and soldering (or clipping) wires to debug pads on the PCB.

### What You Need

- USB-to-TTL serial adapter (FTDI or CP2102 — about $5 on Amazon)
  - **IMPORTANT**: Must be 3.3V logic level. A 5V adapter can permanently damage the board.
- 3x jumper wires (female-to-female usually)
- Soldering iron OR test clips (for pads without headers)
- Terminal program: `minicom`, `screen`, or PuTTY

### Find the UART Pads

Open the router case. Look for 4 pads (sometimes 3) labeled:
- **TX** (transmit from router)
- **RX** (receive to router)
- **GND** (ground)
- **VCC** (power — do NOT connect this)

If they're not labeled, search your exact model on the OpenWrt wiki or on [https://fccid.io](https://fccid.io) — FCC teardown photos usually show the PCB clearly.

### Wire It Up

| Router Pad | Serial Adapter Pin |
|---|---|
| TX | RX |
| RX | TX |
| GND | GND |
| VCC | **Leave disconnected** |

Yes, TX goes to RX and vice versa. They're crossed.

### Connect

```bash
# Linux
sudo minicom -D /dev/ttyUSB0 -b 115200

# macOS
screen /dev/tty.usbserial-* 115200

# Settings: 115200 baud, 8N1, no flow control
```

### Interrupt the Bootloader

1. Power on the router.
2. Watch the serial output — you'll see U-Boot messages scrolling.
3. **Press a key** (usually Enter, Space, or a specific key combo) when you see a message like:
   ```
   Hit any key to stop autoboot: 3
   ```
4. You're now in the U-Boot command prompt.

### Flash from U-Boot

**Option A — TFTP from U-Boot:**
```
# Set network addresses
setenv ipaddr 192.168.1.1
setenv serverip 192.168.1.2

# Download firmware to RAM
tftpboot 0x46000000 openwrt-sysupgrade.bin

# Write to flash (command varies by platform)
# For NAND flash (most modern routers):
nand erase 0x0 0x4000000
nand write 0x46000000 0x0 ${filesize}

# For NOR flash:
sf probe
sf erase 0x0 +${filesize}
sf write 0x46000000 0x0 ${filesize}

# Reboot
reset
```

**Option B — USB from U-Boot (if supported):**
```
usb start
fatload usb 0:1 0x46000000 openwrt-sysupgrade.bin
# Then write to flash as above
```

The exact memory addresses and flash commands depend on your specific hardware. Check the OpenWrt wiki page for your exact model — they usually document the full U-Boot recovery procedure.

---

## Method 4: SPI Flash Programmer (Nuclear Option)

If U-Boot itself is corrupted (very rare — usually only happens with a bad bootloader flash), you need to reprogram the flash chip directly.

### What You Need

- CH341A SPI programmer ($5 on Amazon)
- SOIC8 test clip (to grab the flash chip without desoldering)
- `flashrom` software

### Steps

1. Identify the flash chip on the PCB (usually a small 8-pin SOIC chip near the CPU).
2. Clip the CH341A onto the flash chip. **Router must be powered OFF and unplugged.**
3. Read the current (corrupted) flash as a backup:
   ```bash
   sudo flashrom -p ch341a_spi -r backup.bin
   ```
4. Write the full firmware image:
   ```bash
   sudo flashrom -p ch341a_spi -w openwrt-factory.bin
   ```
5. Disconnect the programmer, reassemble, power on.

---

## Prevention: Don't Brick It Again

- **Always use sysupgrade images**, not factory images, when upgrading from OpenWrt
- **Don't flash from Wi-Fi** — always use a wired ethernet connection
- **Don't interrupt a flash in progress** — no unplugging, no power outages
- **Keep a known-good firmware image** on your computer at all times
- **Write down the TFTP recovery procedure for your specific model** before you need it

---

## Model-Specific Recovery Links

| Router | OpenWrt Wiki Recovery Page |
|---|---|
| GL.iNet GL-MT3000 | [https://openwrt.org/toh/gl.inet/gl-mt3000](https://openwrt.org/toh/gl.inet/gl-mt3000) |
| GL.iNet GL-MT1300 | [https://openwrt.org/toh/gl.inet/gl-mt1300](https://openwrt.org/toh/gl.inet/gl-mt1300) |
| TP-Link Archer C7 | [https://openwrt.org/toh/tp-link/archer_c7](https://openwrt.org/toh/tp-link/archer_c7) |
| Netgear R7800 | [https://openwrt.org/toh/netgear/r7800](https://openwrt.org/toh/netgear/r7800) |

For any router: search `openwrt.org/toh/BRAND/MODEL` — nearly every supported device has documented recovery steps.
