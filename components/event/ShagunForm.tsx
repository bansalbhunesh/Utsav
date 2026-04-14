'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Card } from '@/components/ui/card'
import { paymentService } from '@/lib/services/PaymentService'
import { Sparkles, Check, Send, Heart, Info } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Event } from '@/types'
import { supabase } from '@/lib/supabase/client'

interface ShagunFormProps {
  event: Event
  hostName: string
}

const PRESET_AMOUNTS = [501, 1100, 2100, 5100];

const RELATIONSHIP_SUGGESTIONS = [
  { id: 'immediate', label: 'Immediate Family', range: '₹11,000 - ₹51,000', icon: Heart },
  { id: 'close', label: 'Close Friend/Relative', range: '₹2,100 - ₹11,000', icon: Sparkles },
  { id: 'standard', label: 'Friend/Colleague', range: '₹501 - ₹2,100', icon: Check },
]

export function ShagunForm({ event, hostName }: ShagunFormProps) {
  const [amount, setAmount] = useState<string>('')
  const [senderName, setSenderName] = useState('')
  const [message, setMessage] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isSuccess, setIsSuccess] = useState(false)

  const handlePayment = async () => {
    if (!amount || !senderName) {
      alert('Please enter your name and an amount.')
      return
    }

    if (!event.upi_id) {
      alert('Host has not configured a UPI ID yet.')
      return
    }

    setIsSubmitting(true)

    // 1. Initiate UPI intent
    const success = await paymentService.processPayment({
      amount: parseFloat(amount),
      receiverUpiId: event.upi_id,
      receiverName: hostName,
      transactionNote: `Shagun from ${senderName} for ${event.title}`
    })

    if (success) {
      // 2. Record the transaction in Supabase
      try {
        const { error } = await supabase
          .from('shagun')
          .insert({
            event_id: event.id,
            sender_name: senderName,
            amount: parseFloat(amount),
            message: message,
            payment_method: 'UPI',
            status: 'GUEST_REPORTED'
          })
        
        if (error) throw error
        setIsSuccess(true)
      } catch (err) {
        console.error('Failed to record shagun', err)
      }
    }

    setIsSubmitting(false)
  }

  if (isSuccess) {
    return (
      <Card className="p-8 text-center space-y-6 rounded-3xl border-orange-100 bg-orange-50/20">
        <div className="w-20 h-20 bg-orange-600 rounded-full flex items-center justify-center text-white mx-auto shadow-lg shadow-orange-200">
          <Heart className="w-10 h-10 fill-current" />
        </div>
        <div className="space-y-2">
          <h2 className="text-3xl font-bold font-heading text-zinc-900">A Blessing has Arrived!</h2>
          <p className="text-zinc-600">
            Thank you, <span className="font-bold text-zinc-900">{senderName}</span>. 
            Your shagun of <span className="font-bold text-orange-600">{paymentService.formatINR(parseFloat(amount))}</span> has been recorded.
          </p>
        </div>
        <Button 
          variant="outline" 
          onClick={() => setIsSuccess(false)}
          className="rounded-xl"
        >
          Send another blessing
        </Button>
      </Card>
    )
  }

  return (
    <div className="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-700">
      {/* Smart Suggestion Panel */}
      <div className="bg-zinc-900 rounded-3xl p-6 text-white space-y-4 shadow-xl">
        <div className="flex items-center gap-2 text-zinc-400 text-xs font-bold uppercase tracking-wider">
          <Sparkles className="h-4 w-4 text-orange-500" />
          Smart Suggestions (AI-Lite)
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          {RELATIONSHIP_SUGGESTIONS.map((rel) => (
            <div key={rel.id} className="p-3 rounded-xl bg-zinc-800/50 border border-zinc-700 hover:border-orange-500/50 transition-colors">
              <div className="flex items-center gap-2 mb-1">
                <rel.icon className="h-3.5 w-3.5 text-orange-500" />
                <span className="text-[10px] font-bold text-zinc-400 uppercase">{rel.label}</span>
              </div>
              <p className="text-sm font-bold text-zinc-100 font-heading">{rel.range}</p>
            </div>
          ))}
        </div>
        <div className="flex items-start gap-2 text-[10px] text-zinc-500 italic">
          <Info className="h-3 w-3 mt-0.5 shrink-0" />
          Suggestions based on typical regional norms for {event.type.toLowerCase()} events.
        </div>
      </div>

      <div className="space-y-6">
        <div className="space-y-4">
          <label className="text-sm font-bold text-zinc-700 uppercase tracking-wide">Enter Amount</label>
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
            {PRESET_AMOUNTS.map((amt) => (
              <Button
                key={amt}
                variant="outline"
                onClick={() => setAmount(amt.toString())}
                className={cn(
                  "h-14 rounded-2xl font-bold text-lg transition-all",
                  amount === amt.toString() 
                    ? "border-orange-600 bg-orange-50 text-orange-600 ring-2 ring-orange-200" 
                    : "border-zinc-100 hover:border-zinc-200"
                )}
              >
                ₹{amt}
              </Button>
            ))}
          </div>
          <div className="relative">
             <span className="absolute left-4 top-1/2 -translate-y-1/2 text-zinc-400 font-bold text-xl">₹</span>
             <Input 
                type="number"
                placeholder="Other Amount" 
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                className="h-16 pl-10 rounded-2xl text-xl font-bold shadow-sm focus:ring-orange-600 focus:border-orange-600"
             />
          </div>
        </div>

        <div className="space-y-4">
          <label className="text-sm font-bold text-zinc-700 uppercase tracking-wide">Your Details</label>
          <Input 
            placeholder="Your Full Name" 
            value={senderName}
            onChange={(e) => setSenderName(e.target.value)}
            className="h-14 rounded-2xl font-medium" 
          />
          <Textarea 
            placeholder="Write a warm blessing message (optional)..." 
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            className="min-h-[100px] rounded-2xl resize-none p-4" 
          />
        </div>

        <Button 
          onClick={handlePayment}
          disabled={isSubmitting || !amount || !senderName}
          className="w-full h-16 bg-orange-600 hover:bg-orange-700 text-white font-bold text-lg rounded-2xl shadow-lg shadow-orange-100 transition-all flex items-center justify-center gap-2"
        >
          {isSubmitting ? 'Processing...' : (
            <>
              Send Shagun via UPI
              <Send className="h-5 w-5" />
            </>
          )}
        </Button>
      </div>
    </div>
  )
}
