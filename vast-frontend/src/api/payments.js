const PAYMENTS_URL = import.meta.env.VITE_PAYMENTS_URL || "http://localhost:8083";
const SIMBRIDGE_URL = import.meta.env.VITE_SIMBRIDGE_URL || "http://localhost:9090";

async function request(base, path, opts = {}) {
  const res = await fetch(`${base}${path}`, {
    headers: { "Content-Type": "application/json", ...opts.headers },
    ...opts,
  });
  const body = res.status === 204 ? null : await res.json().catch(() => null);
  if (!res.ok) {
    const msg = body?.message || body?.code || res.statusText;
    throw new Error(msg);
  }
  return body;
}

export function fetchCatalog() {
  return request(PAYMENTS_URL, "/payments/catalog");
}

export function fetchPaymentMethods(userId) {
  return request(PAYMENTS_URL, `/payments/users/${encodeURIComponent(userId)}/methods`);
}

export function savePaymentMethod(userId, method) {
  return request(PAYMENTS_URL, `/payments/users/${encodeURIComponent(userId)}/methods`, {
    method: "POST",
    body: JSON.stringify(method),
  });
}

export function deletePaymentMethod(userId, methodId) {
  return request(PAYMENTS_URL, `/payments/users/${encodeURIComponent(userId)}/methods/${methodId}`, {
    method: "DELETE",
  });
}

export function startTopup(payload) {
  return request(PAYMENTS_URL, "/payments/topup", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function confirmTopup(topupId) {
  return request(PAYMENTS_URL, `/payments/${topupId}/confirm`, { method: "POST" });
}

/** Dial USSD on the local device via simbridge (phone or laptop SIM modem). */
export async function dialUSSD({ carrier, phone, amount, ussd }) {
  return request(SIMBRIDGE_URL, "/simbridge/dial", {
    method: "POST",
    body: JSON.stringify({ carrier, phone, amount, ussd }),
  });
}

export async function fetchLocalSIMs() {
  try {
    return await request(SIMBRIDGE_URL, "/simbridge/sims");
  } catch {
    return [];
  }
}

export function isSimbridgeAvailable() {
  return fetchLocalSIMs().then((sims) => sims.length > 0).catch(() => false);
}

/** Build USSD client-side when simbridge is offline (display-only fallback). */
export function buildUSSD(carrier, phone, amountUnits) {
  const digits = String(phone).replace(/\D/g, "");
  const templates = {
    mtn: "*126*16*{phone}*{amount}#",
    orange: "*144*16*{phone}*{amount}#",
  };
  const tpl = templates[String(carrier).toLowerCase()];
  if (!tpl || !digits || amountUnits <= 0) return null;
  return tpl.replace("{phone}", digits).replace("{amount}", String(Math.round(amountUnits)));
}
