'use client'

import { useState, useEffect } from 'react'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Plus, Image as ImageIcon, Heart, Info, Camera } from 'lucide-react'
import { supabase } from '@/lib/supabase/client'
import { cn } from '@/lib/utils'

const SAMPLE_MODERN_WEDDING = [
  'https://images.unsplash.com/photo-1519741497674-611481863552?auto=format&fit=crop&q=80&w=600',
  'https://images.unsplash.com/photo-1511795409834-ef04bbd61622?auto=format&fit=crop&q=80&w=600',
  'https://images.unsplash.com/photo-1502635385003-ee1e6a1a742d?auto=format&fit=crop&q=80&w=600',
  'https://images.unsplash.com/photo-1519225421980-715cb0215aed?auto=format&fit=crop&q=80&w=600',
]

export function GalleryGrid({ eventId }: { eventId: string }) {
  const [activeTab, setActiveTab] = useState<'official' | 'guests'>('official')
  const [media, setMedia] = useState<any[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    async function fetchMedia() {
      const { data } = await supabase
        .from('event_media')
        .select('*')
        .eq('event_id', eventId)
        .order('created_at', { ascending: false })
      setMedia(data || [])
      setLoading(false)
    }
    fetchMedia()
  }, [eventId])

  const filteredMedia = activeTab === 'official' 
    ? (media.filter(m => m.is_official).length > 0 ? media.filter(m => m.is_official) : SAMPLE_MODERN_WEDDING.map(url => ({ url, is_official: true, id: url })))
    : media.filter(m => !m.is_official)

  return (
    <div className="space-y-8 animate-in fade-in duration-700">
      <div className="flex flex-col sm:flex-row items-center justify-between gap-6">
        <div className="flex p-1.5 bg-zinc-100 rounded-2xl w-full sm:w-auto">
          <button
            onClick={() => setActiveTab('official')}
            className={cn(
              "flex-1 sm:flex-none px-6 py-2.5 rounded-xl text-sm font-bold transition-all",
              activeTab === 'official' ? "bg-white text-zinc-900 shadow-sm" : "text-zinc-500"
            )}
          >
            Official Album
          </button>
          <button
            onClick={() => setActiveTab('guests')}
            className={cn(
              "flex-1 sm:flex-none px-6 py-2.5 rounded-xl text-sm font-bold transition-all",
              activeTab === 'guests' ? "bg-white text-zinc-900 shadow-sm" : "text-zinc-500"
            )}
          >
            Guest Photos
          </button>
        </div>

        <Button className="w-full sm:w-auto h-12 px-6 bg-orange-600 hover:bg-orange-700 text-white font-bold rounded-xl shadow-lg shadow-orange-100">
           <Camera className="mr-2 h-5 w-5" />
           Share a Moment
        </Button>
      </div>

      {loading ? (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
           {[1, 2, 3, 4].map(i => <div key={i} className="aspect-square bg-zinc-100 animate-pulse rounded-2xl" />)}
        </div>
      ) : filteredMedia.length === 0 ? (
        <div className="py-20 text-center bg-zinc-50 rounded-[32px] border border-dashed border-zinc-200">
           <div className="w-16 h-16 bg-white rounded-full flex items-center justify-center mx-auto mb-4 border border-zinc-100 shadow-sm">
              <ImageIcon className="h-8 w-8 text-zinc-300" />
           </div>
           <p className="text-zinc-500 font-medium">No photos have been shared in this category yet.</p>
        </div>
      ) : (
        <div className="columns-2 md:columns-4 gap-4 space-y-4">
           {filteredMedia.map((m, i) => (
             <div key={m.id || i} className="relative group overflow-hidden rounded-[24px] break-inside-avoid shadow-sm hover:shadow-xl transition-all duration-500">
                <img 
                  src={m.url || m} 
                  alt="" 
                  className="w-full object-cover transition-transform duration-700 group-hover:scale-110" 
                />
                <div className="absolute inset-0 bg-linear-to-t from-black/60 via-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300 flex flex-col justify-end p-4">
                   <div className="flex items-center justify-between text-white">
                      <div className="flex items-center gap-2">
                         <div className="w-6 h-6 rounded-full bg-white/20 backdrop-blur-md border border-white/30 flex items-center justify-center">
                            <Heart className="w-3 h-3 fill-white" />
                         </div>
                         <span className="text-[10px] font-bold uppercase tracking-widest">A Blessing</span>
                      </div>
                      <span className="text-[10px] text-white/60">Module 6</span>
                   </div>
                </div>
             </div>
           ))}
        </div>
      )}

      <div className="flex items-start gap-4 p-6 bg-orange-50 border border-orange-100 rounded-3xl">
         <Info className="h-5 w-5 text-orange-600 mt-0.5 shrink-0" />
         <p className="text-sm text-orange-800 leading-relaxed">
            <strong>v1.5 Moderation</strong>: Guest photos are automatically scanned for safety. Only the host can delete or "Pin" photos to the official highlight reel.
         </p>
      </div>
    </div>
  )
}
