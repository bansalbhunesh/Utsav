'use client'

import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { format } from 'date-fns'
import { CalendarIcon, Loader2 } from 'lucide-react'
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
import { Textarea } from '@/components/ui/textarea'
import { Calendar } from '@/components/ui/calendar'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { useEventCreationStore } from '@/store/event-creation-store'
import { cn } from '@/lib/utils'
import { MuhuratHelper } from './MuhuratHelper'
import { Separator } from '@/components/ui/separator'
import { Users } from 'lucide-react'

const formSchema = z.object({
  start_date: z.date(),
  end_date: z.date(),
  description: z.string().optional(),
  cover_image: z.string().url('Please enter a valid image URL (placeholder for MVP)').or(z.string().length(0)),
  co_owner_name: z.string().optional(),
  co_owner_contact: z.string().optional(),
}).refine((data) => data.end_date >= data.start_date, {
  message: "End date cannot be before start date",
  path: ["end_date"],
})

export function DetailsStep() {
  const { eventData, setEventData, setCurrentStep } = useEventCreationStore()

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      start_date: eventData.start_date ? new Date(eventData.start_date as string) : undefined as any,
      end_date: eventData.end_date ? new Date(eventData.end_date as string) : undefined as any,
      description: eventData.description || '',
      cover_image: eventData.cover_image || '',
      co_owner_name: eventData.co_owner_name || '',
      co_owner_contact: eventData.co_owner_contact || '',
    },
  })

  function onSubmit(values: z.infer<typeof formSchema>) {
    setEventData({
      ...values,
      start_date: values.start_date.toISOString(),
      end_date: values.end_date.toISOString(),
    })
    setCurrentStep(3)
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
        {/* Muhurat Helper */}
        <div className="bg-orange-50/50 p-4 rounded-2xl border border-orange-100">
           <MuhuratHelper 
              selectedDate={form.watch('start_date')}
              onSelect={(date) => {
                form.setValue('start_date', date)
                // Default end date to same day if not set
                if (!form.getValues('end_date')) {
                  form.setValue('end_date', date)
                }
              }} 
           />
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
          <FormField
            control={form.control}
            name="start_date"
            render={({ field }) => (
              <FormItem className="flex flex-col">
                <FormLabel className="text-zinc-700 font-semibold text-base mb-1">Start Date</FormLabel>
                <Popover>
                  <PopoverTrigger>
                    <FormControl>
                      <Button
                        variant={"outline"}
                        className={cn(
                          "h-12 rounded-xl pl-3 text-left font-normal",
                          !field.value && "text-zinc-400"
                        )}
                      >
                        {field.value ? (
                          format(field.value, "PPP")
                        ) : (
                          <span>Pick a date</span>
                        )}
                        <CalendarIcon className="ml-auto h-4 w-4 opacity-50" />
                      </Button>
                    </FormControl>
                  </PopoverTrigger>
                  <PopoverContent className="w-auto p-0" align="start">
                    <Calendar
                      mode="single"
                      selected={field.value}
                      onSelect={field.onChange}
                      disabled={(date) =>
                        date < new Date(new Date().setHours(0, 0, 0, 0))
                      }
                      initialFocus
                    />
                  </PopoverContent>
                </Popover>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="end_date"
            render={({ field }) => (
              <FormItem className="flex flex-col">
                <FormLabel className="text-zinc-700 font-semibold text-base mb-1">End Date</FormLabel>
                <Popover>
                  <PopoverTrigger>
                    <FormControl>
                      <Button
                        variant={"outline"}
                        className={cn(
                          "h-12 rounded-xl pl-3 text-left font-normal",
                          !field.value && "text-zinc-400"
                        )}
                      >
                        {field.value ? (
                          format(field.value, "PPP")
                        ) : (
                          <span>Pick a date</span>
                        )}
                        <CalendarIcon className="ml-auto h-4 w-4 opacity-50" />
                      </Button>
                    </FormControl>
                  </PopoverTrigger>
                  <PopoverContent className="w-auto p-0" align="start">
                    <Calendar
                      mode="single"
                      selected={field.value}
                      onSelect={field.onChange}
                      disabled={(date) =>
                        date < form.getValues('start_date') || date < new Date(new Date().setHours(0, 0, 0, 0))
                      }
                      initialFocus
                    />
                  </PopoverContent>
                </Popover>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>

        <FormField
          control={form.control}
          name="cover_image"
          render={({ field }) => (
            <FormItem>
              <FormLabel className="text-zinc-700 font-semibold text-base">Cover Image URL</FormLabel>
              <FormControl>
                <Input placeholder="https://images.unsplash.com/photo-..." {...field} className="h-12 rounded-xl" />
              </FormControl>
              <FormDescription>
                Add a high-quality photo for the event banner.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="description"
          render={({ field }) => (
            <FormItem>
              <FormLabel className="text-zinc-700 font-semibold text-base">Description (Love Story / Theme)</FormLabel>
              <FormControl>
                <Textarea 
                  placeholder="Share a short welcome message or the story behind the celebration..." 
                  className="min-h-[120px] rounded-xl resize-none"
                  {...field} 
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <Separator className="bg-zinc-100" />

        {/* Co-owner Section */}
        <div className="space-y-6">
          <div className="flex items-center gap-2">
            <div className="h-8 w-8 rounded-lg bg-zinc-100 flex items-center justify-center text-zinc-600">
              <Users className="h-4 w-4" />
            </div>
            <div>
              <h3 className="text-lg font-bold text-zinc-900">Invite Co-owner</h3>
              <p className="text-xs text-zinc-500">Add someone from the family to help manage the event.</p>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
            <FormField
              control={form.control}
              name="co_owner_name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel className="text-zinc-700 font-semibold text-sm">Full Name</FormLabel>
                  <FormControl>
                    <Input placeholder="Enter co-owner name" {...field} className="h-11 rounded-xl" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="co_owner_contact"
              render={({ field }) => (
                <FormItem>
                  <FormLabel className="text-zinc-700 font-semibold text-sm">Phone or Email</FormLabel>
                  <FormControl>
                    <Input placeholder="+91 98765 43210" {...field} className="h-11 rounded-xl" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        </div>

        <div className="flex gap-4 pt-4">
          <Button 
            type="button" 
            variant="outline"
            onClick={() => setCurrentStep(1)}
            className="flex-1 h-12 rounded-xl font-bold"
          >
            Back
          </Button>
          <Button 
            type="submit" 
            className="flex-1 h-12 bg-orange-600 hover:bg-orange-700 text-white font-bold rounded-xl shadow-lg shadow-orange-200"
          >
            Next Step
          </Button>
        </div>
      </form>
    </Form>
  )
}
