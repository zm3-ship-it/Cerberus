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

type Capture struct {
	ID        string `json:"id"`
	BSSID     string `json:"bssid"`
	SSID      string `json:"ssid"`
	Channel   int    `json:"channel"`
	Iface     string `json:"iface"`
	FilePath  string `json:"file_path"`
	HasShake  bool   `json:"has_handshake"`
	Running   bool   `json:"running"`
	StartedAt int64  `json:"started_at"`
	Packets   int    `json:"packets"`
}

type Manager struct {
	captures map[string]*captureState
	saveDir  string
	mu       sync.RWMutex
}

type captureState struct {
	capture Capture
	cmd     *exec.Cmd
	stop    chan struct{}
}

func NewManager(saveDir string) *Manager {
	os.MkdirAll(saveDir, 0755)
	return &Manager{
		captures: make(map[string]*captureState),
		saveDir:  saveDir,
	}
}

// Start begins capturing on a specific channel targeting a BSSID
func (m *Manager) Start(id, bssid, ssid, iface string, channel int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.captures[id]; exists {
		return fmt.Errorf("capture %s already running", id)
	}

	if iface == "" {
		return fmt.Errorf("no monitor interface specified")
	}
	if bssid == "" {
		return fmt.Errorf("no BSSID specified")
	}

	outFile := filepath.Join(m.saveDir, fmt.Sprintf("cerberus-%s", time.Now().Format("20060102-150405")))

	// Lock to target channel
	if channel > 0 {
		exec.Command("iw", "dev", iface, "set", "channel", fmt.Sprintf("%d", channel)).Run()
	}

	// No --output-format flag — let airodump use its default (cap)
	cmd := exec.Command("airodump-ng",
		"--bssid", bssid,
		"--channel", fmt.Sprintf("%d", channel),
		"--write", outFile,
		iface,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	state := &captureState{
		capture: Capture{
			ID:        id,
			BSSID:     bssid,
			SSID:      ssid,
			Channel:   channel,
			Iface:     iface,
			FilePath:  outFile + "-01.cap",
			Running:   true,
			StartedAt: time.Now().Unix(),
		},
		cmd:  cmd,
		stop: make(chan struct{}),
	}

	log.Printf("handshake: starting capture for %s (%s) ch%d iface:%s", bssid, ssid, channel, iface)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("airodump-ng failed to start: %w", err)
	}

	m.captures[id] = state
	go m.monitorHandshake(id, state)

	return nil
}

func (m *Manager) Stop(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, ok := m.captures[id]
	if !ok {
		return fmt.Errorf("capture %s not found", id)
	}

	if state.capture.Running {
		close(state.stop)
		if state.cmd.Process != nil {
			state.cmd.Process.Kill()
			state.cmd.Wait()
		}
		state.capture.Running = false
	}

	return nil
}

func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, state := range m.captures {
		if state.capture.Running {
			close(state.stop)
			if state.cmd.Process != nil {
				state.cmd.Process.Kill()
				state.cmd.Wait()
			}
			state.capture.Running = false
		}
	}
}

func (m *Manager) List() []Capture {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]Capture, 0, len(m.captures))
	for _, s := range m.captures {
		result = append(result, s.capture)
	}
	return result
}

func (m *Manager) Get(id string) (*Capture, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if s, ok := m.captures[id]; ok {
		c := s.capture
		return &c, true
	}
	return nil, false
}

// GetCapFile returns the path to the .cap file for download
func (m *Manager) GetCapFile(id string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, ok := m.captures[id]
	if !ok {
		return "", fmt.Errorf("capture %s not found", id)
	}

	if _, err := os.Stat(state.capture.FilePath); err != nil {
		return "", fmt.Errorf("cap file not found: %w", err)
	}

	return state.capture.FilePath, nil
}

func (m *Manager) monitorHandshake(id string, state *captureState) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-state.stop:
			return
		case <-ticker.C:
			capFile := state.capture.FilePath
			if _, err := os.Stat(capFile); err != nil {
				continue
			}

			// Use aircrack-ng to check for handshake
			cmd := exec.Command("aircrack-ng", capFile)
			out, err := cmd.CombinedOutput()
			if err != nil {
				continue
			}

			outStr := strings.ToLower(string(out))
			// Check multiple patterns for different aircrack-ng versions
			hasHandshake := strings.Contains(outStr, "1 handshake") ||
				strings.Contains(outStr, "handshake") && strings.Contains(outStr, "wpa") ||
				strings.Contains(outStr, "eapol") && !strings.Contains(outStr, "0 handshake")

			if hasHandshake {
				m.mu.Lock()
				if s, ok := m.captures[id]; ok {
					s.capture.HasShake = true
					log.Printf("handshake: CAPTURED for %s (%s)!", state.capture.BSSID, state.capture.SSID)
				}
				m.mu.Unlock()
				return
			}

			// Update packet count from file size
			if fi, err := os.Stat(capFile); err == nil {
				m.mu.Lock()
				if s, ok := m.captures[id]; ok {
					s.capture.Packets = int(fi.Size() / 100)
				}
				m.mu.Unlock()
			}
		}
	}
}

func (m *Manager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok := m.captures[id]; ok {
		if s.capture.Running {
			close(s.stop)
			if s.cmd.Process != nil {
				s.cmd.Process.Kill()
			}
		}
		delete(m.captures, id)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
