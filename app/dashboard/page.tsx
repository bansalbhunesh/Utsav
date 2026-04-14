'use client'

import { useMemo } from 'react'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { 
  Plus, 
  Calendar, 
  Users, 
  Settings, 
  ChevronRight,
  Sparkles,
  PartyPopper,
  Loader2,
  LayoutDashboard
} from 'lucide-react'
import Link from 'next/link'
import { apiFetch } from '@/lib/api'
import { format } from 'date-fns'
import { useQuery } from '@tanstack/react-query'
import { getUserFacingError } from '@/lib/error-messages'

interface DashboardEvent {
  id: string
  slug: string
  title: string
  event_type: string
  date_start: string
}

export default function DashboardPage() {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['dashboard-events'],
    queryFn: () => apiFetch<{ events: DashboardEvent[] }>('/v1/events'),
  })
  const events = useMemo(() => data?.events || [], [data])
  const errorMessage = useMemo(
    () => (error ? getUserFacingError(error, 'Failed to load events. Please try again.') : null),
    [error]
  )

  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] space-y-4">
        <Loader2 className="w-10 h-10 animate-spin text-orange-600" />
        <p className="text-zinc-400 font-bold uppercase text-[10px] tracking-widest">Waking up UTSAV...</p>
      </div>
    )
  }

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-10 animate-in fade-in duration-700">
      <header className="flex flex-col md:flex-row justify-between items-start md:items-center gap-6">
        <div className="space-y-1">
          <div className="flex items-center gap-2 text-orange-600 mb-1">
            <LayoutDashboard className="w-5 h-5" />
            <span className="text-[10px] font-bold uppercase tracking-[0.2em]">Management Console</span>
          </div>
          <h1 className="text-4xl md:text-5xl font-bold text-zinc-900 tracking-tight font-heading">Event OS</h1>
          <p className="text-zinc-500 font-medium">Your authoritative ledger for wedding celebrations.</p>
        </div>
        <Link href="/events/create">
          <Button className="bg-orange-600 hover:bg-orange-700 text-white rounded-2xl h-14 px-8 font-bold shadow-2xl shadow-orange-200 group transition-all hover:scale-[1.02]">
            <Plus className="w-5 h-5 mr-2 group-hover:rotate-90 transition-transform" />
            Create Celebration
          </Button>
        </Link>
      </header>

      {errorMessage ? (
        <Card className="p-12 border-none bg-red-50 text-center rounded-[40px] space-y-4">
           <div className="text-red-600 font-bold uppercase text-xs tracking-widest">{errorMessage}</div>
           <Button variant="outline" onClick={() => void refetch()} className="rounded-xl border-red-200 text-red-600">Retry Fetch</Button>
        </Card>
      ) : events.length === 0 ? (
        <Card className="p-20 text-center rounded-[40px] border-none bg-zinc-50/50 space-y-6">
          <div className="w-24 h-24 bg-white rounded-[32px] shadow-xl shadow-zinc-200/50 flex items-center justify-center mx-auto text-orange-600 animate-bounce">
             <PartyPopper className="w-12 h-12" />
          </div>
          <div className="space-y-2">
            <h2 className="text-3xl font-bold text-zinc-900">No events found</h2>
            <p className="text-zinc-500 max-w-sm mx-auto font-medium">
              Start your digital journey by creating your first authoritative event.
            </p>
          </div>
          <Link href="/events/create">
             <Button variant="outline" className="rounded-xl h-12 px-8 border-orange-200 text-orange-600 font-bold hover:bg-orange-50">Launch Wizard</Button>
          </Link>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          {events.map((event) => (
            <Link key={event.id} href={`/events/${event.id}/manage`}>
              <Card className="p-8 rounded-[40px] border border-zinc-100 hover:border-orange-200 shadow-sm hover:shadow-2xl transition-all duration-500 group relative overflow-hidden bg-white h-full flex flex-col justify-between">
                <div className="space-y-6 relative z-10">
                  <div className="flex justify-between items-start">
                    <Badge className="bg-orange-600/10 text-orange-600 border-none px-4 py-1.5 font-bold uppercase text-[9px] tracking-widest rounded-full">
                      {event.event_type || 'Wedding'}
                    </Badge>
                    <div className="w-10 h-10 rounded-2xl bg-zinc-50 flex items-center justify-center text-zinc-300 group-hover:text-orange-600 group-hover:bg-orange-50 transition-all">
                       <ChevronRight className="w-6 h-6" />
                    </div>
                  </div>
                  
                  <div>
                    <h3 className="text-2xl font-bold text-zinc-900 group-hover:text-orange-600 transition-colors tracking-tight leading-tight mb-2">
                      {event.title}
                    </h3>
                    <div className="flex items-center gap-2 text-zinc-400 text-sm font-semibold">
                      <div className="w-6 h-6 rounded-lg bg-zinc-100 flex items-center justify-center">
                        <Calendar className="w-3.5 h-3.5" />
                      </div>
                      {event.date_start ? format(new Date(event.date_start), 'PPP') : 'Date TBD'}
                    </div>
                  </div>
                </div>

                <div className="mt-8 pt-6 flex items-center gap-6 text-[10px] font-bold text-zinc-400 uppercase tracking-[0.15em] border-t border-zinc-50 relative z-10">
                   <div className="flex items-center gap-2 group-hover:text-zinc-600 transition-colors">
                     <Users className="w-4 h-4" /> Guest List
                   </div>
                   <div className="flex items-center gap-2 group-hover:text-zinc-600 transition-colors">
                     <Settings className="w-4 h-4" /> Manage
                   </div>
                </div>
                
                {/* Decorative Elements */}
                <div className="absolute top-0 right-0 p-8 opacity-[0.02] group-hover:opacity-[0.05] transition-opacity">
                   <Sparkles className="w-40 h-40 rotate-12" />
                </div>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
