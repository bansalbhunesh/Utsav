import { apiFetch, clearTokens } from './api'

export const signInWithPhone = async (phone: string) => {
  try {
    const data = await apiFetch<{ ok: boolean }>('/v1/auth/otp/request', {
      method: 'POST',
      json: { phone },
    })
    return { data, error: null }
  } catch (error: unknown) {
    return { data: null, error }
  }
}

export const verifyOtp = async (phone: string, token: string) => {
  try {
    const data = await apiFetch<{ user_id: string; authenticated: boolean }>('/v1/auth/otp/verify', {
      method: 'POST',
      json: { phone, code: token },
    })
    return { data, error: null }
  } catch (error: unknown) {
    return { data: null, error }
  }
}

export const signOut = async () => {
  clearTokens()
  return { error: null }
}
