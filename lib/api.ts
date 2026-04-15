// Authoritative API bridge for UTSAV v1.5
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

export function clearTokens() {
  if (typeof window === "undefined") return
  fetch(`${API_BASE_URL}/v1/auth/logout`, {
    method: 'POST',
    credentials: 'include',
  }).catch(() => {
    // ignore logout network errors on client cleanup
  })
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
  options: ApiOptions
): Promise<T> {
  const headers = new Headers(options.headers)

  if (options.json !== undefined) {
    headers.set('Content-Type', 'application/json')
    options.body = JSON.stringify(options.json)
  }

  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers,
    credentials: 'include',
  })

  if (!response.ok) throw response
  return response.json() as Promise<T>
}

async function parseApiError(response: Response): Promise<ApiError> {
  const payload = (await response.json().catch(() => ({ error: 'Unknown API error' }))) as {
    error?: string | { code?: string; message?: string }
    code?: string
    message?: string
  }

  if (typeof payload.error === 'object' && payload.error) {
    const message = payload.error.message || `HTTP ${response.status}`
    return new ApiError(message, response.status, payload.error.code || payload.code)
  }

  const message = payload.message || payload.error || `HTTP ${response.status}`
  return new ApiError(message, response.status, payload.code)
}

async function refreshAccessToken(): Promise<boolean> {
  try {
    const response = await fetch(`${API_BASE_URL}/v1/auth/refresh`, {
      method: 'POST',
      credentials: 'include',
    })
    if (!response.ok) return false
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
    return await performFetch<T>(endpoint, { ...options })
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
      return await performFetch<T>(endpoint, { ...options })
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
    throw await parseApiError(response)
  }

  return response.json()
}
