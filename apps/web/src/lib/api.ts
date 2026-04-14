const tokenKey = "utsav_access_token";
const refreshKey = "utsav_refresh_token";
const guestTokenKey = "utsav_guest_token";

export function getAccessToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(tokenKey);
}

export function setTokens(access: string, refresh?: string) {
  localStorage.setItem(tokenKey, access);
  if (refresh) localStorage.setItem(refreshKey, refresh);
}

export function getGuestToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(guestTokenKey);
}

export function setGuestToken(t: string) {
  localStorage.setItem(guestTokenKey, t);
}

export async function guestApiFetch<T>(
  path: string,
  init: RequestInit & { json?: unknown } = {},
): Promise<T> {
  const headers = new Headers(init.headers);
  if (init.json !== undefined) headers.set("Content-Type", "application/json");
  const g = getGuestToken();
  if (g) headers.set("Authorization", `Bearer ${g}`);
  const res = await fetch(path, {
    ...init,
    headers,
    body: init.json !== undefined ? JSON.stringify(init.json) : init.body,
  });
  if (!res.ok) throw new Error(await res.text());
  return (await res.json()) as T;
}

export async function apiFetch<T>(
  path: string,
  init: RequestInit & { json?: unknown } = {},
): Promise<T> {
  const headers = new Headers(init.headers);
  if (init.json !== undefined) headers.set("Content-Type", "application/json");
  const token = getAccessToken();
  if (token) headers.set("Authorization", `Bearer ${token}`);
  const res = await fetch(path, {
    ...init,
    headers,
    body: init.json !== undefined ? JSON.stringify(init.json) : init.body,
  });
  if (!res.ok) throw new Error(await res.text());
  return (await res.json()) as T;
}
