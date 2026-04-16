'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { 
  LayoutDashboard, 
  Calendar, 
  Users, 
  Briefcase, 
  MessageSquare, 
  Settings, 
  Sparkles
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'


export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const pathname = usePathname()

  const navItems = [
    { name: 'Overview', href: '/dashboard', icon: LayoutDashboard },
    { name: 'My Events', href: '/dashboard/events', icon: Calendar },
    { name: 'Guest Lists', href: '/dashboard/guests', icon: Users },
    { name: 'Vendors', href: '/dashboard/vendors', icon: Briefcase },
    { name: 'Broadcasts', href: '/dashboard/broadcasts', icon: MessageSquare },
  ]

  return (
    <div className="flex min-h-screen bg-zinc-50">
      {/* Sidebar */}
      <aside className="w-72 bg-white border-r border-zinc-200 hidden lg:flex flex-col sticky top-0 h-screen">
        <div className="p-8">
          <Link href="/dashboard" className="flex items-center gap-2 mb-10">
            <div className="w-8 h-8 bg-orange-600 rounded-lg flex items-center justify-center text-white font-bold text-xl shadow-lg shadow-orange-200">
              U
            </div>
            <span className="text-xl font-bold font-heading tracking-tight text-zinc-900 border-b-2 border-orange-600">
              UTSAV
            </span>
          </Link>

          <nav className="space-y-1">
            {navItems.map((item) => {
              const isActive = pathname === item.href
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className={cn(
                    "flex items-center justify-between px-4 py-3 rounded-xl transition-all duration-200 group",
                    isActive 
                      ? "bg-orange-50 text-orange-600" 
                      : "text-zinc-500 hover:bg-zinc-50 hover:text-zinc-900"
                  )}
                >
                  <div className="flex items-center gap-3">
                    <item.icon className={cn("h-5 w-5", isActive ? "text-orange-600" : "text-zinc-400 group-hover:text-zinc-600")} />
                    <span className="font-bold text-sm tracking-tight">{item.name}</span>
                  </div>
                  {isActive && <div className="w-1.5 h-1.5 rounded-full bg-orange-600" />}
                </Link>
              )
            })}
          </nav>
        </div>

        <div className="mt-auto p-6 space-y-4">
           <div className="bg-zinc-900 rounded-[24px] p-6 text-white space-y-4 relative overflow-hidden group">
              <div className="absolute -top-4 -right-4 opacity-20 blur-xl group-hover:scale-110 transition-transform">
                <Sparkles className="w-24 h-24 text-orange-500" />
              </div>
              <p className="text-xs font-bold text-orange-500 uppercase tracking-widest">v1.5 Premium</p>
              <h4 className="text-sm font-bold">Organiser Mode Active</h4>
              <p className="text-[10px] text-zinc-500 leading-relaxed font-bold">You are currently in the Investor & Hackathon Edition.</p>
              <Button size="sm" className="w-full h-8 bg-white text-zinc-900 rounded-lg font-bold text-[10px]">Upgrade to v2.0</Button>
           </div>
           
           <div className="flex items-center justify-between px-2 text-zinc-400">
              <span className="text-[10px] font-bold tracking-widest uppercase">Support</span>
              <Settings className="h-4 w-4 hover:text-zinc-900 cursor-pointer" />
           </div>
        </div>
      </aside>

      {/* Main Content */}
      <div className="flex-1 flex flex-col">
        {children}
      </div>
    </div>
  )
}
