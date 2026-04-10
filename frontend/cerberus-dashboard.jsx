const { useState, useEffect, useRef } = React;

// ═══════════════════ API ═══════════════════
const API = "";
async function api(path, method, body) {
  try {
    const opts = { method: method || "GET", headers: { "Content-Type": "application/json" } };
    if (body) opts.body = JSON.stringify(body);
    const r = await fetch(API + "/api" + path, opts);
    return await r.json();
  } catch (e) { console.warn("API error:", path, e); return null; }
}

// ═══════════════════ HASH ═══════════════════
async function sha256(msg) { let h=0; for(let i=0;i<msg.length;i++){h=((h<<5)-h)+msg.charCodeAt(i);h|=0;} return h.toString(16); }

// ═══════════════════ STORAGE ═══════════════════
async function storageGet(key) {
  try { const v = localStorage.getItem("cerberus_" + key); return v ? JSON.parse(v) : null; }
  catch { return null; }
}
async function storageSet(key, val) {
  try { localStorage.setItem("cerberus_" + key, JSON.stringify(val)); return true; }
  catch { return false; }
}

// ═══════════════════ THEME ═══════════════════
const K={bg:"#050810",sf:"#0a0f19",cd:"#0f1520",ra:"#151d2c",bd:"rgba(0,255,200,0.06)",bx:"rgba(255,255,255,0.035)",cy:"#00ffc8",rd:"#ff4757",am:"#ffc107",gn:"#22c55e",pu:"#a78bfa",bl:"#60a5fa",tx:"#dfe6f0",dm:"rgba(255,255,255,0.28)",f:"'JetBrains Mono','Fira Code','SF Mono',monospace"};

// ═══════════════════ DEVICE TYPES ═══════════════════
const DV={"Dell Inc.":["\uD83D\uDCBB","Laptop"],"Apple Inc.":["\uD83D\uDCF1","iPhone"],"Amazon Tech":["\uD83D\uDCF7","IoT Cam"],"Sony Corp.":["\uD83C\uDFAE","Console"],"Samsung":["\uD83D\uDCF1","Tablet"],"Roku Inc.":["\uD83D\uDCFA","Stream"],"Google LLC":["\uD83D\uDD0A","Speaker"],"TP-Link":["\uD83C\uDF10","IoT"]};
const dv = (v) => DV[v]||["?","Unknown"];

// ═══════════════════ COMPONENTS ═══════════════════
const Dot = ({c=K.cy,p}) => <span style={{position:"relative",display:"inline-flex",width:7,height:7}}><span style={{width:7,height:7,borderRadius:"50%",background:c,transition:"background 0.3s"}}/>{p&&<span style={{position:"absolute",inset:-3,borderRadius:"50%",border:`1.5px solid ${c}`,animation:"cpls 1.5s ease-out infinite"}}/>}</span>;
const Bdg = ({children,c=K.cy}) => <span style={{fontSize:8,fontWeight:700,letterSpacing:1,padding:"2px 7px",borderRadius:3,background:`${c}12`,color:c,whiteSpace:"nowrap",transition:"all 0.3s"}}>{children}</span>;
const Tog = ({on,fn,c=K.cy}) => <div onClick={fn} style={{width:40,height:20,borderRadius:10,background:on?c:"rgba(255,255,255,0.06)",position:"relative",cursor:"pointer",transition:"background 0.3s",flexShrink:0}}><div style={{width:14,height:14,borderRadius:"50%",background:"#fff",position:"absolute",top:3,left:on?23:3,transition:"left 0.25s cubic-bezier(.4,0,.2,1)",boxShadow:on?`0 0 5px ${c}`:"none"}}/></div>;
const Ck = ({on,fn,c=K.cy}) => <div onClick={fn} style={{width:15,height:15,borderRadius:3,border:`1.5px solid ${on?c:"rgba(255,255,255,0.1)"}`,background:on?`${c}18`:"transparent",cursor:"pointer",display:"flex",alignItems:"center",justifyContent:"center",flexShrink:0,transition:"all 0.25s"}}>{on&&<span style={{color:c,fontSize:9,fontWeight:700}}>{"\u2713"}</span>}</div>;
const Lb = ({children}) => <div style={{fontSize:8,fontWeight:700,letterSpacing:3,color:K.dm,marginBottom:8,textTransform:"uppercase"}}>{children}</div>;
const Bx = ({children,s={}}) => <div style={{background:K.cd,borderRadius:8,border:`1px solid ${K.bx}`,padding:16,transition:"all 0.3s ease",...s}}>{children}</div>;
const Sg = ({v}) => { const s=v>-45?4:v>-55?3:v>-65?2:1; const c=s>=3?K.cy:s===2?K.am:K.rd; return <div style={{display:"flex",alignItems:"flex-end",gap:1.5,height:12}}>{[1,2,3,4].map(i => <div key={i} style={{width:2,height:2+i*2.5,borderRadius:1,background:i<=s?c:"rgba(255,255,255,0.05)",transition:"background 0.3s"}}/>)}</div>; };
const Bn = ({children,fn,v="p",dis,sm,sx={}}) => { const vs={p:{bg:`linear-gradient(135deg,${K.cy}18,${K.cy}35)`,c:K.cy,b:`1px solid ${K.cy}30`},d:{bg:`${K.rd}12`,c:K.rd,b:`1px solid ${K.rd}30`},g:{bg:"rgba(255,255,255,0.025)",c:K.dm,b:`1px solid ${K.bx}`}}; return <button onClick={fn} disabled={dis} style={{fontFamily:K.f,fontSize:sm?9:10,fontWeight:600,letterSpacing:1,padding:sm?"5px 10px":"7px 16px",borderRadius:5,cursor:dis?"not-allowed":"pointer",border:vs[v].b,background:vs[v].bg,color:vs[v].c,opacity:dis?0.3:1,transition:"all 0.3s ease",...sx}}>{children}</button>; };
const Rw = ({children,s={}}) => <div style={{display:"flex",alignItems:"center",gap:7,...s}}>{children}</div>;
const Mt = ({t}) => <div style={{padding:24,textAlign:"center",color:K.dm,fontSize:10}}>{t}</div>;
const Sel = ({value,onChange,options,ph}) => <select value={value} onChange={e => onChange(e.target.value)} style={{background:K.ra,color:K.tx,border:`1px solid ${K.bx}`,borderRadius:5,padding:"7px 10px",fontSize:10,fontFamily:K.f,width:"100%",outline:"none",transition:"border 0.3s"}}>{ph&&<option value="">{ph}</option>}{options.map(o => <option key={o.v} value={o.v}>{o.l}</option>)}</select>;

// ═══════════════════ AUTH ═══════════════════
const AuthScreen = ({onAuth}) => {
  const [mode, setMode] = useState("loading");
  const [u, setU] = useState("");
  const [p, setP] = useState("");
  const [p2, setP2] = useState("");
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(false);

  useEffect(() => { (async () => { const c = await storageGet("cerberus-creds"); setMode(c ? "login" : "setup"); })(); }, []);

  const doSetup = async () => {
    if (!u.trim()) { setErr("Username required"); return; }
    if (p.length < 3) { setErr("Password too short"); return; }
    if (p !== p2) { setErr("Passwords don't match"); return; }
    setLoading(true); setErr("");
    const hash = await sha256(p);
    await storageSet("cerberus-creds", { user: u, hash });
    await storageSet("cerberus-session", { active: true, user: u });
    onAuth();
  };

  const doLogin = async () => {
    setLoading(true); setErr("");
    const creds = await storageGet("cerberus-creds");
    if (!creds) { setErr("No account found"); setLoading(false); return; }
    const hash = await sha256(p);
    if (u === creds.user && hash === creds.hash) {
      await storageSet("cerberus-session", { active: true, user: u });
      onAuth();
    } else { setErr("Invalid credentials"); setLoading(false); }
  };

  if (mode === "loading") return <div style={{minHeight:"100vh",background:K.bg,display:"flex",alignItems:"center",justifyContent:"center",fontFamily:K.f,color:K.cy}}>Loading...</div>;

  const isSetup = mode === "setup";
  const [showPw, setShowPw] = useState(false);

  const inp = (val, set, placeholder, type) => {
    const isPw = type === "password";
    return <div style={{position:"relative"}}>
      <input value={val} onChange={e => set(e.target.value)} type={isPw && !showPw ? "password" : "text"} placeholder={placeholder} onKeyDown={e => e.key==="Enter"&&(isSetup?doSetup():doLogin())} style={{background:"rgba(255,255,255,0.04)",border:`1px solid ${err?"rgba(255,75,87,0.4)":"rgba(255,255,255,0.08)"}`,borderRadius:8,padding:"12px 16px",paddingRight:isPw?"44px":"16px",color:K.tx,fontFamily:K.f,fontSize:13,outline:"none",width:"100%",boxSizing:"border-box",transition:"border 0.3s, box-shadow 0.3s"}}/>
      {isPw && <div onClick={() => setShowPw(!showPw)} style={{position:"absolute",right:12,top:"50%",transform:"translateY(-50%)",cursor:"pointer",fontSize:16,opacity:0.4,userSelect:"none",transition:"opacity 0.2s"}}>{showPw ? "\uD83D\uDC41" : "\uD83D\uDC41\u200D\uD83D\uDDE8"}</div>}
    </div>;
  };

  return <div style={{minHeight:"100vh",background:K.bg,display:"flex",alignItems:"center",justifyContent:"center",fontFamily:K.f,position:"relative",overflow:"hidden"}}>
    <style>{`@import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@300;400;500;600;700&display=swap');
      @keyframes gp{0%{transform:translateY(0)}100%{transform:translateY(40px)}}
      @keyframes glow{0%,100%{opacity:.12}50%{opacity:.3}}
      @keyframes fadeUp{from{opacity:0;transform:translateY(12px)}to{opacity:1;transform:translateY(0)}}
      input::placeholder{color:rgba(255,255,255,0.18)}`}</style>
    <div style={{position:"absolute",inset:0,opacity:0.03,backgroundImage:"linear-gradient(rgba(0,255,200,0.5) 1px,transparent 1px),linear-gradient(90deg,rgba(0,255,200,0.5) 1px,transparent 1px)",backgroundSize:"40px 40px",animation:"gp 8s linear infinite"}}/>
    <div style={{position:"absolute",width:400,height:400,borderRadius:"50%",background:`radial-gradient(circle,${K.cy}12 0%,transparent 70%)`,top:"20%",left:"50%",transform:"translateX(-50%)",animation:"glow 4s ease-in-out infinite"}}/>
    <div style={{position:"relative",zIndex:1,width:380,animation:"fadeUp 0.6s ease-out"}}>
      <div style={{textAlign:"center",marginBottom:36}}>
        <div style={{fontSize:32,marginBottom:10}}>{"\uD83D\uDC15"}</div>
        <div style={{fontSize:26,fontWeight:700,letterSpacing:10,color:K.cy,textShadow:`0 0 30px ${K.cy}25`}}>CERBERUS</div>
        <div style={{fontSize:8,letterSpacing:4,color:K.dm,marginTop:5}}>NETWORK CONTROL SYSTEM</div>
      </div>
      <div style={{background:"rgba(15,21,32,0.85)",backdropFilter:"blur(20px)",border:"1px solid rgba(0,255,200,0.07)",borderRadius:14,padding:"28px 24px"}}>
        <div style={{fontSize:11,fontWeight:600,letterSpacing:2,color:K.dm,marginBottom:18,textAlign:"center"}}>{isSetup?"CREATE ACCOUNT":"SIGN IN"}</div>
        <div style={{marginBottom:14}}><div style={{fontSize:8,fontWeight:600,letterSpacing:2,color:K.dm,marginBottom:5}}>USERNAME</div>{inp(u, setU, isSetup?"Choose username":"Enter username")}</div>
        <div style={{marginBottom:isSetup?14:20}}><div style={{fontSize:8,fontWeight:600,letterSpacing:2,color:K.dm,marginBottom:5}}>PASSWORD</div>{inp(p, setP, isSetup?"Choose password":"Enter password", "password")}</div>
        {isSetup && <div style={{marginBottom:20}}><div style={{fontSize:8,fontWeight:600,letterSpacing:2,color:K.dm,marginBottom:5}}>CONFIRM PASSWORD</div>{inp(p2, setP2, "Confirm password", "password")}</div>}
        {err && <div style={{fontSize:9,color:K.rd,marginBottom:12,textAlign:"center",animation:"fadeUp 0.2s ease"}}>{err}</div>}
        <button onClick={isSetup?doSetup:doLogin} disabled={loading} style={{width:"100%",padding:"11px",background:loading?K.ra:`linear-gradient(135deg,${K.cy}20,${K.cy}45)`,border:`1px solid ${K.cy}35`,borderRadius:8,color:loading?K.dm:K.cy,fontFamily:K.f,fontSize:11,fontWeight:600,letterSpacing:3,cursor:loading?"wait":"pointer",transition:"all 0.3s"}}>
          {loading ? "\u27F3 "+(isSetup?"CREATING...":"AUTHENTICATING...") : isSetup?"CREATE ACCOUNT":"LOGIN"}
        </button>
      </div>
      <div style={{textAlign:"center",marginTop:16,fontSize:8,color:"rgba(255,255,255,0.1)"}}>v0.2.0</div>
    </div>
  </div>;
};

// ═══════════════════ AP DROPDOWN ═══════════════════
const APDrop = ({aps,sel,onSel}) => {
  const [open,setOpen] = useState(false);
  const ref = useRef(null);
  useEffect(() => { const h = (e) => { if (ref.current && !ref.current.contains(e.target)) setOpen(false); }; document.addEventListener("mousedown",h); return () => document.removeEventListener("mousedown",h); },[]);
  return <div ref={ref} style={{position:"relative",minWidth:280}}>
    <div onClick={() => { if (aps.length) setOpen(!open); }} style={{display:"flex",alignItems:"center",gap:7,padding:"5px 11px",background:K.ra,border:`1px solid ${sel?K.cy+"30":K.bx}`,borderRadius:5,cursor:aps.length?"pointer":"default",height:30,transition:"border 0.3s"}}>
      {sel ? <><Dot c={K.rd} p/><span style={{fontSize:10,fontWeight:600,color:K.cy}}>{sel.ssid||sel.SSID}</span><Bdg c={(sel.enc||sel.Enc)==="Open"?K.rd:K.gn}>{sel.enc||sel.Enc||"?"}</Bdg><span style={{fontSize:8,color:K.dm}}>CH{sel.ch||sel.channel||sel.Channel}</span><Sg v={sel.signal||sel.Signal||-50}/><span style={{marginLeft:"auto",fontSize:8,color:K.dm}}>{"\u25BE"}</span></> : <span style={{fontSize:9,color:K.dm}}>{aps.length?"Select target AP...":"Run recon"}</span>}
    </div>
    {open && <div style={{position:"absolute",top:"calc(100% + 2px)",left:0,right:0,background:K.cd,border:`1px solid ${K.bd}`,borderRadius:5,zIndex:200,maxHeight:240,overflowY:"auto",boxShadow:"0 10px 36px rgba(0,0,0,0.6)",animation:"fadeUp 0.2s ease"}}>
      {aps.map((a,i) => { const ssid=a.ssid||a.SSID||"Hidden"; const bssid=a.bssid||a.BSSID||""; const enc=a.enc||a.Enc||"?"; const ch=a.ch||a.channel||a.Channel||0; const sig=a.signal||a.Signal||-50;
        return <div key={bssid||i} onClick={() => { onSel(a); setOpen(false); }} style={{display:"flex",alignItems:"center",gap:7,padding:"8px 11px",cursor:"pointer",borderBottom:`1px solid ${K.bx}`,transition:"background 0.2s"}}>
        <Sg v={sig}/><div style={{flex:1}}><div style={{fontSize:10,fontWeight:600,color:K.tx}}>{ssid}</div><div style={{fontSize:7,color:K.dm,marginTop:1}}>{bssid}</div></div>
        <Bdg c={enc==="Open"?K.rd:K.gn}>{enc}</Bdg><span style={{fontSize:8,color:K.dm}}>CH{ch}</span>
      </div>; })}
    </div>}
  </div>;
};

// ═══════════════════ NAV ═══════════════════
const NAV=[
  {id:"dash",ic:"\u25C9",l:"Overview"},{id:"recon",ic:"\uD83D\uDCE1",l:"Recon"},
  {id:"targets",ic:"\uD83C\uDFAF",l:"Targets"},{id:"mitm",ic:"\uD83D\uDD00",l:"MITM"},
  {id:"twin",ic:"\uD83D\uDC7B",l:"Evil Twin"},{id:"portal",ic:"\uD83C\uDFA3",l:"Captive"},
  {id:"log",ic:"\uD83D\uDCCB",l:"Logging"},{id:"cfg",ic:"\u2699\uFE0F",l:"Settings"},
];

// ═══════════════════ MAIN APP ═══════════════════
function App() {
  const [auth, setAuth] = useState(false);
  const [checking, setChecking] = useState(true);
  useEffect(() => { (async () => { const s = await storageGet("cerberus-session"); if (s && s.active) setAuth(true); setChecking(false); })(); }, []);
  if (checking) return <div style={{minHeight:"100vh",background:K.bg,display:"flex",alignItems:"center",justifyContent:"center",fontFamily:K.f,color:K.cy,fontSize:12}}>
    <style>{`@import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@300;400;500;600;700&display=swap');`}</style>Loading...</div>;
  if (!auth) return <AuthScreen onAuth={() => setAuth(true)} />;
  return <Dashboard onLogout={async () => { await storageSet("cerberus-session", { active: false }); setAuth(false); }} />;
}

function Dashboard({ onLogout }) {
  const [pg, setPg] = useState("dash");
  const [aps, setAps] = useState([]);
  const [tAP, setTAP] = useState(null);
  const [cli, setCli] = useState([]);
  const [rcing, setRcing] = useState(false);
  const [scing, setScing] = useState(false);
  const [sel, setSel] = useState(new Set());
  const [mS, setMS] = useState(new Set());
  const [dS, setDS] = useState(new Set());
  const [dns, setDns] = useState([]);
  const [et, setEt] = useState({on:false,ssid:"",ch:"6"});
  const [cap, setCap] = useState({on:false,tpl:"google"});
  const [doh, setDoh] = useState([
    {ip:"1.1.1.1",n:"Cloudflare",on:true},{ip:"1.0.0.1",n:"Cloudflare Alt",on:true},
    {ip:"8.8.8.8",n:"Google",on:true},{ip:"8.8.4.4",n:"Google Alt",on:true},
    {ip:"9.9.9.9",n:"Quad9",on:true},{ip:"149.112.112.112",n:"Quad9 Alt",on:false},
  ]);
  const [prb, setPrb] = useState([]);
  const [mCfg, setMCfg] = useState({arp:true,dns:true,ssl:false,spoof:false});
  const [lf, setLf] = useState("all");
  const [ls, setLs] = useState("");
  const [hsState, setHsState] = useState("idle");
  const [hsAP, setHsAP] = useState("");
  const [adRoles, setAdRoles] = useState({scan:"wlan1",attack:"wlan2",upstream:"wlan0"});
  const [adapters, setAdapters] = useState([]);
  const [toast, setToast] = useState("");
  const ML = 5000;
  const pollRef = useRef(null);

  const showToast = (msg) => { setToast(msg); setTimeout(() => setToast(""), 3000); };

  const tSel = (m) => setSel(p => { const n=new Set(p); n.has(m)?n.delete(m):n.add(m); return n; });
  const aAll = () => setSel(cli.length===sel.size?new Set():new Set(cli.map(c => c.mac||c.MAC)));
  
  // ═══ REAL API CALLS ═══
  
  const doRecon = async () => {
    setRcing(true); setAps([]); setTAP(null); setCli([]); setPrb([]);
    showToast("Scanning networks...");
    await api("/scan", "POST");
    // Poll for results
    let attempts = 0;
    const iv = setInterval(async () => {
      const nets = await api("/networks");
      const probes = await api("/probes");
      if (nets && Array.isArray(nets)) setAps(nets);
      if (probes && Array.isArray(probes)) setPrb(probes);
      attempts++;
      if (attempts > 15) { clearInterval(iv); setRcing(false); showToast("Recon complete: " + (nets?nets.length:0) + " APs found"); }
    }, 2000);
    pollRef.current = iv;
  };

  const doScan = async () => {
    if (!tAP) return;
    setScing(true); setCli([]); setSel(new Set());
    showToast("Scanning clients...");
    await api("/scan", "POST");
    let attempts = 0;
    const iv = setInterval(async () => {
      const clients = await api("/clients");
      if (clients && Array.isArray(clients)) setCli(clients);
      attempts++;
      if (attempts > 10) { clearInterval(iv); setScing(false); showToast("Found " + (clients?clients.length:0) + " clients"); }
    }, 2000);
  };

  const toggleMitm = async (mac, ip) => {
    if (mS.has(mac)) {
      await api("/mitm/stop", "POST", { mac });
      setMS(p => { const n=new Set(p); n.delete(mac); return n; });
      showToast("MITM stopped on " + mac);
    } else {
      await api("/mitm/start", "POST", { mac, ip });
      setMS(p => { const n=new Set(p); n.add(mac); return n; });
      showToast("MITM started on " + mac);
    }
  };

  const toggleDeauth = async (mac, bssid) => {
    if (dS.has(mac)) {
      await api("/deauth/stop", "POST", { mac });
      setDS(p => { const n=new Set(p); n.delete(mac); return n; });
      showToast("Deauth stopped");
    } else {
      await api("/deauth/start", "POST", { mac, bssid: bssid || (tAP&&tAP.bssid)||"" });
      setDS(p => { const n=new Set(p); n.add(mac); return n; });
      showToast("Deauth started");
    }
  };

  // Poll DNS when MITM is active
  useEffect(() => {
    if (mS.size === 0) return;
    const iv = setInterval(async () => {
      const log = await api("/mitm/dns");
      if (log && Array.isArray(log)) {
        setDns(prev => {
          const combined = [...prev, ...log.slice(prev.length)];
          return combined.length > ML ? combined.slice(combined.length - ML) : combined;
        });
      }
    }, 2000);
    return () => clearInterval(iv);
  }, [mS.size]);

  // Load adapters on mount
  useEffect(() => {
    (async () => {
      const a = await api("/adapters");
      if (a && Array.isArray(a)) setAdapters(a);
    })();
  }, []);

  // Handshake
  const doHandshake = async () => {
    if (!hsAP) return;
    setHsState("listening");
    showToast("Capturing handshake...");
    const ap = aps.find(a => (a.ssid||a.SSID) === hsAP);
    await api("/handshake/start", "POST", { bssid: ap?.bssid||ap?.BSSID||"", ssid: hsAP, channel: String(ap?.ch||ap?.channel||ap?.Channel||6) });
    const iv = setInterval(async () => {
      const st = await api("/handshake/status");
      if (st) {
        setHsState(st.state || st.State || "listening");
        if (st.state === "captured" || st.state === "failed" || st.State === "captured" || st.State === "failed") {
          clearInterval(iv);
          showToast(st.state === "captured" ? "Handshake captured!" : "Capture failed");
        }
      }
    }, 2000);
  };

  const downloadCap = async () => {
    const st = await api("/handshake/status");
    if (st && (st.cap_file || st.CapFile)) {
      window.open(API + "/api/handshake/download/" + (st.cap_file || st.CapFile));
    } else { showToast("No capture file available"); }
  };

  const fDns = dns.filter(e => {
    const status = e.s || e.status || e.Status || "";
    const domain = e.d || e.domain || e.Domain || "";
    if (lf==="blocked" && status!=="blocked") return false;
    if (lf==="passed" && status!=="pass" && status!=="passed") return false;
    if (ls && !domain.toLowerCase().includes(ls.toLowerCase())) return false;
    return true;
  });

  const getField = (obj, ...keys) => { for (const k of keys) { if (obj[k] !== undefined) return obj[k]; } return ""; };

  // ═══ PAGES ═══

  const Dash = () => <div>
    <div style={{display:"grid",gridTemplateColumns:"repeat(4,1fr)",gap:8,marginBottom:14}}>
      {[{l:"CLIENTS",v:cli.length,c:K.cy,i:"\uD83D\uDCE1"},{l:"MITM",v:mS.size,c:K.gn,i:"\uD83D\uDD00"},{l:"DEAUTH",v:dS.size,c:K.rd,i:"\uD83D\uDC80"},{l:"DNS",v:dns.length,c:K.pu,i:"\uD83D\uDCCB"}].map(s => <Bx key={s.l} s={{textAlign:"center",padding:12}}><div style={{fontSize:18,marginBottom:2}}>{s.i}</div><div style={{fontSize:22,fontWeight:700,color:s.c,transition:"all 0.3s"}}>{s.v}</div><div style={{fontSize:7,letterSpacing:2,color:K.dm,marginTop:2}}>{s.l}</div></Bx>)}
    </div>
    <div style={{display:"grid",gridTemplateColumns:"1fr 1fr",gap:8,marginBottom:8}}>
      <Bx s={{padding:14}}><Lb>Target</Lb>{tAP ? <div><div style={{fontSize:13,fontWeight:700,color:K.cy}}>{tAP.ssid||tAP.SSID}</div><div style={{fontSize:9,color:K.dm,marginTop:2}}>{tAP.bssid||tAP.BSSID} | CH{tAP.ch||tAP.channel||tAP.Channel}</div></div> : <span style={{color:K.dm,fontSize:10}}>No target</span>}</Bx>
      <Bx s={{padding:14}}><Lb>Active</Lb>
        {mS.size>0&&<Rw s={{marginBottom:3}}><Dot c={K.cy} p/><span style={{fontSize:10}}>MITM: {mS.size}</span></Rw>}
        {dS.size>0&&<Rw s={{marginBottom:3}}><Dot c={K.rd} p/><span style={{fontSize:10}}>Deauth: {dS.size}</span></Rw>}
        {et.on&&<Rw s={{marginBottom:3}}><Dot c={K.pu} p/><span style={{fontSize:10}}>Twin: {et.ssid}</span></Rw>}
        {cap.on&&<Rw><Dot c={K.am} p/><span style={{fontSize:10}}>Portal</span></Rw>}
        {mS.size===0&&dS.size===0&&!et.on&&!cap.on&&<span style={{color:K.dm,fontSize:10}}>None</span>}
      </Bx>
    </div>
    <Bx s={{padding:14,marginBottom:8}}><Lb>Connected Devices</Lb>
      <div style={{maxHeight:200,overflowY:"auto"}}>{cli.length>0 ? cli.map((c,i) => { const mac=getField(c,"mac","MAC"); const host=getField(c,"host","hostname","Hostname")||mac; const ip=getField(c,"ip","IP"); const vnd=getField(c,"vnd","vendor","Vendor"); const sig=getField(c,"sig","signal","Signal")||-50; const [ic,ty]=dv(vnd);
        return <div key={mac||i} style={{display:"grid",gridTemplateColumns:"22px 1.4fr 80px 55px 55px",gap:5,padding:"6px 0",borderBottom:`1px solid ${K.bx}`,alignItems:"center",fontSize:10}}>
          <span>{ic}</span><div><span style={{fontWeight:600}}>{host}</span><div style={{fontSize:8,color:K.dm}}>{ty}</div></div>
          <span style={{color:K.dm,fontSize:9}}>{ip}</span><Sg v={sig}/>
          <Bdg c={K.gn}>LIVE</Bdg>
        </div>; }) : <Mt t="Scan to see devices"/>}</div>
    </Bx>
    <Bx s={{padding:14}}><Lb>Adapters</Lb>
      {(adapters.length > 0 ? adapters : [{id:"wlan0",chip:"Built-in"},{id:"wlan1",chip:"USB"},{id:"wlan2",chip:"USB"}]).map((a,i) => <Rw key={a.id||i} s={{justifyContent:"space-between",padding:"5px 0",borderBottom:`1px solid ${K.bx}`}}>
        <Rw><span style={{fontSize:11,fontWeight:600,color:K.cy}}>{a.id||a.ID}</span><span style={{fontSize:8,color:K.dm}}>{a.chip||a.Chip||""}</span></Rw>
        <Bdg c={adRoles.scan===(a.id||a.ID)?K.cy:adRoles.attack===(a.id||a.ID)?K.rd:K.gn}>{adRoles.scan===(a.id||a.ID)?"SCAN":adRoles.attack===(a.id||a.ID)?"ATTACK":"UPSTREAM"}</Bdg>
      </Rw>)}
    </Bx>
  </div>;

  const Recon = () => <div>
    <Rw s={{justifyContent:"space-between",marginBottom:10}}><Lb>Recon</Lb><Bn fn={doRecon} dis={rcing}>{rcing?"\u27F3 SCANNING...":"\u26A1 RECON"}</Bn></Rw>
    <Bx s={{marginBottom:8}}><Lb>Access Points ({aps.length})</Lb>
      {aps.map((a,i) => { const ssid=getField(a,"ssid","SSID")||"Hidden"; const bssid=getField(a,"bssid","BSSID"); const ch=getField(a,"ch","channel","Channel"); const enc=getField(a,"enc","Enc")||"?"; const sig=getField(a,"signal","Signal")||-50;
        return <div key={bssid||i} onClick={() => setTAP(a)} style={{display:"grid",gridTemplateColumns:"1.6fr 1.1fr 42px 48px 55px",gap:5,padding:"7px 0",borderBottom:`1px solid ${K.bx}`,alignItems:"center",cursor:"pointer",background:tAP===a?`${K.cy}05`:"transparent",transition:"background 0.2s"}}>
        <span style={{fontSize:10,fontWeight:600,color:tAP===a?K.cy:K.tx}}>{ssid}</span>
        <span style={{fontSize:8,color:K.dm}}>{bssid}</span><span style={{fontSize:8}}>CH{ch}</span>
        <Bdg c={enc==="Open"?K.rd:K.gn}>{enc}</Bdg><Rw><Sg v={sig}/></Rw>
      </div>; })}
      {aps.length===0&&!rcing&&<Mt t="Hit RECON to discover networks"/>}
      {rcing&&aps.length===0&&<Mt t="\u27F3 Scanning..."/>}
    </Bx>
    <Bx s={{marginBottom:8,borderColor:hsState==="captured"?`${K.gn}25`:K.bx}}><Lb>WPA Handshake Capture</Lb>
      <div style={{marginBottom:8}}><Sel value={hsAP} onChange={setHsAP} options={aps.filter(a => (a.enc||a.Enc)!=="Open").map(a => ({v:a.ssid||a.SSID,l:`${a.ssid||a.SSID} [${a.enc||a.Enc}]`}))} ph="Select target AP..."/></div>
      <Rw s={{gap:6}}>
        <Bn fn={doHandshake} dis={!hsAP||hsState==="listening"||hsState==="deauthing"} sm>{hsState==="idle"?"Capture Handshake":hsState==="listening"?"\u27F3 Listening...":hsState==="deauthing"?"\u27F3 Deauthing...":hsState==="captured"?"\u2713 Captured!":"Capture"}</Bn>
        {hsState==="captured"&&<Bn fn={downloadCap} sm v="g">{"\u2B07"} Download .cap</Bn>}
        {hsState!=="idle"&&hsState!=="captured"&&<Bn fn={() => { api("/handshake/stop","POST"); setHsState("idle"); }} sm v="d">Cancel</Bn>}
      </Rw>
      {hsState!=="idle"&&<div style={{marginTop:8,padding:"6px 10px",borderRadius:4,background:hsState==="captured"?`${K.gn}08`:`${K.am}08`,border:`1px solid ${hsState==="captured"?K.gn:K.am}18`,fontSize:9,color:hsState==="captured"?K.gn:K.am}}>
        {hsState==="listening"&&<><Dot c={K.am} p/> Listening for EAPOL frames...</>}
        {hsState==="deauthing"&&<><Dot c={K.rd} p/> Sending deauth to force reconnect...</>}
        {hsState==="captured"&&<><Dot c={K.gn} p/> 4-way handshake captured!</>}
        {hsState==="failed"&&<><Dot c={K.rd}/> Capture failed — try again</>}
      </div>}
    </Bx>
    <div style={{display:"grid",gridTemplateColumns:"1fr 1fr",gap:8}}>
      <Bx><Lb>Clients ({cli.length})</Lb><div style={{maxHeight:200,overflowY:"auto"}}>{cli.map((c,i) => { const [ic,ty]=dv(getField(c,"vnd","vendor","Vendor")); return <Rw key={i} s={{padding:"5px 0",borderBottom:`1px solid ${K.bx}`}}><span style={{fontSize:12}}>{ic}</span><div style={{flex:1}}><div style={{fontSize:10,fontWeight:600}}>{getField(c,"host","hostname","Hostname")||getField(c,"mac","MAC")}</div><div style={{fontSize:8,color:K.dm}}>{getField(c,"ip","IP")}</div></div><Sg v={getField(c,"sig","signal","Signal")||-50}/></Rw>; })}{cli.length===0&&<Mt t="Select AP then scan"/>}</div></Bx>
      <Bx><Lb>Probes ({prb.length})</Lb><div style={{maxHeight:200,overflowY:"auto"}}>{prb.map((p,i) => <div key={i} style={{display:"flex",justifyContent:"space-between",padding:"5px 0",borderBottom:`1px solid ${K.bx}`,fontSize:10}}><span style={{color:K.dm}}>{p.t||p.time||p.Time||""}</span><span style={{fontWeight:500}}>{p.ssid||p.SSID||""}</span></div>)}{prb.length===0&&<Mt t="Appear during scan"/>}</div></Bx>
    </div>
  </div>;

  const Targets = () => <div>
    <Rw s={{justifyContent:"space-between",marginBottom:10,flexWrap:"wrap",gap:5}}>
      <Lb>Targets{tAP?` \u2014 ${tAP.ssid||tAP.SSID}`:""}</Lb>
      <Rw s={{gap:4}}>
        <Bn fn={doScan} dis={scing||!tAP} sm>{scing?"\u27F3":"\u26A1"} SCAN</Bn>
        <Bn fn={() => sel.forEach(m => { const c=cli.find(x=>(x.mac||x.MAC)===m); if(c&&!mS.has(m)) toggleMitm(m,c.ip||c.IP); })} sm v="g">MITM</Bn>
        <Bn fn={() => sel.forEach(m => { if(!dS.has(m)) toggleDeauth(m); })} sm v="d">Deauth</Bn>
      </Rw>
    </Rw>
    {!tAP ? <Bx><Mt t="Select target AP from top bar"/></Bx> :
    <Bx s={{padding:0,overflow:"hidden"}}>
      <div style={{overflowX:"auto"}}>
        <div style={{display:"grid",gridTemplateColumns:"28px 20px 1.5fr 80px 115px 40px 45px 42px 42px",gap:4,padding:"8px 10px",fontSize:7,fontWeight:700,letterSpacing:2,color:K.dm,borderBottom:`1px solid ${K.bx}`,alignItems:"center",minWidth:580}}>
          <Ck on={sel.size===cli.length&&cli.length>0} fn={aAll}/><span/><span>DEVICE</span><span>IP</span><span>MAC</span><span>SIG</span><span>STAT</span><span>MITM</span><span>DEAU</span>
        </div>
        <div style={{maxHeight:340,overflowY:"auto"}}>
          {cli.map((c,x) => { const mac=getField(c,"mac","MAC"); const ip=getField(c,"ip","IP"); const host=getField(c,"host","hostname","Hostname")||mac; const vnd=getField(c,"vnd","vendor","Vendor"); const sig=getField(c,"sig","signal","Signal")||-50; const [ic,ty]=dv(vnd); const s=sel.has(mac); const m=mS.has(mac); const d=dS.has(mac);
            return <div key={mac||x} style={{display:"grid",gridTemplateColumns:"28px 20px 1.5fr 80px 115px 40px 45px 42px 42px",gap:4,padding:"7px 10px",borderBottom:`1px solid ${K.bx}`,alignItems:"center",background:s?`${K.cy}04`:"transparent",minWidth:580,transition:"background 0.2s"}}>
              <Ck on={s} fn={() => tSel(mac)}/><span style={{fontSize:11}}>{ic}</span>
              <div><div style={{fontSize:10,fontWeight:600,color:s?K.cy:K.tx}}>{host}</div><div style={{fontSize:7,color:K.dm}}>{ty}</div></div>
              <span style={{fontSize:9}}>{ip}</span><span style={{fontSize:7,color:K.dm}}>{mac}</span>
              <Sg v={sig}/><Bdg c={K.gn}>LIVE</Bdg>
              <Tog on={m} fn={() => toggleMitm(mac,ip)} c={K.cy}/>
              <Tog on={d} fn={() => toggleDeauth(mac)} c={K.rd}/>
            </div>; })}
          {cli.length===0&&<Mt t="Hit SCAN"/>}
        </div>
      </div>
    </Bx>}
    {dS.size>0&&<Bx s={{marginTop:7,borderColor:`${K.rd}18`,padding:10}}><Rw><Dot c={K.rd} p/><span style={{fontSize:9,color:K.rd}}>Deauth active on {dS.size} target{dS.size>1?"s":""}</span></Rw></Bx>}
  </div>;

  const MitmPg = () => { const act=cli.filter(c => mS.has(c.mac||c.MAC)); return <div>
    <Lb>MITM</Lb>
    <div style={{display:"grid",gridTemplateColumns:"1fr 1fr",gap:8}}>
      <Bx><Lb>Add Target</Lb><Sel value="" onChange={v => { const c=cli.find(x=>(x.mac||x.MAC)===v); if(c) toggleMitm(v,c.ip||c.IP); }} options={cli.map(c => ({v:c.mac||c.MAC,l:`${dv(c.vnd||c.vendor||c.Vendor||"")[0]} ${c.host||c.hostname||c.Hostname||c.mac||c.MAC} \u2014 ${c.ip||c.IP}`}))} ph="+ Add..."/>
        <div style={{marginTop:8}}>{act.map((c,i) => { const mac=c.mac||c.MAC; const host=c.host||c.hostname||c.Hostname||mac; return <Rw key={mac} s={{justifyContent:"space-between",padding:"5px 0",borderBottom:`1px solid ${K.bx}`}}><Rw><Dot c={K.cy} p/><span style={{fontSize:10,fontWeight:600}}>{host}</span></Rw><Bn fn={() => toggleMitm(mac,c.ip||c.IP)} sm v="d">Stop</Bn></Rw>; })}{act.length===0&&<Mt t="Select targets"/>}</div>
      </Bx>
      <Bx><Lb>Config</Lb>
        {[{k:"arp",l:"ARP Spoofing",d:"Intercept via ARP"},{k:"dns",l:"DNS Logging",d:"Log DNS"},{k:"ssl",l:"SSL Strip",d:"Downgrade HTTPS"},{k:"spoof",l:"DNS Spoofing",d:"Redirect domains"}].map((a,i) => <Rw key={a.k} s={{justifyContent:"space-between",padding:"5px 0",borderBottom:`1px solid ${K.bx}`}}><div><div style={{fontSize:10,fontWeight:600}}>{a.l}</div><div style={{fontSize:8,color:K.dm}}>{a.d}</div></div><Tog on={mCfg[a.k]} fn={() => setMCfg(p => ({...p,[a.k]:!p[a.k]}))}/></Rw>)}
      </Bx>
    </div>
  </div>; };

  const TwinPg = () => <div>
    <Lb>Evil Twin</Lb>
    <div style={{display:"grid",gridTemplateColumns:"1fr 1fr",gap:8}}>
      <Bx><Lb>Clone AP</Lb>
        <div style={{marginBottom:7}}><div style={{fontSize:8,color:K.dm,marginBottom:3}}>SSID</div><Sel value={et.ssid} onChange={v => setEt(p => ({...p,ssid:v}))} options={aps.map(a => ({v:a.ssid||a.SSID,l:`${a.ssid||a.SSID} [CH${a.ch||a.channel||a.Channel}]`}))} ph="Select..."/></div>
        <div style={{marginBottom:7}}><div style={{fontSize:8,color:K.dm,marginBottom:3}}>Channel</div><Sel value={et.ch} onChange={v => setEt(p => ({...p,ch:v}))} options={[1,6,11,36,40,44].map(c => ({v:String(c),l:`CH ${c}`}))}/></div>
        <Bn fn={async () => { if(et.on){await api("/eviltwin/stop","POST");setEt(p=>({...p,on:false}));showToast("Twin stopped");}else{await api("/eviltwin/start","POST",{SSID:et.ssid,Channel:et.ch,Iface:"wlan0"});setEt(p=>({...p,on:true}));showToast("Twin launched: "+et.ssid);} }} v={et.on?"d":"p"} sx={{width:"100%"}}>{et.on?"\u25FC STOP":"\u25B6 LAUNCH"}</Bn>
        {et.on&&et.ssid&&<div style={{marginTop:7,padding:"6px 9px",borderRadius:4,background:`${K.cy}06`,border:`1px solid ${K.cy}15`,fontSize:9,color:K.cy}}><Dot c={K.cy} p/> "{et.ssid}" CH{et.ch}</div>}
      </Bx>
      <Bx><Lb>Probes</Lb>
        {prb.length>0 ? [...new Set(prb.map(p => p.ssid||p.SSID))].map((s,i) => <Rw key={s} s={{justifyContent:"space-between",padding:"4px 0",borderBottom:`1px solid ${K.bx}`}}><span style={{fontSize:10}}>{s}</span><Bdg c={K.pu}>PROBE</Bdg></Rw>) : <Mt t="Run recon"/>}
        <div style={{marginTop:10}}><Lb>Force Reconnect</Lb><Sel value="" onChange={v => { if(v) toggleDeauth(v); }} options={cli.map(c => ({v:c.mac||c.MAC,l:`${dv(c.vnd||c.vendor||c.Vendor||"")[0]} ${c.host||c.hostname||c.Hostname||""}`}))} ph="Deauth target..."/></div>
      </Bx>
    </div>
  </div>;

  const PortalPg = () => <div>
    <Lb>Captive Portal</Lb>
    <div style={{display:"grid",gridTemplateColumns:"1fr 1fr",gap:8}}>
      <Bx><Lb>Config</Lb>
        <div style={{marginBottom:7}}><div style={{fontSize:8,color:K.dm,marginBottom:3}}>Template</div><Sel value={cap.tpl} onChange={v => setCap(p => ({...p,tpl:v}))} options={[{v:"google",l:"Google Sign-In"},{v:"facebook",l:"Facebook Login"},{v:"wifi",l:"WiFi Terms"},{v:"hotel",l:"Hotel Portal"},{v:"custom",l:"Custom HTML"}]}/></div>
        <div style={{fontSize:9,color:K.dm,marginBottom:10}}>Runs on Evil Twin AP</div>
        <Bn fn={async () => { if(cap.on){await api("/captive/stop","POST");setCap(p=>({...p,on:false}));showToast("Portal stopped");}else{await api("/captive/start","POST",{Template:cap.tpl});setCap(p=>({...p,on:true}));showToast("Portal launched");} }} v={cap.on?"d":"p"} sx={{width:"100%"}}>{cap.on?"\u25FC STOP":"\u25B6 LAUNCH"}</Bn>
      </Bx>
      <Bx><Lb>Captured Creds</Lb><Mt t={cap.on?"Waiting...":"Portal not active"}/></Bx>
    </div>
  </div>;

  const LogPg = () => <div>
    <Rw s={{justifyContent:"space-between",marginBottom:8,flexWrap:"wrap",gap:5}}>
      <Rw s={{gap:4}}><Lb>DNS Log ({fDns.length}/{dns.length})</Lb>{dns.length>=ML&&<Bdg c={K.am}>MAX</Bdg>}</Rw>
      <Rw s={{gap:3}}>
        {["all","passed","blocked"].map(f => <button key={f} onClick={() => setLf(f)} style={{fontFamily:K.f,fontSize:8,fontWeight:600,letterSpacing:1,padding:"3px 8px",borderRadius:3,cursor:"pointer",border:`1px solid ${lf===f?(f==="blocked"?K.rd:f==="passed"?K.cy:K.bl)+"35":K.bx}`,background:lf===f?(f==="blocked"?K.rd:f==="passed"?K.cy:K.bl)+"10":"transparent",color:lf===f?(f==="blocked"?K.rd:f==="passed"?K.cy:K.bl):K.dm,textTransform:"uppercase",transition:"all 0.25s"}}>{f}</button>)}
        <Bn fn={() => { setDns([]); }} sm v="g">Clear</Bn>
      </Rw>
    </Rw>
    <div style={{marginBottom:6}}><input value={ls} onChange={e => setLs(e.target.value)} placeholder="Filter domain..." style={{background:K.ra,color:K.tx,border:`1px solid ${K.bx}`,borderRadius:4,padding:"6px 9px",fontSize:9,fontFamily:K.f,width:"100%",outline:"none",boxSizing:"border-box",transition:"border 0.3s"}}/></div>
    <Bx s={{padding:0}}>
      <div style={{display:"grid",gridTemplateColumns:"55px 1fr 42px",gap:5,padding:"7px 12px",fontSize:7,fontWeight:700,letterSpacing:2,color:K.dm,borderBottom:`1px solid ${K.bx}`}}><span>TIME</span><span>DOMAIN</span><span>STATUS</span></div>
      <div style={{maxHeight:440,overflowY:"auto"}}>
        {fDns.map((e,i) => { const time=e.t||e.time||e.Time||""; const domain=e.d||e.domain||e.Domain||""; const status=e.s||e.status||e.Status||"";
          return <div key={i} style={{display:"grid",gridTemplateColumns:"55px 1fr 42px",gap:5,padding:"5px 12px",fontSize:10,borderBottom:`1px solid ${K.bx}`}}><span style={{color:K.dm}}>{time}</span><span style={{fontWeight:500}}>{domain}</span><Bdg c={status==="blocked"?K.rd:K.cy}>{status==="blocked"?"BLOCK":"PASS"}</Bdg></div>; })}
        {fDns.length===0&&<Mt t={dns.length===0?"No DNS logged — start MITM to capture":"No matches"}/>}
      </div>
    </Bx>
  </div>;

  const CfgPg = () => <div>
    <Lb>Settings</Lb>
    <div style={{display:"grid",gridTemplateColumns:"1fr 1fr",gap:8}}>
      <Bx><Lb>Adapter Roles</Lb>
        {(adapters.length>0?adapters:[{id:"wlan0",chip:"Built-in",mac:"",modes:[],band:""},{id:"wlan1",chip:"USB",mac:"",modes:[],band:""},{id:"wlan2",chip:"USB",mac:"",modes:[],band:""}]).map((a,i) => { const id=a.id||a.ID; const role=adRoles.scan===id?"scan":adRoles.attack===id?"attack":"upstream";
          return <div key={id||i} style={{padding:"8px 0",borderBottom:`1px solid ${K.bx}`}}>
            <Rw s={{justifyContent:"space-between",marginBottom:4}}>
              <div><span style={{fontSize:11,fontWeight:700,color:K.cy}}>{id}</span><span style={{fontSize:8,color:K.dm,marginLeft:6}}>{a.chip||a.Chip||""}</span></div>
              <Bdg c={role==="scan"?K.cy:role==="attack"?K.rd:K.gn}>{role.toUpperCase()}</Bdg>
            </Rw>
            <div style={{fontSize:8,color:K.dm}}>{a.mac||a.MAC||""} {a.band||a.Band||""}</div>
            <div style={{marginTop:4}}><Sel value={role} onChange={async v => {
              const nr={...adRoles}; const old=Object.entries(nr).find(([,vid]) => vid===id);
              const swapTo=Object.entries(nr).find(([r]) => r===v);
              if (old&&swapTo) { nr[old[0]]=swapTo[1]; }
              nr[v]=id; setAdRoles(nr);
              await api("/adapters/role","POST",{Adapter:id,Role:v});
              showToast(`${id} → ${v}`);
            }} options={[{v:"scan",l:"Scanning"},{v:"attack",l:"Attacking"},{v:"upstream",l:"Upstream"}]}/></div>
          </div>; })}
      </Bx>
      <Bx><Lb>DoH / DoT Blocking</Lb>
        {doh.map((s,i) => <Rw key={s.ip} s={{justifyContent:"space-between",padding:"4px 0",borderBottom:`1px solid ${K.bx}`}}><div><span style={{fontSize:10,fontWeight:600}}>{s.n}</span><span style={{fontSize:8,color:K.dm,marginLeft:5}}>{s.ip}</span></div><Tog on={s.on} fn={() => { const n=[...doh]; n[i]={...n[i],on:!n[i].on}; setDoh(n); }} c={K.rd}/></Rw>)}
        <div style={{marginTop:14,borderTop:`1px solid ${K.bx}`,paddingTop:10}}>
          <Bn fn={onLogout} v="d" sx={{width:"100%"}}>Logout</Bn>
        </div>
      </Bx>
    </div>
  </div>;

  const pages={dash:Dash,recon:Recon,targets:Targets,mitm:MitmPg,twin:TwinPg,portal:PortalPg,log:LogPg,cfg:CfgPg};
  const Pg=pages[pg];

  return <div style={{display:"flex",flexDirection:"column",height:"100vh",background:K.bg,color:K.tx,fontFamily:K.f}}>
    <style>{`@import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@300;400;500;600;700&display=swap');
      @keyframes cpls{0%{transform:scale(1);opacity:.4}100%{transform:scale(2.2);opacity:0}}
      @keyframes fadeUp{from{opacity:0;transform:translateY(8px)}to{opacity:1;transform:translateY(0)}}
      @keyframes slideIn{from{opacity:0;transform:translateX(-8px)}to{opacity:1;transform:translateX(0)}}
      @keyframes toastIn{from{opacity:0;transform:translateY(20px)}to{opacity:1;transform:translateY(0)}}
      @keyframes toastOut{from{opacity:1}to{opacity:0;transform:translateY(10px)}}
      @keyframes shimmer{0%{background-position:-200% 0}100%{background-position:200% 0}}
      *{scrollbar-width:thin;scrollbar-color:${K.ra} transparent;transition-timing-function:cubic-bezier(.4,0,.2,1)}
      ::selection{background:${K.cy}28}
      input::placeholder{color:rgba(255,255,255,0.16)}
      div,span,button{transition-duration:0.15s}`}</style>

    {/* Toast */}
    {toast && <div style={{position:"fixed",bottom:24,left:"50%",transform:"translateX(-50%)",background:K.cd,border:`1px solid ${K.cy}30`,borderRadius:8,padding:"10px 20px",fontSize:11,color:K.cy,fontFamily:K.f,zIndex:9999,boxShadow:`0 4px 20px rgba(0,0,0,0.5), 0 0 15px ${K.cy}15`,animation:"toastIn 0.3s ease",letterSpacing:1}}>{toast}</div>}

    {/* TOP BAR */}
    <div style={{display:"flex",alignItems:"center",gap:8,padding:"8px 14px",background:K.sf,borderBottom:`1px solid ${K.bx}`,flexShrink:0}}>
      <div style={{marginRight:4}}><div style={{fontSize:13,fontWeight:700,letterSpacing:5,color:K.cy}}>CERBERUS</div><div style={{fontSize:6,letterSpacing:2,color:K.dm}}>v0.2.0</div></div>
      <div style={{width:1,height:22,background:K.bx}}/>
      <APDrop aps={aps} sel={tAP} onSel={setTAP}/>
      <Bn fn={doRecon} dis={rcing} sm>{rcing?"\u27F3":"\uD83D\uDCE1"} RECON</Bn>
      <Bn fn={doScan} dis={scing||!tAP} sm>{scing?"\u27F3":"\u26A1"} SCAN</Bn>
      <div style={{flex:1}}/>
      <Rw><Dot c={K.gn} p/><span style={{fontSize:8,color:K.dm}}>Online</span></Rw>
      <div onClick={onLogout} style={{cursor:"pointer",fontSize:9,color:K.dm,padding:"4px 10px",borderRadius:4,border:`1px solid ${K.bx}`,transition:"all 0.25s"}} onMouseOver={e=>e.target.style.color=K.rd} onMouseOut={e=>e.target.style.color=K.dm}>{"\u23FB"}</div>
    </div>

    <div style={{display:"flex",flex:1,overflow:"hidden"}}>
      {/* SIDEBAR */}
      <div style={{width:140,background:K.sf,borderRight:`1px solid ${K.bx}`,padding:"8px 0",display:"flex",flexDirection:"column",flexShrink:0}}>
        {NAV.map((n,i) => { const a=pg===n.id; return <div key={n.id} onClick={() => setPg(n.id)} style={{display:"flex",alignItems:"center",gap:7,padding:"7px 12px",cursor:"pointer",fontSize:10,fontWeight:a?600:400,color:a?K.cy:K.dm,background:a?`${K.cy}05`:"transparent",borderLeft:a?`2px solid ${K.cy}`:"2px solid transparent",transition:"all 0.25s"}}><span style={{fontSize:11,width:15,textAlign:"center"}}>{n.ic}</span>{n.l}</div>; })}
      </div>

      {/* MAIN */}
      <div style={{flex:1,padding:14,overflowY:"auto"}}><Pg/></div>
    </div>
  </div>;
}
