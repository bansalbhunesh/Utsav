'use client'

import { useState, useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useEventCreationStore } from '@/store/event-creation-store'
import { Loader2, Check, X } from 'lucide-react'
import { useDebounce } from '@/hooks/use-debounce'

const formSchema = z.object({
  title: z.string().min(2, 'Title must be at least 2 characters'),
  event_type: z.enum(['WEDDING', 'BIRTHDAY', 'PARTY', 'GET_TOGETHER']),
  slug: z.string().min(3, 'Slug must be at least 3 characters').regex(/^[a-z0-9-]+$/, 'Slug must only contain lowercase letters, numbers, and hyphens'),
  upi_id: z.string().min(3, 'UPI ID is required').regex(/^[\w.-]+@[\w.-]+$/, 'Please enter a valid UPI ID (e.g. name@upi)'),
})

export function BasicsStep() {
  const { eventData, setEventData, setCurrentStep } = useEventCreationStore()
  const [isCheckingSlug, setIsCheckingSlug] = useState(false)
  const [isSlugAvailable, setIsSlugAvailable] = useState<boolean | null>(null)

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      title: eventData.title || '',
      event_type: (eventData.event_type as z.infer<typeof formSchema>['event_type']) || 'WEDDING',
      slug: eventData.slug || '',
      upi_id: eventData.upi_id || '',
    },
  })

  const slugValue = form.watch('slug')
  const debouncedSlug = useDebounce(slugValue, 500)

  useEffect(() => {
    async function checkSlug() {
      if (debouncedSlug && debouncedSlug.length >= 3 && /^[a-z0-9-]+$/.test(debouncedSlug)) {
        setIsCheckingSlug(true)
        try {
          const res = await fetch(`/api/events/check-slug?slug=${debouncedSlug}`)
          const data = await res.json()
          setIsSlugAvailable(data.available)
        } catch (err) {
          console.error('Slug check failed', err)
        } finally {
          setIsCheckingSlug(false)
        }
      } else {
        setIsSlugAvailable(null)
      }
    }
    checkSlug()
  }, [debouncedSlug])

  function onSubmit(values: z.infer<typeof formSchema>) {
    if (isSlugAvailable === false) return
    setEventData(values)
    setCurrentStep(2)
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
        <FormField
          control={form.control}
          name="title"
          render={({ field }) => (
            <FormItem>
              <FormLabel className="text-zinc-700 font-semibold text-base">Event Title</FormLabel>
              <FormControl>
                <Input placeholder="Ankur & Priya's Wedding" {...field} className="h-12 rounded-xl" />
              </FormControl>
              <FormDescription>
                This will be displayed at the top of your event page.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="event_type"
          render={({ field }) => (
            <FormItem>
              <FormLabel className="text-zinc-700 font-semibold text-base">Event Type</FormLabel>
              <Select onValueChange={field.onChange} defaultValue={field.value}>
                <FormControl>
                  <SelectTrigger className="h-12 rounded-xl">
                    <SelectValue placeholder="Select event type" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  <SelectItem value="WEDDING">Wedding</SelectItem>
                  <SelectItem value="BIRTHDAY">Birthday</SelectItem>
                  <SelectItem value="PARTY">Party</SelectItem>
                  <SelectItem value="GET_TOGETHER">Get-Together</SelectItem>
                </SelectContent>
              </Select>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="slug"
          render={({ field }) => (
            <FormItem>
              <FormLabel className="text-zinc-700 font-semibold text-base">Custom URL Slug</FormLabel>
              <div className="relative">
                <FormControl>
                  <Input 
                    placeholder="ankur-priya" 
                    {...field} 
                    className="h-12 rounded-xl pr-10" 
                    onChange={(e) => {
                      const val = e.target.value.toLowerCase().replace(/\s+/g, '-')
                      field.onChange(val)
                    }}
                  />
                </FormControl>
                <div className="absolute right-3 top-1/2 -translate-y-1/2">
                  {isCheckingSlug ? (
                    <Loader2 className="h-4 w-4 animate-spin text-zinc-400" />
                  ) : isSlugAvailable === true ? (
                    <Check className="h-4 w-4 text-green-500" />
                  ) : isSlugAvailable === false ? (
                    <X className="h-4 w-4 text-red-500" />
                  ) : null}
                </div>
              </div>
              <FormDescription>
                Your event will be at: utsav.app/{field.value || 'your-slug'}
              </FormDescription>
              <FormMessage />
              {isSlugAvailable === false && (
                <p className="text-sm font-medium text-red-500">This slug is already taken.</p>
              )}
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="upi_id"
          render={({ field }) => (
            <FormItem>
              <FormLabel className="text-zinc-700 font-semibold text-base">Host UPI ID (VPA)</FormLabel>
              <FormControl>
                <Input placeholder="name@okaxis" {...field} className="h-12 rounded-xl" />
              </FormControl>
              <FormDescription>
                Shagun will be sent directly to this UPI ID.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <div className="pt-4">
          <Button 
            type="submit" 
            disabled={isSlugAvailable === false || isCheckingSlug}
            className="w-full h-12 bg-orange-600 hover:bg-orange-700 text-white font-bold rounded-xl shadow-lg shadow-orange-200"
          >
            Next Step
          </Button>
        </div>
      </form>
    </Form>
  )
}
