'use client'

import { useEventCreationStore } from '@/store/event-creation-store'
import { BasicsStep } from './BasicsStep'
import { DetailsStep } from './DetailsStep'
import { SubEventsStep } from './SubEventsStep'
import { BrandingStep } from './BrandingStep'
import { Card } from '@/components/ui/card'
import { cn } from '@/lib/utils'

const steps = [
  { id: 1, name: 'Basics' },
  { id: 2, name: 'Details' },
  { id: 3, name: 'Schedule' },
  { id: 4, name: 'Branding' },
]

export function EventWizard() {
  const { currentStep } = useEventCreationStore()

  return (
    <div className="w-full max-w-2xl mx-auto space-y-8">
      {/* Step Indicator */}
      <div className="flex items-center justify-between px-2">
        {steps.map((step, i) => (
          <div key={step.id} className="flex items-center gap-3">
            <div 
              className={cn(
                "w-10 h-10 rounded-full flex items-center justify-center font-bold transition-all",
                currentStep === step.id 
                  ? "bg-orange-600 text-white shadow-lg shadow-orange-200 scale-110" 
                  : currentStep > step.id 
                    ? "bg-green-100 text-green-600" 
                    : "bg-zinc-100 text-zinc-400"
              )}
            >
              {currentStep > step.id ? (
                <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
                </svg>
              ) : (
                step.id
              )}
            </div>
            <span className={cn(
              "hidden sm:block font-semibold text-sm",
              currentStep === step.id ? "text-orange-600" : "text-zinc-400"
            )}>
              {step.name}
            </span>
            {i < steps.length - 1 && (
              <div className="hidden sm:block w-8 h-[2px] bg-zinc-100 mx-1" />
            )}
          </div>
        ))}
      </div>

      <Card className="p-8 rounded-3xl border-zinc-100 shadow-2xl shadow-zinc-200/50">
        {currentStep === 1 && <BasicsStep />}
        {currentStep === 2 && <DetailsStep />}
        {currentStep === 3 && <SubEventsStep />}
        {currentStep === 4 && <BrandingStep />}
      </Card>
    </div>
  )
}
