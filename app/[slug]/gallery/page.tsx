import { supabase } from '@/lib/supabase/client'
import { notFound } from 'next/navigation'
import { GalleryGrid } from '@/components/event/GalleryGrid'
import { ImageIcon, ChevronLeft } from 'lucide-react'
import Link from 'next/link'
import { Button } from '@/components/ui/button'

interface GalleryPageProps {
  params: {
    slug: string
  }
}

async function getEventData(slug: string) {
  const { data: event } = await supabase
    .from('events')
    .select('id, title, items:sub_events(name)')
    .eq('slug', slug)
    .single()
  return event
}

export default async function GalleryPage({ params }: GalleryPageProps) {
  const event = await getEventData(params.slug)
  if (!event) notFound()

  return (
    <main className="min-h-screen bg-white pb-20">
      {/* Header */}
      <div className="bg-white border-b border-zinc-100 px-6 py-4 sticky top-0 z-20 backdrop-blur-md bg-white/80">
        <div className="max-w-5xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link href={`/${params.slug}`}>
               <Button variant="ghost" size="icon" className="rounded-full">
                 <ChevronLeft className="h-5 w-5" />
               </Button>
            </Link>
            <div>
               <h1 className="font-bold text-lg text-zinc-900 leading-none">{event.title}</h1>
               <p className="text-[10px] text-orange-600 font-bold uppercase tracking-widest mt-1">Live Gallery</p>
            </div>
          </div>
        </div>
      </div>

      <div className="max-w-6xl mx-auto p-6 md:p-10 space-y-12">
        <div className="text-center space-y-4 max-w-2xl mx-auto">
           <div className="w-16 h-16 bg-orange-100 rounded-[24px] flex items-center justify-center mx-auto text-orange-600">
              <ImageIcon className="w-8 h-8" />
           </div>
           <h2 className="text-3xl sm:text-5xl font-bold font-heading text-zinc-900 tracking-tight">Capture & Relive</h2>
           <p className="text-zinc-500 font-medium">Every smile, every dance, and every blessing captured by the ones who matter most.</p>
        </div>

        <GalleryGrid eventId={event.id} />
      </div>
    </main>
  )
}
