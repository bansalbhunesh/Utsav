'use client'

import { useEffect } from 'react'
import { supabase } from '@/lib/supabase/client'
import { useAuthStore } from '@/store/auth-store'

export default function AuthProvider({ children }: { children: React.ReactNode }) {
  const { setUser, setLoading } = useAuthStore()

  useEffect(() => {
    // Initial fetch
    supabase.auth.getUser().then(({ data: { user } }) => {
      setUser(user)
      setLoading(false)
    })

    // Listen for changes
    const { data: { subscription } } = supabase.auth.onAuthStateChange((_event, session) => {
      setUser(session?.user ?? null)
      setLoading(false)
    })

    return () => {
      subscription.unsubscribe()
    }
  }, [setUser, setLoading])

  return <>{children}</>
}
