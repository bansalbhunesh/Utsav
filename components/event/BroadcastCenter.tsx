'use client'

import { useState } from 'react'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { MessageSquare, Share2, Copy, Check, Info } from 'lucide-react'
import { cn } from '@/lib/utils'

interface BroadcastCenterProps {
  eventTitle: string
  eventSlug: string
}

export function BroadcastCenter({ eventTitle, eventSlug }: BroadcastCenterProps) {
  const [copiedIndex, setCopiedIndex] = useState<number | null>(null)
  
  const eventLink = `https://utsav.app/${eventSlug}`
  
  const templates = [
    {
      title: 'Save the Date',
      message: `Namaste! We are excited to invite you to the celebration of ${eventTitle}. View the details and RSVP here: ${eventLink}`,
      icon: MessageSquare
    },
    {
      title: 'Schedule Update',
      message: `Update for ${eventTitle}: Check out the full event schedule and venue locations here: ${eventLink}`,
      icon: Share2
    },
    {
      title: 'Digital Shagun',
      message: `In case you missed it, you can send your virtual blessings and shagun for ${eventTitle} via our digital box here: ${eventLink}/shagun`,
      icon: MessageSquare
    }
  ]

  const handleCopy = (text: string, index: number) => {
    navigator.clipboard.writeText(text)
    setCopiedIndex(index)
    setTimeout(() => setCopiedIndex(null), 2000)
  }

  const handleWhatsApp = (text: string) => {
     const url = `https://wa.me/?text=${encodeURIComponent(text)}`
     window.open(url, '_blank')
  }

  return (
    <Card className="p-6 border-zinc-200 shadow-xl rounded-[32px] bg-white space-y-6">
      <div className="flex items-center gap-3">
        <div className="w-10 h-10 rounded-2xl bg-orange-100 flex items-center justify-center text-orange-700">
          <MessageSquare className="h-5 w-5" />
        </div>
        <div>
          <h3 className="font-bold text-zinc-900">Broadcast Center</h3>
          <p className="text-[10px] text-zinc-500 uppercase font-bold tracking-wider">Module 9 · v1.5</p>
        </div>
      </div>

      <div className="space-y-4">
        {templates.map((tpl, i) => (
          <div key={i} className="p-4 bg-zinc-50 rounded-2xl border border-zinc-100 space-y-3 group hover:border-orange-200 transition-colors">
            <div className="flex justify-between items-center">
               <span className="text-xs font-bold text-zinc-400 uppercase tracking-widest">{tpl.title}</span>
               <div className="flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                  <Button variant="ghost" size="icon" className="h-8 w-8 rounded-lg" onClick={() => handleCopy(tpl.message, i)}>
                     {copiedIndex === i ? <Check className="h-4 w-4 text-green-600" /> : <Copy className="h-4 w-4" />}
                  </Button>
               </div>
            </div>
            <p className="text-xs text-zinc-600 leading-relaxed italic line-clamp-2">
               "{tpl.message}"
            </p>
            <Button 
               onClick={() => handleWhatsApp(tpl.message)}
               className="w-full h-10 rounded-xl bg-green-600 hover:bg-green-700 text-white font-bold text-xs"
            >
               Share on WhatsApp
            </Button>
          </div>
        ))}
      </div>

      <div className="flex items-start gap-3 p-4 bg-zinc-50 rounded-2xl">
         <Info className="h-4 w-4 text-zinc-400 mt-0.5 shrink-0" />
         <p className="text-[10px] text-zinc-500 leading-relaxed">
           <strong>Smart Reminders</strong>: v1.5 uses local browser intents. Phase 2 will include automatic WhatsApp Business API integration.
         </p>
      </div>
    </Card>
  )
}
