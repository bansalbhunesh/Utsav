'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card } from '@/components/ui/card'
<<<<<<< HEAD
import { supabase } from '@/lib/supabase/client'
import { Plus, Banknote, User, Check, Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'

interface FastCashLoggerProps {
  eventId: string
  onSuccess?: () => void
}

const COMMON_DENOMINATIONS = [101, 201, 501, 1100, 2100, 5100]

export function FastCashLogger({ eventId, onSuccess }: FastCashLoggerProps) {
  const [amount, setAmount] = useState('')
  const [guestName, setGuestName] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [showSuccess, setShowSuccess] = useState(false)

  const handleLog = async () => {
    if (!amount || !guestName) return

    setIsSubmitting(true)
    try {
      const { error } = await supabase
        .from('shagun_entries')
        .insert({
          event_id: eventId,
          channel: 'CASH',
          amount_paise: Math.round(parseFloat(amount) * 100),
          status: 'verified',
          meta: { sender_name: guestName }
        })

      if (error) throw error
      
      setShowSuccess(true)
      setTimeout(() => {
        setShowSuccess(false)
        setAmount('')
        setGuestName('')
        if (onSuccess) onSuccess()
      }, 1500)
    } catch (err) {
      console.error('Failed to log cash shagun', err)
      alert('Error saving entry.')
=======
import { 
  IndianRupee, 
  User, 
  Phone, 
  Send,
  CheckCircle2,
  Loader2,
  Lock
} from 'lucide-react'
import { apiFetch } from '@/lib/api'
import { Confetti } from '@/components/ui/Confetti'

export function FastCashLogger({ eventId }: { eventId: string }) {
  const [amount, setAmount] = useState('')
  const [name, setName] = useState('')
  const [phone, setPhone] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isSuccess, setIsSuccess] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleLog = async () => {
    if (!amount || !name) return
    setIsSubmitting(true)
    setError(null)
    try {
      await apiFetch(`/v1/events/${eventId}/cash-shagun`, {
        method: 'POST',
        json: {
          amount_paise: Math.round(parseFloat(amount) * 100),
          guest_name: name,
          guest_phone: phone,
          channel: 'CASH'
        }
      })
      setIsSuccess(true)
      setTimeout(() => {
        setIsSuccess(false)
        setAmount('')
        setName('')
        setPhone('')
      }, 3000)
    } catch (err: any) {
      setError('Failed to log entry. Please try again.')
>>>>>>> f7494df (feat: Architectural Level Up - Go-Authoritative Backend, RSVP OTP Flow, and Frontend Consolidation (v1.5 Final))
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
<<<<<<< HEAD
    <Card className="p-6 border-zinc-200 shadow-xl rounded-[32px] bg-white space-y-6">
      <div className="flex items-center gap-3 mb-2">
        <div className="w-10 h-10 rounded-2xl bg-green-100 flex items-center justify-center text-green-700">
          <Banknote className="h-5 w-5" />
        </div>
        <div>
          <h3 className="font-bold text-zinc-900">Fast Cash Logger</h3>
          <p className="text-[10px] text-zinc-500 uppercase font-bold tracking-wider">One-Handed UI</p>
=======
    <Card className="p-6 border-orange-100 shadow-xl rounded-[32px] bg-white relative overflow-hidden">
      {isSuccess && <Confetti />}
      
      <div className="flex items-center gap-3 mb-6">
        <div className="w-10 h-10 rounded-2xl bg-orange-100 flex items-center justify-center text-orange-600">
          <IndianRupee className="h-5 w-5" />
        </div>
        <div>
          <h3 className="font-bold text-zinc-900 font-heading">Hand-Cash Logger</h3>
          <p className="text-[10px] text-zinc-400 uppercase font-bold tracking-widest flex items-center gap-1">
            <Lock className="w-2 h-2" /> Secure Authoritative Entry
          </p>
>>>>>>> f7494df (feat: Architectural Level Up - Go-Authoritative Backend, RSVP OTP Flow, and Frontend Consolidation (v1.5 Final))
        </div>
      </div>

      <div className="space-y-4">
<<<<<<< HEAD
        {/* Amount Input */}
        <div className="space-y-4">
          <div className="relative">
            <span className="absolute left-4 top-1/2 -translate-y-1/2 text-zinc-400 font-bold text-xl">₹</span>
            <Input 
              type="number"
              placeholder="Amount" 
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              className="h-14 pl-10 rounded-2xl text-xl font-bold border-zinc-200 focus:ring-green-600 focus:border-green-600"
            />
          </div>
          
          <div className="grid grid-cols-3 gap-2">
            {COMMON_DENOMINATIONS.map(den => (
              <Button
                key={den}
                type="button"
                variant="outline"
                size="sm"
                onClick={() => setAmount(den.toString())}
                className={cn(
                  "h-10 rounded-xl font-bold transition-all",
                  amount === den.toString() ? "bg-green-50 border-green-600 text-green-600" : "hover:border-green-200"
                )}
              >
                ₹{den}
              </Button>
            ))}
          </div>
        </div>

        {/* Guest Name */}
        <div className="relative">
          <User className="absolute left-4 top-1/2 -translate-y-1/2 text-zinc-400 h-5 w-5" />
          <Input 
            placeholder="Guest / Family Name" 
            value={guestName}
            onChange={(e) => setGuestName(e.target.value)}
            className="h-14 pl-12 rounded-2xl font-medium border-zinc-200" 
          />
        </div>

        <Button 
          onClick={handleLog}
          disabled={isSubmitting || !amount || !guestName}
          className={cn(
            "w-full h-14 font-bold text-lg rounded-2xl transition-all duration-300",
            showSuccess 
              ? "bg-green-500 hover:bg-green-500 text-white" 
              : "bg-zinc-900 hover:bg-black text-white shadow-lg"
          )}
        >
          {isSubmitting ? (
            <Loader2 className="h-6 w-6 animate-spin" />
          ) : showSuccess ? (
            <span className="flex items-center gap-2">
              <Check className="h-6 w-6" />
              Logged!
            </span>
          ) : (
            <span className="flex items-center gap-2">
              <Plus className="h-5 w-5" />
              Log Shagun
            </span>
=======
        <div className="space-y-1.5">
           <label className="text-[10px] font-bold text-zinc-400 uppercase tracking-widest ml-1">Gift Amount (₹)</label>
           <div className="relative">
              <IndianRupee className="absolute left-4 top-1/2 -translate-y-1/2 h-5 w-5 text-zinc-400" />
              <Input 
                type="number" 
                placeholder="501" 
                value={amount}
                onChange={e => setAmount(e.target.value)}
                className="pl-12 h-14 text-xl font-bold rounded-2xl border-zinc-100 bg-zinc-50/50"
              />
           </div>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div className="space-y-1.5">
             <label className="text-[10px] font-bold text-zinc-400 uppercase tracking-widest ml-1">Member Name</label>
             <div className="relative">
                <User className="absolute left-4 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-400" />
                <Input 
                  placeholder="Rahul G." 
                  value={name}
                  onChange={e => setName(e.target.value)}
                  className="pl-10 h-12 rounded-xl border-zinc-100 bg-zinc-50/50 text-sm"
                />
             </div>
          </div>
          <div className="space-y-1.5">
             <label className="text-[10px] font-bold text-zinc-400 uppercase tracking-widest ml-1">Phone (Optional)</label>
             <div className="relative">
                <Phone className="absolute left-4 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-400" />
                <Input 
                  placeholder="9876543210" 
                  value={phone}
                  onChange={e => setPhone(e.target.value)}
                  className="pl-10 h-12 rounded-xl border-zinc-100 bg-zinc-50/50 text-sm"
                />
             </div>
          </div>
        </div>

        {error && <p className="text-[10px] text-red-500 font-bold uppercase text-center">{error}</p>}

        <Button 
          onClick={handleLog}
          disabled={isSubmitting || !amount || !name}
          className="w-full h-14 bg-orange-600 hover:bg-orange-700 text-white font-bold rounded-2xl shadow-xl shadow-orange-100 mt-2 flex items-center justify-center gap-2 group"
        >
          {isSubmitting ? (
            <Loader2 className="h-5 w-5 animate-spin" />
          ) : isSuccess ? (
            <><CheckCircle2 className="h-5 w-5" /> Entry Verified</>
          ) : (
            <><Send className="h-5 w-5 group-hover:translate-x-1 group-hover:-translate-y-1 transition-transform" /> Log Cash Entry</>
>>>>>>> f7494df (feat: Architectural Level Up - Go-Authoritative Backend, RSVP OTP Flow, and Frontend Consolidation (v1.5 Final))
          )}
        </Button>
      </div>
    </Card>
  )
}
