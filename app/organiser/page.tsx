"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { apiFetch, getAccessToken } from "@/lib/api";
import { Button } from "@/components/ui/card"; // Wait, card? No.
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { 
  Users, 
  Plus, 
  Link as LinkIcon, 
  CheckCircle2, 
  ArrowLeft,
  Briefcase,
  Search,
  Mail,
  Phone
} from "lucide-react";
import { Badge } from "@/components/ui/badge";

type Client = {
  id: string;
  name: string;
  contact_email?: string;
  contact_phone?: string;
  notes?: string;
};

type EventRow = { id: string; title: string; slug: string };

export default function OrganiserPage() {
  const [err, setErr] = useState<string | null>(null);
  const companyName = "Utsav Planner";
  const description = "Premium Event Management";
  const [clients, setClients] = useState<Client[]>([]);
  const [events, setEvents] = useState<EventRow[]>([]);
  const [clientName, setClientName] = useState("");
  const [clientEmail, setClientEmail] = useState("");
  const [clientPhone, setClientPhone] = useState("");
  const [selectedClientId, setSelectedClientId] = useState("");
  const [selectedEventId, setSelectedEventId] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const load = useCallback(async () => {
    try {
      const [c, e] = await Promise.all([
        apiFetch<{ clients: Client[] }>("/v1/organiser/clients"),
        apiFetch<{ events: EventRow[] }>("/v1/events"),
      ]);
      setClients(c.clients || []);
      setEvents(e.events || []);
      if (!selectedClientId && c.clients?.length) setSelectedClientId(c.clients[0].id);
      if (!selectedEventId && e.events?.length) setSelectedEventId(e.events[0].id);
    } catch (err: any) {
      setErr(err.message || "Failed to load organiser data");
    }
  }, [selectedClientId, selectedEventId]);

  useEffect(() => {
    if (!getAccessToken()) {
      window.location.href = "/login";
      return;
    }
    void (async () => {
      try {
        await load();
      } catch {
        // Auto-create profile if missing
        await apiFetch("/v1/organiser/profile", {
          method: "POST",
          json: { company_name: companyName, description },
        });
        await load();
      }
    })();
  }, [companyName, description, load]);

  async function createClient() {
    setErr(null);
    setIsSubmitting(true);
    try {
      await apiFetch("/v1/organiser/clients", {
        method: "POST",
        json: { name: clientName, contact_email: clientEmail, contact_phone: clientPhone, notes: "" },
      });
      setClientName("");
      setClientEmail("");
      setClientPhone("");
      await load();
    } catch (e: any) {
      setErr(e.message);
    } finally {
      setIsSubmitting(false);
    }
  }

  async function linkEvent() {
    setErr(null);
    if (!selectedClientId || !selectedEventId) return;
    try {
      await apiFetch(`/v1/organiser/clients/${selectedClientId}/events`, {
        method: "POST",
        json: { event_id: selectedEventId },
      });
      await load();
    } catch (e: any) {
      setErr(e.message);
    }
  }

  return (
    <div className="min-h-screen bg-zinc-50 p-6 lg:p-12">
      <div className="max-w-6xl mx-auto space-y-10">
        
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-6">
           <div className="space-y-2">
             <Link href="/dashboard" className="inline-flex items-center text-xs font-bold text-zinc-400 uppercase tracking-widest hover:text-orange-600 transition-colors">
               <ArrowLeft className="w-3 h-3 mr-1" />
               Back to Dashboard
             </Link>
             <h1 className="text-4xl font-bold font-heading tracking-tight text-zinc-900">Organiser Console</h1>
             <p className="text-zinc-500 font-medium">Manage your professional service clients and event associations.</p>
           </div>
           <div className="flex items-center gap-3">
              <Badge className="bg-orange-100 text-orange-700 border-none font-bold py-1 px-3">
                <Briefcase className="w-3 h-3 mr-1" /> Professional v1.5
              </Badge>
           </div>
        </div>

        {err && (
          <div className="p-4 bg-red-50 border border-red-100 rounded-2xl text-red-600 text-sm font-medium">
            {err}
          </div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          
          {/* Left: Create Client */}
          <div className="lg:col-span-1 space-y-6">
            <Card className="p-6 rounded-[32px] border-zinc-200 shadow-xl shadow-zinc-200/50 space-y-6">
               <div className="flex items-center gap-3">
                 <div className="w-10 h-10 rounded-2xl bg-orange-100 flex items-center justify-center text-orange-600">
                    <Plus className="w-5 h-5" />
                 </div>
                 <h2 className="font-bold text-lg text-zinc-900">New Client</h2>
               </div>

               <div className="space-y-4">
                  <div className="space-y-1.5">
                    <label className="text-[10px] font-bold text-zinc-400 uppercase tracking-wider ml-1">Client Name</label>
                    <Input placeholder="e.g. Aditi Rao" value={clientName} onChange={e => setClientName(e.target.value)} className="rounded-xl border-zinc-100" />
                  </div>
                  <div className="space-y-1.5">
                    <label className="text-[10px] font-bold text-zinc-400 uppercase tracking-wider ml-1">Email (Optional)</label>
                    <Input placeholder="aditi@example.com" value={clientEmail} onChange={e => setClientEmail(e.target.value)} className="rounded-xl border-zinc-100" />
                  </div>
                  <div className="space-y-1.5">
                    <label className="text-[10px] font-bold text-zinc-400 uppercase tracking-wider ml-1">Phone</label>
                    <Input placeholder="+91 98765..." value={clientPhone} onChange={e => setClientPhone(e.target.value)} className="rounded-xl border-zinc-100" />
                  </div>
                  <button 
                    onClick={createClient} 
                    disabled={isSubmitting || !clientName}
                    className="w-full h-12 bg-zinc-900 hover:bg-black text-white font-bold rounded-xl shadow-lg transition-all disabled:opacity-50"
                  >
                    Add to Portfolio
                  </button>
               </div>
            </Card>

            <Card className="p-6 rounded-[32px] border-zinc-200 shadow-xl shadow-zinc-200/50 space-y-6">
               <div className="flex items-center gap-3">
                 <div className="w-10 h-10 rounded-2xl bg-blue-100 flex items-center justify-center text-blue-600">
                    <LinkIcon className="w-5 h-5" />
                 </div>
                 <h2 className="font-bold text-lg text-zinc-900">Link to Event</h2>
               </div>
               
               <p className="text-sm text-zinc-500 pr-4">Associate a managed client with one of your active events.</p>

               <div className="space-y-4">
                  <select 
                    className="w-full h-12 rounded-xl border border-zinc-100 bg-zinc-50 px-4 text-sm font-medium focus:ring-orange-500"
                    value={selectedClientId}
                    onChange={e => setSelectedClientId(e.target.value)}
                  >
                    <option value="">Select client...</option>
                    {clients.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
                  </select>

                  <select 
                    className="w-full h-12 rounded-xl border border-zinc-100 bg-zinc-50 px-4 text-sm font-medium focus:ring-orange-500"
                    value={selectedEventId}
                    onChange={e => setSelectedEventId(e.target.value)}
                  >
                    <option value="">Select event...</option>
                    {events.map(e => <option key={e.id} value={e.id}>{e.title}</option>)}
                  </select>

                  <button 
                    onClick={linkEvent}
                    className="w-full h-12 bg-blue-600 hover:bg-blue-700 text-white font-bold rounded-xl shadow-lg transition-all"
                  >
                    Confirm Association
                  </button>
               </div>
            </Card>
          </div>

          {/* Right: Clients List */}
          <div className="lg:col-span-2 space-y-6">
             <div className="flex items-center justify-between px-2">
                <h3 className="text-xl font-bold text-zinc-900 flex items-center gap-2">
                   <Users className="w-6 h-6 text-zinc-400" />
                   Managed Clients ({clients.length})
                </h3>
             </div>

             <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                {clients.length === 0 ? (
                  <div className="col-span-full py-20 text-center bg-white rounded-[40px] border-2 border-dashed border-zinc-100">
                     <p className="text-zinc-400 font-medium">Your portfolio is empty.</p>
                  </div>
                ) : (
                  clients.map(c => (
                    <Card key={c.id} className="p-6 rounded-[32px] border-zinc-100 hover:border-orange-200 hover:shadow-2xl hover:shadow-orange-100/30 transition-all duration-500 group">
                       <div className="flex items-center justify-between mb-4">
                          <div className="w-12 h-12 rounded-2xl bg-zinc-100 border border-zinc-200 flex items-center justify-center font-bold text-zinc-400 text-xl group-hover:bg-orange-600 group-hover:text-white group-hover:border-orange-600 transition-all">
                             {c.name.charAt(0)}
                          </div>
                          <Badge className="bg-green-50 text-green-700 border-none font-bold text-[10px] uppercase">Active</Badge>
                       </div>
                       <h4 className="text-lg font-bold text-zinc-900 mb-1">{c.name}</h4>
                       <div className="space-y-2">
                          <div className="flex items-center gap-2 text-xs font-medium text-zinc-500">
                             <Mail className="w-3.5 h-3.5" />
                             {c.contact_email || "No email"}
                          </div>
                          <div className="flex items-center gap-2 text-xs font-medium text-zinc-500">
                             <Phone className="w-3.5 h-3.5" />
                             {c.contact_phone || "No phone"}
                          </div>
                       </div>
                    </Card>
                  ))
                )}
             </div>
          </div>

        </div>
      </div>
    </div>
  );
}
