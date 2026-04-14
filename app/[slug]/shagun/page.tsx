import { notFound } from 'next/navigation'
import { ShagunForm } from '@/components/event/ShagunForm'
import { Heart } from 'lucide-react'
import { guestApiFetch } from '@/lib/api'
import { parsePublicEventResponse } from '@/lib/contracts/public'

interface ShagunPageProps {
  params: {
    slug: string
  }
}

async function getEventBySlug(slug: string) {
  try {
    const data = await guestApiFetch<unknown>(`/v1/public/events/${slug}`)
    return parsePublicEventResponse(data).event
  } catch (err) {
    console.error('Failed to resolve event for shagun:', err)
    return null
  }
}

export default async function ShagunPage({ params }: ShagunPageProps) {
  const event = await getEventBySlug(params.slug)
  if (!event) notFound()
  
  const hostName = event.profiles?.full_name || 'the Host'

  return (
    <main className="min-h-screen bg-[#FAFAFA] pb-20">
      <div 
        className="h-[300px] w-full bg-zinc-900 flex items-center justify-center relative overflow-hidden"
        style={{
          backgroundImage: (event.cover_image_url || event.cover_image) ? `url(${event.cover_image_url || event.cover_image})` : 'none',
          backgroundSize: 'cover',
          backgroundPosition: 'center'
        }}
      >
        <div className="absolute inset-0 bg-black/60 backdrop-blur-[2px]" />
        <div className="relative text-center px-4 space-y-4">
          <div className="w-16 h-16 bg-white/10 backdrop-blur-md rounded-full flex items-center justify-center mx-auto border border-white/20">
            <Heart className="w-8 h-8 text-white fill-white/20" />
          </div>
          <div className="space-y-1">
            <h1 className="text-3xl font-bold font-heading text-white">{event.title}</h1>
            <p className="text-zinc-300 font-medium">Digital Shagun Box</p>
          </div>
        </div>
      </div>

      <div className="max-w-2xl mx-auto px-4 -mt-10 relative z-10">
        <div className="bg-white rounded-[32px] p-6 sm:p-10 shadow-xl shadow-zinc-200/50 border border-zinc-100">
          <div className="text-center mb-10 space-y-2">
            <h2 className="text-2xl font-bold text-zinc-900">Send Your Blessings</h2>
            <p className="text-zinc-500 font-medium">
              Your shagun will be sent directly to <span className="font-bold text-zinc-800">{hostName}</span> via UPI.
            </p>
          </div>
          <ShagunForm
            event={{ ...event, upi_id: event.upi_id || event.host_upi_vpa }}
            hostName={hostName}
          />
        </div>
      </div>
    </main>
  )
}
