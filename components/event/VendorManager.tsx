'use client'

import { useCallback, useEffect, useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { 
  Briefcase, 
  Plus, 
  Trash2, 
  Loader2
} from 'lucide-react'
import { apiFetch } from '@/lib/api'
import { paymentService } from '@/lib/services/PaymentService'

interface Vendor {
  id: string
  name: string
  category: string
  total_paise: number
  advance_paise: number
  status: string
  notes?: string
}

export function VendorManager({ eventId }: { eventId: string }) {
  const [vendors, setVendors] = useState<Vendor[]>([])
  const [isAdding, setIsAdding] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  
  const [newName, setNewName] = useState('')
  const [newCategory, setNewCategory] = useState('')
  const [newBudget, setNewBudget] = useState('')

  const fetchVendors = useCallback(async () => {
    try {
      const data = await apiFetch<{ vendors: Vendor[] }>(`/v1/events/${eventId}/vendors`)
      setVendors(data.vendors || [])
    } catch {
      setError('Failed to load vendors')
    } finally {
      setLoading(false)
    }
  }, [eventId])

  useEffect(() => {
    void fetchVendors()
  }, [fetchVendors])

  const handleAdd = async () => {
    if (!newName) return
    setError(null)
    try {
      await apiFetch(`/v1/events/${eventId}/vendors`, {
        method: 'POST',
        json: {
          name: newName,
          category: newCategory,
          total_paise: Math.round((parseFloat(newBudget) || 0) * 100),
          advance_paise: 0
        }
      })
      setNewName('')
      setNewCategory('')
      setNewBudget('')
      setIsAdding(false)
      fetchVendors()
    } catch {
      setError('Failed to add vendor')
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await apiFetch(`/v1/events/${eventId}/vendors/${id}`, { method: 'DELETE' })
      fetchVendors()
    } catch {
      setError('Failed to delete vendor')
    }
  }

  if (loading) return (
    <div className="p-8 text-center bg-white rounded-[32px] border-2 border-dashed border-zinc-100 animate-pulse">
       <Loader2 className="h-6 w-6 animate-spin mx-auto text-blue-600 mb-2" />
       <p className="text-zinc-400 text-xs font-bold uppercase tracking-widest">Loading Vendors...</p>
    </div>
  )

  const totalBudget = vendors.reduce((acc, v) => acc + (v.total_paise || 0), 0) / 100
  const totalPaid = vendors.reduce((acc, v) => acc + (v.advance_paise || 0), 0) / 100

  return (
    <Card className="p-6 border-zinc-200 shadow-xl rounded-[32px] bg-white space-y-6">
      <div className="flex items-center justify-between gap-3 mb-2">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-2xl bg-blue-100 flex items-center justify-center text-blue-700">
            <Briefcase className="h-5 w-5" />
          </div>
          <div>
            <h3 className="font-bold text-zinc-900 font-heading text-lg tracking-tight">Vendor Control</h3>
            <p className="text-[10px] text-zinc-500 uppercase font-bold tracking-wider flex items-center gap-1">
              <span className="w-1 h-1 bg-green-500 rounded-full" /> Authorized API Access
            </p>
          </div>
        </div>
        {!isAdding && (
          <Button size="sm" onClick={() => setIsAdding(true)} className="rounded-xl h-9 bg-zinc-900 hover:bg-black shadow-lg">
            <Plus className="h-4 w-4 mr-1" />
            Add Vendor
          </Button>
        )}
      </div>

      {error && (
        <div className="p-3 bg-red-50 text-red-600 text-[10px] font-bold uppercase tracking-wider rounded-xl text-center">
          {error}
        </div>
      )}

      <div className="grid grid-cols-2 gap-4 pb-4 border-b border-zinc-100">
         <div className="space-y-1">
            <p className="text-2xl font-bold font-heading tracking-tight">{paymentService.formatINR(totalBudget)}</p>
            <p className="text-[10px] text-zinc-400 uppercase font-bold tracking-widest">Planned Budget</p>
         </div>
         <div className="border-l border-zinc-100 pl-4 space-y-1">
            <p className="text-2xl font-bold font-heading text-blue-600 tracking-tight">{paymentService.formatINR(totalPaid)}</p>
            <p className="text-[10px] text-zinc-400 uppercase font-bold tracking-widest">Total Paid</p>
         </div>
      </div>

      {isAdding && (
        <div className="p-5 bg-zinc-50 border border-zinc-100 rounded-3xl space-y-4 animate-in fade-in slide-in-from-top-2">
          <div className="space-y-1.5">
            <label className="text-[10px] font-bold text-zinc-400 uppercase tracking-wider ml-1">Vendor Name</label>
            <Input placeholder="e.g. Royal Caterers" value={newName} onChange={e => setNewName(e.target.value)} className="rounded-xl border-zinc-200" />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <label className="text-[10px] font-bold text-zinc-400 uppercase tracking-wider ml-1">Category</label>
              <Input placeholder="Catering" value={newCategory} onChange={e => setNewCategory(e.target.value)} className="rounded-xl border-zinc-200" />
            </div>
            <div className="space-y-1.5">
              <label className="text-[10px] font-bold text-zinc-400 uppercase tracking-wider ml-1">Budget (₹)</label>
              <Input type="number" placeholder="50000" value={newBudget} onChange={e => setNewBudget(e.target.value)} className="rounded-xl border-zinc-200" />
            </div>
          </div>
          <div className="flex gap-2 pt-2">
            <Button onClick={handleAdd} className="flex-1 rounded-xl bg-blue-600 hover:bg-blue-700 text-white font-bold">Save Vendor</Button>
            <Button variant="ghost" onClick={() => setIsAdding(false)} className="rounded-xl text-zinc-500 hover:bg-zinc-100 transition-colors">Cancel</Button>
          </div>
        </div>
      )}

      <div className="space-y-3">
        {vendors.length === 0 && !isAdding && (
          <div className="py-12 text-center border-2 border-dashed border-zinc-100 rounded-3xl">
            <p className="text-zinc-400 text-xs font-bold uppercase tracking-widest">Your vendor list is empty.</p>
          </div>
        )}
        {vendors.map((vendor) => (
          <div key={vendor.id} className="p-4 bg-white rounded-2xl border border-zinc-100 flex items-center justify-between group hover:border-blue-200 hover:shadow-xl hover:shadow-blue-50/50 transition-all duration-300">
            <div className="flex items-center gap-4">
              <div className="w-12 h-12 rounded-2xl bg-zinc-50 border border-zinc-100 flex items-center justify-center font-bold text-zinc-300 text-lg group-hover:bg-blue-600 group-hover:text-white group-hover:border-blue-600 transition-all">
                {vendor.name.charAt(0)}
              </div>
              <div className="space-y-1">
                <p className="font-bold text-zinc-900 text-sm">{vendor.name}</p>
                <div className="flex items-center gap-2">
                  <Badge variant="ghost" className="px-0 h-auto text-[10px] text-zinc-400 group-hover:text-zinc-600">{vendor.category}</Badge>
                  <span className="text-[10px] text-zinc-200 group-hover:text-zinc-300">·</span>
                  <Badge className="bg-zinc-100 text-zinc-500 border-none h-4 text-[8px] uppercase font-bold rounded-md">
                    {vendor.status || 'Active'}
                  </Badge>
                </div>
              </div>
            </div>
            <div className="text-right flex items-center gap-6">
              <div className="space-y-0.5">
                 <p className="font-bold text-base text-zinc-900 tracking-tight">{paymentService.formatINR(vendor.total_paise / 100)}</p>
                 <div className="flex items-center justify-end gap-1.5 text-[9px] font-bold text-blue-600">
                   <div className="w-1.5 h-1.5 bg-blue-500 rounded-full animate-pulse" />
                   {paymentService.formatINR(vendor.advance_paise / 100)} PAID
                 </div>
              </div>
              <Button variant="ghost" size="icon" onClick={() => handleDelete(vendor.id)} className="h-9 w-9 text-zinc-300 hover:text-red-500 hover:bg-red-50 rounded-xl opacity-0 group-hover:opacity-100 transition-all">
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          </div>
        ))}
      </div>
    </Card>
  )
}
