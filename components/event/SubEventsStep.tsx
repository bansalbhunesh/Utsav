'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { format } from 'date-fns'
import { CalendarIcon, Plus, Trash2, Clock, MapPin } from 'lucide-react'
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
import { Calendar } from '@/components/ui/calendar'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogFooter,
} from '@/components/ui/dialog'
import { useEventCreationStore } from '@/store/event-creation-store'
import { cn } from '@/lib/utils'

const subEventSchema = z.object({
  name: z.string().min(2, 'Name is required'),
  type: z.string().min(2, 'Type is required (e.g. Sangeet, Haldi)'),
  date_time: z.date(),
  venue_name: z.string().min(2, 'Venue name is required'),
  venue_address: z.string().optional(),
  dress_code: z.string().optional(),
  description: z.string().optional(),
})

export function SubEventsStep() {
  const { subEvents, addSubEvent, removeSubEvent, setCurrentStep } = useEventCreationStore()
  const [isDialogOpen, setIsDialogOpen] = useState(false)

  const form = useForm<z.infer<typeof subEventSchema>>({
    resolver: zodResolver(subEventSchema),
    defaultValues: {
      name: '',
      type: '',
      venue_name: '',
      venue_address: '',
      dress_code: '',
      description: '',
    },
  })

  function onAddSubEvent(values: z.infer<typeof subEventSchema>) {
    addSubEvent({
      ...values,
      date_time: values.date_time.toISOString(),
    })
    form.reset()
    setIsDialogOpen(false)
  }

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <div className="space-y-1">
          <h3 className="text-xl font-bold font-heading text-zinc-900">Event Schedule</h3>
          <p className="text-sm text-zinc-500">Add different functions or sessions for your event.</p>
        </div>
        
        <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
          <DialogTrigger
            render={
              <Button className="bg-orange-600 hover:bg-orange-700 text-white rounded-xl shadow-lg shadow-orange-100">
                <Plus className="mr-2 h-4 w-4" />
                Add Function
              </Button>
            }
          />
          <DialogContent className="sm:max-w-[500px] rounded-3xl">
            <DialogHeader>
              <DialogTitle className="text-2xl font-bold font-heading">Add New Function</DialogTitle>
            </DialogHeader>
            <Form {...form}>
              <form onSubmit={form.handleSubmit(onAddSubEvent)} className="space-y-4 py-4">
                <FormField
                  control={form.control}
                  name="name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Function Name</FormLabel>
                      <FormControl>
                        <Input placeholder="Engagement / Pool Party" {...field} className="rounded-xl" />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="date_time"
                  render={({ field }) => (
                    <FormItem className="flex flex-col">
                      <FormLabel>Date & Time</FormLabel>
                      <Popover>
                        <PopoverTrigger
                          className={cn(
                            "inline-flex h-10 w-full items-center justify-start rounded-xl border border-input bg-transparent px-3 py-2 text-left text-sm font-normal shadow-xs",
                            !field.value && "text-zinc-400"
                          )}
                        >
                          {field.value ? (
                            format(field.value, "PPP p")
                          ) : (
                            <span>Pick a date & time</span>
                          )}
                          <CalendarIcon className="ml-auto h-4 w-4 opacity-50" />
                        </PopoverTrigger>
                        <PopoverContent className="w-auto p-0" align="start">
                          <Calendar
                            mode="single"
                            selected={field.value}
                            onSelect={field.onChange}
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
                  name="venue_name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Venue Name</FormLabel>
                      <FormControl>
                        <Input placeholder="The Grand Ballroom" {...field} className="rounded-xl" />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="venue_address"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Venue Address</FormLabel>
                      <FormControl>
                        <Input placeholder="123 MG Road, Delhi" {...field} className="rounded-xl" />
                      </FormControl>
                      <FormDescription className="text-[10px]">
                        Google Places autocomplete enabled in Phase 2
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <div className="grid grid-cols-2 gap-4">
                  <FormField
                    control={form.control}
                    name="dress_code"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Dress Code</FormLabel>
                        <FormControl>
                          <Input placeholder="Ethnic Formal" {...field} className="rounded-xl" />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name="type"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Type</FormLabel>
                        <FormControl>
                          <Input placeholder="Sangeet / Reception" {...field} className="rounded-xl" />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>

                <FormField
                  control={form.control}
                  name="description"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Description</FormLabel>
                      <FormControl>
                        <Input placeholder="Special notes for guests..." {...field} className="rounded-xl" />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <DialogFooter className="pt-4">
                  <Button type="submit" className="w-full bg-orange-600 hover:bg-orange-700 text-white font-bold rounded-xl h-11">
                    Add to Schedule
                  </Button>
                </DialogFooter>
              </form>
            </Form>
          </DialogContent>
        </Dialog>
      </div>

      <div className="space-y-4">
        {subEvents.length === 0 ? (
          <div className="text-center py-12 bg-zinc-50 rounded-3xl border-2 border-dashed border-zinc-200">
            <p className="text-zinc-400 font-medium">No functions added yet.</p>
          </div>
        ) : (
          subEvents.map((sub, i) => (
            <div key={i} className="flex items-center justify-between p-4 bg-white rounded-2xl border border-zinc-100 shadow-sm hover:shadow-md transition-shadow">
              <div className="flex gap-4 items-center text-zinc-900 font-bold">
                 <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-orange-100 text-orange-600">
                    <Clock className="h-5 w-5" />
                 </div>
                 <div className="space-y-1">
                    <p className="font-bold leading-tight">{sub.name}</p>
                    <div className="flex items-center gap-3 text-xs text-zinc-500 font-medium">
                       <span className="flex items-center gap-1">
                          <CalendarIcon className="h-3 w-3" />
                          {format(new Date(sub.date_time as string), "MMM d, p")}
                       </span>
                       <span className="flex items-center gap-1">
                          <MapPin className="h-3 w-3" />
                          {sub.venue_name}
                       </span>
                    </div>
                 </div>
              </div>
              <Button 
                variant="ghost" 
                size="icon" 
                onClick={() => removeSubEvent(i)}
                className="text-zinc-400 hover:text-red-500 rounded-lg"
              >
                <Trash2 className="h-5 w-5" />
              </Button>
            </div>
          ))
        )}
      </div>

      <div className="flex gap-4 pt-4">
        <Button 
          type="button" 
          variant="outline"
          onClick={() => setCurrentStep(2)}
          className="flex-1 h-12 rounded-xl font-bold"
        >
          Back
        </Button>
        <Button 
          type="button" 
          onClick={() => setCurrentStep(4)}
          disabled={subEvents.length === 0}
          className="flex-1 h-12 bg-orange-600 hover:bg-orange-700 text-white font-bold rounded-xl shadow-lg shadow-orange-200"
        >
          Branding & Finish
        </Button>
      </div>
    </div>
  )
}
