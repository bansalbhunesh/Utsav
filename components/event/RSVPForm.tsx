'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { guestApiFetch, setGuestToken } from '@/lib/api'
import { 
  Check, 
  X, 
  Users, 
  Loader2, 
  Phone, 
  Lock, 
  ChevronRight, 
  UtensilsCrossed, 
  Calendar
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Confetti } from '@/components/ui/Confetti'
import { format } from 'date-fns'

interface SubEvent {
  id: string
  name: string
  sub_type?: string
  starts_at?: string
  date_time?: string
}

interface RSVPFormProps {
  eventTitle: string
  eventSlug: string
  subEvents: SubEvent[]
}

type RSVPStep = 'IDENTITY' | 'VERIFY' | 'DETAILS' | 'SUCCESS'

export function RSVPForm({ eventTitle, eventSlug, subEvents = [] }: RSVPFormProps) {
  const [step, setStep] = useState<RSVPStep>('IDENTITY')
  const [phone, setPhone] = useState('')
  const [otp, setOtp] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // RSVP Details state
  const [responses, setResponses] = useState<Record<string, { 
    status: 'CONFIRMED' | 'DECLINED',
    meal_pref?: string,
    plus_one_names?: string 
  }>>({})

  const handleRequestOTP = async () => {
    if (!phone) return
    setIsSubmitting(true)
    setError(null)
    try {
      await guestApiFetch(`/v1/public/events/${eventSlug}/rsvp/otp/request`, {
        method: 'POST',
        json: { phone }
      })
      setStep('VERIFY')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to request OTP')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleVerifyOTP = async () => {
    if (!otp) return
    setIsSubmitting(true)
    setError(null)
    try {
      const data = await guestApiFetch<{ guest_access_token: string }>(
        `/v1/public/events/${eventSlug}/rsvp/otp/verify`,
        { method: 'POST', json: { phone, code: otp } }
      )
      setGuestToken(data.guest_access_token)
      setStep('DETAILS')
    } catch {
      setError('Invalid OTP code. Please try again.')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleSubmitRSVP = async () => {
    const items = Object.entries(responses).map(([subId, data]) => ({
      sub_event_id: subId,
      status: data.status,
      meal_pref: data.meal_pref || 'VEG',
      plus_one_names: data.plus_one_names || ''
    }))

    if (items.length === 0) {
      setError('Please select attendance for at least one event.')
      return
    }

    setIsSubmitting(true)
    setError(null)
    try {
      await guestApiFetch(`/v1/public/events/${eventSlug}/rsvp`, {
        method: 'POST',
        json: { items }
      })
      setStep('SUCCESS')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to submit RSVP')
    } finally {
      setIsSubmitting(false)
    }
  }

  const toggleSubEvent = (subId: string, status: 'CONFIRMED' | 'DECLINED') => {
    setResponses(prev => ({
      ...prev,
      [subId]: { ...prev[subId], status }
    }))
  }

  if (step === 'SUCCESS') {
    return (
      <Card className="p-8 text-center space-y-4 rounded-[40px] border-none bg-orange-50 relative overflow-hidden">
        <Confetti />
        <div className="w-20 h-20 bg-orange-600 rounded-2xl flex items-center justify-center text-white mx-auto shadow-xl shadow-orange-200">
          <Check className="w-10 h-10" />
        </div>
        <div className="space-y-2">
          <h2 className="text-3xl font-bold text-zinc-900 font-heading">Blessings Received!</h2>
          <p className="text-zinc-600 font-medium">
            Your RSVP for <span className="font-bold">{eventTitle}</span> has been securely recorded.
          </p>
        </div>
        <Button 
          variant="outline" 
          onClick={() => setStep('IDENTITY')} 
          className="rounded-xl border-orange-200 text-orange-600 hover:bg-orange-100"
        >
          Update Response
        </Button>
      </Card>
    )
  }

  return (
    <div className="space-y-6 text-left">
      {error && (
        <div className="p-4 bg-red-50 border border-red-100 rounded-2xl text-red-600 text-sm font-bold animate-in fade-in zoom-in-95">
          {error}
        </div>
      )}

      {step === 'IDENTITY' && (
        <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4">
           <div className="space-y-2 text-center">
              <div className="w-12 h-12 bg-white rounded-2xl flex items-center justify-center mx-auto mb-4 shadow-sm border border-zinc-100">
                <Phone className="w-6 h-6 text-zinc-400" />
              </div>
              <h3 className="text-xl font-bold text-white">Join the Celebration</h3>
              <p className="text-zinc-400 text-sm">Enter your phone number to find your invitation.</p>
           </div>
           
           <div className="relative">
              <Phone className="absolute left-4 top-1/2 -translate-y-1/2 text-zinc-400 h-5 w-5" />
              <Input 
                placeholder="+91 9876543210" 
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                className="h-14 pl-12 rounded-2xl border-zinc-700 bg-zinc-800 text-white placeholder:text-zinc-500 text-lg focus:ring-orange-600"
              />
           </div>

           <Button 
             onClick={handleRequestOTP}
             disabled={isSubmitting || !phone}
             className="w-full h-14 bg-orange-600 hover:bg-orange-700 text-white font-bold rounded-2xl shadow-xl transition-all"
           >
             {isSubmitting ? <Loader2 className="h-6 h-6 animate-spin" /> : 'Get OTP Code'}
           </Button>
        </div>
      )}

      {step === 'VERIFY' && (
        <div className="space-y-6 animate-in fade-in slide-in-from-right-4">
           <div className="space-y-2 text-center">
              <div className="w-12 h-12 bg-white rounded-2xl flex items-center justify-center mx-auto mb-4 shadow-sm border border-zinc-100">
                <Lock className="w-6 h-6 text-zinc-400" />
              </div>
              <h3 className="text-xl font-bold text-white">Security Check</h3>
              <p className="text-zinc-400 text-sm">We sent a verification code to {phone}</p>
           </div>

           <div className="relative">
              <Input 
                placeholder="Enter 6-digit OTP" 
                value={otp}
                maxLength={6}
                onChange={(e) => setOtp(e.target.value)}
                className="h-14 rounded-2xl border-zinc-700 bg-zinc-800 text-white text-center text-2xl font-bold tracking-[0.5em] focus:ring-orange-600"
              />
           </div>

           <div className="grid grid-cols-2 gap-4">
              <Button 
                variant="ghost" 
                onClick={() => setStep('IDENTITY')}
                className="h-14 rounded-2xl text-zinc-400 hover:text-white"
              >
                Change Phone
              </Button>
              <Button 
                onClick={handleVerifyOTP}
                disabled={isSubmitting || otp.length < 4}
                className="h-14 bg-orange-600 hover:bg-orange-700 text-white font-bold rounded-2xl shadow-xl"
              >
                {isSubmitting ? <Loader2 className="h-6 h-6 animate-spin" /> : 'Verify & Continue'}
              </Button>
           </div>
        </div>
      )}

      {step === 'DETAILS' && (
        <div className="space-y-8 animate-in fade-in slide-in-from-bottom-8">
           <div className="flex items-center justify-between">
              <h3 className="text-xl font-bold text-white">Your Schedule</h3>
              <Badge className="bg-orange-600/10 text-orange-500 border-none font-bold">Relational Ledger v1.5</Badge>
           </div>

           <div className="space-y-4 max-h-[50vh] overflow-y-auto pr-2 scrollbar-hide">
              {subEvents.map((sub) => {
                const config = responses[sub.id] || { status: 'DECLINED' }
                return (
                  <div key={sub.id} className="p-5 rounded-[28px] bg-zinc-800 border border-zinc-700 space-y-4">
                    <div className="flex justify-between items-start">
                      <div className="space-y-1">
                        <div className="flex items-center gap-2 text-[10px] font-bold text-zinc-500 uppercase tracking-widest">
                          <Calendar className="w-3 h-3" />
                          {format(new Date(sub.starts_at || sub.date_time || new Date()), 'do MMM')} · {format(new Date(sub.starts_at || sub.date_time || new Date()), 'p')}
                        </div>
                        <h4 className="font-bold text-white">{sub.name}</h4>
                      </div>
                      <Badge variant="outline" className="border-zinc-600 text-zinc-400 rounded-lg text-[10px]">
                        {sub.sub_type || 'Standard'}
                      </Badge>
                    </div>

                    <div className="grid grid-cols-2 gap-3">
                       <button 
                         onClick={() => toggleSubEvent(sub.id, 'CONFIRMED')}
                         className={cn(
                           "h-12 rounded-xl border flex items-center justify-center gap-2 font-bold text-sm transition-all",
                           config.status === 'CONFIRMED' 
                             ? "bg-green-600 border-green-600 text-white shadow-lg shadow-green-900/20" 
                             : "border-zinc-700 bg-zinc-900/50 text-zinc-500 hover:border-zinc-600"
                         )}
                       >
                         <Check className="w-4 h-4" /> I&apos;m Joining
                       </button>
                       <button 
                         onClick={() => toggleSubEvent(sub.id, 'DECLINED')}
                         className={cn(
                           "h-12 rounded-xl border flex items-center justify-center gap-2 font-bold text-sm transition-all",
                           config.status === 'DECLINED' 
                             ? "bg-zinc-700 border-zinc-600 text-zinc-300 shadow-inner" 
                             : "border-zinc-700 bg-zinc-900/50 text-zinc-500 hover:border-zinc-600"
                         )}
                       >
                         <X className="w-4 h-4" /> Unable
                       </button>
                    </div>

                    {config.status === 'CONFIRMED' && (
                       <div className="pt-2 grid grid-cols-2 gap-4 animate-in slide-in-from-top-2">
                          <div className="space-y-1">
                             <label className="text-[10px] font-bold text-zinc-500 uppercase tracking-widest flex items-center gap-1">
                               <UtensilsCrossed className="w-3 h-3" /> Meal
                             </label>
                             <select 
                               value={config.meal_pref}
                               onChange={(e) => setResponses(p => ({ ...p, [sub.id]: { ...p[sub.id], meal_pref: e.target.value } }))}
                               className="w-full bg-zinc-900 border-zinc-700 rounded-lg text-xs p-2 text-white"
                             >
                               <option value="VEG">Vegetarian</option>
                               <option value="NON_VEG">Non-Veg</option>
                               <option value="VEGAN">Vegan</option>
                             </select>
                          </div>
                          <div className="space-y-1">
                             <label className="text-[10px] font-bold text-zinc-500 uppercase tracking-widest flex items-center gap-1">
                               <Users className="w-3 h-3" /> Plus One
                             </label>
                             <Input 
                               placeholder="Name..."
                               value={config.plus_one_names}
                               onChange={(e) => setResponses(p => ({ ...p, [sub.id]: { ...p[sub.id], plus_one_names: e.target.value } }))}
                               className="h-8 bg-zinc-900 border-zinc-700 rounded-lg text-xs text-white"
                             />
                          </div>
                       </div>
                    )}
                  </div>
                )
              })}
           </div>

           <Button 
             onClick={handleSubmitRSVP}
             disabled={isSubmitting}
             className="w-full h-14 bg-orange-600 hover:bg-orange-700 text-white font-bold rounded-2xl shadow-xl flex items-center justify-center gap-2 group"
           >
             {isSubmitting ? (
               <Loader2 className="h-6 w-6 animate-spin" />
             ) : (
               <>
                 Submit Formal RSVP <ChevronRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
               </>
             )}
           </Button>
        </div>
      )}
    </div>
  )
}
