"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { apiFetch, getAccessToken } from "@/lib/api";

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
  const description = "";
  const [clients, setClients] = useState<Client[]>([]);
  const [events, setEvents] = useState<EventRow[]>([]);
  const [clientName, setClientName] = useState("");
  const [clientEmail, setClientEmail] = useState("");
  const [clientPhone, setClientPhone] = useState("");
  const [selectedClientId, setSelectedClientId] = useState("");
  const [selectedEventId, setSelectedEventId] = useState("");

  const load = useCallback(async () => {
    const [c, e] = await Promise.all([
      apiFetch<{ clients: Client[] }>("/v1/organiser/clients"),
      apiFetch<{ events: EventRow[] }>("/v1/events"),
    ]);
    setClients(c.clients || []);
    setEvents(e.events || []);
    if (!selectedClientId && c.clients?.length) setSelectedClientId(c.clients[0].id);
    if (!selectedEventId && e.events?.length) setSelectedEventId(e.events[0].id);
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
        await apiFetch("/v1/organiser/profile", {
          method: "POST",
          json: { company_name: companyName, description },
        });
        await load();
      }
    })().catch((e) => setErr(String(e)));
  }, [companyName, description, load]);

  async function createClient() {
    setErr(null);
    try {
      await apiFetch("/v1/organiser/clients", {
        method: "POST",
        json: { name: clientName, contact_email: clientEmail, contact_phone: clientPhone, notes: "" },
      });
      setClientName("");
      setClientEmail("");
      setClientPhone("");
      await load();
    } catch (e) {
      setErr(String(e));
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
    } catch (e) {
      setErr(String(e));
    }
  }

  return (
    <main className="mx-auto max-w-5xl space-y-6 px-6 py-10 text-zinc-100">
      <Link href="/dashboard" className="text-sm text-zinc-400 hover:text-white">
        ← Dashboard
      </Link>
      <h1 className="text-2xl font-semibold text-white">Organiser console</h1>
      {err ? <p className="text-sm text-red-400">{err}</p> : null}

      <section className="rounded-xl border border-zinc-800 bg-zinc-900/40 p-4">
        <h2 className="text-sm font-medium text-zinc-300">Create client</h2>
        <div className="mt-3 grid gap-2 sm:grid-cols-3">
          <input className="rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm" placeholder="Name" value={clientName} onChange={(e) => setClientName(e.target.value)} />
          <input className="rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm" placeholder="Email" value={clientEmail} onChange={(e) => setClientEmail(e.target.value)} />
          <input className="rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm" placeholder="Phone" value={clientPhone} onChange={(e) => setClientPhone(e.target.value)} />
        </div>
        <button type="button" onClick={() => void createClient()} className="mt-3 rounded bg-amber-500 px-4 py-2 text-sm font-semibold text-black">
          Add client
        </button>
      </section>

      <section className="rounded-xl border border-zinc-800 bg-zinc-900/40 p-4">
        <h2 className="text-sm font-medium text-zinc-300">Link client to event</h2>
        <div className="mt-3 grid gap-2 sm:grid-cols-2">
          <select className="rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm" value={selectedClientId} onChange={(e) => setSelectedClientId(e.target.value)}>
            <option value="">Select client</option>
            {clients.map((c) => (
              <option key={c.id} value={c.id}>{c.name}</option>
            ))}
          </select>
          <select className="rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm" value={selectedEventId} onChange={(e) => setSelectedEventId(e.target.value)}>
            <option value="">Select event</option>
            {events.map((e) => (
              <option key={e.id} value={e.id}>{e.title}</option>
            ))}
          </select>
        </div>
        <button type="button" onClick={() => void linkEvent()} className="mt-3 rounded border border-zinc-600 px-4 py-2 text-sm">
          Link
        </button>
      </section>

      <section className="rounded-xl border border-zinc-800 bg-zinc-900/40 p-4">
        <h2 className="text-sm font-medium text-zinc-300">Clients</h2>
        <ul className="mt-3 space-y-2 text-sm">
          {clients.map((c) => (
            <li key={c.id} className="rounded border border-zinc-700 bg-zinc-950/60 p-3">
              <p className="font-medium text-white">{c.name}</p>
              <p className="text-zinc-400">{c.contact_email || c.contact_phone || "No contact details"}</p>
            </li>
          ))}
          {clients.length === 0 ? <li className="text-zinc-500">No clients yet.</li> : null}
        </ul>
      </section>
    </main>
  );
}
