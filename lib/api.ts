// Authoritative API bridge for UTSAV v1.5
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

// --- Host Auth Management ---
const tokenKey = "utsav_access_token";
const refreshKey = "utsav_refresh_token";

export function getAccessToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(tokenKey);
}

export function setTokens(access: string, refresh?: string) {
  if (typeof window === "undefined") return;
  localStorage.setItem(tokenKey, access);
  if (refresh) localStorage.setItem(refreshKey, refresh);
}

// --- Guest Session Management ---
const guestTokenKey = "utsav_guest_token";

export const setGuestToken = (token: string) => {
  if (typeof window !== "undefined") {
    localStorage.setItem(guestTokenKey, token);
  }
}

export const getGuestToken = (): string | null => {
  if (typeof window !== "undefined") {
    return localStorage.getItem(guestTokenKey);
  }
  return null;
}

// --- Authorized API Fetcher (Host) ---
export async function apiFetch<T>(
  endpoint: string, 
  options: RequestInit & { json?: Record<string, unknown> | unknown } = {}
): Promise<T> {
  const headers = new Headers(options.headers)
  const token = getAccessToken()
  
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  if (options.json) {
    headers.set('Content-Type', 'application/json')
    options.body = JSON.stringify(options.json)
  }

  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers,
  })

  if (!response.ok) {
    const error = (await response.json().catch(() => ({ error: 'Unknown API error' }))) as { error?: string }
    throw new Error(error.error || `HTTP ${response.status}`)
  }

  return response.json()
}

// --- Guest API Fetcher (RSVP/OTP) ---
export async function guestApiFetch<T>(
  endpoint: string, 
  options: RequestInit & { json?: Record<string, unknown> | unknown } = {}
): Promise<T> {
  const headers = new Headers(options.headers)
  const token = getGuestToken()
  
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }
  
  if (options.json !== undefined) {
    headers.set('Content-Type', 'application/json')
    options.body = JSON.stringify(options.json)
  }

  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers,
  })

  if (!response.ok) {
    const error = (await response.json().catch(() => ({ error: 'Unknown API error' }))) as { error?: string }
    throw new Error(error.error || `HTTP ${response.status}`)
  }

  return response.json()
}
