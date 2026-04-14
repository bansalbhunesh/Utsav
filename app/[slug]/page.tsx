import { notFound } from 'next/navigation'
import { RSVPForm } from '@/components/event/RSVPForm'
import { guestApiFetch } from '@/lib/api'
import { 
  Calendar, 
  MapPin, 
  Clock, 
  Heart, 
  IndianRupee, 
  ChevronRight,
  Sparkles,
  PartyPopper,
  Image as ImageIcon
} from 'lucide-react'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { format } from 'date-fns'

interface EventPageProps {
  params: {
    slug: string
  }
}

async function getEventData(slug: string) {
  try {
    const data = await guestApiFetch<{ event: any }>(`/v1/public/events/${slug}`)
    return data.event
  } catch (err) {
    console.error('Failed to load event data:', err)
    return null
  }
}

export default async function GuestEventPage({ params }: EventPageProps) {
  const event = await getEventData(params.slug)

  if (!event) {
    notFound()
  }

  const hostName = (event.profiles as any)?.full_name || 'the Hosts'
  const subEvents = event.sub_events || []
  const themeColor = event.branding_color || '#EA580C'

  return (
    <div className="min-h-screen bg-white" style={{ '--theme-color': themeColor } as any}>
      <style jsx global>{`
        .bg-theme { background-color: var(--theme-color); }
        .text-theme { color: var(--theme-color); }
        .border-theme { border-color: var(--theme-color); }
        .ring-theme { --tw-ring-color: var(--theme-color); }
      `}</style>
      <section className="relative h-[60vh] min-h-[500px] w-full flex items-center justify-center overflow-hidden">
        {(event.cover_image_url || event.cover_image) && (
          <img
            src={event.cover_image_url || event.cover_image}
            alt="Cover"
            className="absolute inset-0 w-full h-full object-cover"
          />
        )}
        <div className="absolute inset-0 bg-black/40 backdrop-blur-[1px]" />
        <div className="relative text-center px-6 space-y-6 max-w-3xl">
          <div className="inline-flex items-center gap-2 bg-white/10 backdrop-blur-md border border-white/20 px-4 py-2 rounded-full text-white text-xs font-bold uppercase tracking-widest animate-in fade-in slide-in-from-top-4 duration-1000">
            <Sparkles className="h-4 w-4 text-theme/70" />
            Official Invitation
          </div>
          <h1 className="text-4xl sm:text-7xl font-bold text-white tracking-tight animate-in fade-in slide-in-from-bottom-4 duration-1000">
            {event.title}
          </h1>
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4 sm:gap-8 text-white/90 font-medium animate-in fade-in duration-1000 delay-300">
            <div className="flex items-center gap-2">
              <Calendar className="h-5 w-5 text-theme/70" />
              {event.date_start || event.start_date
                ? format(new Date(event.date_start || event.start_date), 'PPP')
                : 'Date TBD'}
            </div>
            <div className="flex items-center gap-2">
              <Heart className="h-5 w-5 text-theme/70" />
              Hosted by {hostName}
            </div>
          </div>
        </div>

        <div className="absolute bottom-10 left-0 right-0 px-6 sm:hidden">
          <div className="flex gap-4">
            <Link href="#rsvp" className="flex-1">
              <Button className="w-full h-14 bg-white text-zinc-900 font-bold rounded-2xl shadow-xl">RSVP Now</Button>
            </Link>
            <Link href={`/${params.slug}/shagun`} className="flex-1">
              <Button className="w-full h-14 bg-theme text-white font-bold rounded-2xl shadow-xl">Gifting</Button>
            </Link>
          </div>
        </div>
      </section>

      <main className="max-w-4xl mx-auto px-6 py-16 space-y-20">
        {event.description && (
          <section className="text-center space-y-4">
            <h2 className="text-zinc-400 text-xs font-bold uppercase tracking-widest">A Note from the Family</h2>
            <p className="text-2xl sm:text-3xl text-zinc-900 font-serif leading-relaxed line-clamp-4">
              "{event.description}"
            </p>
          </section>
        )}

        <section className="text-center space-y-10">
          <h2 className="text-2xl font-bold text-zinc-900">Are you joining us?</h2>
          <RSVPForm
            eventId={event.id}
            eventTitle={event.title}
            eventSlug={params.slug}
            subEvents={subEvents}
          />
        </section>
        <section className="space-y-10">
          <div className="flex items-center justify-between">
            <h2 className="text-2xl font-bold text-zinc-900 flex items-center gap-3">
              <Clock className="w-8 h-8 text-theme" />
              Event Schedule
            </h2>
            <div className="h-px flex-1 bg-zinc-100 ml-6 hidden sm:block" />
          </div>

          <div className="grid grid-cols-1 gap-6">
            {subEvents.length > 0 ? (
              subEvents.map((sub: any) => (
                <div key={sub.id} className="group flex flex-col sm:flex-row gap-6 p-6 rounded-[32px] border border-zinc-100 bg-white hover:border-orange-100 hover:shadow-xl hover:shadow-orange-50/50 transition-all duration-500">
                  <div className="w-full sm:w-48 h-32 sm:h-auto bg-zinc-50 rounded-2xl flex flex-col items-center justify-center text-center p-4">
                     <span className="text-zinc-400 text-[10px] font-bold uppercase tracking-widest mb-1">
                       {format(new Date(sub.date_time || sub.starts_at), 'EEE')}
                     </span>
                     <span className="text-2xl font-bold text-zinc-900">
                       {format(new Date(sub.date_time || sub.starts_at), 'do MMM')}
                     </span>
                     <span className="text-zinc-500 text-xs mt-1">
                       {format(new Date(sub.date_time || sub.starts_at), 'p')}
                     </span>
                  </div>
                  
                  <div className="flex-1 space-y-4">
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <Badge variant="outline" className="text-[10px] uppercase font-bold text-theme border-orange-100 bg-orange-50/50">
                          {sub.type || 'Main Event'}
                        </Badge>
                        <Heart className="w-3 h-3 text-zinc-200" />
                      </div>
                      <h3 className="text-xl font-bold text-zinc-900">{sub.name}</h3>
                    </div>

                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                      <div className="flex items-start gap-3">
                        <MapPin className="h-5 w-5 text-zinc-400 shrink-0" />
                        <div>
                          <p className="text-sm font-bold text-zinc-800">{sub.venue_name || sub.venue_label}</p>
                          <p className="text-xs text-zinc-500">{sub.venue_address || 'Address provided on card'}</p>
                        </div>
                      </div>
                      {sub.dress_code && (
                        <div className="flex items-start gap-3">
                          <PartyPopper className="h-5 w-5 text-zinc-400 shrink-0" />
                          <div>
                            <p className="text-xs font-bold text-zinc-400 uppercase tracking-wide">Dress Code</p>
                            <p className="text-sm font-bold text-zinc-800">{sub.dress_code}</p>
                          </div>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              ))
            ) : (
              <div className="py-12 text-center text-zinc-400 bg-zinc-50 rounded-3xl border border-dashed border-zinc-200">
                Schedule items will be posted soon.
              </div>
            )}
          </div>
        </section>

        <section id="rsvp" className="bg-zinc-900 rounded-[40px] p-8 sm:p-16 text-white text-center space-y-10 relative overflow-hidden">
          <div className="absolute top-0 right-0 p-10 opacity-10 blur-xl">
             <Heart className="w-64 h-64 text-orange-500" />
          </div>
          
          <div className="max-w-xl mx-auto space-y-4 relative z-10">
            <h2 className="text-3xl sm:text-5xl font-bold font-heading tracking-tight">Are you joining us?</h2>
            <p className="text-zinc-400 font-medium">Please confirm your presence to help us plan the arrangements perfectly.</p>
          </div>

          <div className="max-w-lg mx-auto relative z-10">
            <RSVPForm eventId={event.id} eventTitle={event.title} eventSlug={params.slug} subEvents={subEvents} />
          </div>
        </section>

        <section className="grid grid-cols-1 sm:grid-cols-2 gap-6">
           <Link href={`/${params.slug}/gallery`} className="group">
              <div className="bg-zinc-900 rounded-[32px] p-8 text-white space-y-4 hover:bg-black transition-all">
                 <div className="w-12 h-12 bg-white/10 rounded-2xl flex items-center justify-center border border-white/20">
                    <ImageIcon className="w-6 h-6" />
                 </div>
                 <h3 className="text-xl font-bold">Event Gallery</h3>
                 <p className="text-zinc-400 text-sm">Relive the moments through the lens of all our guests.</p>
                 <div className="pt-2 flex items-center gap-2 text-xs font-bold text-orange-500">
                    View Photos <ChevronRight className="w-3 h-3" />
                 </div>
              </div>
           </Link>

           <Link href={`/${params.slug}/memory-book`} className="group">
              <div className="bg-white border border-zinc-200 rounded-[32px] p-8 space-y-4 hover:border-orange-200 transition-all shadow-sm">
                 <div className="w-12 h-12 bg-orange-100 rounded-2xl flex items-center justify-center text-theme">
                    <Heart className="w-6 h-6 fill-current" />
                 </div>
                 <h3 className="text-xl font-bold text-zinc-900">Memory Book</h3>
                 <p className="text-zinc-500 text-sm">A digital keepsake of all the blessings and highlights.</p>
                 <div className="pt-2 flex items-center gap-2 text-xs font-bold text-theme">
                    Open Souvenir <ChevronRight className="w-3 h-3" />
                 </div>
              </div>
           </Link>
        </section>

        <section className="bg-orange-50 rounded-[40px] p-8 sm:p-12 border border-orange-100 flex flex-col sm:flex-row items-center gap-8">
           <div className="w-20 h-20 bg-theme rounded-2xl flex items-center justify-center text-white shadow-xl shadow-orange-200">
              <IndianRupee className="w-10 h-10" />
           </div>
           <div className="flex-1 text-center sm:text-left space-y-2">
              <h3 className="text-2xl font-bold text-zinc-900">Digital Shagun</h3>
              <p className="text-zinc-600">Send your blessings and gifts directly to the hosts via secure UPI payment.</p>
           </div>
           <Link href={`/${params.slug}/shagun`}>
              <Button className="h-14 px-8 bg-theme hover:bg-orange-700 text-white font-bold rounded-2xl group">
                 Open Gifting Box
                 <ChevronRight className="ml-2 h-5 w-5 group-hover:translate-x-1 transition-transform" />
              </Button>
           </Link>
        </section>

      </main>

      <footer className="py-10 text-center border-t border-zinc-100 bg-zinc-50/50">
        <div className="flex items-center justify-center gap-2 mb-4">
           <div className="w-6 h-6 bg-orange-600 rounded flex items-center justify-center text-white text-[10px] font-bold">U</div>
           <p className="text-xs font-bold text-zinc-900 tracking-tighter uppercase">UTSAV PLATFORM</p>
        </div>
        <p className="text-[10px] text-zinc-400 uppercase tracking-widest font-bold">Operating System for India's Events</p>
      </footer>
    </div>
  )
}
