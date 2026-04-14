<<<<<<< HEAD
import Image from "next/image";
import { Button } from "@/components/ui/button";
import { MoveRight, Sparkles, PartyPopper } from "lucide-react";
import Link from "next/link";

export default function Home() {
  return (
    <div className="flex flex-col min-h-screen bg-white">
      {/* Navigation */}
      <nav className="flex items-center justify-between px-6 py-4 border-b border-zinc-100">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 bg-orange-600 rounded-lg flex items-center justify-center text-white font-bold text-xl">
            U
          </div>
          <span className="text-xl font-bold font-heading tracking-tight text-orange-600">
            UTSAV
          </span>
        </div>
        <div className="flex items-center gap-4">
          <Button variant="ghost" className="font-medium">Sign in</Button>
          <Button className="bg-orange-600 hover:bg-orange-700 text-white font-semibold">
            Create Event
          </Button>
        </div>
      </nav>

      {/* Hero Section */}
      <main className="flex-1 flex flex-col items-center justify-center px-6 text-center py-20 lg:py-32 bg-[radial-gradient(circle_at_top,_var(--tw-gradient-stops))] from-orange-100/40 via-white to-white">
        <div className="space-y-8 max-w-4xl mx-auto">
          <div className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full bg-linear-to-r from-orange-600 to-rose-600 text-white text-xs font-bold tracking-widest uppercase shadow-lg shadow-orange-200 animate-in fade-in slide-in-from-top-4 duration-1000">
            <Sparkles className="h-3 w-3" />
            v1.5 Investor Edition
          </div>
          
          <h1 className="text-5xl lg:text-8xl font-bold font-heading tracking-tighter text-zinc-900 leading-[0.9] animate-in fade-in slide-in-from-bottom-8 duration-1000 delay-200">
            Events, <br />
            <span className="text-transparent bg-clip-text bg-linear-to-r from-orange-600 via-rose-600 to-orange-600">Reimagined.</span>
          </h1>
          
          <p className="text-xl lg:text-2xl text-zinc-500 max-w-2xl mx-auto leading-relaxed font-medium animate-in fade-in duration-1000 delay-500">
            The all-in-one <span className="text-zinc-900 font-bold underline decoration-orange-500 decoration-2 underline-offset-4">Operating System</span> for India's biggest celebrations. 
            Replacing WhatsApp chaos with digital elegance.
          </p>

          <div className="flex flex-col sm:flex-row items-center justify-center gap-6 pt-6 animate-in fade-in duration-1000 delay-700">
            <Link href="/events/new">
              <Button size="lg" className="h-16 px-10 bg-zinc-900 hover:bg-black text-white font-bold text-xl rounded-2xl shadow-2xl shadow-zinc-200 group transition-all hover:-translate-y-1">
                Launch Your Event
                <MoveRight className="ml-3 h-6 w-6 group-hover:translate-x-1 transition-transform" />
              </Button>
            </Link>
            <Button size="lg" variant="outline" className="h-16 px-10 border-2 border-zinc-200 font-bold text-xl rounded-2xl hover:bg-zinc-50 transition-all">
              Watch Demo
            </Button>
          </div>
        </div>

        {/* Feature Preview Pill */}
        <div className="mt-24 lg:mt-32 w-full max-w-6xl rounded-[48px] border-8 border-white bg-zinc-100 p-2 shadow-2xl shadow-zinc-200/50 animate-in fade-in zoom-in-95 duration-1000 delay-1000">
          <div className="h-[400px] sm:h-[600px] w-full rounded-[40px] bg-white border border-zinc-100 flex items-center justify-center relative overflow-hidden">
             <div className="absolute inset-0 bg-[radial-gradient(circle_at_center,_var(--tw-gradient-stops))] from-orange-50 via-transparent to-transparent opacity-50" />
             <div className="relative text-center space-y-4">
                <div className="w-20 h-20 bg-orange-600 rounded-3xl flex items-center justify-center text-white mx-auto shadow-xl shadow-orange-200 animate-bounce">
                   <PartyPopper className="h-10 w-10" />
                </div>
                <h3 className="text-2xl font-bold text-zinc-900">Dashboard Preview</h3>
                <p className="text-zinc-500 text-sm font-medium">Coming to life in v1.5...</p>
             </div>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="py-10 text-center border-t border-zinc-100">
        <p className="text-zinc-500 text-sm">
          &copy; 2026 UTSAV Technologies. Built for India's biggest celebrations.
        </p>
      </footer>
    </div>
  );
=======
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { 
  Sparkles, 
  PartyPopper, 
  ChevronRight, 
  Heart, 
  Calendar, 
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
        <Link href="/login">
          <Button variant="ghost" className="font-bold text-zinc-900 rounded-xl hover:bg-orange-50 hover:text-orange-600 transition-all">Sign In</Button>
        </Link>
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
                 <div key={i} className="w-12 h-12 rounded-full border-4 border-white bg-zinc-100 overflow-hidden">
                    <img src={`https://i.pravatar.cc/150?u=${i}`} alt="user" />
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
>>>>>>> f7494df (feat: Architectural Level Up - Go-Authoritative Backend, RSVP OTP Flow, and Frontend Consolidation (v1.5 Final))
}
