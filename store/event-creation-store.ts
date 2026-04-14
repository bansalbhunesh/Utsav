import { create } from 'zustand'
import { Event, SubEvent } from '@/types'

interface EventCreationState {
  currentStep: number
  eventData: Partial<Event> & {
    co_owner_name?: string;
    co_owner_contact?: string;
    upi_id?: string;
  }
  subEvents: Partial<SubEvent>[]
  setCurrentStep: (step: number) => void
  setEventData: (data: Partial<Event> & { co_owner_name?: string; co_owner_contact?: string; upi_id?: string }) => void
  setSubEvents: (subEvents: Partial<SubEvent>[]) => void
  addSubEvent: (subEvent: Partial<SubEvent>) => void
  removeSubEvent: (index: number) => void
  reset: () => void
}

export const useEventCreationStore = create<EventCreationState>((set) => ({
  currentStep: 1,
  eventData: {
    event_type: 'WEDDING',
    settings: {
      shagun_enabled: true,
      gallery_enabled: true,
      rsvp_enabled: true,
    },
    is_public: false,
  },
  subEvents: [],
  setCurrentStep: (step) => set({ currentStep: step }),
  setEventData: (data) => set((state) => ({ 
    eventData: { ...state.eventData, ...data } 
  })),
  setSubEvents: (subEvents) => set({ subEvents }),
  addSubEvent: (subEvent) => set((state) => ({ 
    subEvents: [...state.subEvents, subEvent] 
  })),
  removeSubEvent: (index) => set((state) => ({ 
    subEvents: state.subEvents.filter((_, i) => i !== index) 
  })),
  reset: () => set({ 
    currentStep: 1, 
    eventData: { 
      event_type: 'WEDDING',
      co_owner_name: '',
      co_owner_contact: '',
      upi_id: ''
    }, 
    subEvents: [] 
  }),
}))
