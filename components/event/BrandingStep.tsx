'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { useEventCreationStore } from '@/store/event-creation-store'
import { Check, Loader2, Sparkles, Palette } from 'lucide-react'
import { cn } from '@/lib/utils'
import { supabase } from '@/lib/supabase/client'
import { useAuthStore } from '@/store/auth-store'

const themes = [
  { id: 'default', name: 'UTSAV Classic', primary: '#EA580C', secondary: '#FFF7ED' },
  { id: 'royal', name: 'Royal Saffron', primary: '#F59E0B', secondary: '#FEF3C7' },
  { id: 'modern', name: 'Modern White', primary: '#18181B', secondary: '#F4F4F5' },
  { id: 'rose', name: 'Rose Quartz', primary: '#E11D48', secondary: '#FFF1F2' },
  { id: 'emerald', name: 'Emerald Garden', primary: '#059669', secondary: '#ECFDF5' },
  { id: 'marigold', name: 'Vibrant Marigold', primary: '#F97316', secondary: '#FFFBEB' },
  { id: 'midnight', name: 'Midnight Gala', primary: '#1E3A8A', secondary: '#EFF6FF' },
  { id: 'lavender', name: 'Lavender Dream', primary: '#8B5CF6', secondary: '#F5F3FF' },
  { id: 'sandstone', name: 'Sandstone Earth', primary: '#A16207', secondary: '#FEFCE8' },
  { id: 'starlight', name: 'Starlight Silver', primary: '#4B5563', secondary: '#F9FAFB' },
]

export function BrandingStep() {
  const router = useRouter()
  const { user } = useAuthStore()
  const { eventData, subEvents, setEventData, setCurrentStep, reset } = useEventCreationStore()
  const [selectedTheme, setSelectedTheme] = useState(eventData.branding?.theme_name || 'default')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleFinish = async () => {
    if (!user) {
      setError('You must be signed in to create an event.')
      return
    }

    setIsSubmitting(true)
    setError(null)

    const finalBranding = themes.find(t => t.id === selectedTheme)

    try {
      // 1. Create the event
      const { data: event, error: eventError } = await supabase
        .from('events')
        .insert({
          ...eventData,
          upi_id: eventData.upi_id,
          owner_user_id: user.id,
          branding_color: finalBranding?.primary,
          branding: {
            theme_name: selectedTheme,
          }
        })
        .select()
        .single()

      if (eventError) throw eventError

      // 2. Add sub-events
      if (subEvents.length > 0) {
        const { error: subError } = await supabase
          .from('sub_events')
          .insert(
            subEvents.map(sub => ({
              ...sub,
              event_id: event.id
            }))
          )
        if (subError) throw subError
      }

      // 3. Add owner as OWNER in event_members
      const { error: memberError } = await supabase
        .from('event_members')
        .insert({
          event_id: event.id,
          user_id: user.id,
          role: 'OWNER'
        })
      if (memberError) throw memberError

      // 4. Add co-owner if provided
      if (eventData.co_owner_name && eventData.co_owner_contact) {
        // NOTE: In a real app, this would send an invite or look up a user by contact.
        // For MVP, we'll store the intent/contact. For now, we skip if no user ID found.
        // But we can add it to a 'guest' list or a dedicated 'invites' table.
        // For now, we'll just log it as a guest with the CO_OWNER side-info or similar.
        const { error: coOwnerError } = await supabase
        .from('event_members')
        .insert({
          event_id: event.id,
          user_id: user.id, // Using current user for now as a placeholder for co-owner logic
          role: 'CO_OWNER',
          // Metadata or temporary contact info could go here in Phase 2
        })
        if (coOwnerError) console.error('Co-owner addition skipped/failed', coOwnerError)
      }

      // Success!
      reset()
      router.push(`/dashboard?success=true&event=${event.slug}`)

    } catch (err: any) {
      console.error('Event creation failed', err)
      setError(err.message || 'An unexpected error occurred.')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="space-y-8">
      <div className="space-y-2">
        <h3 className="text-xl font-bold font-heading text-zinc-900">Choose Your Theme</h3>
        <p className="text-sm text-zinc-500">Pick a visual style that matches your celebration.</p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        {themes.map((theme) => (
          <button
            key={theme.id}
            onClick={() => setSelectedTheme(theme.id)}
            className={cn(
              "relative flex flex-col p-4 rounded-2xl border-2 transition-all text-left group",
              selectedTheme === theme.id 
                ? "border-orange-600 bg-orange-50/30 ring-4 ring-orange-50" 
                : "border-zinc-100 bg-white hover:border-zinc-200"
            )}
          >
            <div className="flex items-center justify-between mb-3">
               <div 
                 className="w-10 h-10 rounded-lg shadow-sm flex items-center justify-center"
                 style={{ backgroundColor: theme.primary }}
               >
                 <Palette className="w-5 h-5 text-white opacity-80" />
               </div>
               {selectedTheme === theme.id && (
                 <div className="bg-orange-600 text-white rounded-full p-1 shadow-md">
                   <Check className="w-4 h-4" />
                 </div>
               )}
            </div>
            <p className={cn(
              "font-bold transition-colors",
              selectedTheme === theme.id ? "text-orange-900" : "text-zinc-900"
            )}>
              {theme.name}
            </p>
            <div className="flex gap-1 mt-2">
               <div className="h-2 w-8 rounded-full" style={{ backgroundColor: theme.primary }} />
               <div className="h-2 w-4 rounded-full" style={{ backgroundColor: theme.secondary }} />
            </div>
          </button>
        ))}
      </div>

      {error && (
        <p className="text-sm text-red-500 font-medium bg-red-50 p-3 rounded-xl border border-red-100 text-center">
          {error}
        </p>
      )}

      <div className="flex gap-4 pt-4 border-t border-zinc-100 mt-8">
        <Button 
          type="button" 
          variant="outline"
          onClick={() => setCurrentStep(3)}
          className="flex-1 h-12 rounded-xl font-bold"
          disabled={isSubmitting}
        >
          Back
        </Button>
        <Button 
          type="button" 
          onClick={handleFinish}
          disabled={isSubmitting}
          className="flex-1 h-12 bg-orange-600 hover:bg-orange-700 text-white font-bold rounded-xl shadow-lg shadow-orange-200"
        >
          {isSubmitting ? (
            <Loader2 className="h-5 w-5 animate-spin" />
          ) : (
            <>
              Launch Event
              <Sparkles className="ml-2 h-4 w-4" />
            </>
          )}
        </Button>
      </div>
    </div>
  )
}
