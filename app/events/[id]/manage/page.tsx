'use client'

import { useCallback } from 'react'
import { useParams, useRouter } from 'next/navigation'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { 
  ChevronLeft, 
  Users, 
  IndianRupee, 
  Share2, 
  Settings, 
  QrCode,
  TrendingUp,
  Clock
} from 'lucide-react'
import { FastCashLogger } from '@/components/event/FastCashLogger'
import { VendorManager } from '@/components/event/VendorManager'
import { BroadcastCenter } from '@/components/event/BroadcastCenter'
import { paymentService } from '@/lib/services/PaymentService'
import { apiFetch } from '@/lib/api'
import { cn } from '@/lib/utils'
import { useQuery } from '@tanstack/react-query'
import { getUserFacingError } from '@/lib/error-messages'
import { parseHostEvent, parseHostShagunResponse } from '@/lib/contracts/host'

interface EventDetails {
  id: string
  slug: string
  title: string
}

interface ShagunItem {
  id: string
  channel: string
  amount_paise: number
  status: string
  created_at: string
  meta?: {
    sender_name?: string
  }
}

export default function EventManagePage() {
  const params = useParams()
  const router = useRouter()
  const eventId = params.id as string
  
  const fetchData = useCallback(async () => {
    const [eventData, shagunData] = await Promise.all([
      apiFetch<unknown>(`/v1/events/${eventId}`),
      apiFetch<unknown>(`/v1/events/${eventId}/shagun`),
    ])
    return {
      event: parseHostEvent(eventData),
      shagun: parseHostShagunResponse(shagunData).shagun as ShagunItem[],
    }
  }, [eventId])

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['event-manage', eventId],
    queryFn: fetchData,
  })

  const event = data?.event || null
  const shagun = data?.shagun || []
  const errorMessage = error ? getUserFacingError(error, 'Failed to load event management data.') : null

  const totalShagun = shagun.reduce((acc, curr) => acc + (Number(curr.amount_paise) || 0), 0) / 100
  const digitalCount = shagun.filter(s => s.channel === 'UPI').length
  const cashCount = shagun.filter(s => s.channel === 'CASH').length

  if (isLoading) return <div className="p-10 text-center animate-pulse">Loading event control...</div>
  if (errorMessage) {
    return (
      <div className="p-10 text-center space-y-4">
        <p className="text-red-600 font-bold">{errorMessage}</p>
        <Button variant="outline" onClick={() => void refetch()} className="rounded-xl">
          Retry
        </Button>
      </div>
    )
  }
  if (!event) return <div className="p-10 text-center">Event not found.</div>

  return (
    <div className="min-h-screen bg-zinc-50 pb-20">
      {/* Top Header */}
      <div className="bg-white border-b border-zinc-200 px-6 py-4 sticky top-0 z-20">
        <div className="max-w-5xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Button variant="ghost" size="icon" onClick={() => router.back()} className="rounded-full">
              <ChevronLeft className="h-5 w-5" />
            </Button>
            <h1 className="font-bold text-xl text-zinc-900 border-l pl-4 border-zinc-200">{event.title}</h1>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" className="rounded-xl hidden sm:flex">
              <Settings className="h-4 w-4 mr-2" />
              Settings
            </Button>
            <Link href={`/${event.slug}`} target="_blank">
              <Button size="sm" className="rounded-xl bg-orange-600 hover:bg-orange-700">
                <Share2 className="h-4 w-4 mr-2" />
                Live Page
              </Button>
            </Link>
          </div>
        </div>
      </div>

      <main className="max-w-5xl mx-auto p-6 space-y-8">
        {/* Quick Stats */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <Card className="p-6 rounded-[24px] border-none shadow-sm bg-zinc-900 text-white space-y-2">
            <div className="flex justify-between items-start">
               <IndianRupee className="h-5 w-5 text-orange-500" />
               <TrendingUp className="h-4 w-4 text-zinc-500" />
            </div>
            <p className="text-3xl font-bold font-heading">{paymentService.formatINR(totalShagun)}</p>
            <p className="text-xs text-zinc-400 font-medium">Total Shagun Received</p>
          </Card>
          
          <Card className="p-6 rounded-[24px] border-none shadow-sm bg-white space-y-2 border border-zinc-100">
            <Users className="h-5 w-5 text-blue-500" />
            <p className="text-3xl font-bold font-heading">{shagun.length}</p>
            <p className="text-xs text-zinc-500 font-medium">Blessings from Guests</p>
          </Card>

          <Card className="p-6 rounded-[24px] border-none shadow-sm bg-white space-y-2 border border-zinc-100">
            <div className="flex gap-4">
              <div>
                <p className="text-xl font-bold">{digitalCount}</p>
                <p className="text-[10px] text-zinc-500 uppercase font-bold tracking-wider">Digital</p>
              </div>
              <div className="border-l border-zinc-100 pl-4">
                <p className="text-xl font-bold">{cashCount}</p>
                <p className="text-[10px] text-zinc-500 uppercase font-bold tracking-wider">Physical</p>
              </div>
            </div>
            <div className="pt-2">
               <Badge className="bg-zinc-100 text-zinc-600 border-none text-[10px]">v1.5 Tracking</Badge>
            </div>
          </Card>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-5 gap-8">
          {/* Left: Logger & Actions */}
          <div className="lg:col-span-2 space-y-6">
            <FastCashLogger eventId={eventId} onSuccess={fetchData} />
            
            <VendorManager eventId={eventId} />

            <BroadcastCenter eventTitle={event.title} eventSlug={event.slug} />

            <Card className="p-6 rounded-[24px] border-zinc-200 bg-white space-y-4">
              <h3 className="font-bold text-zinc-900 flex items-center gap-2">
                <QrCode className="h-4 w-4 text-orange-600" />
                Event QR Code
              </h3>
              <div className="aspect-square bg-zinc-50 rounded-2xl flex items-center justify-center border-2 border-dashed border-zinc-200">
                 <p className="text-[10px] text-zinc-400 text-center px-6 uppercase font-bold tracking-widest">QR Code Generation Logic Post-Launch</p>
              </div>
              <Button variant="outline" className="w-full rounded-xl">Download QR Kit</Button>
            </Card>
          </div>

          {/* Right: Activity List */}
          <div className="lg:col-span-3 space-y-4">
            <div className="flex items-center justify-between px-2">
              <h3 className="font-bold text-lg text-zinc-900 flex items-center gap-2">
                <Clock className="h-5 w-5 text-zinc-400" />
                Recent Activity
              </h3>
              <Button variant="ghost" size="sm" className="text-xs font-bold text-orange-600">View All</Button>
            </div>

            <div className="space-y-3">
              {shagun.length === 0 ? (
                <div className="py-20 text-center bg-white rounded-3xl border border-dashed border-zinc-200">
                   <p className="text-sm text-zinc-400">No shagun recorded yet.</p>
                </div>
              ) : (
                  shagun.map((item) => (
                    <Card key={item.id} className="p-4 rounded-2xl border-none shadow-sm bg-white hover:bg-zinc-50 transition-colors flex items-center justify-between">
                      <div className="flex items-center gap-4">
                        <div className={cn(
                          "w-12 h-12 rounded-xl flex items-center justify-center font-bold text-lg",
                          item.channel === 'UPI' ? "bg-orange-100 text-orange-700" : "bg-green-100 text-green-700"
                        )}>
                          {(item.meta?.sender_name || 'G').charAt(0)}
                        </div>
                        <div>
                          <p className="font-bold text-zinc-900">{item.meta?.sender_name || 'Guest'}</p>
                          <p className="text-xs text-zinc-500">{item.channel} · {new Date(item.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</p>
                        </div>
                      </div>
                      <div className="text-right">
                         <p className="font-bold text-lg text-zinc-900">₹{item.amount_paise / 100}</p>
                         <Badge variant="outline" className="text-[10px] rounded-md h-5">
                            {item.status.toLowerCase().replace('_', ' ')}
                         </Badge>
                      </div>
                    </Card>
                  ))
              )}
            </div>
          </div>
        </div>
      </main>
    </div>
  )
}
