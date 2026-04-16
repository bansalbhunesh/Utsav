import { notFound } from 'next/navigation'
import { Heart, Sparkles, Image as ImageIcon, IndianRupee, MessageCircle } from 'lucide-react'
import { paymentService } from '@/lib/services/PaymentService'
import { guestApiFetch } from '@/lib/api'
import { parseMemoryPayloadResponse, parsePublicEventResponse, parsePublicGalleryResponse } from '@/lib/contracts/public'

interface MemoryBookProps {
  params: Promise<{
    slug: string
  }>
}

async function getStats(slug: string) {
  try {
    const [eventData, galleryData, memoryData] = await Promise.all([
      guestApiFetch<unknown>(`/v1/public/events/${slug}`),
      guestApiFetch<unknown>(`/v1/public/events/${slug}/gallery`),
      guestApiFetch<unknown>(`/v1/public/memory/${slug}-memory`).catch(() => null),
    ])

    const event = parsePublicEventResponse(eventData).event
    if (!event) return null

    const gallery = parsePublicGalleryResponse(galleryData)
    const parsedMemory = memoryData ? parseMemoryPayloadResponse(memoryData) : null
    const highlights = parsedMemory?.payload?.highlights || {}
    return {
      event,
      totalWishes: Number(highlights.shagun_count || 0),
      totalShagunPaise: Number(highlights.shagun_total_paise || 0),
      totalPhotos: Number(gallery.assets?.length || 0),
      featuredWishes: parsedMemory?.payload?.featured_wishes || [],
    }
  } catch (err) {
    console.error('Failed to load memory book stats:', err)
    return null
  }
}

export default async function MemoryBookPage({ params }: MemoryBookProps) {
  const { slug } = await params
  const data = await getStats(slug)
  if (!data) notFound()

  const { event, totalWishes, totalShagunPaise, totalPhotos, featuredWishes } = data

  return (
    <main className="min-h-screen bg-linear-to-b from-orange-50/50 to-white pb-32">
      {/* Premium Keepsake Header */}
      <div className="relative pt-20 pb-20 text-center space-y-6 overflow-hidden">
        <div className="absolute top-10 left-1/2 -translate-x-1/2 w-64 h-64 bg-orange-400/10 rounded-full blur-3xl -z-10" />
        
        <div className="inline-flex items-center gap-2 bg-orange-100 text-orange-700 px-4 py-2 rounded-full text-[10px] font-bold uppercase tracking-widest animate-in fade-in duration-1000">
           <Heart className="h-4 w-4 fill-current" />
           Official Memory Book
        </div>
        
        <h1 className="text-4xl sm:text-7xl font-bold font-heading text-zinc-900 tracking-tighter px-4 text-balance">
          A Celebration for <br />
          <span className="italic text-orange-600">the Ages</span>
        </h1>
        
        <p className="text-zinc-500 font-medium max-w-lg mx-auto px-6">
          Thank you for making {event.title} so special. Your presence and blessings mean the world to us.
        </p>
      </div>

      <div className="max-w-4xl mx-auto px-6 space-y-16">
        {/* Stats Grid */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-6">
           <div className="p-8 rounded-[40px] bg-white border border-orange-100 shadow-xl shadow-orange-100/20 text-center space-y-2 group hover:-translate-y-2 transition-transform duration-500">
              <div className="w-12 h-12 bg-orange-600 rounded-2xl flex items-center justify-center text-white mx-auto mb-4">
                 <Heart className="w-6 h-6 fill-current" />
              </div>
              <p className="text-4xl font-bold font-heading text-zinc-900">{totalWishes}</p>
              <p className="text-xs font-bold text-zinc-400 uppercase tracking-widest">Total Blessings</p>
           </div>

           <div className="p-8 rounded-[40px] bg-zinc-900 shadow-xl shadow-zinc-200 text-center space-y-2 group hover:-translate-y-2 transition-transform duration-500">
              <div className="w-12 h-12 bg-white/10 backdrop-blur-md rounded-2xl flex items-center justify-center text-white mx-auto mb-4 border border-white/20">
                 <IndianRupee className="w-6 h-6" />
              </div>
              <p className="text-4xl font-bold font-heading text-white">{paymentService.formatINR(totalShagunPaise / 100)}</p>
              <p className="text-xs font-bold text-zinc-500 uppercase tracking-widest">Digital Shagun</p>
           </div>

           <div className="p-8 rounded-[40px] bg-white border border-orange-100 shadow-xl shadow-orange-100/20 text-center space-y-2 group hover:-translate-y-2 transition-transform duration-500">
              <div className="w-12 h-12 bg-orange-100 rounded-2xl flex items-center justify-center text-orange-600 mx-auto mb-4">
                 <ImageIcon className="w-6 h-6" />
              </div>
              <p className="text-4xl font-bold font-heading text-zinc-900">{totalPhotos}</p>
              <p className="text-xs font-bold text-zinc-400 uppercase tracking-widest">Captured Moments</p>
           </div>
        </div>

        {/* Featured Messages */}
        <section className="space-y-8">
           <div className="flex items-center gap-4">
              <div className="h-px flex-1 bg-zinc-100" />
              <h2 className="text-lg font-bold text-zinc-900 uppercase tracking-widest flex items-center gap-2">
                 <MessageCircle className="w-5 h-5 text-orange-600" />
                 Guest Wishes
              </h2>
              <div className="h-px flex-1 bg-zinc-100" />
           </div>

           <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {featuredWishes.filter((s) => s.blessing_note).slice(0, 10).map((shagun) => (
                <div key={shagun.id} className="p-8 rounded-[32px] bg-white border border-zinc-100 shadow-sm relative overflow-hidden group hover:border-orange-200 transition-colors">
                   <div className="absolute -top-4 -right-4 opacity-5 group-hover:opacity-10 transition-opacity">
                      <Heart className="w-24 h-24 text-orange-600" />
                   </div>
                   <p className="text-lg text-zinc-700 font-serif leading-relaxed line-clamp-4 relative z-10 italic">
                     &quot;{shagun.blessing_note}&quot;
                   </p>
                   <div className="mt-6 flex items-center gap-3 relative z-10">
                      <div className="w-8 h-8 rounded-full bg-orange-100 flex items-center justify-center text-orange-700 text-[10px] font-bold">
                         {shagun.meta?.sender_name?.charAt(0) || 'G'}
                      </div>
                      <p className="text-sm font-bold text-zinc-900">{shagun.meta?.sender_name || 'Guest'}</p>
                   </div>
                </div>
              ))}
           </div>
        </section>

        {/* Closing Photo Highlight */}
        <div className="rounded-[48px] bg-zinc-100 aspect-video flex flex-col items-center justify-center text-center space-y-4 px-10 relative overflow-hidden border-8 border-white shadow-2xl">
           <Sparkles className="w-12 h-12 text-orange-400 animate-pulse" />
           <p className="text-2xl font-bold font-heading text-zinc-400 uppercase tracking-widest">Official Film Launching Soon</p>
           <p className="text-sm text-zinc-400 max-w-sm font-medium">
             The full event highlights and official video will be uploaded here by the vendors.
           </p>
        </div>
      </div>
    </main>
  )
}
