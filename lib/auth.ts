import { supabase } from './supabase/client'

export const signInWithPhone = async (phone: string) => {
  const { data, error } = await supabase.auth.signInWithOtp({
    phone,
    options: {
      // For India, we usually need the +91 prefix.
      // We can handle formatting in the UI.
    }
  })
  return { data, error }
}

export const verifyOtp = async (phone: string, token: string) => {
  const { data, error } = await supabase.auth.verifyOtp({
    phone,
    token,
    type: 'sms'
  })
  return { data, error }
}

export const signOut = async () => {
  const { error } = await supabase.auth.signOut()
  return { error }
}
