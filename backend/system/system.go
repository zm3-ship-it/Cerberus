package system

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Info struct {
	Hostname    string  `json:"hostname"`
	Uptime      string  `json:"uptime"`
	UptimeSecs  int64   `json:"uptime_secs"`
	Firmware    string  `json:"firmware"`
	Kernel      string  `json:"kernel"`
	Arch        string  `json:"arch"`
	CPUModel    string  `json:"cpu_model"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemTotal    int64   `json:"mem_total_kb"`
	MemUsed     int64   `json:"mem_used_kb"`
	MemPercent  float64 `json:"mem_percent"`
	DiskTotal   int64   `json:"disk_total_kb"`
	DiskUsed    int64   `json:"disk_used_kb"`
	DiskPercent float64 `json:"disk_percent"`
	LocalTime   string  `json:"local_time"`
}

func GetInfo() Info {
	info := Info{
		Arch:      runtime.GOARCH,
		LocalTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	// Hostname
	if h, err := os.Hostname(); err == nil {
		info.Hostname = h
	}

	// Uptime
	if data, err := os.ReadFile("/proc/uptime"); err == nil {
		parts := strings.Fields(string(data))
		if len(parts) > 0 {
			if secs, err := strconv.ParseFloat(parts[0], 64); err == nil {
				info.UptimeSecs = int64(secs)
				d := int(secs) / 86400
				h := (int(secs) % 86400) / 3600
				m := (int(secs) % 3600) / 60
				if d > 0 {
					info.Uptime = fmt.Sprintf("%dd %dh %dm", d, h, m)
				} else if h > 0 {
					info.Uptime = fmt.Sprintf("%dh %dm", h, m)
				} else {
					info.Uptime = fmt.Sprintf("%dm", m)
				}
			}
		}
	}

	// Firmware version
	if data, err := os.ReadFile("/etc/openwrt_release"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "DISTRIB_DESCRIPTION=") {
				info.Firmware = strings.Trim(strings.TrimPrefix(line, "DISTRIB_DESCRIPTION="), "'\"")
			}
		}
	}
	if info.Firmware == "" {
		if data, err := os.ReadFile("/etc/os-release"); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(line, "PRETTY_NAME=") {
					info.Firmware = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
				}
			}
		}
	}

	// Kernel
	if out, err := exec.Command("uname", "-r").Output(); err == nil {
		info.Kernel = strings.TrimSpace(string(out))
	}

	// CPU model
	if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "model name") || strings.HasPrefix(line, "cpu model") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					info.CPUModel = strings.TrimSpace(parts[1])
					break
				}
			}
		}
	}
	if info.CPUModel == "" {
		info.CPUModel = "MediaTek MT7981B"
	}

	// CPU usage (simple calc from /proc/stat)
	info.CPUUsage = getCPUUsage()

	// Memory
	if data, err := os.ReadFile("/proc/meminfo"); err == nil {
		var total, free, buffers, cached int64
		for _, line := range strings.Split(string(data), "\n") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				val, _ := strconv.ParseInt(fields[1], 10, 64)
				switch fields[0] {
				case "MemTotal:":
					total = val
				case "MemFree:":
					free = val
				case "Buffers:":
					buffers = val
				case "Cached:":
					cached = val
				}
			}
		}
		info.MemTotal = total
		info.MemUsed = total - free - buffers - cached
		if total > 0 {
			info.MemPercent = float64(info.MemUsed) / float64(total) * 100
		}
	}

	// Disk
	if out, err := exec.Command("df", "/").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) >= 2 {
			fields := strings.Fields(lines[1])
			if len(fields) >= 5 {
				info.DiskTotal, _ = strconv.ParseInt(fields[1], 10, 64)
				info.DiskUsed, _ = strconv.ParseInt(fields[2], 10, 64)
				if info.DiskTotal > 0 {
					info.DiskPercent = float64(info.DiskUsed) / float64(info.DiskTotal) * 100
				}
			}
		}
	}

	return info
}

func getCPUUsage() float64 {
	read := func() (idle, total int64) {
		data, err := os.ReadFile("/proc/stat")
		if err != nil {
			return
		}
		lines := strings.Split(string(data), "\n")
		if len(lines) == 0 {
			return
		}
		fields := strings.Fields(lines[0])
		if len(fields) < 5 {
			return
		}
		var sum int64
		for i := 1; i < len(fields); i++ {
			v, _ := strconv.ParseInt(fields[i], 10, 64)
			sum += v
			if i == 4 {
				idle = v
			}
		}
		return idle, sum
	}

	idle1, total1 := read()
	time.Sleep(200 * time.Millisecond)
	idle2, total2 := read()

	idleDelta := float64(idle2 - idle1)
	totalDelta := float64(total2 - total1)
	if totalDelta == 0 {
		return 0
	}
	return (1.0 - idleDelta/totalDelta) * 100
}

// Reboot the router
func Reboot() error {
	log.Println("[!] Rebooting system...")
	return exec.Command("reboot").Start()
}

// FlashFirmware handles firmware upload and sysupgrade
func FlashFirmware(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only", 405)
		return
	}

	// Parse multipart form (max 64MB)
	r.ParseMultipartForm(64 << 20)
	file, header, err := r.FormFile("firmware")
	if err != nil {
		jsonError(w, "No firmware file provided", 400)
		return
	}
	defer file.Close()

	log.Printf("[*] Firmware upload: %s (%d bytes)", header.Filename, header.Size)

	// Validate file
	if !strings.HasSuffix(header.Filename, ".bin") && !strings.HasSuffix(header.Filename, ".img") {
		jsonError(w, "File must be .bin or .img", 400)
		return
	}

	// Save to /tmp
	tmpPath := "/tmp/firmware_upload.bin"
	out, err := os.Create(tmpPath)
	if err != nil {
		jsonError(w, "Failed to save firmware", 500)
		return
	}

	written, err := io.Copy(out, file)
	out.Close()
	if err != nil {
		jsonError(w, "Failed to write firmware", 500)
		return
	}

	log.Printf("[*] Firmware saved: %s (%d bytes)", tmpPath, written)

	// Verify with sysupgrade --test
	testCmd := exec.Command("sysupgrade", "--test", tmpPath)
	if testOut, err := testCmd.CombinedOutput(); err != nil {
		os.Remove(tmpPath)
		jsonError(w, fmt.Sprintf("Firmware validation failed: %s", strings.TrimSpace(string(testOut))), 400)
		return
	}

	log.Println("[*] Firmware validated, ready to flash")

	// Check if keepSettings was requested
	keepSettings := r.FormValue("keep_settings") == "true"

	// Return success — actual flash happens after response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "ready",
		"filename": header.Filename,
		"size":     written,
		"message":  "Firmware validated. Flashing will begin in 3 seconds. Router will reboot.",
	})

	// Flash in background after response is sent
	go func() {
		time.Sleep(3 * time.Second)
		args := []string{"-v"}
		if !keepSettings {
			args = append(args, "-n")
		}
		args = append(args, tmpPath)
		log.Printf("[!] FLASHING FIRMWARE: sysupgrade %s", strings.Join(args, " "))
		exec.Command("sysupgrade", args...).Run()
	}()
}

// SetHostname changes the system hostname via UCI
func SetHostname(hostname string) error {
	if err := exec.Command("uci", "set", fmt.Sprintf("system.@system[0].hostname=%s", hostname)).Run(); err != nil {
		return err
	}
	exec.Command("uci", "commit", "system").Run()
	exec.Command("/etc/init.d/system", "reload").Run()
	return nil
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
