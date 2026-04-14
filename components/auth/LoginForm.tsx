'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { signInWithPhone, verifyOtp } from '@/lib/auth'
import { Loader2, Phone, KeyRound } from 'lucide-react'

export function LoginForm() {
  const [phone, setPhone] = useState('')
  const [otp, setOtp] = useState('')
  const [step, setStep] = useState<'phone' | 'otp'>('phone')
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSendOtp = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)
    setError(null)

    // Ensure +91 prefix for Indian numbers if not present
    const formattedPhone = phone.startsWith('+') ? phone : `+91${phone}`
    
    const { error: signInError } = await signInWithPhone(formattedPhone)
    
    setIsLoading(false)
    if (signInError) {
      setError(signInError.message)
    } else {
      setStep('otp')
    }
  }

  const handleVerifyOtp = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)
    setError(null)

    const formattedPhone = phone.startsWith('+') ? phone : `+91${phone}`
    const { error: verifyError } = await verifyOtp(formattedPhone, otp)

    setIsLoading(false)
    if (verifyError) {
      setError(verifyError.message)
    } else {
      // Success will be handled by the AuthProvider listener
    }
  }

  return (
    <div className="w-full max-w-sm mx-auto p-6 space-y-8 bg-white rounded-3xl border border-zinc-100 shadow-xl shadow-orange-100/50">
      <div className="space-y-2 text-center">
        <h2 className="text-3xl font-bold font-heading tracking-tight text-zinc-900 leading-tight">
          {step === 'phone' ? 'Welcome to UTSAV' : 'Verify Phone'}
        </h2>
        <p className="text-zinc-500">
          {step === 'phone' 
            ? 'Enter your mobile number to get started' 
            : `Enter the 6-digit code sent to ${phone}`}
        </p>
      </div>

      <form onSubmit={step === 'phone' ? handleSendOtp : handleVerifyOtp} className="space-y-6">
        {step === 'phone' ? (
          <div className="space-y-2">
            <Label htmlFor="phone" className="text-zinc-700 font-medium">Mobile Number</Label>
            <div className="relative">
              <Phone className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-400" />
              <Input
                id="phone"
                type="tel"
                placeholder="9876543210"
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                className="pl-10 h-12 rounded-xl border-zinc-200 focus:border-orange-500 focus:ring-orange-500"
                required
              />
            </div>
          </div>
        ) : (
          <div className="space-y-2">
            <Label htmlFor="otp" className="text-zinc-700 font-medium">One-Time Password</Label>
            <div className="relative">
              <KeyRound className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-400" />
              <Input
                id="otp"
                type="text"
                maxLength={6}
                placeholder="000000"
                value={otp}
                onChange={(e) => setOtp(e.target.value)}
                className="pl-10 h-12 rounded-xl border-zinc-200 focus:border-orange-500 focus:ring-orange-500 tracking-[0.5em] text-center font-bold"
                required
              />
            </div>
          </div>
        )}

        {error && (
          <p className="text-sm text-red-500 font-medium bg-red-50 p-3 rounded-xl border border-red-100">
            {error}
          </p>
        )}

        <Button 
          type="submit" 
          disabled={isLoading}
          className="w-full h-12 bg-orange-600 hover:bg-orange-700 text-white font-bold rounded-xl transition-all shadow-lg shadow-orange-200"
        >
          {isLoading ? (
            <Loader2 className="h-5 w-5 animate-spin" />
          ) : (
            step === 'phone' ? 'Get OTP' : 'Verify & Continue'
          )}
        </Button>

        {step === 'otp' && (
          <Button
            type="button"
            variant="ghost"
            onClick={() => setStep('phone')}
            className="w-full text-zinc-500 hover:text-orange-600"
          >
            Change Phone Number
          </Button>
        )}
      </form>
      
      <p className="text-xs text-center text-zinc-400 px-6">
        By continuing, you agree to Utsav's Terms of Service and Privacy Policy.
      </p>
    </div>
  )
}
