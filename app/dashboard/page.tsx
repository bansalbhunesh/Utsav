'use client'

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
      if (!user) return
      
      const { data, error } = await supabase
        .from('events')
        .select('*')
        .eq('owner_id', user.id)
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
                   {event.cover_image && (
                     <img src={event.cover_image} alt="" className="w-full h-full object-cover" />
                   )}
                   <div className="absolute top-4 right-4">
                      <Badge className="bg-white/90 backdrop-blur-sm text-zinc-900 font-bold border-none">
                         {event.type}
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
    </div>
  )
}
