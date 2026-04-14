'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { supabase } from '@/lib/supabase/client'
import { Check, X, Users, Heart, Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Confetti } from '@/components/ui/Confetti'

interface RSVPFormProps {
  eventId: string
  eventTitle: string
}

export function RSVPForm({ eventId, eventTitle }: RSVPFormProps) {
  const [status, setStatus] = useState<'CONFIRMED' | 'DECLINED' | null>(null)
  const [name, setName] = useState('')
  const [guestCount, setGuestCount] = useState(1)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isSuccess, setIsSuccess] = useState(false)

  const handleSubmit = async () => {
    if (!name || !status) return

    setIsSubmitting(true)
    try {
      const { error } = await supabase
        .from('guests')
        .insert({
          event_id: eventId,
          name: name,
          status: status,
          notes: status === 'CONFIRMED' ? `Bringing ${guestCount} guests` : '',
          // Simplified family RSVP: we store count in notes or can expand schema
        })

      if (error) throw error
      setIsSuccess(true)
    } catch (err) {
      console.error('RSVP failed', err)
      alert('Error saving RSVP.')
    } finally {
      setIsSubmitting(false)
    }
  }

  if (isSuccess) {
    return (
      <Card className="p-8 text-center space-y-4 rounded-3xl border-green-100 bg-green-50/20">
        <div className="w-16 h-16 bg-green-500 rounded-full flex items-center justify-center text-white mx-auto">
          <Check className="w-8 h-8" />
        </div>
        <div>
          <h2 className="text-2xl font-bold text-zinc-900">Thank You!</h2>
          <p className="text-zinc-600">
            Your RSVP has been sent to the hosts of <span className="font-bold">{eventTitle}</span>.
          </p>
        </div>
      </Card>
    )
  }

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-2 gap-4">
        <button
          onClick={() => setStatus('CONFIRMED')}
          className={cn(
            "p-6 rounded-[24px] border-2 transition-all flex flex-col items-center gap-2",
            status === 'CONFIRMED' 
              ? "border-orange-600 bg-orange-50 text-orange-600" 
              : "border-zinc-100 bg-white hover:border-zinc-200"
          )}
        >
          <div className={cn(
            "w-10 h-10 rounded-full flex items-center justify-center",
            status === 'CONFIRMED' ? "bg-orange-600 text-white" : "bg-zinc-100 text-zinc-400"
          )}>
            <Check className="h-5 w-5" />
          </div>
          <span className="font-bold">Attending</span>
        </button>

        <button
          onClick={() => setStatus('DECLINED')}
          className={cn(
            "p-6 rounded-[24px] border-2 transition-all flex flex-col items-center gap-2",
            status === 'DECLINED' 
              ? "border-zinc-400 bg-zinc-50 text-zinc-900" 
              : "border-zinc-100 bg-white hover:border-zinc-200"
          )}
        >
          <div className={cn(
            "w-10 h-10 rounded-full flex items-center justify-center",
            status === 'DECLINED' ? "bg-zinc-900 text-white" : "bg-zinc-100 text-zinc-400"
          )}>
            <X className="h-5 w-5" />
          </div>
          <span className="font-bold text-zinc-500">Regret</span>
        </button>
      </div>

      <div className="space-y-4 animate-in fade-in slide-in-from-top-2">
        <div className="space-y-2">
          <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider ml-1">Main Contact Name</label>
          <Input 
            placeholder="e.g. Rahul Sharma" 
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="h-12 rounded-xl"
          />
        </div>

        {status === 'CONFIRMED' && (
          <div className="space-y-2">
            <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider ml-1">Total Guests (Including you)</label>
            <div className="flex items-center gap-4">
               <div className="flex-1 flex overflow-hidden border border-zinc-200 rounded-xl bg-white h-12">
                  {[1, 2, 3, 4, 5].map((num) => (
                    <button
                      key={num}
                      type="button"
                      onClick={() => setGuestCount(num)}
                      className={cn(
                        "flex-1 font-bold transition-colors",
                        guestCount === num ? "bg-zinc-900 text-white" : "hover:bg-zinc-50 text-zinc-600"
                      )}
                    >
                      {num}{num === 5 ? '+' : ''}
                    </button>
                  ))}
               </div>
            </div>
          </div>
        )}

        <Button 
          onClick={handleSubmit}
          disabled={isSubmitting || !status || !name}
          className="w-full h-12 bg-orange-600 hover:bg-orange-700 text-white font-bold rounded-xl mt-4"
        >
          {isSubmitting ? <Loader2 className="h-5 w-5 animate-spin" /> : 'Submit RSVP'}
        </Button>
      </div>
    </div>
  )
}
