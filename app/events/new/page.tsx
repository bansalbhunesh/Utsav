import { EventWizard } from "@/components/event/EventWizard";
import Link from "next/link";
import { ChevronLeft } from "lucide-react";

export default function NewEventPage() {
  return (
    <div className="min-h-screen flex flex-col bg-zinc-50 lg:bg-white pb-20">
      {/* Header */}
      <header className="p-6 flex items-center justify-between max-w-7xl mx-auto w-full">
        <Link 
          href="/dashboard" 
          className="inline-flex items-center text-sm font-medium text-zinc-500 hover:text-orange-600 transition-colors"
        >
          <ChevronLeft className="mr-1 h-4 w-4" />
          Back to Dashboard
        </Link>
        <div className="flex items-center gap-2">
           <div className="w-8 h-8 bg-orange-600 rounded-lg flex items-center justify-center text-white font-bold">
              U
           </div>
           <span className="font-bold text-zinc-900 tracking-tight">UTSAV</span>
        </div>
      </header>

      {/* Main Content */}
      <main className="flex-1 px-6 space-y-10">
        <div className="max-w-2xl mx-auto text-center space-y-3">
          <h1 className="text-4xl font-bold font-heading tracking-tight text-zinc-900 leading-tight">
            Create New Event
          </h1>
          <p className="text-zinc-500 max-w-md mx-auto">
            Set up your event details in just a few minutes. You can always edit 
            these later from your dashboard.
          </p>
        </div>

        <EventWizard />
      </main>
    </div>
  );
}
