'use client'

import { useState } from 'react'
import Link from 'next/link'
import { ChevronLeft, Loader2, Sparkles, ShieldCheck, Smartphone } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { apiFetch } from '@/lib/api'

export default function LoginPage() {
  const [phone, setPhone] = useState('')
  const [code, setCode] = useState('')
  const [step, setStep] = useState<'PHONE' | 'OTP'>('PHONE')
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleRequestOTP = async () => {
    if (!phone) return
    setIsLoading(true)
    setError(null)
    try {
      await apiFetch('/v1/auth/otp/request', { 
        method: 'POST', 
        json: { phone } 
      })
      setStep('OTP')
    } catch {
      setError('Failed to send OTP. Is your phone number correct?')
    } finally {
      setIsLoading(false)
    }
  }

  const handleVerifyOTP = async () => {
    if (!code) return
    setIsLoading(true)
    setError(null)
    try {
      await apiFetch<{ user_id: string; authenticated: boolean }>(
        '/v1/auth/otp/verify',
        { method: 'POST', json: { phone, code } }
      )
      // Redirect to dashboard
      window.location.href = '/dashboard'
    } catch {
      setError('Invalid code. Please try again.')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex flex-col lg:flex-row bg-white">
      {/* Brand Side */}
      <div className="hidden lg:flex lg:w-1/2 bg-orange-600 p-12 flex-col justify-between text-white relative overflow-hidden">
        <div className="relative z-10">
          <Link href="/" className="flex items-center gap-2 mb-12">
            <div className="w-10 h-10 bg-white rounded-xl flex items-center justify-center text-orange-600 font-bold text-2xl">U</div>
            <span className="text-2xl font-bold tracking-tight">UTSAV</span>
          </Link>
          
          <div className="space-y-6 max-w-lg">
             <h1 className="text-6xl font-bold leading-tight tracking-tight">
               Manage Events with <br />
               <span className="text-orange-200 italic underline decoration-white/20 underline-offset-8">Precision.</span>
             </h1>
             <p className="text-xl text-orange-100 font-medium">
               The Relational Ledger v1.5 enables secure auth for organisers and hosts.
             </p>
          </div>
        </div>

        <div className="relative z-10 grid grid-cols-2 gap-8">
           <div className="space-y-2">
              <ShieldCheck className="w-8 h-8 text-orange-200" />
              <h3 className="font-bold text-lg">Authoritative API</h3>
              <p className="text-sm text-orange-100">Every action is verified by our Go-Backend.</p>
           </div>
           <div className="space-y-2">
              <Smartphone className="w-8 h-8 text-orange-200" />
              <h3 className="font-bold text-lg">Instant OTP</h3>
              <p className="text-sm text-orange-100">Zero password vulnerability for your guests.</p>
           </div>
        </div>

        {/* Decorative Circles */}
        <div className="absolute top-0 right-0 w-[800px] h-[800px] bg-orange-500 rounded-full -translate-y-1/2 translate-x-1/2 opacity-50" />
        <div className="absolute bottom-0 left-0 w-[400px] h-[400px] bg-orange-700 rounded-full translate-y-1/2 -translate-x-1/2 opacity-30" />
      </div>

      {/* Form Side */}
      <div className="flex-1 flex flex-col p-6 lg:p-12">
        <header className="flex justify-between items-center mb-12 lg:mb-20">
          <Link href="/" className="inline-flex items-center text-sm font-bold text-zinc-400 hover:text-orange-600 transition-colors uppercase tracking-widest">
            <ChevronLeft className="mr-1 h-4 w-4" />
            Home
          </Link>
          <div className="lg:hidden flex items-center gap-2">
            <div className="w-8 h-8 bg-orange-600 rounded-lg flex items-center justify-center text-white font-bold">U</div>
            <span className="font-bold tracking-tight text-zinc-900">UTSAV</span>
          </div>
        </header>

        <main className="flex-1 flex flex-col justify-center max-w-sm mx-auto w-full space-y-8">
           <div className="space-y-2">
              <div className="inline-flex items-center gap-2 px-3 py-1 bg-orange-50 text-orange-600 rounded-full text-[10px] font-bold uppercase tracking-widest mb-2">
                <Sparkles className="w-3 h-3" /> Secure Access
              </div>
              <h2 className="text-4xl font-bold text-zinc-900 tracking-tight">Step Into Your Event</h2>
              <p className="text-zinc-500 font-medium">Verify your phone to access your dashboard.</p>
           </div>

           {error && (
             <div className="p-4 bg-red-50 border border-red-100 rounded-2xl text-red-600 text-sm font-bold animate-in fade-in zoom-in-95">
                {error}
             </div>
           )}

           <div className="space-y-6">
              {step === 'PHONE' ? (
                <div className="space-y-4">
                  <div className="space-y-2">
                    <label className="text-[10px] font-bold text-zinc-400 uppercase tracking-widest ml-1">Phone Number</label>
                    <Input 
                      placeholder="+91 9876543210" 
                      value={phone}
                      onChange={(e) => setPhone(e.target.value)}
                      className="h-14 rounded-2xl border-zinc-100 bg-zinc-50 text-lg font-medium focus:ring-orange-600"
                    />
                  </div>
                  <Button 
                    onClick={handleRequestOTP}
                    disabled={isLoading || !phone}
                    className="w-full h-14 bg-orange-600 hover:bg-orange-700 text-white font-bold rounded-2xl shadow-xl shadow-orange-100"
                  >
                    {isLoading ? <Loader2 className="w-6 h-6 animate-spin" /> : 'Get Verification Code'}
                  </Button>
                </div>
              ) : (
                <div className="space-y-4 animate-in slide-in-from-right-4 duration-500">
                  <div className="space-y-2">
                    <label className="text-[10px] font-bold text-zinc-400 uppercase tracking-widest ml-1">6-Digit OTP</label>
                    <Input 
                      placeholder="000000" 
                      value={code}
                      onChange={(e) => setCode(e.target.value)}
                      className="h-14 rounded-2xl border-zinc-100 bg-zinc-50 text-center text-3xl font-bold tracking-[0.5em] focus:ring-orange-600"
                    />
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <Button 
                      variant="ghost" 
                      onClick={() => setStep('PHONE')}
                      className="h-14 rounded-2xl text-zinc-400 font-bold hover:bg-zinc-100"
                    >
                      Back
                    </Button>
                    <Button 
                      onClick={handleVerifyOTP}
                      disabled={isLoading || !code}
                      className="h-14 bg-zinc-900 hover:bg-black text-white font-bold rounded-2xl shadow-xl"
                    >
                      {isLoading ? <Loader2 className="w-6 h-6 animate-spin" /> : 'Verify'}
                    </Button>
                  </div>
                </div>
              )}
           </div>

           <p className="text-zinc-400 text-[10px] font-bold uppercase tracking-widest text-center">
              Authoritative Auth v1.5 · Powered by Go API
           </p>
        </main>

        <footer className="mt-auto pt-12 text-center lg:text-left">
           <p className="text-zinc-300 text-[10px] font-bold uppercase tracking-widest">
             &copy; 2026 UTSAV Technologies
           </p>
        </footer>
      </div>
    </div>
  )
}
