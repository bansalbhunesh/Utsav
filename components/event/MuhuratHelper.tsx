'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { CalendarIcon, Sparkles, Check } from 'lucide-react'
import { cn } from '@/lib/utils'
import { format } from 'date-fns'

interface Muhurat {
  date: Date
  description: string
  label: string
}

const AUSPICIOUS_DATES: Muhurat[] = [
  { date: new Date('2026-05-18'), label: 'Siddha Yoga', description: 'Highly auspicious for weddings and new beginnings.' },
  { date: new Date('2026-05-24'), label: 'Amrit Siddhi', description: 'Best for long-lasting prosperity and happiness.' },
  { date: new Date('2026-06-12'), label: 'Pushya Nakshatra', description: 'Excellent for religious and social ceremonies.' },
  { date: new Date('2026-11-04'), label: 'Dev Uthani Gyaras', description: 'Major wedding dates begin from this day.' },
  { date: new Date('2026-11-21'), label: 'Shubha Muhurat', description: 'A perfectly balanced day for celebrations.' },
  { date: new Date('2026-12-08'), label: 'Vivah Muhurat', description: 'The most popular auspicious window of the month.' },
]

interface MuhuratHelperProps {
  onSelect: (date: Date) => void
  selectedDate?: Date
}

export function MuhuratHelper({ onSelect, selectedDate }: MuhuratHelperProps) {
  const [isOpen, setIsOpen] = useState(false)

  return (
    <div className="space-y-4">
      <Button
        type="button"
        variant="ghost"
        onClick={() => setIsOpen(!isOpen)}
        className="text-orange-600 hover:text-orange-700 hover:bg-orange-50 px-0 h-auto font-semibold flex items-center gap-2"
      >
        <Sparkles className="h-4 w-4" />
        {isOpen ? 'Hide Auspicious Dates' : 'View Auspicious Dates (Muhurat)'}
      </Button>

      {isOpen && (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 animate-in fade-in slide-in-from-top-2">
          {AUSPICIOUS_DATES.map((muhurat, index) => {
            const isSelected = selectedDate?.toDateString() === muhurat.date.toDateString()
            return (
              <button
                key={index}
                type="button"
                onClick={() => onSelect(muhurat.date)}
                className={cn(
                  "p-3 rounded-xl border text-left transition-all hover:shadow-md",
                  isSelected 
                    ? "border-orange-600 bg-orange-50 ring-2 ring-orange-200" 
                    : "border-zinc-100 bg-white hover:border-zinc-200"
                )}
              >
                <div className="flex justify-between items-start mb-1">
                  <span className="text-xs font-bold text-orange-600 uppercase tracking-wider">
                    {muhurat.label}
                  </span>
                  {isSelected && <Check className="h-4 w-4 text-orange-600" />}
                </div>
                <div className="font-bold text-zinc-900 flex items-center gap-1.5 mb-1">
                  <CalendarIcon className="h-3.5 w-3.5 text-zinc-400" />
                  {format(muhurat.date, 'PPP')}
                </div>
                <p className="text-[10px] text-zinc-500 leading-relaxed">
                  {muhurat.description}
                </p>
              </button>
            )
          })}
        </div>
      )}
    </div>
  )
}
