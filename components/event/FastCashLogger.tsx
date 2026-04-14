'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card } from '@/components/ui/card'
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
        .from('shagun')
        .insert({
          event_id: eventId,
          sender_name: guestName,
          amount: parseFloat(amount),
          payment_method: 'CASH',
          status: 'VERIFIED'
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
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Card className="p-6 border-zinc-200 shadow-xl rounded-[32px] bg-white space-y-6">
      <div className="flex items-center gap-3 mb-2">
        <div className="w-10 h-10 rounded-2xl bg-green-100 flex items-center justify-center text-green-700">
          <Banknote className="h-5 w-5" />
        </div>
        <div>
          <h3 className="font-bold text-zinc-900">Fast Cash Logger</h3>
          <p className="text-[10px] text-zinc-500 uppercase font-bold tracking-wider">One-Handed UI</p>
        </div>
      </div>

      <div className="space-y-4">
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
          )}
        </Button>
      </div>
    </Card>
  )
}
