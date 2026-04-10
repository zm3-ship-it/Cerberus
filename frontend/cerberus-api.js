// cerberus-api.js
// Drop-in API client for Cerberus dashboard
// Replace mock data hooks with these functions

const BASE = window.location.port === '1471' 
  ? '' 
  : 'http://192.168.1.1:1471';

async function api(path, method = 'GET', body = null) {
  const opts = {
    method,
    headers: { 'Content-Type': 'application/json' },
  };
  if (body) opts.body = JSON.stringify(body);

  const res = await fetch(`${BASE}/api${path}`, opts);
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || 'Request failed');
  }
  return res.json();
}

// ═══════════════════════════════════
// SCANNER
// ═══════════════════════════════════

export async function startScan() {
  return api('/scan', 'POST');
}

export async function getClients() {
  return api('/clients');
}

export async function getNetworks() {
  return api('/networks');
}

export async function getProbes() {
  return api('/probes');
}

export async function getScanStatus() {
  return api('/status');
}

// ═══════════════════════════════════
// MITM
// ═══════════════════════════════════

export async function startMitm(mac, ip) {
  return api('/mitm/start', 'POST', { mac, ip });
}

export async function stopMitm(mac) {
  return api('/mitm/stop', 'POST', { mac });
}

export async function getMitmTargets() {
  return api('/mitm/targets');
}

export async function getDnsLog() {
  return api('/mitm/dns');
}

// ═══════════════════════════════════
// DEAUTH
// ═══════════════════════════════════

export async function startDeauth(mac, bssid) {
  return api('/deauth/start', 'POST', { mac, bssid });
}

export async function stopDeauth(mac) {
  return api('/deauth/stop', 'POST', { mac });
}

export async function getDeauthTargets() {
  return api('/deauth/targets');
}

// ═══════════════════════════════════
// EVIL TWIN
// ═══════════════════════════════════

export async function startEvilTwin(ssid, channel, iface) {
  return api('/eviltwin/start', 'POST', { ssid, channel, interface: iface });
}

export async function stopEvilTwin() {
  return api('/eviltwin/stop', 'POST');
}

export async function getEvilTwinStatus() {
  return api('/eviltwin/status');
}

// ═══════════════════════════════════
// CAPTIVE PORTAL
// ═══════════════════════════════════

export async function startCaptive(template) {
  return api('/captive/start', 'POST', { template });
}

export async function stopCaptive() {
  return api('/captive/stop', 'POST');
}

export async function getCapturedCreds() {
  return api('/captive/creds');
}

// ═══════════════════════════════════
// HANDSHAKE CAPTURE
// ═══════════════════════════════════

export async function startHandshake(bssid, channel, iface) {
  return api('/handshake/start', 'POST', { bssid, channel, interface: iface });
}

export async function getHandshakeStatus() {
  return api('/handshake/status');
}

export async function downloadHandshake(filename) {
  const res = await fetch(`${BASE}/api/handshake/download/${filename}`);
  if (!res.ok) throw new Error('Download failed');
  return res.blob();
}

// ═══════════════════════════════════
// ADAPTERS
// ═══════════════════════════════════

export async function getAdapters() {
  return api('/adapters');
}

export async function setAdapterRole(adapterId, role) {
  return api('/adapters/role', 'POST', { adapter: adapterId, role });
}

// ═══════════════════════════════════
// SYSTEM
// ═══════════════════════════════════

export async function getSystemStatus() {
  return api('/status');
}

export async function getDohConfig() {
  return api('/config/doh');
}

export async function setDohBlock(ip, blocked) {
  return api('/config/doh', 'POST', { ip, blocked });
}

// ═══════════════════════════════════
// POLLING HELPER
// ═══════════════════════════════════

// Use this to poll endpoints at intervals
// Returns a cleanup function to stop polling
export function poll(fn, intervalMs = 2000) {
  let active = true;
  const run = async () => {
    while (active) {
      try { await fn(); } catch (e) { console.warn('Poll error:', e); }
      await new Promise(r => setTimeout(r, intervalMs));
    }
  };
  run();
  return () => { active = false; };
}

// Example usage in React:
// useEffect(() => {
//   const stop = poll(async () => {
//     const log = await getDnsLog();
//     setDns(log);
//   }, 1500);
//   return stop;
// }, [mitmActive]);
