'use client'

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { apiFetch, getAccessToken } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { 
  Users, 
  Plus, 
  Link as LinkIcon, 
  ArrowLeft,
  Briefcase,
  Mail,
  Phone,
  Loader2,
  CheckCircle2
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
  const [description] = useState("Premium Event Management");
  const [clients, setClients] = useState<Client[]>([]);
  const [events, setEvents] = useState<EventRow[]>([]);
  const [clientName, setClientName] = useState("");
  const [clientEmail, setClientEmail] = useState("");
  const [clientPhone, setClientPhone] = useState("");
  const [selectedClientId, setSelectedClientId] = useState("");
  const [selectedEventId, setSelectedEventId] = useState("");
  const [isLoading, setIsLoading] = useState(true);
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
    } catch (err: unknown) {
      setErr(err instanceof Error ? err.message : "Failed to load organiser data");
    } finally {
      setIsLoading(false);
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
        try {
          await apiFetch("/v1/organiser/profile", {
            method: "POST",
            json: { company_name: companyName, description },
          });
          await load();
        } catch {
          setErr("Failed to initialize organiser profile.");
        }
      }
    })();
  }, [load, companyName, description]);

  async function createClient() {
    if (!clientName) return;
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
    } catch (e: unknown) {
      setErr(e instanceof Error ? e.message : "Failed to create client");
    } finally {
      setIsSubmitting(false);
    }
  }

  async function linkEvent() {
    setErr(null);
    if (!selectedClientId || !selectedEventId) return;
    setIsSubmitting(true);
    try {
      await apiFetch(`/v1/organiser/clients/${selectedClientId}/events`, {
        method: "POST",
        json: { event_id: selectedEventId },
      });
      await load();
    } catch (e: unknown) {
      setErr(e instanceof Error ? e.message : "Failed to link event");
    } finally {
      setIsSubmitting(false);
    }
  }

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-zinc-50">
        <Loader2 className="w-10 h-10 animate-spin text-orange-600" />
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-zinc-50 p-6 lg:p-12 animate-in fade-in duration-700">
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
              <Badge className="bg-orange-100 text-orange-700 border-none font-bold py-1 px-3 rounded-full">
                <Briefcase className="w-3 h-3 mr-1" /> Professional v1.5
              </Badge>
           </div>
        </div>

        {err && (
          <div className="p-4 bg-red-50 border border-red-100 rounded-2xl text-red-600 text-sm font-bold animate-in zoom-in-95">
            {err}
          </div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          
          {/* Left: Management Tools */}
          <div className="lg:col-span-1 space-y-6">
            <Card className="p-8 rounded-[40px] border-zinc-200 shadow-xl shadow-zinc-200/50 space-y-6 bg-white overflow-hidden relative">
               <div className="relative z-10 flex items-center gap-3">
                 <div className="w-10 h-10 rounded-2xl bg-orange-100 flex items-center justify-center text-orange-600">
                    <Plus className="w-5 h-5" />
                 </div>
                 <h2 className="font-bold text-xl text-zinc-900 tracking-tight">New Client</h2>
               </div>

               <div className="relative z-10 space-y-4">
                  <div className="space-y-1.5">
                    <label className="text-[10px] font-bold text-zinc-400 uppercase tracking-widest ml-1">Client Name</label>
                    <Input placeholder="e.g. Aditi Rao" value={clientName} onChange={e => setClientName(e.target.value)} className="rounded-2xl border-zinc-100 h-12" />
                  </div>
                  <div className="space-y-1.5">
                    <label className="text-[10px] font-bold text-zinc-400 uppercase tracking-widest ml-1">Email</label>
                    <Input placeholder="aditi@example.com" value={clientEmail} onChange={e => setClientEmail(e.target.value)} className="rounded-2xl border-zinc-100 h-12" />
                  </div>
                  <div className="space-y-1.5">
                    <label className="text-[10px] font-bold text-zinc-400 uppercase tracking-widest ml-1">Phone</label>
                    <Input placeholder="+91 98765..." value={clientPhone} onChange={e => setClientPhone(e.target.value)} className="rounded-2xl border-zinc-100 h-12" />
                  </div>
                  <Button 
                    onClick={createClient} 
                    disabled={isSubmitting || !clientName}
                    className="w-full h-14 bg-zinc-900 hover:bg-black text-white font-bold rounded-2xl shadow-xl transition-all"
                  >
                    {isSubmitting ? <Loader2 className="w-5 h-5 animate-spin" /> : 'Add to Portfolio'}
                  </Button>
               </div>
            </Card>

            <Card className="p-8 rounded-[40px] border-zinc-200 shadow-xl shadow-zinc-200/50 space-y-6 bg-white">
               <div className="flex items-center gap-3">
                 <div className="w-10 h-10 rounded-2xl bg-blue-100 flex items-center justify-center text-blue-600">
                    <LinkIcon className="w-5 h-5" />
                 </div>
                 <h2 className="font-bold text-xl text-zinc-900 tracking-tight">Link Event</h2>
               </div>
               
               <p className="text-xs text-zinc-500 font-medium">Connect a client to one of your active authoritative events.</p>

               <div className="space-y-4">
                  <select 
                    className="w-full h-14 rounded-2xl border border-zinc-100 bg-zinc-50 px-4 text-sm font-bold text-zinc-700 outline-none focus:ring-2 focus:ring-orange-500 transition-all"
                    value={selectedClientId}
                    onChange={e => setSelectedClientId(e.target.value)}
                  >
                    <option value="">Select client...</option>
                    {clients.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
                  </select>

                  <select 
                    className="w-full h-14 rounded-2xl border border-zinc-100 bg-zinc-50 px-4 text-sm font-bold text-zinc-700 outline-none focus:ring-2 focus:ring-orange-500 transition-all"
                    value={selectedEventId}
                    onChange={e => setSelectedEventId(e.target.value)}
                  >
                    <option value="">Select event...</option>
                    {events.map(e => <option key={e.id} value={e.id}>{e.title}</option>)}
                  </select>

                  <Button 
                    onClick={linkEvent}
                    disabled={isSubmitting || !selectedClientId || !selectedEventId}
                    className="w-full h-14 bg-blue-600 hover:bg-blue-700 text-white font-bold rounded-2xl shadow-xl transition-all"
                  >
                     {isSubmitting ? <Loader2 className="w-5 h-5 animate-spin" /> : 'Confirm Association'}
                  </Button>
               </div>
            </Card>
          </div>

          {/* Right: Clients List */}
          <div className="lg:col-span-2 space-y-6">
             <div className="flex items-center justify-between px-2">
                <h3 className="text-2xl font-bold text-zinc-900 flex items-center gap-3">
                   <Users className="w-6 h-6 text-zinc-300" />
                   Managed Clients ({clients.length})
                </h3>
             </div>

             <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                {clients.length === 0 ? (
                  <div className="col-span-full py-32 text-center bg-white rounded-[40px] border-2 border-dashed border-zinc-100">
                     <p className="text-zinc-400 font-bold uppercase text-xs tracking-widest">Your portfolio is empty.</p>
                  </div>
                ) : (
                  clients.map(c => (
                    <Card key={c.id} className="p-8 rounded-[40px] border-zinc-100 hover:border-orange-200 hover:shadow-2xl hover:shadow-orange-100/30 transition-all duration-500 group bg-white">
                       <div className="flex items-center justify-between mb-6">
                          <div className="w-14 h-14 rounded-2xl bg-zinc-50 border border-zinc-100 flex items-center justify-center font-bold text-zinc-300 text-2xl group-hover:bg-orange-600 group-hover:text-white group-hover:border-orange-600 transition-all">
                             {c.name.charAt(0)}
                          </div>
                          <Badge className="bg-green-50 text-green-600 border-none font-bold text-[10px] uppercase px-3 py-1 rounded-full">
                            <CheckCircle2 className="w-3 h-3 mr-1" /> Profile Active
                          </Badge>
                       </div>
                       <h4 className="text-xl font-bold text-zinc-900 mb-4 tracking-tight">{c.name}</h4>
                       <div className="space-y-3">
                          <div className="flex items-center gap-3 text-xs font-bold text-zinc-400 uppercase tracking-widest group-hover:text-zinc-600 transition-colors">
                             <Mail className="w-4 h-4" />
                             {c.contact_email || "No email"}
                          </div>
                          <div className="flex items-center gap-3 text-xs font-bold text-zinc-400 uppercase tracking-widest group-hover:text-zinc-600 transition-colors">
                             <Phone className="w-4 h-4" />
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
