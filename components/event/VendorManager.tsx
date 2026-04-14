'use client'

import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { 
  Briefcase, 
  Plus, 
  Trash2, 
  Phone, 
  IndianRupee, 
  CheckCircle2, 
  Clock,
  MoreVertical
} from 'lucide-react'
import { supabase } from '@/lib/supabase/client'
import { paymentService } from '@/lib/services/PaymentService'

interface Vendor {
  id: string
  name: string
  category: string
  budget_amount: number
  paid_amount: number
  status: string
  contact_info: string
}

export function VendorManager({ eventId }: { eventId: string }) {
  const [vendors, setVendors] = useState<Vendor[]>([])
  const [isAdding, setIsAdding] = useState(false)
  const [loading, setLoading] = useState(true)
  
  const [newName, setNewName] = useState('')
  const [newCategory, setNewCategory] = useState('')
  const [newBudget, setNewBudget] = useState('')

  const fetchVendors = async () => {
    const { data } = await supabase
      .from('vendors')
      .select('*')
      .eq('event_id', eventId)
      .order('created_at', { ascending: false })
    
    setVendors(data || [])
    setLoading(false)
  }

  useEffect(() => {
    fetchVendors()
  }, [eventId])

  const handleAdd = async () => {
    if (!newName) return
    const { error } = await supabase.from('vendors').insert({
      event_id: eventId,
      name: newName,
      category: newCategory,
      budget_amount: parseFloat(newBudget) || 0,
      status: 'PENDING'
    })

    if (!error) {
       setNewName('')
       setNewCategory('')
       setNewBudget('')
       setIsAdding(false)
       fetchVendors()
    }
  }

  const handleDelete = async (id: string) => {
    await supabase.from('vendors').delete().eq('id', id)
    fetchVendors()
  }

  if (loading) return <div className="p-4 text-center animate-pulse">Loading vendors...</div>

  const totalBudget = vendors.reduce((acc, v) => acc + Number(v.budget_amount), 0)
  const totalPaid = vendors.reduce((acc, v) => acc + Number(v.paid_amount), 0)

  return (
    <Card className="p-6 border-zinc-200 shadow-xl rounded-[32px] bg-white space-y-6">
      <div className="flex items-center justify-between gap-3 mb-2">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-2xl bg-blue-100 flex items-center justify-center text-blue-700">
            <Briefcase className="h-5 w-5" />
          </div>
          <div>
            <h3 className="font-bold text-zinc-900">Vendor Quick-Assign</h3>
            <p className="text-[10px] text-zinc-500 uppercase font-bold tracking-wider">Module 8 · v1.5</p>
          </div>
        </div>
        {!isAdding && (
          <Button size="sm" onClick={() => setIsAdding(true)} className="rounded-xl h-9 bg-zinc-900">
            <Plus className="h-4 w-4 mr-1" />
            Add Vendor
          </Button>
        )}
      </div>

      <div className="grid grid-cols-2 gap-4 pb-4 border-b border-zinc-100">
         <div>
            <p className="text-xl font-bold font-heading">{paymentService.formatINR(totalBudget)}</p>
            <p className="text-[10px] text-zinc-400 uppercase font-bold tracking-widest">Planned Budget</p>
         </div>
         <div className="border-l border-zinc-100 pl-4">
            <p className="text-xl font-bold font-heading text-blue-600">{paymentService.formatINR(totalPaid)}</p>
            <p className="text-[10px] text-zinc-400 uppercase font-bold tracking-widest">Total Paid</p>
         </div>
      </div>

      {isAdding && (
        <div className="p-4 bg-zinc-50 rounded-2xl space-y-4 animate-in fade-in slide-in-from-top-2">
          <Input placeholder="Vendor Name (e.g. Royal Caterers)" value={newName} onChange={e => setNewName(e.target.value)} />
          <div className="grid grid-cols-2 gap-4">
            <Input placeholder="Category" value={newCategory} onChange={e => setNewCategory(e.target.value)} />
            <Input type="number" placeholder="Budget Amount" value={newBudget} onChange={e => setNewBudget(e.target.value)} />
          </div>
          <div className="flex gap-2">
            <Button onClick={handleAdd} className="flex-1 rounded-xl bg-blue-600">Save Vendor</Button>
            <Button variant="ghost" onClick={() => setIsAdding(false)} className="rounded-xl">Cancel</Button>
          </div>
        </div>
      )}

      <div className="space-y-3">
        {vendors.length === 0 && !isAdding && (
          <div className="py-10 text-center border border-dashed border-zinc-200 rounded-2xl">
            <p className="text-zinc-400 text-sm">No vendors assigned yet.</p>
          </div>
        )}
        {vendors.map((vendor) => (
          <div key={vendor.id} className="p-4 bg-zinc-50/50 rounded-2xl border border-zinc-100 flex items-center justify-between group">
            <div className="flex items-center gap-4">
              <div className="w-10 h-10 rounded-xl bg-white border border-zinc-100 flex items-center justify-center font-bold text-zinc-400">
                {vendor.name.charAt(0)}
              </div>
              <div>
                <p className="font-bold text-zinc-900 text-sm">{vendor.name}</p>
                <div className="flex items-center gap-2">
                  <Badge variant="ghost" className="px-0 h-auto text-[10px] text-zinc-500">{vendor.category}</Badge>
                  <span className="text-[10px] text-zinc-300">·</span>
                  <Badge className="bg-white text-zinc-600 border-zinc-100 h-5 text-[9px] uppercase font-bold">
                    {vendor.status}
                  </Badge>
                </div>
              </div>
            </div>
            <div className="text-right flex items-center gap-4">
              <div>
                 <p className="font-bold text-sm text-zinc-900">{paymentService.formatINR(vendor.budget_amount)}</p>
                 <div className="flex items-center justify-end gap-1 text-[9px] font-bold text-green-600">
                   <div className="w-1 h-1 bg-green-500 rounded-full" />
                   PAID {Math.round((vendor.paid_amount/vendor.budget_amount)*100 || 0)}%
                 </div>
              </div>
              <Button variant="ghost" size="icon" onClick={() => handleDelete(vendor.id)} className="h-8 w-8 text-zinc-300 hover:text-red-500 opacity-0 group-hover:opacity-100 transition-opacity">
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          </div>
        ))}
      </div>
    </Card>
  )
}
