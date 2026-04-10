import { useState, useEffect, useRef } from "react";

// ═══════════════════ DATA ═══════════════════
const APS=[{ssid:"HomeNet-5G",bssid:"A0:B1:C2:D3:E4:01",ch:36,enc:"WPA2",signal:-38},{ssid:"Linksys_Guest",bssid:"F0:E1:D2:C3:B4:02",ch:6,enc:"WPA2",signal:-52},{ssid:"xfinitywifi",bssid:"11:22:33:44:55:03",ch:1,enc:"Open",signal:-61},{ssid:"NETGEAR-2.4",bssid:"AA:CC:EE:11:33:04",ch:11,enc:"WPA3",signal:-70},{ssid:"TP-Link_8842",bssid:"CC:DD:EE:FF:00:05",ch:6,enc:"WPA2",signal:-74}];
const CLI_MAP={"HomeNet-5G":[{ip:"192.168.1.12",mac:"A4:83:E7:2F:11:B0",host:"Jake-Laptop",vnd:"Dell Inc.",sig:-42,bw:"2.4 MB/s",seen:"now"},{ip:"192.168.1.15",mac:"F0:18:98:44:CC:01",host:"iPhone-Sarah",vnd:"Apple Inc.",sig:-58,bw:"340 KB/s",seen:"now"},{ip:"192.168.1.23",mac:"DC:A6:32:AA:BB:CC",host:"Ring-Doorbell",vnd:"Amazon Tech",sig:-61,bw:"18 KB/s",seen:"2m"},{ip:"192.168.1.31",mac:"B8:27:EB:12:34:56",host:"PS5-Living",vnd:"Sony Corp.",sig:-39,bw:"12.1 MB/s",seen:"now"},{ip:"192.168.1.44",mac:"00:1A:2B:3C:4D:5E",host:"Galaxy-Tab-A7",vnd:"Samsung",sig:-55,bw:"890 KB/s",seen:"now"},{ip:"192.168.1.50",mac:"E8:6F:38:11:22:33",host:"Roku-Bedroom",vnd:"Roku Inc.",sig:-67,bw:"4.2 MB/s",seen:"5m"}],"Linksys_Guest":[{ip:"192.168.2.10",mac:"AA:11:22:33:44:55",host:"MacBook-Pro",vnd:"Apple Inc.",sig:-54,bw:"1.8 MB/s",seen:"now"},{ip:"192.168.2.15",mac:"BB:22:33:44:55:66",host:"Pixel-7",vnd:"Google LLC",sig:-60,bw:"450 KB/s",seen:"now"}],"xfinitywifi":[{ip:"10.0.0.12",mac:"CC:33:44:55:66:77",host:"iPhone-12",vnd:"Apple Inc.",sig:-63,bw:"200 KB/s",seen:"now"}]};
const PRB=[{mac:"A4:83:E7:2F:11:B0",ssid:"Starbucks_WiFi",t:"14:31:02"},{mac:"F0:18:98:44:CC:01",ssid:"ATT-Passpoint",t:"14:31:08"},{mac:"00:1A:2B:3C:4D:5E",ssid:"McDonalds_Free",t:"14:31:15"},{mac:"A4:83:E7:2F:11:B0",ssid:"SchoolWiFi",t:"14:31:22"}];
const DNS_Q=[{t:"14:32:07",d:"www.youtube.com",s:"pass"},{t:"14:32:08",d:"i.ytimg.com",s:"pass"},{t:"14:32:09",d:"fonts.googleapis.com",s:"pass"},{t:"14:32:11",d:"ads.doubleclick.net",s:"blocked"},{t:"14:32:14",d:"www.reddit.com",s:"pass"},{t:"14:32:16",d:"tracker.analytics.com",s:"blocked"},{t:"14:32:19",d:"discord.com",s:"pass"},{t:"14:32:22",d:"cdn.discord.com",s:"pass"},{t:"14:32:25",d:"api.snapchat.com",s:"pass"},{t:"14:32:28",d:"graph.instagram.com",s:"pass"},{t:"14:32:30",d:"telemetry.microsoft.com",s:"blocked"},{t:"14:32:33",d:"www.tiktok.com",s:"pass"},{t:"14:32:36",d:"api.twitter.com",s:"pass"},{t:"14:32:39",d:"static.xx.fbcdn.net",s:"pass"},{t:"14:32:41",d:"pixel.adsafeprotected.com",s:"blocked"},{t:"14:32:44",d:"play.google.com",s:"pass"},{t:"14:32:47",d:"ocsp.pki.goog",s:"pass"},{t:"14:32:50",d:"beacons.gvt2.com",s:"blocked"}];
const ADAPTERS=[{id:"wlan0",chip:"MediaTek MT7981B",mac:"AA:BB:CC:11:22:33",modes:["Managed","Monitor"],band:"2.4/5 GHz"},{id:"wlan1",chip:"Realtek RTL8812AU",mac:"DD:EE:FF:44:55:66",modes:["Managed","Monitor","Injection"],band:"2.4/5 GHz"},{id:"wlan2",chip:"Atheros AR9271",mac:"11:22:33:AA:BB:CC",modes:["Managed","Monitor","Injection"],band:"2.4 GHz"}];
const DOH_INIT=[{ip:"1.1.1.1",n:"Cloudflare",on:true},{ip:"1.0.0.1",n:"Cloudflare Alt",on:true},{ip:"8.8.8.8",n:"Google",on:true},{ip:"8.8.4.4",n:"Google Alt",on:true},{ip:"9.9.9.9",n:"Quad9",on:true},{ip:"149.112.112.112",n:"Quad9 Alt",on:false},{ip:"208.67.222.222",n:"OpenDNS",on:false},{ip:"94.140.14.14",n:"AdGuard",on:false}];
const DV={"Dell Inc.":["\uD83D\uDCBB","Laptop"],"Apple Inc.":["\uD83D\uDCF1","iPhone"],"Amazon Tech":["\uD83D\uDCF7","IoT Cam"],"Sony Corp.":["\uD83C\uDFAE","Console"],"Samsung":["\uD83D\uDCF1","Tablet"],"Roku Inc.":["\uD83D\uDCFA","Stream"],"Google LLC":["\uD83D\uDD0A","Speaker"],"TP-Link":["\uD83C\uDF10","IoT"]};
const dv = (v) => DV[v]||["?","Unknown"];

// ═══════════════════ THEME ═══════════════════
const K={bg:"#050810",sf:"#0a0f19",cd:"#0f1520",ra:"#151d2c",bd:"rgba(0,255,200,0.06)",bx:"rgba(255,255,255,0.035)",cy:"#00ffc8",rd:"#ff4757",am:"#ffc107",gn:"#22c55e",pu:"#a78bfa",bl:"#60a5fa",tx:"#dfe6f0",dm:"rgba(255,255,255,0.28)",f:"'JetBrains Mono','Fira Code','SF Mono',monospace"};

// ═══════════════════ HASH UTIL ═══════════════════
async function sha256(msg) {
  const buf = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(msg));
  return Array.from(new Uint8Array(buf)).map(b => b.toString(16).padStart(2,"0")).join("");
}

// ═══════════════════ STORAGE HELPERS ═══════════════════
async function storageGet(key) {
  try { const r = await window.storage.get(key); return r ? JSON.parse(r.value) : null; }
  catch { return null; }
}
async function storageSet(key, val) {
  try { await window.storage.set(key, JSON.stringify(val)); return true; }
  catch { return false; }
}

// ═══════════════════ MICRO COMPONENTS ═══════════════════
const Dot = ({c=K.cy,p}) => <span style={{position:"relative",display:"inline-flex",width:7,height:7}}><span style={{width:7,height:7,borderRadius:"50%",background:c}}/>{p&&<span style={{position:"absolute",inset:-3,borderRadius:"50%",border:`1.5px solid ${c}`,animation:"cpls 1.5s ease-out infinite"}}/>}</span>;
const Bdg = ({children,c=K.cy}) => <span style={{fontSize:8,fontWeight:700,letterSpacing:1,padding:"2px 7px",borderRadius:3,background:`${c}12`,color:c,whiteSpace:"nowrap"}}>{children}</span>;
const Tog = ({on,fn,c=K.cy}) => <div onClick={fn} style={{width:40,height:20,borderRadius:10,background:on?c:"rgba(255,255,255,0.06)",position:"relative",cursor:"pointer",transition:"background 0.3s",flexShrink:0}}><div style={{width:14,height:14,borderRadius:"50%",background:"#fff",position:"absolute",top:3,left:on?23:3,transition:"left 0.2s cubic-bezier(.4,0,.2,1)",boxShadow:on?`0 0 5px ${c}`:"none"}}/></div>;
const Ck = ({on,fn,c=K.cy}) => <div onClick={fn} style={{width:15,height:15,borderRadius:3,border:`1.5px solid ${on?c:"rgba(255,255,255,0.1)"}`,background:on?`${c}18`:"transparent",cursor:"pointer",display:"flex",alignItems:"center",justifyContent:"center",flexShrink:0}}>{on&&<span style={{color:c,fontSize:9,fontWeight:700}}>{"\u2713"}</span>}</div>;
const Lb = ({children}) => <div style={{fontSize:8,fontWeight:700,letterSpacing:3,color:K.dm,marginBottom:8,textTransform:"uppercase"}}>{children}</div>;
const Bx = ({children,s={}}) => <div style={{background:K.cd,borderRadius:8,border:`1px solid ${K.bx}`,padding:16,...s}}>{children}</div>;
const Sg = ({v}) => { const s=v>-45?4:v>-55?3:v>-65?2:1; const c=s>=3?K.cy:s===2?K.am:K.rd; return <div style={{display:"flex",alignItems:"flex-end",gap:1.5,height:12}}>{[1,2,3,4].map(i => <div key={i} style={{width:2,height:2+i*2.5,borderRadius:1,background:i<=s?c:"rgba(255,255,255,0.05)"}}/>)}</div>; };
const Bn = ({children,fn,v="p",dis,sm,sx={}}) => { const vs={p:{bg:`linear-gradient(135deg,${K.cy}18,${K.cy}35)`,c:K.cy,b:`1px solid ${K.cy}30`},d:{bg:`${K.rd}12`,c:K.rd,b:`1px solid ${K.rd}30`},g:{bg:"rgba(255,255,255,0.025)",c:K.dm,b:`1px solid ${K.bx}`}}; return <button onClick={fn} disabled={dis} style={{fontFamily:K.f,fontSize:sm?9:10,fontWeight:600,letterSpacing:1,padding:sm?"5px 10px":"7px 16px",borderRadius:5,cursor:dis?"not-allowed":"pointer",border:vs[v].b,background:vs[v].bg,color:vs[v].c,opacity:dis?0.3:1,transition:"all 0.25s",...sx}}>{children}</button>; };
const Rw = ({children,s={}}) => <div style={{display:"flex",alignItems:"center",gap:7,...s}}>{children}</div>;
const Mt = ({t}) => <div style={{padding:24,textAlign:"center",color:K.dm,fontSize:10}}>{t}</div>;
const Sel = ({value,onChange,options,ph}) => <select value={value} onChange={e => onChange(e.target.value)} style={{background:K.ra,color:K.tx,border:`1px solid ${K.bx}`,borderRadius:5,padding:"7px 10px",fontSize:10,fontFamily:K.f,width:"100%",outline:"none"}}>{ph&&<option value="">{ph}</option>}{options.map(o => <option key={o.v} value={o.v}>{o.l}</option>)}</select>;

// ═══════════════════ AUTH SCREEN ═══════════════════
const AuthScreen = ({onAuth}) => {
  const [mode, setMode] = useState("loading"); // loading, setup, login
  const [u, setU] = useState("");
  const [p, setP] = useState("");
  const [p2, setP2] = useState("");
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(false);

  // On mount, check if account exists in persistent storage
  useEffect(() => {
    (async () => {
      const creds = await storageGet("cerberus-creds");
      setMode(creds ? "login" : "setup");
    })();
  }, []);

  const doSetup = async () => {
    if (!u.trim()) { setErr("Username required"); return; }
    if (p.length < 3) { setErr("Password too short"); return; }
    if (p !== p2) { setErr("Passwords don't match"); return; }
    setLoading(true); setErr("");
    const hash = await sha256(p);
    const ok = await storageSet("cerberus-creds", { user: u, hash });
    if (ok) {
      await storageSet("cerberus-session", { active: true, user: u });
      onAuth();
    } else {
      setErr("Storage error"); setLoading(false);
    }
  };

  const doLogin = async () => {
    setLoading(true); setErr("");
    const creds = await storageGet("cerberus-creds");
    if (!creds) { setErr("No account found"); setLoading(false); return; }
    const hash = await sha256(p);
    if (u === creds.user && hash === creds.hash) {
      await storageSet("cerberus-session", { active: true, user: u });
      onAuth();
    } else {
      setErr("Invalid credentials"); setLoading(false);
    }
  };

  if (mode === "loading") {
    return <div style={{minHeight:"100vh",background:K.bg,display:"flex",alignItems:"center",justifyContent:"center",fontFamily:K.f,color:K.cy}}>Loading...</div>;
  }

  const isSetup = mode === "setup";
  const inp = (val, set, placeholder, type="text") => <input value={val} onChange={e => set(e.target.value)} type={type} placeholder={placeholder} onKeyDown={e => e.key==="Enter"&&(isSetup?doSetup():doLogin())} style={{background:"rgba(255,255,255,0.04)",border:`1px solid ${err?"rgba(255,75,87,0.4)":"rgba(255,255,255,0.08)"}`,borderRadius:8,padding:"12px 16px",color:K.tx,fontFamily:K.f,fontSize:13,outline:"none",width:"100%",boxSizing:"border-box",transition:"border 0.3s"}}/>;

  return <div style={{minHeight:"100vh",background:K.bg,display:"flex",alignItems:"center",justifyContent:"center",fontFamily:K.f,position:"relative",overflow:"hidden"}}>
    <style>{`@import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@300;400;500;600;700&display=swap');
      @keyframes gp{0%{transform:translateY(0)}100%{transform:translateY(40px)}}
      @keyframes glow{0%,100%{opacity:.12}50%{opacity:.3}}
      @keyframes fu{from{opacity:0;transform:translateY(16px)}to{opacity:1;transform:translateY(0)}}
      input::placeholder{color:rgba(255,255,255,0.18)}`}</style>
    <div style={{position:"absolute",inset:0,opacity:0.03,backgroundImage:"linear-gradient(rgba(0,255,200,0.5) 1px,transparent 1px),linear-gradient(90deg,rgba(0,255,200,0.5) 1px,transparent 1px)",backgroundSize:"40px 40px",animation:"gp 8s linear infinite"}}/>
    <div style={{position:"absolute",width:400,height:400,borderRadius:"50%",background:`radial-gradient(circle,${K.cy}12 0%,transparent 70%)`,top:"20%",left:"50%",transform:"translateX(-50%)",animation:"glow 4s ease-in-out infinite"}}/>
    <div style={{position:"relative",zIndex:1,width:380,animation:"fu 0.6s ease-out"}}>
      <div style={{textAlign:"center",marginBottom:36}}>
        <div style={{fontSize:32,marginBottom:10}}>{"\uD83D\uDC15"}</div>
        <div style={{fontSize:26,fontWeight:700,letterSpacing:10,color:K.cy,textShadow:`0 0 30px ${K.cy}25`}}>CERBERUS</div>
        <div style={{fontSize:8,letterSpacing:4,color:K.dm,marginTop:5}}>NETWORK CONTROL SYSTEM</div>
      </div>
      <div style={{background:"rgba(15,21,32,0.85)",backdropFilter:"blur(20px)",border:"1px solid rgba(0,255,200,0.07)",borderRadius:14,padding:"28px 24px"}}>
        <div style={{fontSize:11,fontWeight:600,letterSpacing:2,color:K.dm,marginBottom:18,textAlign:"center"}}>{isSetup?"CREATE ACCOUNT":"SIGN IN"}</div>
        <div style={{marginBottom:14}}>
          <div style={{fontSize:8,fontWeight:600,letterSpacing:2,color:K.dm,marginBottom:5}}>USERNAME</div>
          {inp(u, setU, isSetup?"Choose username":"Enter username")}
        </div>
        <div style={{marginBottom:isSetup?14:20}}>
          <div style={{fontSize:8,fontWeight:600,letterSpacing:2,color:K.dm,marginBottom:5}}>PASSWORD</div>
          {inp(p, setP, isSetup?"Choose password":"Enter password", "password")}
        </div>
        {isSetup && <div style={{marginBottom:20}}>
          <div style={{fontSize:8,fontWeight:600,letterSpacing:2,color:K.dm,marginBottom:5}}>CONFIRM PASSWORD</div>
          {inp(p2, setP2, "Confirm password", "password")}
        </div>}
        {err && <div style={{fontSize:9,color:K.rd,marginBottom:12,textAlign:"center"}}>{err}</div>}
        <button onClick={isSetup?doSetup:doLogin} disabled={loading} style={{width:"100%",padding:"11px",background:loading?K.ra:`linear-gradient(135deg,${K.cy}20,${K.cy}45)`,border:`1px solid ${K.cy}35`,borderRadius:8,color:loading?K.dm:K.cy,fontFamily:K.f,fontSize:11,fontWeight:600,letterSpacing:3,cursor:loading?"wait":"pointer",transition:"all 0.3s"}}>
          {loading ? "\u27F3 "+(isSetup?"CREATING...":"AUTHENTICATING...") : isSetup?"CREATE ACCOUNT":"LOGIN"}
        </button>
      </div>
      <div style={{textAlign:"center",marginTop:16,fontSize:8,color:"rgba(255,255,255,0.1)"}}>v0.1.0</div>
    </div>
  </div>;
};

// ═══════════════════ AP DROPDOWN ═══════════════════
const APDrop = ({aps,sel,onSel}) => {
  const [open,setOpen] = useState(false);
  const ref = useRef(null);
  useEffect(() => { const h = (e) => { if (ref.current && !ref.current.contains(e.target)) setOpen(false); }; document.addEventListener("mousedown",h); return () => document.removeEventListener("mousedown",h); },[]);
  return <div ref={ref} style={{position:"relative",minWidth:280}}>
    <div onClick={() => { if (aps.length) setOpen(!open); }} style={{display:"flex",alignItems:"center",gap:7,padding:"5px 11px",background:K.ra,border:`1px solid ${sel?K.cy+"30":K.bx}`,borderRadius:5,cursor:aps.length?"pointer":"default",height:30}}>
      {sel ? <><Dot c={K.rd} p/><span style={{fontSize:10,fontWeight:600,color:K.cy}}>{sel.ssid}</span><Bdg c={sel.enc==="Open"?K.rd:K.gn}>{sel.enc}</Bdg><span style={{fontSize:8,color:K.dm}}>CH{sel.ch}</span><Sg v={sel.signal}/><span style={{marginLeft:"auto",fontSize:8,color:K.dm}}>{"\u25BE"}</span></> : <span style={{fontSize:9,color:K.dm}}>{aps.length?"Select target AP...":"Run recon"}</span>}
    </div>
    {open && <div style={{position:"absolute",top:"calc(100% + 2px)",left:0,right:0,background:K.cd,border:`1px solid ${K.bd}`,borderRadius:5,zIndex:200,maxHeight:240,overflowY:"auto",boxShadow:"0 10px 36px rgba(0,0,0,0.6)"}}>
      {aps.map(a => <div key={a.bssid} onClick={() => { onSel(a); setOpen(false); }} style={{display:"flex",alignItems:"center",gap:7,padding:"8px 11px",cursor:"pointer",borderBottom:`1px solid ${K.bx}`,background:sel?.bssid===a.bssid?`${K.cy}05`:"transparent"}}>
        <Sg v={a.signal}/><div style={{flex:1}}><div style={{fontSize:10,fontWeight:600,color:sel?.bssid===a.bssid?K.cy:K.tx}}>{a.ssid}</div><div style={{fontSize:7,color:K.dm,marginTop:1}}>{a.bssid}</div></div>
        <Bdg c={a.enc==="Open"?K.rd:K.gn}>{a.enc}</Bdg><span style={{fontSize:8,color:K.dm}}>CH{a.ch}</span>
      </div>)}
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
export default function App() {
  const [auth, setAuth] = useState(false);
  const [checking, setChecking] = useState(true);

  // Check for existing session on mount
  useEffect(() => {
    (async () => {
      const session = await storageGet("cerberus-session");
      if (session && session.active) setAuth(true);
      setChecking(false);
    })();
  }, []);

  if (checking) return <div style={{minHeight:"100vh",background:K.bg,display:"flex",alignItems:"center",justifyContent:"center",fontFamily:K.f,color:K.cy,fontSize:12}}>
    <style>{`@import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@300;400;500;600;700&display=swap');`}</style>
    Loading...
  </div>;

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
  const [di, setDi] = useState(0);
  const [et, setEt] = useState({on:false,ssid:"",ch:"6"});
  const [cap, setCap] = useState({on:false,tpl:"google"});
  const [doh, setDoh] = useState(DOH_INIT);
  const [prb, setPrb] = useState([]);
  const [mCfg, setMCfg] = useState({arp:true,dns:true,ssl:false,spoof:false});
  const [lf, setLf] = useState("all");
  const [ls, setLs] = useState("");
  const [hsState, setHsState] = useState("idle");
  const [hsAP, setHsAP] = useState("");
  const [adRoles, setAdRoles] = useState({scan:"wlan1",attack:"wlan2",upstream:"wlan0"});
  const ML = 5000;

  const tSel = (m) => setSel(p => { const n=new Set(p); n.has(m)?n.delete(m):n.add(m); return n; });
  const aAll = () => setSel(cli.length===sel.size?new Set():new Set(cli.map(c => c.mac)));
  const tM = (m) => setMS(p => { const n=new Set(p); n.has(m)?n.delete(m):n.add(m); return n; });
  const tD = (m) => setDS(p => { const n=new Set(p); n.has(m)?n.delete(m):n.add(m); return n; });

  const doRecon = () => {
    setRcing(true); setAps([]); setTAP(null); setCli([]); setSel(new Set()); setMS(new Set()); setDS(new Set()); setDns([]); setDi(0); setPrb([]); setHsState("idle");
    let i=0; const iv=setInterval(() => { if (i<APS.length) { const a=APS[i]; i++; setAps(p => [...p,a]); } else { clearInterval(iv); setRcing(false); } },300);
  };

  const doScan = () => {
    if (!tAP) return;
    setScing(true); setCli([]); setSel(new Set()); setMS(new Set()); setDS(new Set()); setDns([]); setDi(0); setPrb([]);
    const pool=CLI_MAP[tAP.ssid]||[]; let i=0; let pi=0;
    const iv=setInterval(() => { if (i<pool.length) { const c=pool[i]; i++; setCli(p => [...p,c]); if (pi<PRB.length&&Math.random()>0.4) { const pr=PRB[pi]; pi++; setPrb(p => [...p,pr]); } } else { clearInterval(iv); setScing(false); } },400);
  };

  const doHandshake = () => {
    if (!hsAP) return;
    setHsState("listening");
    setTimeout(() => { setHsState("deauthing"); setTimeout(() => { setHsState("captured"); },2500); },3000);
  };

  const downloadCap = () => {
    const data = `CERBERUS HANDSHAKE CAPTURE\n========================\nTarget: ${hsAP}\nTime: ${new Date().toISOString()}\nType: WPA2 4-Way Handshake\nAdapter: ${adRoles.scan}\n\n[Binary .cap data in production]`;
    const blob = new Blob([data],{type:"application/octet-stream"});
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url; a.download = `${hsAP.replace(/[^a-zA-Z0-9]/g,"_")}_handshake.cap`; a.click(); URL.revokeObjectURL(url);
  };

  useEffect(() => {
    if (mS.size>0 && di<DNS_Q.length) {
      const t = setTimeout(() => { setDns(p => { const n=[...p,DNS_Q[di]]; return n.length>ML?n.slice(n.length-ML):n; }); setDi(i => i+1); },500+Math.random()*900);
      return () => clearTimeout(t);
    }
  },[mS.size,di]);

  const fDns = dns.filter(e => {
    if (lf==="blocked"&&e.s!=="blocked") return false;
    if (lf==="passed"&&e.s!=="pass") return false;
    if (ls&&!e.d.toLowerCase().includes(ls.toLowerCase())) return false;
    return true;
  });

  // ═══ PAGES ═══
  const Dash = () => <div>
    <div style={{display:"grid",gridTemplateColumns:"repeat(4,1fr)",gap:8,marginBottom:14}}>
      {[{l:"CLIENTS",v:cli.length,c:K.cy,i:"\uD83D\uDCE1"},{l:"MITM",v:mS.size,c:K.gn,i:"\uD83D\uDD00"},{l:"DEAUTH",v:dS.size,c:K.rd,i:"\uD83D\uDC80"},{l:"DNS",v:dns.length,c:K.pu,i:"\uD83D\uDCCB"}].map(s => <Bx key={s.l} s={{textAlign:"center",padding:12}}><div style={{fontSize:18,marginBottom:2}}>{s.i}</div><div style={{fontSize:22,fontWeight:700,color:s.c}}>{s.v}</div><div style={{fontSize:7,letterSpacing:2,color:K.dm,marginTop:2}}>{s.l}</div></Bx>)}
    </div>
    <div style={{display:"grid",gridTemplateColumns:"1fr 1fr",gap:8,marginBottom:8}}>
      <Bx s={{padding:14}}><Lb>Target</Lb>{tAP ? <div><div style={{fontSize:13,fontWeight:700,color:K.cy}}>{tAP.ssid}</div><div style={{fontSize:9,color:K.dm,marginTop:2}}>{tAP.bssid} | CH{tAP.ch} | {tAP.enc}</div></div> : <span style={{color:K.dm,fontSize:10}}>No target</span>}</Bx>
      <Bx s={{padding:14}}><Lb>Active</Lb>
        {mS.size>0&&<Rw s={{marginBottom:3}}><Dot c={K.cy} p/><span style={{fontSize:10}}>MITM: {mS.size}</span></Rw>}
        {dS.size>0&&<Rw s={{marginBottom:3}}><Dot c={K.rd} p/><span style={{fontSize:10}}>Deauth: {dS.size}</span></Rw>}
        {et.on&&<Rw s={{marginBottom:3}}><Dot c={K.pu} p/><span style={{fontSize:10}}>Twin: {et.ssid}</span></Rw>}
        {cap.on&&<Rw><Dot c={K.am} p/><span style={{fontSize:10}}>Portal</span></Rw>}
        {mS.size===0&&dS.size===0&&!et.on&&!cap.on&&<span style={{color:K.dm,fontSize:10}}>None</span>}
      </Bx>
    </div>
    <Bx s={{padding:14,marginBottom:8}}><Lb>Connected Devices</Lb>
      <div style={{maxHeight:200,overflowY:"auto"}}>{cli.length>0 ? cli.map(c => { const [ic,ty]=dv(c.vnd); return <div key={c.mac} style={{display:"grid",gridTemplateColumns:"22px 1.4fr 80px 100px 55px 50px 55px",gap:5,padding:"6px 0",borderBottom:`1px solid ${K.bx}`,alignItems:"center",fontSize:10}}>
        <span>{ic}</span><div><span style={{fontWeight:600}}>{c.host}</span><div style={{fontSize:8,color:K.dm}}>{ty}</div></div>
        <span style={{color:K.dm,fontSize:9}}>{c.ip}</span><span style={{color:K.dm,fontSize:8}}>{c.mac}</span>
        <Sg v={c.sig}/><span style={{fontSize:9}}>{c.bw}</span><Bdg c={c.seen==="now"?K.gn:K.dm}>{c.seen==="now"?"LIVE":c.seen}</Bdg>
      </div>; }) : <Mt t="Scan to see devices"/>}</div>
    </Bx>
    <Bx s={{padding:14}}><Lb>Adapters</Lb>
      {ADAPTERS.map(a => <Rw key={a.id} s={{justifyContent:"space-between",padding:"5px 0",borderBottom:`1px solid ${K.bx}`}}>
        <Rw><span style={{fontSize:11,fontWeight:600,color:K.cy}}>{a.id}</span><span style={{fontSize:8,color:K.dm}}>{a.chip}</span></Rw>
        <Bdg c={adRoles.scan===a.id?K.cy:adRoles.attack===a.id?K.rd:K.gn}>{adRoles.scan===a.id?"SCAN":adRoles.attack===a.id?"ATTACK":"UPSTREAM"}</Bdg>
      </Rw>)}
    </Bx>
  </div>;

  const Recon = () => <div>
    <Rw s={{justifyContent:"space-between",marginBottom:10}}><Lb>Recon</Lb><Bn fn={doRecon} dis={rcing}>{rcing?"\u27F3 SCANNING...":"\u26A1 RECON"}</Bn></Rw>
    <Bx s={{marginBottom:8}}><Lb>Access Points ({aps.length})</Lb>
      {aps.map(a => <div key={a.bssid} onClick={() => setTAP(a)} style={{display:"grid",gridTemplateColumns:"1.6fr 1.1fr 42px 48px 55px",gap:5,padding:"7px 0",borderBottom:`1px solid ${K.bx}`,alignItems:"center",cursor:"pointer",background:tAP?.bssid===a.bssid?`${K.cy}05`:"transparent"}}>
        <span style={{fontSize:10,fontWeight:600,color:tAP?.bssid===a.bssid?K.cy:K.tx}}>{a.ssid}</span>
        <span style={{fontSize:8,color:K.dm}}>{a.bssid}</span><span style={{fontSize:8}}>CH{a.ch}</span>
        <Bdg c={a.enc==="Open"?K.rd:K.gn}>{a.enc}</Bdg><Rw><Sg v={a.signal}/><span style={{fontSize:8,color:K.dm}}>{a.signal}</span></Rw>
      </div>)}{aps.length===0&&<Mt t="Hit RECON"/>}
    </Bx>
    <Bx s={{marginBottom:8,borderColor:hsState==="captured"?`${K.gn}25`:K.bx}}><Lb>WPA Handshake Capture</Lb>
      <div style={{marginBottom:8}}><Sel value={hsAP} onChange={setHsAP} options={aps.filter(a => a.enc!=="Open").map(a => ({v:a.ssid,l:`${a.ssid} [${a.enc}] CH${a.ch}`}))} ph="Select target AP..."/></div>
      <Rw s={{gap:6}}>
        <Bn fn={doHandshake} dis={!hsAP||hsState==="listening"||hsState==="deauthing"} sm>{hsState==="idle"?"Capture Handshake":hsState==="listening"?"\u27F3 Listening...":hsState==="deauthing"?"\u27F3 Deauthing...":"\u2713 Captured!"}</Bn>
        {hsState==="captured"&&<Bn fn={downloadCap} sm v="g">{"\u2B07"} Download .cap</Bn>}
        {hsState!=="idle"&&hsState!=="captured"&&<Bn fn={() => setHsState("idle")} sm v="d">Cancel</Bn>}
      </Rw>
      {hsState!=="idle"&&<div style={{marginTop:8,padding:"6px 10px",borderRadius:4,background:hsState==="captured"?`${K.gn}08`:`${K.am}08`,border:`1px solid ${hsState==="captured"?K.gn:K.am}18`,fontSize:9,color:hsState==="captured"?K.gn:K.am}}>
        {hsState==="listening"&&<><Dot c={K.am} p/> Listening for EAPOL on {hsAP}... ({adRoles.scan})</>}
        {hsState==="deauthing"&&<><Dot c={K.rd} p/> Deauthing to force reconnect... ({adRoles.attack})</>}
        {hsState==="captured"&&<><Dot c={K.gn} p/> 4-way handshake captured! Ready to download.</>}
      </div>}
    </Bx>
    <div style={{display:"grid",gridTemplateColumns:"1fr 1fr",gap:8}}>
      <Bx><Lb>Clients ({cli.length})</Lb><div style={{maxHeight:200,overflowY:"auto"}}>{cli.map(c => { const [ic,ty]=dv(c.vnd); return <Rw key={c.mac} s={{padding:"5px 0",borderBottom:`1px solid ${K.bx}`}}><span style={{fontSize:12}}>{ic}</span><div style={{flex:1}}><div style={{fontSize:10,fontWeight:600}}>{c.host}</div><div style={{fontSize:8,color:K.dm}}>{c.ip} {"\u00B7"} {ty}</div></div><Sg v={c.sig}/></Rw>; })}{cli.length===0&&<Mt t="Select AP then scan"/>}</div></Bx>
      <Bx><Lb>Probes ({prb.length})</Lb><div style={{maxHeight:200,overflowY:"auto"}}>{prb.map((p,i) => <div key={i} style={{display:"flex",justifyContent:"space-between",padding:"5px 0",borderBottom:`1px solid ${K.bx}`,fontSize:10}}><span style={{color:K.dm}}>{p.t}</span><span style={{fontWeight:500}}>{p.ssid}</span><span style={{fontSize:8,color:K.dm}}>{p.mac.slice(-8)}</span></div>)}{prb.length===0&&<Mt t="Appear during scan"/>}</div></Bx>
    </div>
  </div>;

  const Targets = () => <div>
    <Rw s={{justifyContent:"space-between",marginBottom:10,flexWrap:"wrap",gap:5}}>
      <Lb>Targets{tAP?` \u2014 ${tAP.ssid}`:""}</Lb>
      <Rw s={{gap:4}}><Bn fn={doScan} dis={scing||!tAP} sm>{scing?"\u27F3":"\u26A1"} SCAN</Bn><Bn fn={() => sel.forEach(m => { if (!mS.has(m)) tM(m); })} sm v="g">MITM</Bn><Bn fn={() => sel.forEach(m => { if (!dS.has(m)) tD(m); })} sm v="d">Deauth</Bn></Rw>
    </Rw>
    {!tAP ? <Bx><Mt t="Select target AP from top bar"/></Bx> :
    <Bx s={{padding:0,overflow:"hidden"}}>
      <div style={{overflowX:"auto"}}>
        <div style={{display:"grid",gridTemplateColumns:"28px 20px 1.5fr 80px 115px 40px 45px 42px 42px",gap:4,padding:"8px 10px",fontSize:7,fontWeight:700,letterSpacing:2,color:K.dm,borderBottom:`1px solid ${K.bx}`,alignItems:"center",minWidth:580}}>
          <Ck on={sel.size===cli.length&&cli.length>0} fn={aAll}/><span/><span>DEVICE</span><span>IP</span><span>MAC</span><span>SIG</span><span>STAT</span><span>MITM</span><span>DEAU</span>
        </div>
        <div style={{maxHeight:340,overflowY:"auto"}}>
          {cli.filter(Boolean).map((c,x) => { const [ic,ty]=dv(c.vnd); const s=sel.has(c.mac); const m=mS.has(c.mac); const d=dS.has(c.mac);
            return <div key={c.mac} style={{display:"grid",gridTemplateColumns:"28px 20px 1.5fr 80px 115px 40px 45px 42px 42px",gap:4,padding:"7px 10px",borderBottom:`1px solid ${K.bx}`,alignItems:"center",background:s?`${K.cy}04`:"transparent",minWidth:580,animation:`fi 0.2s ease ${x*40}ms both`}}>
              <Ck on={s} fn={() => tSel(c.mac)}/><span style={{fontSize:11}}>{ic}</span>
              <div><div style={{fontSize:10,fontWeight:600,color:s?K.cy:K.tx}}>{c.host}</div><div style={{fontSize:7,color:K.dm}}>{ty}</div></div>
              <span style={{fontSize:9}}>{c.ip}</span><span style={{fontSize:7,color:K.dm}}>{c.mac}</span>
              <Sg v={c.sig}/><Bdg c={c.seen==="now"?K.gn:K.dm}>{c.seen==="now"?"LIVE":c.seen}</Bdg>
              <Tog on={m} fn={() => tM(c.mac)} c={K.cy}/><Tog on={d} fn={() => tD(c.mac)} c={K.rd}/>
            </div>; })}
          {cli.length===0&&<Mt t="Hit SCAN"/>}
        </div>
      </div>
    </Bx>}
    {dS.size>0&&<Bx s={{marginTop:7,borderColor:`${K.rd}18`,padding:10}}><Rw><Dot c={K.rd} p/><span style={{fontSize:9,color:K.rd}}>Deauth: {cli.filter(c => dS.has(c.mac)).map(c => c.host).join(", ")}</span></Rw></Bx>}
  </div>;

  const MitmPg = () => { const act=cli.filter(c => mS.has(c.mac)); return <div>
    <Lb>MITM</Lb>
    <div style={{display:"grid",gridTemplateColumns:"1fr 1fr",gap:8}}>
      <Bx><Lb>Add Target</Lb><Sel value="" onChange={v => { if (v) tM(v); }} options={cli.map(c => ({v:c.mac,l:`${dv(c.vnd)[0]} ${c.host} \u2014 ${c.ip}`}))} ph="+ Add..."/>
        <div style={{marginTop:8}}>{act.map(c => <Rw key={c.mac} s={{justifyContent:"space-between",padding:"5px 0",borderBottom:`1px solid ${K.bx}`}}><Rw><Dot c={K.cy} p/><span style={{fontSize:10,fontWeight:600}}>{dv(c.vnd)[0]} {c.host}</span><span style={{fontSize:8,color:K.dm}}>{c.ip}</span></Rw><Bn fn={() => tM(c.mac)} sm v="d">Stop</Bn></Rw>)}{act.length===0&&<Mt t="Select targets"/>}</div>
      </Bx>
      <Bx><Lb>Config</Lb>
        {[{k:"arp",l:"ARP Spoofing",d:"Intercept via ARP"},{k:"dns",l:"DNS Logging",d:"Log DNS"},{k:"ssl",l:"SSL Strip",d:"Downgrade HTTPS"},{k:"spoof",l:"DNS Spoofing",d:"Redirect domains"}].map(a => <Rw key={a.k} s={{justifyContent:"space-between",padding:"5px 0",borderBottom:`1px solid ${K.bx}`}}><div><div style={{fontSize:10,fontWeight:600}}>{a.l}</div><div style={{fontSize:8,color:K.dm}}>{a.d}</div></div><Tog on={mCfg[a.k]} fn={() => setMCfg(p => ({...p,[a.k]:!p[a.k]}))}/></Rw>)}
      </Bx>
    </div>
  </div>; };

  const TwinPg = () => <div>
    <Lb>Evil Twin</Lb>
    <div style={{display:"grid",gridTemplateColumns:"1fr 1fr",gap:8}}>
      <Bx><Lb>Clone AP</Lb>
        <div style={{marginBottom:7}}><div style={{fontSize:8,color:K.dm,marginBottom:3}}>SSID</div><Sel value={et.ssid} onChange={v => setEt(p => ({...p,ssid:v}))} options={aps.map(a => ({v:a.ssid,l:`${a.ssid} [CH${a.ch}]`}))} ph="Select..."/></div>
        <div style={{marginBottom:7}}><div style={{fontSize:8,color:K.dm,marginBottom:3}}>Channel</div><Sel value={et.ch} onChange={v => setEt(p => ({...p,ch:v}))} options={[1,6,11,36,40,44].map(c => ({v:String(c),l:`CH ${c}`}))}/></div>
        <Bn fn={() => setEt(p => ({...p,on:!p.on}))} v={et.on?"d":"p"} sx={{width:"100%"}}>{et.on?"\u25FC STOP":"\u25B6 LAUNCH"}</Bn>
        {et.on&&et.ssid&&<div style={{marginTop:7,padding:"6px 9px",borderRadius:4,background:`${K.cy}06`,border:`1px solid ${K.cy}15`,fontSize:9,color:K.cy}}><Dot c={K.cy} p/> "{et.ssid}" CH{et.ch}</div>}
      </Bx>
      <Bx><Lb>Probes</Lb>
        {prb.length>0 ? [...new Set(prb.map(p => p.ssid))].map(s => <Rw key={s} s={{justifyContent:"space-between",padding:"4px 0",borderBottom:`1px solid ${K.bx}`}}><span style={{fontSize:10}}>{s}</span><Bdg c={K.pu}>PROBE</Bdg></Rw>) : <Mt t="Run recon"/>}
        <div style={{marginTop:10}}><Lb>Force Reconnect</Lb><Sel value="" onChange={v => { if (v) tD(v); }} options={cli.map(c => ({v:c.mac,l:`${dv(c.vnd)[0]} ${c.host}`}))} ph="Deauth target..."/></div>
      </Bx>
    </div>
  </div>;

  const PortalPg = () => <div>
    <Lb>Captive Portal</Lb>
    <div style={{display:"grid",gridTemplateColumns:"1fr 1fr",gap:8}}>
      <Bx><Lb>Config</Lb>
        <div style={{marginBottom:7}}><div style={{fontSize:8,color:K.dm,marginBottom:3}}>Template</div><Sel value={cap.tpl} onChange={v => setCap(p => ({...p,tpl:v}))} options={[{v:"google",l:"Google Sign-In"},{v:"facebook",l:"Facebook Login"},{v:"wifi",l:"WiFi Terms"},{v:"hotel",l:"Hotel Portal"},{v:"custom",l:"Custom HTML"}]}/></div>
        <div style={{fontSize:9,color:K.dm,marginBottom:10}}>Runs on Evil Twin AP</div>
        <Bn fn={() => setCap(p => ({...p,on:!p.on}))} v={cap.on?"d":"p"} sx={{width:"100%"}}>{cap.on?"\u25FC STOP":"\u25B6 LAUNCH"}</Bn>
      </Bx>
      <Bx><Lb>Captured Creds</Lb><Mt t={cap.on?"Waiting...":"Portal not active"}/></Bx>
    </div>
  </div>;

  const LogPg = () => <div>
    <Rw s={{justifyContent:"space-between",marginBottom:8,flexWrap:"wrap",gap:5}}>
      <Rw s={{gap:4}}><Lb>DNS Log ({fDns.length}/{dns.length})</Lb>{dns.length>=ML&&<Bdg c={K.am}>MAX</Bdg>}</Rw>
      <Rw s={{gap:3}}>
        {["all","passed","blocked"].map(f => <button key={f} onClick={() => setLf(f)} style={{fontFamily:K.f,fontSize:8,fontWeight:600,letterSpacing:1,padding:"3px 8px",borderRadius:3,cursor:"pointer",border:`1px solid ${lf===f?(f==="blocked"?K.rd:f==="passed"?K.cy:K.bl)+"35":K.bx}`,background:lf===f?(f==="blocked"?K.rd:f==="passed"?K.cy:K.bl)+"10":"transparent",color:lf===f?(f==="blocked"?K.rd:f==="passed"?K.cy:K.bl):K.dm,textTransform:"uppercase"}}>{f}</button>)}
        <Bn fn={() => { setDns([]); setDi(0); }} sm v="g">Clear</Bn>
      </Rw>
    </Rw>
    <div style={{marginBottom:6}}><input value={ls} onChange={e => setLs(e.target.value)} placeholder="Filter domain..." style={{background:K.ra,color:K.tx,border:`1px solid ${K.bx}`,borderRadius:4,padding:"6px 9px",fontSize:9,fontFamily:K.f,width:"100%",outline:"none",boxSizing:"border-box"}}/></div>
    <Bx s={{padding:0}}>
      <div style={{display:"grid",gridTemplateColumns:"55px 1fr 42px",gap:5,padding:"7px 12px",fontSize:7,fontWeight:700,letterSpacing:2,color:K.dm,borderBottom:`1px solid ${K.bx}`}}><span>TIME</span><span>DOMAIN</span><span>STATUS</span></div>
      <div style={{maxHeight:440,overflowY:"auto"}}>
        {fDns.map((e,i) => <div key={i} style={{display:"grid",gridTemplateColumns:"55px 1fr 42px",gap:5,padding:"5px 12px",fontSize:10,borderBottom:`1px solid ${K.bx}`}}><span style={{color:K.dm}}>{e.t}</span><span style={{fontWeight:500}}>{e.d}</span><Bdg c={e.s==="blocked"?K.rd:K.cy}>{e.s==="blocked"?"BLOCK":"PASS"}</Bdg></div>)}
        {fDns.length===0&&<Mt t={dns.length===0?"No DNS logged":"No matches"}/>}
      </div>
    </Bx>
  </div>;

  const CfgPg = () => <div>
    <Lb>Settings</Lb>
    <div style={{display:"grid",gridTemplateColumns:"1fr 1fr",gap:8,marginBottom:8}}>
      <Bx><Lb>Adapter Roles</Lb><div style={{fontSize:9,color:K.dm,marginBottom:8}}>Assign adapters to roles</div>
        {ADAPTERS.map(a => { const role=adRoles.scan===a.id?"scan":adRoles.attack===a.id?"attack":"upstream";
          return <div key={a.id} style={{padding:"8px 0",borderBottom:`1px solid ${K.bx}`}}>
            <Rw s={{justifyContent:"space-between",marginBottom:4}}>
              <div><span style={{fontSize:11,fontWeight:700,color:K.cy}}>{a.id}</span><span style={{fontSize:8,color:K.dm,marginLeft:6}}>{a.chip}</span></div>
              <Bdg c={role==="scan"?K.cy:role==="attack"?K.rd:K.gn}>{role.toUpperCase()}</Bdg>
            </Rw>
            <div style={{display:"grid",gridTemplateColumns:"1fr 1fr",gap:4,fontSize:8,color:K.dm}}>
              <div>MAC: {a.mac}</div><div>Band: {a.band}</div>
              <div style={{gridColumn:"1/3"}}>Modes: {a.modes.join(", ")}</div>
            </div>
            <div style={{marginTop:4}}><Sel value={role} onChange={v => {
              const nr={...adRoles}; const old=Object.entries(nr).find(([,id]) => id===a.id);
              const swapTo=Object.entries(nr).find(([r]) => r===v);
              if (old&&swapTo) { nr[old[0]]=swapTo[1]; }
              nr[v]=a.id;
              setAdRoles(nr);
            }} options={[{v:"scan",l:"Scanning"},{v:"attack",l:"Attacking"},{v:"upstream",l:"Upstream"}]}/></div>
          </div>; })}
      </Bx>
      <Bx><Lb>DoH / DoT Blocking</Lb><div style={{fontSize:9,color:K.dm,marginBottom:6}}>Block encrypted DNS</div>
        {doh.map((s,i) => <Rw key={s.ip} s={{justifyContent:"space-between",padding:"4px 0",borderBottom:`1px solid ${K.bx}`}}><div><span style={{fontSize:10,fontWeight:600}}>{s.n}</span><span style={{fontSize:8,color:K.dm,marginLeft:5}}>{s.ip}</span></div><Tog on={s.on} fn={() => { const n=[...doh]; n[i]={...n[i],on:!n[i].on}; setDoh(n); }} c={K.rd}/></Rw>)}
        <div style={{marginTop:10}}><Lb>Port 853</Lb><Rw s={{justifyContent:"space-between"}}><span style={{fontSize:10}}>Block DoT</span><Tog on={true} fn={() => {}} c={K.rd}/></Rw></div>
        <div style={{marginTop:14,borderTop:`1px solid ${K.bx}`,paddingTop:10}}>
          <Bn fn={onLogout} v="d" sx={{width:"100%"}}>Logout</Bn>
        </div>
      </Bx>
    </div>
  </div>;

  const pages={dash:Dash,recon:Recon,targets:Targets,mitm:MitmPg,twin:TwinPg,portal:PortalPg,log:LogPg,cfg:CfgPg};
  const Pg=pages[pg];

  return <div style={{display:"flex",flexDirection:"column",height:"100vh",background:K.bg,color:K.tx,fontFamily:K.f}}>
    <style>{`@import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@300;400;500;600;700&display=swap');@keyframes cpls{0%{transform:scale(1);opacity:.4}100%{transform:scale(2.2);opacity:0}}@keyframes fi{from{opacity:0;transform:translateY(3px)}to{opacity:1;transform:translateY(0)}}*{scrollbar-width:thin;scrollbar-color:${K.ra} transparent}::selection{background:${K.cy}28}input::placeholder{color:rgba(255,255,255,0.16)}`}</style>
    <div style={{display:"flex",alignItems:"center",gap:8,padding:"8px 14px",background:K.sf,borderBottom:`1px solid ${K.bx}`,flexShrink:0}}>
      <div style={{marginRight:4}}><div style={{fontSize:13,fontWeight:700,letterSpacing:5,color:K.cy}}>CERBERUS</div><div style={{fontSize:6,letterSpacing:2,color:K.dm}}>v0.1.0</div></div>
      <div style={{width:1,height:22,background:K.bx}}/>
      <APDrop aps={aps} sel={tAP} onSel={setTAP}/>
      <Bn fn={doRecon} dis={rcing} sm>{rcing?"\u27F3":"\uD83D\uDCE1"} RECON</Bn>
      <Bn fn={doScan} dis={scing||!tAP} sm>{scing?"\u27F3":"\u26A1"} SCAN</Bn>
      <div style={{flex:1}}/>
      <Rw><span style={{fontSize:8,color:K.dm}}>scan:{adRoles.scan}</span><span style={{fontSize:8,color:K.dm}}>atk:{adRoles.attack}</span></Rw>
      <Rw><Dot c={K.gn} p/><span style={{fontSize:8,color:K.dm}}>Online</span></Rw>
    </div>
    <div style={{display:"flex",flex:1,overflow:"hidden"}}>
      <div style={{width:140,background:K.sf,borderRight:`1px solid ${K.bx}`,padding:"8px 0",display:"flex",flexDirection:"column",flexShrink:0}}>
        {NAV.map(n => { const a=pg===n.id; return <div key={n.id} onClick={() => setPg(n.id)} style={{display:"flex",alignItems:"center",gap:7,padding:"7px 12px",cursor:"pointer",fontSize:10,fontWeight:a?600:400,color:a?K.cy:K.dm,background:a?`${K.cy}05`:"transparent",borderLeft:a?`2px solid ${K.cy}`:"2px solid transparent",transition:"all 0.15s"}}><span style={{fontSize:11,width:15,textAlign:"center"}}>{n.ic}</span>{n.l}</div>; })}
      </div>
      <div style={{flex:1,padding:14,overflowY:"auto"}}><Pg/></div>
    </div>
  </div>;
}
