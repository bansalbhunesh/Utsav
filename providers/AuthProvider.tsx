'use client'

import { useEffect } from 'react'
import { apiFetch, getAccessToken } from '@/lib/api'
import { useAuthStore } from '@/store/auth-store'
import type { User } from '@supabase/supabase-js'

export default function AuthProvider({ children }: { children: React.ReactNode }) {
  const { setUser, setLoading } = useAuthStore()

  useEffect(() => {
    const bootstrapAuth = async () => {
      const token = getAccessToken()
      if (!token) {
        setUser(null)
        setLoading(false)
        return
      }

      try {
        const me = await apiFetch<{ id: string; phone: string; display_name?: string }>('/v1/me')
        const user = {
          id: me.id,
          phone: me.phone,
          user_metadata: { display_name: me.display_name || '' },
        } as User
        setUser(user)
      } catch (err) {
        console.error('Failed to hydrate auth session:', err)
        setUser(null)
      } finally {
        setLoading(false)
      }
    }

    void bootstrapAuth()
  }, [setUser, setLoading])

  return <>{children}</>
}
