import { useState, useEffect, useRef } from "react";

// ═══ COOKIES ═══
function setCookie(name, value, days) {
  const d = new Date();
  d.setTime(d.getTime() + days * 86400000);
  document.cookie = name + "=" + encodeURIComponent(JSON.stringify(value)) + ";expires=" + d.toUTCString() + ";path=/";
}
function getCookie(name) {
  const match = document.cookie.match(new RegExp("(^| )" + name + "=([^;]+)"));
  if (match) try { return JSON.parse(decodeURIComponent(match[2])); } catch(e) {}
  return null;
}
function delCookie(name) { document.cookie = name + "=;expires=Thu, 01 Jan 1970 00:00:00 GMT;path=/"; }

// ═══ THEME ═══
const K = {
  bg:"#030508",sf:"#080d16",cd:"#0c1420",ra:"#131c2a",
  bd:"rgba(0,255,200,0.06)",bx:"rgba(255,255,255,0.035)",
  cy:"#00ffc8",rd:"#ff4757",am:"#ffc107",gn:"#22c55e",pu:"#a78bfa",bl:"#60a5fa",
  tx:"#dfe6f0",dm:"rgba(255,255,255,0.25)",
  f:"'Courier New',monospace",
};
const CSS = `
@keyframes fi{from{opacity:0;transform:translateY(3px)}to{opacity:1;transform:translateY(0)}}
@keyframes fadeUp{from{opacity:0;transform:translateY(16px)}to{opacity:1;transform:translateY(0)}}
@keyframes shakeX{0%,100%{transform:translateX(0)}20%{transform:translateX(-8px)}40%{transform:translateX(6px)}60%{transform:translateX(-4px)}80%{transform:translateX(2px)}}
*{scrollbar-width:thin;scrollbar-color:${K.ra} transparent;box-sizing:border-box;margin:0;padding:0}
body{background:${K.bg};overflow:hidden}
::selection{background:${K.cy}28}
input::placeholder{color:rgba(255,255,255,0.15)}
input:focus{border-color:${K.cy}40!important;background:rgba(0,255,200,0.02)!important}
select{appearance:none}`;

// ═══ MICRO ═══
const Dot=({c=K.cy})=><span style={{width:7,height:7,borderRadius:"50%",background:c,display:"inline-block",flexShrink:0}}/>;
const Bdg=({children,c=K.cy})=><span style={{fontSize:8,fontWeight:700,letterSpacing:1,padding:"2px 7px",borderRadius:3,background:`${c}12`,color:c,whiteSpace:"nowrap"}}>{children}</span>;
const Tog=({on,fn,c=K.cy})=><div onClick={fn} style={{width:40,height:20,borderRadius:10,background:on?c:"rgba(255,255,255,0.06)",position:"relative",cursor:"pointer",flexShrink:0}}><div style={{width:14,height:14,borderRadius:"50%",background:"#fff",position:"absolute",top:3,left:on?23:3,transition:"left 0.15s"}}/></div>;
const Ck=({on,fn,c=K.cy})=><div onClick={fn} style={{width:15,height:15,borderRadius:3,border:`1.5px solid ${on?c:"rgba(255,255,255,0.1)"}`,background:on?`${c}18`:"transparent",cursor:"pointer",display:"flex",alignItems:"center",justifyContent:"center",flexShrink:0}}>{on&&<span style={{color:c,fontSize:9,fontWeight:700}}>{"\u2713"}</span>}</div>;
const Lb=({children})=><div style={{fontSize:8,fontWeight:700,letterSpacing:3,color:K.dm,marginBottom:8,textTransform:"uppercase"}}>{children}</div>;
const Bx=({children,s={}})=><div style={{background:K.cd,borderRadius:8,border:`1px solid ${K.bx}`,padding:16,...s}}>{children}</div>;
const Sg=({v})=>{const s=v>-45?4:v>-55?3:v>-65?2:1;const c=s>=3?K.cy:s===2?K.am:K.rd;return<div style={{display:"flex",alignItems:"flex-end",gap:1.5,height:12}}>{[1,2,3,4].map(i=><div key={i} style={{width:2,height:2+i*2.5,borderRadius:1,background:i<=s?c:"rgba(255,255,255,0.05)"}}/>)}</div>;};
const Bn=({children,fn,v="p",dis,sm,sx={}})=>{const vs={p:{bg:`linear-gradient(135deg,${K.cy}18,${K.cy}35)`,c:K.cy,b:`1px solid ${K.cy}30`},d:{bg:`${K.rd}12`,c:K.rd,b:`1px solid ${K.rd}30`},g:{bg:"rgba(255,255,255,0.025)",c:K.dm,b:`1px solid ${K.bx}`}};return<button onClick={fn} disabled={dis} style={{fontFamily:K.f,fontSize:sm?9:10,fontWeight:600,letterSpacing:1,padding:sm?"5px 10px":"7px 16px",borderRadius:5,cursor:dis?"not-allowed":"pointer",border:vs[v].b,background:vs[v].bg,color:vs[v].c,opacity:dis?0.3:1,...sx}}>{children}</button>;};
const Rw=({children,s={}})=><div style={{display:"flex",alignItems:"center",gap:7,...s}}>{children}</div>;
const Mt=({t})=><div style={{padding:24,textAlign:"center",color:K.dm,fontSize:10}}>{t}</div>;
const Sel=({value,onChange,options,ph})=><select value={value} onChange={e=>onChange(e.target.value)} style={{background:K.ra,color:K.tx,border:`1px solid ${K.bx}`,borderRadius:5,padding:"7px 10px",fontSize:10,fontFamily:K.f,width:"100%",outline:"none"}}>{ph&&<option value="">{ph}</option>}{options.map(o=><option key={o.v} value={o.v}>{o.l}</option>)}</select>;

// ═══ EYE ═══
const EyeSVG=({open})=><svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">{open?<><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></>:<><path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94"/><path d="M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19"/><line x1="1" y1="1" x2="23" y2="23"/><path d="M14.12 14.12a3 3 0 1 1-4.24-4.24"/></>}</svg>;
const PwInput=({value,onChange,placeholder,onKey})=>{const[show,setShow]=useState(false);return<div style={{position:"relative"}}><input value={value} onChange={e=>onChange(e.target.value)} onKeyDown={onKey} type={show?"text":"password"} placeholder={placeholder} autoComplete="off" style={{width:"100%",boxSizing:"border-box",background:"rgba(255,255,255,0.03)",border:"1px solid rgba(255,255,255,0.06)",borderRadius:8,padding:"12px 44px 12px 14px",color:K.tx,fontFamily:K.f,fontSize:13,outline:"none"}}/><button onClick={()=>setShow(!show)} style={{position:"absolute",right:10,top:"50%",transform:"translateY(-50%)",background:"none",border:"none",cursor:"pointer",color:show?K.cy:"rgba(255,255,255,0.2)",padding:4,display:"flex"}} tabIndex={-1}><EyeSVG open={show}/></button></div>;};

// ═══ DATA ═══
const APS=[{ssid:"HomeNet-5G",bssid:"A0:B1:C2:D3:E4:01",ch:36,enc:"WPA2",signal:-38},{ssid:"Linksys_Guest",bssid:"F0:E1:D2:C3:B4:02",ch:6,enc:"WPA2",signal:-52},{ssid:"xfinitywifi",bssid:"11:22:33:44:55:03",ch:1,enc:"Open",signal:-61},{ssid:"NETGEAR-2.4",bssid:"AA:CC:EE:11:33:04",ch:11,enc:"WPA3",signal:-70},{ssid:"TP-Link_8842",bssid:"CC:DD:EE:FF:00:05",ch:6,enc:"WPA2",signal:-74}];
const CLI_MAP={"HomeNet-5G":[{ip:"192.168.1.12",mac:"A4:83:E7:2F:11:B0",host:"Jake-Laptop",vnd:"Dell",sig:-42,bw:"2.4 MB/s",seen:"now"},{ip:"192.168.1.15",mac:"F0:18:98:44:CC:01",host:"iPhone-Sarah",vnd:"Apple",sig:-58,bw:"340 KB/s",seen:"now"},{ip:"192.168.1.23",mac:"DC:A6:32:AA:BB:CC",host:"Ring-Doorbell",vnd:"Amazon",sig:-61,bw:"18 KB/s",seen:"2m"},{ip:"192.168.1.31",mac:"B8:27:EB:12:34:56",host:"PS5-Living",vnd:"Sony",sig:-39,bw:"12.1 MB/s",seen:"now"},{ip:"192.168.1.44",mac:"00:1A:2B:3C:4D:5E",host:"Galaxy-Tab",vnd:"Samsung",sig:-55,bw:"890 KB/s",seen:"now"},{ip:"192.168.1.50",mac:"E8:6F:38:11:22:33",host:"Roku-Bedroom",vnd:"Roku",sig:-67,bw:"4.2 MB/s",seen:"5m"}],"Linksys_Guest":[{ip:"192.168.2.10",mac:"AA:11:22:33:44:55",host:"MacBook-Pro",vnd:"Apple",sig:-54,bw:"1.8 MB/s",seen:"now"},{ip:"192.168.2.15",mac:"BB:22:33:44:55:66",host:"Pixel-7",vnd:"Google",sig:-60,bw:"450 KB/s",seen:"now"}],"xfinitywifi":[{ip:"10.0.0.12",mac:"CC:33:44:55:66:77",host:"iPhone-12",vnd:"Apple",sig:-63,bw:"200 KB/s",seen:"now"}]};
const PRB=[{mac:"A4:83:E7:2F:11:B0",ssid:"Starbucks_WiFi",t:"14:31:02"},{mac:"F0:18:98:44:CC:01",ssid:"ATT-Passpoint",t:"14:31:08"},{mac:"00:1A:2B:3C:4D:5E",ssid:"McDonalds_Free",t:"14:31:15"},{mac:"A4:83:E7:2F:11:B0",ssid:"SchoolWiFi",t:"14:31:22"}];
const DNS_Q=[{t:"14:32:07",d:"www.youtube.com",s:"pass"},{t:"14:32:08",d:"i.ytimg.com",s:"pass"},{t:"14:32:11",d:"ads.doubleclick.net",s:"blocked"},{t:"14:32:14",d:"www.reddit.com",s:"pass"},{t:"14:32:16",d:"tracker.analytics.com",s:"blocked"},{t:"14:32:19",d:"discord.com",s:"pass"},{t:"14:32:25",d:"api.snapchat.com",s:"pass"},{t:"14:32:30",d:"telemetry.microsoft.com",s:"blocked"},{t:"14:32:33",d:"www.tiktok.com",s:"pass"},{t:"14:32:41",d:"pixel.adsafeprotected.com",s:"blocked"},{t:"14:32:44",d:"play.google.com",s:"pass"},{t:"14:32:50",d:"beacons.gvt2.com",s:"blocked"}];
const ADAPTERS=[{id:"wlan0",chip:"MT7981B",mac:"AA:BB:CC:11:22:33",band:"2.4/5"},{id:"wlan1",chip:"RTL8812AU",mac:"DD:EE:FF:44:55:66",band:"2.4/5"},{id:"wlan2",chip:"AR9271",mac:"11:22:33:AA:BB:CC",band:"2.4"}];
const DOH_D=[{ip:"1.1.1.1",n:"Cloudflare",on:true},{ip:"1.0.0.1",n:"CF Alt",on:true},{ip:"8.8.8.8",n:"Google",on:true},{ip:"8.8.4.4",n:"Google Alt",on:true},{ip:"9.9.9.9",n:"Quad9",on:true}];

// ═══ AP DROPDOWN ═══
const APDrop=({aps,sel,onSel})=>{const[open,setOpen]=useState(false);const ref=useRef(null);useEffect(()=>{const h=e=>{if(ref.current&&!ref.current.contains(e.target))setOpen(false);};document.addEventListener("mousedown",h);return()=>document.removeEventListener("mousedown",h);},[]);
return<div ref={ref} style={{position:"relative",minWidth:240}}><div onClick={()=>{if(aps.length)setOpen(!open);}} style={{display:"flex",alignItems:"center",gap:6,padding:"5px 10px",background:K.ra,border:`1px solid ${sel?K.cy+"30":K.bx}`,borderRadius:5,cursor:aps.length?"pointer":"default",height:28}}>{sel?<><Dot c={K.rd}/><span style={{fontSize:10,fontWeight:600,color:K.cy}}>{sel.ssid}</span><Bdg c={sel.enc==="Open"?K.rd:K.gn}>{sel.enc}</Bdg><span style={{marginLeft:"auto",fontSize:8,color:K.dm}}>{"\u25BE"}</span></>:<span style={{fontSize:9,color:K.dm}}>{aps.length?"Select AP...":"Run recon"}</span>}</div>
{open&&<div style={{position:"absolute",top:"calc(100% + 2px)",left:0,right:0,background:K.cd,border:`1px solid ${K.bd}`,borderRadius:5,zIndex:200,maxHeight:200,overflowY:"auto",boxShadow:"0 8px 24px rgba(0,0,0,0.5)"}}>{aps.map(a=><div key={a.bssid} onClick={()=>{onSel(a);setOpen(false);}} style={{display:"flex",alignItems:"center",gap:6,padding:"7px 10px",cursor:"pointer",borderBottom:`1px solid ${K.bx}`,fontSize:10}}><Sg v={a.signal}/><span style={{flex:1,fontWeight:600,color:sel?.bssid===a.bssid?K.cy:K.tx}}>{a.ssid}</span><Bdg c={a.enc==="Open"?K.rd:K.gn}>{a.enc}</Bdg></div>)}</div>}</div>;};

// ═══ LOGIN — uses cookies ═══
const AuthScreen=({onAuth})=>{
  const[u,setU]=useState("");const[p,setP]=useState("");const[p2,setP2]=useState("");
  const[err,setErr]=useState("");const[loading,setLoading]=useState(false);const[shake,setShake]=useState(false);
  const savedCreds=getCookie("cerberus_creds");
  const[isSetup]=useState(!savedCreds);

  const doSetup=()=>{
    if(!u.trim()){setErr("Username required");return;}
    if(p.length<3){setErr("Min 3 characters");return;}
    if(p!==p2){setErr("Passwords don't match");return;}
    setLoading(true);setErr("");
    setTimeout(()=>{
      setCookie("cerberus_creds",{u:u.trim(),p},365);
      setCookie("cerberus_auth",{user:u.trim()},7);
      onAuth(u.trim());
    },500);
  };
  const doLogin=()=>{
    if(!u||!p)return;setLoading(true);setErr("");setShake(false);
    setTimeout(()=>{
      if(savedCreds&&u===savedCreds.u&&p===savedCreds.p){
        setCookie("cerberus_auth",{user:u},7);
        onAuth(u);
      }else{setErr("Access denied");setShake(true);setLoading(false);setTimeout(()=>setShake(false),500);}
    },600);
  };
  const submit=isSetup?doSetup:doLogin;const handleKey=e=>{if(e.key==="Enter")submit();};

  return<div style={{minHeight:"100vh",background:K.bg,display:"flex",alignItems:"center",justifyContent:"center",fontFamily:K.f}}><style>{CSS}</style>
  <div style={{width:380,animation:`fadeUp 0.5s ease-out${shake?",shakeX 0.4s ease":""}`}}>
    <div style={{textAlign:"center",marginBottom:32}}>
      <div style={{fontSize:28,fontWeight:700,letterSpacing:10,color:K.cy}}>CERBERUS</div>
      <div style={{fontSize:9,letterSpacing:4,color:K.dm,marginTop:6}}>WELCOME TO THE MEME-DETECTOR</div>
    </div>
    <div style={{background:"rgba(12,20,32,0.8)",border:`1px solid ${err?K.rd+"30":"rgba(0,255,200,0.06)"}`,borderRadius:12,padding:"28px 24px"}}>
      <div style={{fontSize:10,fontWeight:600,letterSpacing:3,color:K.dm,marginBottom:18,textAlign:"center"}}>{isSetup?"CREATE ACCOUNT":"SIGN IN"}</div>
      <div style={{marginBottom:14}}><div style={{fontSize:8,fontWeight:600,letterSpacing:2,color:K.dm,marginBottom:5}}>USERNAME</div><input value={u} onChange={e=>{setU(e.target.value);setErr("");}} onKeyDown={handleKey} placeholder={isSetup?"Choose username":"Username"} autoComplete="off" style={{width:"100%",boxSizing:"border-box",background:"rgba(255,255,255,0.03)",border:"1px solid rgba(255,255,255,0.06)",borderRadius:8,padding:"12px 14px",color:K.tx,fontFamily:K.f,fontSize:13,outline:"none"}}/></div>
      <div style={{marginBottom:isSetup?14:20}}><div style={{fontSize:8,fontWeight:600,letterSpacing:2,color:K.dm,marginBottom:5}}>PASSWORD</div><PwInput value={p} onChange={v=>{setP(v);setErr("");}} placeholder={isSetup?"Choose password":"Password"} onKey={handleKey}/></div>
      {isSetup&&<div style={{marginBottom:20}}><div style={{fontSize:8,fontWeight:600,letterSpacing:2,color:K.dm,marginBottom:5}}>CONFIRM</div><PwInput value={p2} onChange={v=>{setP2(v);setErr("");}} placeholder="Confirm password" onKey={handleKey}/></div>}
      {err&&<div style={{fontSize:10,color:K.rd,marginBottom:14,textAlign:"center",padding:"6px 10px",background:`${K.rd}08`,borderRadius:5,border:`1px solid ${K.rd}18`}}>{err}</div>}
      <button onClick={submit} disabled={loading} style={{width:"100%",padding:"12px",background:loading?K.ra:`linear-gradient(135deg,${K.cy}20,${K.cy}40)`,border:`1px solid ${K.cy}${loading?"15":"35"}`,borderRadius:8,color:loading?K.dm:K.cy,fontFamily:K.f,fontSize:12,fontWeight:600,letterSpacing:3,cursor:loading?"wait":"pointer"}}>{loading?(isSetup?"CREATING...":"VERIFYING..."):(isSetup?"CREATE ACCOUNT":"LOGIN")}</button>
    </div>
  </div></div>;
};

// ═══ NAV ═══
const NAV=[{id:"dash",l:"Overview"},{id:"recon",l:"Recon"},{id:"targets",l:"Targets"},{id:"mitm",l:"MITM"},{id:"twin",l:"Evil Twin"},{id:"portal",l:"Captive"},{id:"log",l:"Logging"},{id:"cfg",l:"Settings"}];

// ═══ MAIN — checks cookie on mount ═══
export default function App(){
  const[auth,setAuth]=useState(false);
  const[user,setUser]=useState("");

  useEffect(()=>{
    const session=getCookie("cerberus_auth");
    if(session&&session.user){setUser(session.user);setAuth(true);}
  },[]);

  const logout=()=>{delCookie("cerberus_auth");setAuth(false);setUser("");};

  if(!auth)return<AuthScreen onAuth={u=>{setUser(u);setAuth(true);}}/>;
  return<Dashboard user={user} onLogout={logout}/>;
}

function Dashboard({user,onLogout}){
const[pg,setPg]=useState("dash");const[aps,setAps]=useState([]);const[tAP,setTAP]=useState(null);const[cli,setCli]=useState([]);const[rcing,setRcing]=useState(false);const[scing,setScing]=useState(false);const[sel,setSel]=useState(new Set());const[mS,setMS]=useState(new Set());const[dS,setDS]=useState(new Set());const[dns,setDns]=useState([]);const[di,setDi]=useState(0);const[et,setEt]=useState({on:false,ssid:"",ch:"6"});const[cap,setCap]=useState({on:false,tpl:"google"});const[doh,setDoh]=useState(DOH_D);const[prb,setPrb]=useState([]);const[mCfg,setMCfg]=useState({arp:true,dns:true,ssl:false,spoof:false});const[lf,setLf]=useState("all");const[ls,setLs]=useState("");const[hsState,setHsState]=useState("idle");const[hsAP,setHsAP]=useState("");const[adR,setAdR]=useState({scan:"wlan1",attack:"wlan2",upstream:"wlan0"});const ML=2000;
const tSel=m=>setSel(p=>{const n=new Set(p);n.has(m)?n.delete(m):n.add(m);return n;});const aAll=()=>setSel(cli.length===sel.size?new Set():new Set(cli.map(c=>c.mac)));const tM=m=>setMS(p=>{const n=new Set(p);n.has(m)?n.delete(m):n.add(m);return n;});const tD=m=>setDS(p=>{const n=new Set(p);n.has(m)?n.delete(m):n.add(m);return n;});
const doRecon=()=>{setRcing(true);setAps([]);setTAP(null);setCli([]);setSel(new Set());setMS(new Set());setDS(new Set());setDns([]);setDi(0);setPrb([]);setHsState("idle");let i=0;const iv=setInterval(()=>{if(i<APS.length){const item=APS[i];i++;setAps(p=>[...p,item]);}else{clearInterval(iv);setRcing(false);}},300);};
const doScan=()=>{if(!tAP)return;setScing(true);setCli([]);setSel(new Set());setMS(new Set());setDS(new Set());setDns([]);setDi(0);setPrb([]);const pool=CLI_MAP[tAP.ssid]||[];let i=0;let pi=0;const iv=setInterval(()=>{if(i<pool.length){const item=pool[i];i++;setCli(p=>[...p,item]);if(pi<PRB.length&&Math.random()>0.4){const probe=PRB[pi];pi++;setPrb(p=>[...p,probe]);}}else{clearInterval(iv);setScing(false);}},400);};
const doHS=()=>{if(!hsAP)return;setHsState("listening");setTimeout(()=>{setHsState("deauthing");setTimeout(()=>{setHsState("captured");},2500);},3000);};
const dlCap=()=>{const d=`CERBERUS CAPTURE\nTarget: ${hsAP}\nTime: ${new Date().toISOString()}\nType: WPA2 4-Way\n\n[cap data]`;const b=new Blob([d],{type:"application/octet-stream"});const u=URL.createObjectURL(b);const a=document.createElement("a");a.href=u;a.download=`${hsAP.replace(/[^a-zA-Z0-9]/g,"_")}.cap`;a.click();URL.revokeObjectURL(u);};
useEffect(()=>{if(mS.size>0&&di<DNS_Q.length){const t=setTimeout(()=>{setDns(p=>{const n=[...p,DNS_Q[di]];return n.length>ML?n.slice(n.length-ML):n;});setDi(i=>i+1);},800+Math.random()*1200);return()=>clearTimeout(t);}},[mS.size,di]);
const fDns=dns.filter(e=>{if(lf==="blocked"&&e.s!=="blocked")return false;if(lf==="passed"&&e.s!=="pass")return false;if(ls&&!e.d.toLowerCase().includes(ls.toLowerCase()))return false;return true;});

const Dash=()=><div><div style={{display:"grid",gridTemplateColumns:"repeat(4,1fr)",gap:8,marginBottom:12}}>{[{l:"CLIENTS",v:cli.length,c:K.cy},{l:"MITM",v:mS.size,c:K.gn},{l:"DEAUTH",v:dS.size,c:K.rd},{l:"DNS",v:dns.length,c:K.pu}].map(s=><Bx key={s.l} s={{textAlign:"center",padding:10}}><div style={{fontSize:20,fontWeight:700,color:s.c}}>{s.v}</div><div style={{fontSize:7,letterSpacing:2,color:K.dm,marginTop:2}}>{s.l}</div></Bx>)}</div><div style={{display:"grid",gridTemplateColumns:"1fr 1fr",gap:8,marginBottom:8}}><Bx s={{padding:12}}><Lb>Target</Lb>{tAP?<div><div style={{fontSize:12,fontWeight:700,color:K.cy}}>{tAP.ssid}</div><div style={{fontSize:8,color:K.dm,marginTop:2}}>{tAP.bssid}</div></div>:<span style={{color:K.dm,fontSize:10}}>No target</span>}</Bx><Bx s={{padding:12}}><Lb>Active</Lb>{mS.size>0&&<div style={{fontSize:10,marginBottom:2}}><Dot c={K.cy}/> MITM: {mS.size}</div>}{dS.size>0&&<div style={{fontSize:10,marginBottom:2}}><Dot c={K.rd}/> Deauth: {dS.size}</div>}{et.on&&<div style={{fontSize:10}}><Dot c={K.pu}/> Twin: {et.ssid}</div>}{mS.size===0&&dS.size===0&&!et.on&&<span style={{color:K.dm,fontSize:10}}>None</span>}</Bx></div><Bx s={{padding:12,marginBottom:8}}><Lb>Devices</Lb><div style={{maxHeight:180,overflowY:"auto"}}>{cli.length>0?cli.map(c=><div key={c.mac} style={{display:"flex",justifyContent:"space-between",padding:"4px 0",borderBottom:`1px solid ${K.bx}`,fontSize:10}}><span style={{fontWeight:600}}>{c.host}</span><span style={{color:K.dm}}>{c.ip}</span><Bdg c={c.seen==="now"?K.gn:K.dm}>{c.seen==="now"?"LIVE":c.seen}</Bdg></div>):<Mt t="Scan to see devices"/>}</div></Bx><Bx s={{padding:12}}><Lb>Recent DNS</Lb>{dns.slice(-5).map((e,i)=><div key={i} style={{display:"flex",justifyContent:"space-between",padding:"3px 0",borderBottom:`1px solid ${K.bx}`,fontSize:10}}><span style={{color:K.dm,width:50}}>{e.t}</span><span style={{fontWeight:500,flex:1,marginLeft:6}}>{e.d}</span><Bdg c={e.s==="blocked"?K.rd:K.cy}>{e.s==="blocked"?"BLK":"OK"}</Bdg></div>)}{dns.length===0&&<Mt t="Start MITM to capture"/>}</Bx></div>;
const Recon=()=><div><Rw s={{justifyContent:"space-between",marginBottom:8}}><Lb>Recon</Lb><Bn fn={doRecon} dis={rcing} sm>{rcing?"SCANNING...":"RECON"}</Bn></Rw><Bx s={{marginBottom:8}}><Lb>APs ({aps.length})</Lb>{aps.map(a=><div key={a.bssid} onClick={()=>setTAP(a)} style={{display:"flex",alignItems:"center",gap:6,padding:"6px 0",borderBottom:`1px solid ${K.bx}`,cursor:"pointer",fontSize:10}}><Sg v={a.signal}/><span style={{flex:1,fontWeight:600,color:tAP?.bssid===a.bssid?K.cy:K.tx}}>{a.ssid}</span><span style={{color:K.dm,fontSize:8}}>{a.bssid}</span><Bdg c={a.enc==="Open"?K.rd:K.gn}>{a.enc}</Bdg></div>)}{aps.length===0&&<Mt t="Hit RECON"/>}</Bx><Bx s={{marginBottom:8}}><Lb>Handshake Capture</Lb><div style={{marginBottom:6}}><Sel value={hsAP} onChange={setHsAP} options={aps.filter(a=>a.enc!=="Open").map(a=>({v:a.ssid,l:`${a.ssid} [${a.enc}]`}))} ph="Select AP..."/></div><Rw s={{gap:4}}><Bn fn={doHS} dis={!hsAP||hsState==="listening"||hsState==="deauthing"} sm>{hsState==="idle"?"Capture":hsState==="captured"?"Done!":"..."}</Bn>{hsState==="captured"&&<Bn fn={dlCap} sm v="g">Download .cap</Bn>}</Rw></Bx><div style={{display:"grid",gridTemplateColumns:"1fr 1fr",gap:8}}><Bx><Lb>Clients ({cli.length})</Lb><div style={{maxHeight:180,overflowY:"auto"}}>{cli.map(c=><div key={c.mac} style={{display:"flex",justifyContent:"space-between",padding:"4px 0",borderBottom:`1px solid ${K.bx}`,fontSize:10}}><span style={{fontWeight:600}}>{c.host}</span><Sg v={c.sig}/></div>)}{cli.length===0&&<Mt t="Select AP"/>}</div></Bx><Bx><Lb>Probes ({prb.length})</Lb><div style={{maxHeight:180,overflowY:"auto"}}>{prb.map((p,i)=><div key={i} style={{display:"flex",justifyContent:"space-between",padding:"4px 0",borderBottom:`1px solid ${K.bx}`,fontSize:10}}><span>{p.ssid}</span><span style={{color:K.dm,fontSize:8}}>{p.mac.slice(-8)}</span></div>)}{prb.length===0&&<Mt t="During scan"/>}</div></Bx></div></div>;
const Targets=()=><div><Rw s={{justifyContent:"space-between",marginBottom:8,flexWrap:"wrap",gap:4}}><Lb>Targets{tAP?` — ${tAP.ssid}`:""}</Lb><Rw s={{gap:3}}><Bn fn={doScan} dis={scing||!tAP} sm>SCAN</Bn><Bn fn={()=>sel.forEach(m=>{if(!mS.has(m))tM(m);})} sm v="g">MITM</Bn><Bn fn={()=>sel.forEach(m=>{if(!dS.has(m))tD(m);})} sm v="d">Deauth</Bn></Rw></Rw>{!tAP?<Bx><Mt t="Select AP from top bar"/></Bx>:<Bx s={{padding:0,overflow:"hidden"}}><div style={{overflowX:"auto"}}><div style={{display:"grid",gridTemplateColumns:"24px 1.4fr 70px 100px 35px 38px 38px",gap:4,padding:"6px 8px",fontSize:7,fontWeight:700,letterSpacing:2,color:K.dm,borderBottom:`1px solid ${K.bx}`,minWidth:480}}><Ck on={sel.size===cli.length&&cli.length>0} fn={aAll}/><span>DEVICE</span><span>IP</span><span>MAC</span><span>SIG</span><span>M</span><span>D</span></div><div style={{maxHeight:300,overflowY:"auto"}}>{cli.map((c,x)=>{const s=sel.has(c.mac);return<div key={c.mac} style={{display:"grid",gridTemplateColumns:"24px 1.4fr 70px 100px 35px 38px 38px",gap:4,padding:"5px 8px",borderBottom:`1px solid ${K.bx}`,alignItems:"center",background:s?`${K.cy}04`:"transparent",minWidth:480,fontSize:10}}><Ck on={s} fn={()=>tSel(c.mac)}/><span style={{fontWeight:600,color:s?K.cy:K.tx}}>{c.host}</span><span style={{fontSize:9}}>{c.ip}</span><span style={{fontSize:7,color:K.dm}}>{c.mac}</span><Sg v={c.sig}/><Tog on={mS.has(c.mac)} fn={()=>tM(c.mac)} c={K.cy}/><Tog on={dS.has(c.mac)} fn={()=>tD(c.mac)} c={K.rd}/></div>;})}{cli.length===0&&<Mt t="Hit SCAN"/>}</div></div></Bx>}</div>;
const MitmPg=()=>{const act=cli.filter(c=>mS.has(c.mac));return<div><Lb>MITM</Lb><div style={{display:"grid",gridTemplateColumns:"1fr 1fr",gap:8}}><Bx><Lb>Targets</Lb><Sel value="" onChange={v=>{if(v)tM(v);}} options={cli.map(c=>({v:c.mac,l:`${c.host} — ${c.ip}`}))} ph="+ Add..."/><div style={{marginTop:6}}>{act.map(c=><Rw key={c.mac} s={{justifyContent:"space-between",padding:"4px 0",borderBottom:`1px solid ${K.bx}`}}><span style={{fontSize:10,fontWeight:600}}>{c.host}</span><Bn fn={()=>tM(c.mac)} sm v="d">Stop</Bn></Rw>)}{act.length===0&&<Mt t="Select targets"/>}</div></Bx><Bx><Lb>Config</Lb>{[{k:"arp",l:"ARP Spoof"},{k:"dns",l:"DNS Log"},{k:"ssl",l:"SSL Strip"},{k:"spoof",l:"DNS Spoof"}].map(a=><Rw key={a.k} s={{justifyContent:"space-between",padding:"4px 0",borderBottom:`1px solid ${K.bx}`}}><span style={{fontSize:10}}>{a.l}</span><Tog on={mCfg[a.k]} fn={()=>setMCfg(p=>({...p,[a.k]:!p[a.k]}))}/></Rw>)}</Bx></div></div>;};
const TwinPg=()=><div><Lb>Evil Twin</Lb><div style={{display:"grid",gridTemplateColumns:"1fr 1fr",gap:8}}><Bx><Lb>Clone AP</Lb><div style={{marginBottom:6}}><Sel value={et.ssid} onChange={v=>setEt(p=>({...p,ssid:v}))} options={aps.map(a=>({v:a.ssid,l:`${a.ssid} [CH${a.ch}]`}))} ph="Select..."/></div><div style={{marginBottom:6}}><Sel value={et.ch} onChange={v=>setEt(p=>({...p,ch:v}))} options={[1,6,11,36,40,44].map(c=>({v:String(c),l:`CH ${c}`}))}/></div><Bn fn={()=>setEt(p=>({...p,on:!p.on}))} v={et.on?"d":"p"} sx={{width:"100%"}}>{et.on?"STOP":"LAUNCH"}</Bn></Bx><Bx><Lb>Probes</Lb>{prb.length>0?[...new Set(prb.map(p=>p.ssid))].map(s=><div key={s} style={{padding:"3px 0",borderBottom:`1px solid ${K.bx}`,fontSize:10}}>{s}</div>):<Mt t="Run recon"/>}</Bx></div></div>;
const PortalPg=()=><div><Lb>Captive Portal</Lb><div style={{display:"grid",gridTemplateColumns:"1fr 1fr",gap:8}}><Bx><Lb>Template</Lb><Sel value={cap.tpl} onChange={v=>setCap(p=>({...p,tpl:v}))} options={[{v:"google",l:"Google"},{v:"facebook",l:"Facebook"},{v:"hotel",l:"Hotel"},{v:"custom",l:"Custom"}]}/><div style={{marginTop:8}}><Bn fn={()=>setCap(p=>({...p,on:!p.on}))} v={cap.on?"d":"p"} sx={{width:"100%"}}>{cap.on?"STOP":"LAUNCH"}</Bn></div></Bx><Bx><Lb>Captured</Lb><Mt t={cap.on?"Waiting...":"Not active"}/></Bx></div></div>;
const LogPg=()=><div><Rw s={{justifyContent:"space-between",marginBottom:6,flexWrap:"wrap",gap:4}}><Lb>DNS ({fDns.length})</Lb><Rw s={{gap:3}}>{["all","passed","blocked"].map(f=><button key={f} onClick={()=>setLf(f)} style={{fontFamily:K.f,fontSize:8,fontWeight:600,padding:"2px 6px",borderRadius:3,cursor:"pointer",border:`1px solid ${lf===f?K.cy+"30":K.bx}`,background:lf===f?K.cy+"10":"transparent",color:lf===f?K.cy:K.dm}}>{f}</button>)}<Bn fn={()=>{setDns([]);setDi(0);}} sm v="g">Clear</Bn></Rw></Rw><div style={{marginBottom:5}}><input value={ls} onChange={e=>setLs(e.target.value)} placeholder="Filter..." style={{background:K.ra,color:K.tx,border:`1px solid ${K.bx}`,borderRadius:4,padding:"5px 8px",fontSize:9,fontFamily:K.f,width:"100%",outline:"none",boxSizing:"border-box"}}/></div><Bx s={{padding:0}}><div style={{maxHeight:400,overflowY:"auto"}}>{fDns.map((e,i)=><div key={i} style={{display:"flex",justifyContent:"space-between",padding:"4px 10px",fontSize:10,borderBottom:`1px solid ${K.bx}`}}><span style={{color:K.dm,width:50}}>{e.t}</span><span style={{flex:1,fontWeight:500}}>{e.d}</span><Bdg c={e.s==="blocked"?K.rd:K.cy}>{e.s==="blocked"?"BLK":"OK"}</Bdg></div>)}{fDns.length===0&&<Mt t="No data"/>}</div></Bx></div>;
const CfgPg=()=><div><Lb>Settings</Lb><div style={{display:"grid",gridTemplateColumns:"1fr 1fr",gap:8}}><Bx><Lb>Adapters</Lb>{ADAPTERS.map(a=>{const role=adR.scan===a.id?"scan":adR.attack===a.id?"attack":"upstream";return<div key={a.id} style={{padding:"6px 0",borderBottom:`1px solid ${K.bx}`}}><Rw s={{justifyContent:"space-between"}}><span style={{fontSize:10,fontWeight:700,color:K.cy}}>{a.id}</span><Bdg c={role==="scan"?K.cy:role==="attack"?K.rd:K.gn}>{role}</Bdg></Rw><div style={{fontSize:8,color:K.dm,marginTop:2}}>{a.chip} | {a.band} GHz</div></div>;})}</Bx><Bx><Lb>DoH Blocking</Lb>{doh.map((s,i)=><Rw key={s.ip} s={{justifyContent:"space-between",padding:"3px 0",borderBottom:`1px solid ${K.bx}`}}><span style={{fontSize:10}}>{s.n} <span style={{color:K.dm,fontSize:8}}>{s.ip}</span></span><Tog on={s.on} fn={()=>{const n=[...doh];n[i]={...n[i],on:!n[i].on};setDoh(n);}} c={K.rd}/></Rw>)}</Bx></div><div style={{marginTop:12}}><Bn fn={onLogout} v="d" sx={{width:"100%"}}>LOGOUT</Bn></div></div>;

const pages={dash:Dash,recon:Recon,targets:Targets,mitm:MitmPg,twin:TwinPg,portal:PortalPg,log:LogPg,cfg:CfgPg};const Pg=pages[pg];
return<div style={{display:"flex",flexDirection:"column",height:"100vh",background:K.bg,color:K.tx,fontFamily:K.f}}><style>{CSS}</style>
<div style={{display:"flex",alignItems:"center",gap:6,padding:"6px 12px",background:K.sf,borderBottom:`1px solid ${K.bx}`,flexShrink:0}}>
<span style={{fontSize:12,fontWeight:700,letterSpacing:4,color:K.cy,marginRight:4}}>CERBERUS</span>
<div style={{width:1,height:18,background:K.bx}}/>
<APDrop aps={aps} sel={tAP} onSel={setTAP}/>
<Bn fn={doRecon} dis={rcing} sm>{rcing?"...":"RECON"}</Bn>
<Bn fn={doScan} dis={scing||!tAP} sm>{scing?"...":"SCAN"}</Bn>
<div style={{flex:1}}/>
<Rw><Dot c={K.gn}/><span style={{fontSize:8,color:K.dm}}>{user}</span></Rw>
</div>
<div style={{display:"flex",flex:1,overflow:"hidden"}}>
<div style={{width:110,background:K.sf,borderRight:`1px solid ${K.bx}`,padding:"6px 0",flexShrink:0}}>
{NAV.map(n=>{const a=pg===n.id;return<div key={n.id} onClick={()=>setPg(n.id)} style={{padding:"6px 10px",cursor:"pointer",fontSize:10,fontWeight:a?600:400,color:a?K.cy:K.dm,background:a?`${K.cy}05`:"transparent",borderLeft:a?`2px solid ${K.cy}`:"2px solid transparent"}}>{n.l}</div>;})}
</div>
<div style={{flex:1,padding:12,overflowY:"auto"}}><Pg/></div>
</div></div>;}
