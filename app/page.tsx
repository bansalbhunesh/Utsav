import Link from 'next/link'
import Image from 'next/image'
import { Button } from '@/components/ui/button'
import { 
  Sparkles, 
  PartyPopper, 
  ChevronRight, 
  ShieldCheck, 
  Zap, 
  Star 
} from 'lucide-react'

export default function LandingPage() {
  return (
    <div className="min-h-screen bg-white overflow-hidden">
      {/* Navigation */}
      <nav className="flex items-center justify-between px-6 py-6 max-w-7xl mx-auto relative z-20">
        <div className="flex items-center gap-2">
          <div className="w-10 h-10 bg-orange-600 rounded-2xl flex items-center justify-center text-white font-bold text-xl shadow-lg shadow-orange-200">U</div>
          <span className="text-xl font-bold tracking-tight text-zinc-900 font-heading">UTSAV</span>
        </div>
        <div className="hidden md:flex items-center gap-8 text-sm font-bold text-zinc-500 uppercase tracking-widest">
           <Link href="#" className="hover:text-orange-600 transition-colors">Features</Link>
           <Link href="#" className="hover:text-orange-600 transition-colors">Showcase</Link>
           <Link href="#" className="hover:text-orange-600 transition-colors">Pricing</Link>
        </div>
        <div className="flex items-center gap-4">
          <Link href="/login">
            <Button variant="ghost" className="font-bold text-zinc-900 rounded-xl hover:bg-orange-50 hover:text-orange-600 transition-all">Sign In</Button>
          </Link>
          <Link href="/dashboard">
            <Button className="bg-orange-600 hover:bg-orange-700 text-white font-bold rounded-xl px-6 h-11">Dashboard</Button>
          </Link>
        </div>
      </nav>

      {/* Hero Section */}
      <section className="relative px-6 pt-20 pb-32 text-center space-y-10 max-w-5xl mx-auto">
        <div className="inline-flex items-center gap-2 bg-orange-50 border border-orange-100 px-4 py-2 rounded-full text-orange-600 text-xs font-bold uppercase tracking-widest animate-in fade-in slide-in-from-top-4 duration-1000">
          <Sparkles className="h-4 w-4" />
          Operating System for Events
        </div>
        
        <h1 className="text-6xl sm:text-8xl font-bold tracking-tight text-zinc-900 font-heading leading-tight italic decoration-orange-600/20 underline-offset-8">
          Celebrate with <span className="text-orange-600">Authority.</span>
        </h1>
        
        <p className="text-xl text-zinc-500 max-w-2xl mx-auto font-medium">
          The all-in-one platform for modern hosts to manage guests, logistics, and shagun with precision and elegance.
        </p>

        <div className="flex flex-col sm:flex-row items-center justify-center gap-4 relative z-10 pt-4">
          <Link href="/dashboard" className="w-full sm:w-auto">
            <Button className="w-full sm:w-auto h-16 px-10 bg-orange-600 hover:bg-orange-700 text-white font-bold rounded-2xl shadow-2xl shadow-orange-200 text-lg group">
              Start Your Event
              <ChevronRight className="ml-2 h-6 w-6 group-hover:translate-x-1 transition-transform" />
            </Button>
          </Link>
          <Button variant="outline" className="w-full sm:w-auto h-16 px-10 rounded-2xl border-zinc-200 bg-white text-zinc-900 font-bold hover:bg-zinc-50 text-lg">
             View Demo
          </Button>
        </div>

        {/* Decorative Gradients */}
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[120%] h-[120%] bg-[radial-gradient(circle_at_center,_var(--tw-gradient-stops))] from-orange-50 via-transparent to-transparent -z-10 opacity-60" />
      </section>

      {/* Feature Grid */}
      <section className="px-6 py-24 bg-zinc-50/50">
        <div className="max-w-7xl mx-auto grid grid-cols-1 md:grid-cols-3 gap-12">
           <div className="space-y-4">
              <div className="w-14 h-14 bg-white rounded-2xl flex items-center justify-center shadow-md border border-zinc-100 text-orange-600">
                 <ShieldCheck className="w-7 h-7" />
              </div>
              <h3 className="text-xl font-bold text-zinc-900">Relational Ledger</h3>
              <p className="text-zinc-500 text-sm leading-relaxed">Secure digital auditing for shagun and guest interactions with verifiable authenticity.</p>
           </div>
           
           <div className="space-y-4">
              <div className="w-14 h-14 bg-white rounded-2xl flex items-center justify-center shadow-md border border-zinc-100 text-blue-600">
                 <Zap className="w-7 h-7" />
              </div>
              <h3 className="text-xl font-bold text-zinc-900">Instant OTP RSVP</h3>
              <p className="text-zinc-500 text-sm leading-relaxed">No more guesswork. Guests verify their identity instantly via secure OTP for a seamless log.</p>
           </div>

           <div className="space-y-4">
              <div className="w-14 h-14 bg-white rounded-2xl flex items-center justify-center shadow-md border border-zinc-100 text-purple-600">
                 <PartyPopper className="w-7 h-7" />
              </div>
              <h3 className="text-xl font-bold text-zinc-900">Memories, Refined</h3>
              <p className="text-zinc-500 text-sm leading-relaxed">A digital keepsake that captures blessings, photos, and milestones in a stunning souvenir.</p>
           </div>
        </div>
      </section>

      {/* Social Proof Placeholder */}
      <footer className="py-20 text-center border-t border-zinc-100">
         <div className="flex flex-col items-center gap-6">
            <div className="flex -space-x-4">
               {[1,2,3,4,5].map(i => (
                 <div key={i} className="w-12 h-12 rounded-full border-4 border-white bg-zinc-100 overflow-hidden" title={`User ${i}`}>
                    <Image src={`https://i.pravatar.cc/150?u=${i}`} alt={`User ${i}`} width={48} height={48} className="h-full w-full object-cover" unoptimized />
                 </div>
               ))}
            </div>
            <p className="text-zinc-500 font-medium">Joined by <span className="text-zinc-900 font-bold">200+</span> hosts this wedding season.</p>
            <div className="flex items-center gap-1 text-orange-500">
               {[1,2,3,4,5].map(i => <Star key={i} className="w-4 h-4 fill-current" />)}
            </div>
         </div>
      </footer>
    </div>
  )
}
