'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card } from '@/components/ui/card'
import { 
  User, 
  CheckCircle2, 
  Loader2, 
  Lock,
  Plus,
  Banknote
} from 'lucide-react'
import { apiFetch } from '@/lib/api'
import { Confetti } from '@/components/ui/Confetti'
import { cn } from '@/lib/utils'

interface FastCashLoggerProps {
  eventId: string
  onSuccess?: () => void
}

const COMMON_DENOMINATIONS = [101, 201, 501, 1100, 2100, 5100]

export function FastCashLogger({ eventId, onSuccess }: FastCashLoggerProps) {
  const [amount, setAmount] = useState('')
  const [name, setName] = useState('')
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
          channel: 'CASH'
        }
      })
      
      setIsSuccess(true)
      setTimeout(() => {
        setIsSuccess(false)
        setAmount('')
        setName('')
        if (onSuccess) onSuccess()
      }, 3000)
    } catch (err: unknown) {
      console.error('Failed to log cash shagun', err)
      setError('Failed to log entry. Please try again.')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Card className="p-6 border-zinc-200 shadow-xl rounded-[32px] bg-white relative overflow-hidden">
      {isSuccess && <Confetti />}
      
      <div className="flex items-center gap-3 mb-6">
        <div className="w-10 h-10 rounded-2xl bg-green-100 flex items-center justify-center text-green-700">
          <Banknote className="h-5 w-5" />
        </div>
        <div>
          <h3 className="font-bold text-zinc-900 font-heading tracking-tight">Hand-Cash Logger</h3>
          <p className="text-[10px] text-zinc-400 uppercase font-bold tracking-widest flex items-center gap-1">
            <Lock className="w-2 h-2" /> Authoritative Relational Ledger
          </p>
        </div>
      </div>

      <div className="space-y-6">
        {/* Amount Input & Denominations */}
        <div className="space-y-4">
          <div className="relative">
            <span className="absolute left-4 top-1/2 -translate-y-1/2 text-zinc-400 font-bold text-xl">₹</span>
            <Input 
              type="number"
              placeholder="Amount" 
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              className="h-14 pl-10 rounded-2xl text-xl font-bold border-zinc-100 bg-zinc-50/50 focus:ring-green-600"
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
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="h-14 pl-12 rounded-2xl font-medium border-zinc-100 bg-zinc-50/50" 
          />
        </div>

        {error && <p className="text-[10px] text-red-500 font-bold uppercase text-center">{error}</p>}

        <Button 
          onClick={handleLog}
          disabled={isSubmitting || !amount || !name}
          className={cn(
            "w-full h-14 font-bold text-lg rounded-2xl transition-all duration-300",
            isSuccess 
              ? "bg-green-500 hover:bg-green-500 text-white" 
              : "bg-zinc-900 hover:bg-black text-white shadow-xl"
          )}
        >
          {isSubmitting ? (
            <Loader2 className="h-6 w-6 animate-spin" />
          ) : isSuccess ? (
            <span className="flex items-center gap-2">
              <CheckCircle2 className="h-6 w-6" />
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
