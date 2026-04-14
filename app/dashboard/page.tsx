'use client'

<<<<<<< HEAD
import { useEffect, useState } from 'react'
import Link from 'next/link'
import { useSearchParams } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Plus, Calendar, Settings, Share2, MoreVertical, LogOut } from 'lucide-react'
import { supabase } from '@/lib/supabase/client'
import { useAuthStore } from '@/store/auth-store'
import { signOut } from '@/lib/auth'

export default function DashboardPage() {
  const searchParams = useSearchParams()
  const success = searchParams.get('success')
  const newEventSlug = searchParams.get('event')
  const { user } = useAuthStore()
  const [events, setEvents] = useState<any[]>([])
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    async function fetchEvents() {
      if (!user && !isLoading) {
        window.location.href = '/login'
        return
      }
      if (!user) return
      
      const { data, error } = await supabase
        .from('events')
        .select('*')
        .eq('owner_user_id', user.id)
        .order('created_at', { ascending: false })

      if (error) {
        console.error('Fetch events failed', error)
      } else {
        setEvents(data || [])
      }
      setIsLoading(false)
    }

    fetchEvents()
  }, [user])

  return (
    <div className="min-h-screen bg-zinc-50 flex flex-col">

      {/* Main Content */}
      <main className="flex-1 max-w-7xl mx-auto w-full p-6 lg:p-10 space-y-10">
        {success && (
          <div className="animate-in fade-in slide-in-from-top duration-500 p-4 bg-green-50 border border-green-100 rounded-2xl flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="bg-green-500 text-white rounded-full p-1">
                <Plus className="w-4 h-4" />
              </div>
              <p className="text-sm font-bold text-green-700">
                Event created successfully!
              </p>
            </div>
            <Link href={`/${newEventSlug}`} className="text-sm font-bold text-green-700 underline flex items-center gap-1">
               Go to event page <Share2 className="h-3 w-3" />
            </Link>
          </div>
        )}

        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-6">
          <div className="space-y-1">
            <h1 className="text-3xl font-bold font-heading tracking-tight text-zinc-900">Your Events</h1>
            <p className="text-zinc-500">Manage all your active celebrations and sessions.</p>
          </div>

          <Link href="/events/new">
            <Button className="h-12 px-6 bg-orange-600 hover:bg-orange-700 text-white font-bold rounded-xl shadow-lg shadow-orange-200">
              <Plus className="mr-2 h-5 w-5" />
              New Event
            </Button>
          </Link>
        </div>

        {/* Events Grid */}
        {isLoading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {[1, 2, 3].map(i => (
              <div key={i} className="h-[200px] bg-zinc-200/50 animate-pulse rounded-3xl" />
            ))}
          </div>
        ) : events.length === 0 ? (
          <div className="py-20 text-center space-y-4 bg-white rounded-3xl border border-zinc-100 shadow-sm border-dashed">
             <div className="w-16 h-16 bg-zinc-50 rounded-full flex items-center justify-center mx-auto">
                <Calendar className="h-8 w-8 text-zinc-300" />
             </div>
             <div className="space-y-1">
                <p className="text-lg font-bold text-zinc-900 text-heading">No events found</p>
                <p className="text-zinc-500 max-w-xs mx-auto text-sm">Create your first event to start managing guests, schedule, and shagun.</p>
             </div>
             <Link href="/events/new" className="inline-block">
                <Button variant="outline" className="rounded-xl font-bold">Create First Event</Button>
             </Link>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {events.map((event) => (
              <div key={event.id} className="group bg-white rounded-3xl border border-zinc-100 shadow-sm hover:shadow-xl hover:-translate-y-1 transition-all duration-300 overflow-hidden">
                <div className="h-32 bg-zinc-100 relative overflow-hidden">
                   {event.cover_image_url && (
                     <img src={event.cover_image_url} alt="" className="w-full h-full object-cover" />
                   )}
                   <div className="absolute top-4 right-4">
                      <Badge className="bg-white/90 backdrop-blur-sm text-zinc-900 font-bold border-none">
                         {event.event_type}
                      </Badge>
                   </div>
                </div>
                <div className="p-6 space-y-4">
                  <div className="flex justify-between items-start">
                    <h3 className="text-xl font-bold font-heading text-zinc-900 line-clamp-1">{event.title}</h3>
                    <Button variant="ghost" size="icon" className="h-8 w-8 rounded-full">
                       <MoreVertical className="h-4 w-4" />
                    </Button>
                  </div>

                  <div className="flex items-center gap-2 text-sm text-zinc-500 font-medium">
                     <Calendar className="h-4 w-4" />
                     {new Date(event.start_date).toLocaleDateString('en-IN', { day: 'numeric', month: 'short', year: 'numeric' })}
                  </div>

                  <div className="flex gap-2 pt-2">
                    <Link href={`/events/${event.id}/manage`} className="flex-1">
                       <Button variant="outline" size="sm" className="w-full rounded-lg font-semibold h-9 text-xs">
                          <Settings className="mr-1 h-3 w-3" />
                          Manage
                       </Button>
                    </Link>
                    <Link href={`/${event.slug}`} className="flex-1">
                       <Button size="sm" className="w-full bg-zinc-900 hover:bg-zinc-800 text-white rounded-lg font-semibold h-9 text-xs">
                          <Share2 className="mr-1 h-3 w-3" />
                          Event Page
                       </Button>
                    </Link>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </main>
=======
import { useState, useEffect } from 'react'
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
  Loader2
} from 'lucide-react'
import Link from 'next/link'
import { apiFetch } from '@/lib/api'
import { format } from 'date-fns'

interface DashboardEvent {
  id: string
  slug: string
  title: string
  event_type: string
  date_start: string
}

export default function DashboardPage() {
  const [events, setEvents] = useState<DashboardEvent[]>([])
  const [loading, setLoading] = useState(true)

  const fetchEvents = async () => {
    try {
      const data = await apiFetch<{ events: DashboardEvent[] }>('/v1/events')
      setEvents(data.events || [])
    } catch (err) {
      console.error('Failed to load events:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchEvents()
  }, [])

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Loader2 className="w-8 h-8 animate-spin text-orange-600" />
      </div>
    )
  }

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-10">
      <header className="flex justify-between items-center">
        <div>
          <h1 className="text-4xl font-bold text-zinc-900 tracking-tight font-heading">Event OS</h1>
          <p className="text-zinc-500 font-medium">Manage your celebrations authoritative-first.</p>
        </div>
        <Link href="/events/create">
          <Button className="bg-orange-600 hover:bg-orange-700 text-white rounded-2xl h-12 px-6 font-bold shadow-xl shadow-orange-100 group">
            <Plus className="w-5 h-5 mr-2 group-hover:rotate-90 transition-transform" />
            New Event
          </Button>
        </Link>
      </header>

      {events.length === 0 ? (
        <Card className="p-20 text-center rounded-[40px] border-none bg-zinc-50 space-y-6">
          <div className="w-20 h-20 bg-orange-100 rounded-3xl flex items-center justify-center mx-auto text-orange-600">
             <PartyPopper className="w-10 h-10" />
          </div>
          <div className="space-y-2">
            <h2 className="text-2xl font-bold text-zinc-900">No events found</h2>
            <p className="text-zinc-500 max-w-sm mx-auto">Create your first event to start managing RSVPs, Vendors, and Shagun.</p>
          </div>
          <Link href="/events/create">
             <Button variant="outline" className="rounded-xl border-orange-200 text-orange-600">Get Started</Button>
          </Link>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          {events.map((event) => (
            <Link key={event.id} href={`/events/${event.id}/manage`}>
              <Card className="p-6 rounded-[32px] border border-zinc-100 hover:border-orange-200 shadow-sm hover:shadow-xl transition-all duration-500 group relative overflow-hidden bg-white">
                <div className="space-y-4 relative z-10">
                  <div className="flex justify-between items-start">
                    <Badge className="bg-orange-600/10 text-orange-600 border-none px-3 py-1 font-bold uppercase text-[10px] tracking-widest">
                      {event.event_type}
                    </Badge>
                    <div className="w-8 h-8 rounded-full bg-zinc-50 flex items-center justify-center text-zinc-300 group-hover:text-orange-600 group-hover:bg-orange-50 transition-colors">
                       <ChevronRight className="w-5 h-5" />
                    </div>
                  </div>
                  
                  <div>
                    <h3 className="text-xl font-bold text-zinc-900 group-hover:text-orange-600 transition-colors">{event.title}</h3>
                    <div className="flex items-center gap-2 text-zinc-400 text-xs font-medium mt-1">
                      <Calendar className="w-4 h-4" />
                      {event.date_start ? format(new Date(event.date_start), 'PPP') : 'Date TBD'}
                    </div>
                  </div>

                  <div className="pt-4 flex items-center gap-4 text-xs font-bold text-zinc-400 uppercase tracking-widest border-t border-zinc-50">
                     <span className="flex items-center gap-1.5"><Users className="w-4 h-4" /> Guest List</span>
                     <span className="flex items-center gap-1.5"><Settings className="w-4 h-4" /> Manage</span>
                  </div>
                </div>
                
                {/* Decorative Pattern */}
                <div className="absolute top-0 right-0 p-8 opacity-[0.03] group-hover:opacity-[0.07] transition-opacity">
                   <Sparkles className="w-32 h-32 rotate-12" />
                </div>
              </Card>
            </Link>
          ))}
        </div>
      )}
>>>>>>> f7494df (feat: Architectural Level Up - Go-Authoritative Backend, RSVP OTP Flow, and Frontend Consolidation (v1.5 Final))
    </div>
  )
}
