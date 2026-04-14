// Authoritative API bridge for UTSAV v1.5
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

// --- Host Auth Management ---
const tokenKey = "utsav_access_token";
const refreshKey = "utsav_refresh_token";
const authCookieKey = "utsav_access_token";

export function getAccessToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(tokenKey);
}

export function setTokens(access: string, refresh?: string) {
  if (typeof window === "undefined") return;
  localStorage.setItem(tokenKey, access);
  if (refresh) localStorage.setItem(refreshKey, refresh);
  document.cookie = `${authCookieKey}=${encodeURIComponent(access)}; path=/; samesite=lax`;
}

export function clearTokens() {
  if (typeof window === "undefined") return;
  localStorage.removeItem(tokenKey);
  localStorage.removeItem(refreshKey);
  document.cookie = `${authCookieKey}=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT; samesite=lax`;
}

function getRefreshToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(refreshKey);
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
type ApiOptions = RequestInit & { json?: Record<string, unknown> | unknown }

export class ApiError extends Error {
  status: number
  code?: string

  constructor(message: string, status: number, code?: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
}

async function performFetch<T>(
  endpoint: string,
  options: ApiOptions,
  token: string | null
): Promise<T> {
  const headers = new Headers(options.headers)

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

  if (!response.ok) throw response
  return response.json() as Promise<T>
}

async function parseApiError(response: Response): Promise<ApiError> {
  const payload = (await response.json().catch(() => ({ error: 'Unknown API error' }))) as {
    error?: string
    code?: string
  }
  const message = payload.error || `HTTP ${response.status}`
  return new ApiError(message, response.status, payload.code)
}

async function refreshAccessToken(): Promise<boolean> {
  const refresh = getRefreshToken()
  if (!refresh) return false
  try {
    const response = await fetch(`${API_BASE_URL}/v1/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refresh }),
    })
    if (!response.ok) return false
    const data = (await response.json()) as { access_token: string; refresh_token?: string }
    setTokens(data.access_token, data.refresh_token)
    return true
  } catch {
    return false
  }
}

export async function apiFetch<T>(
  endpoint: string,
  options: ApiOptions = {}
): Promise<T> {
  try {
    return await performFetch<T>(endpoint, { ...options }, getAccessToken())
  } catch (rawErr) {
    const response = rawErr as Response
    if (response?.status !== 401) {
      throw await parseApiError(response)
    }

    const refreshed = await refreshAccessToken()
    if (!refreshed) {
      clearTokens()
      throw new Error('Session expired. Please log in again.')
    }

    try {
      return await performFetch<T>(endpoint, { ...options }, getAccessToken())
    } catch (retryErr) {
      const retryResponse = retryErr as Response
      throw await parseApiError(retryResponse)
    }
  }
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
