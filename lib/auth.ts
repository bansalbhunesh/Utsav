import { apiFetch, clearTokens, setTokens } from './api'

export const signInWithPhone = async (phone: string) => {
  try {
    const data = await apiFetch<{ ok: boolean }>('/v1/auth/otp/request', {
      method: 'POST',
      json: { phone },
    })
    return { data, error: null }
  } catch (error: any) {
    return { data: null, error }
  }
}

export const verifyOtp = async (phone: string, token: string) => {
  try {
    const data = await apiFetch<{ access_token: string; refresh_token?: string }>('/v1/auth/otp/verify', {
      method: 'POST',
      json: { phone, code: token },
    })
    setTokens(data.access_token, data.refresh_token)
    return { data, error: null }
  } catch (error: any) {
    return { data: null, error }
  }
}

export const signOut = async () => {
  clearTokens()
  return { error: null }
}
