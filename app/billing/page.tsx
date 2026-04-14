"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { apiFetch, getAccessToken } from "@/lib/api";

type EventRow = { id: string; title: string };
type Checkout = { id: string; tier: string; status: string; order_id: string; event_id?: string };

export default function BillingPage() {
  const [err, setErr] = useState<string | null>(null);
  const [events, setEvents] = useState<EventRow[]>([]);
  const [items, setItems] = useState<Checkout[]>([]);
  const [eventId, setEventId] = useState("");
  const [tier, setTier] = useState("pro");

  const load = useCallback(async () => {
    const [e, b] = await Promise.all([
      apiFetch<{ events: EventRow[] }>("/v1/events"),
      apiFetch<{ checkouts: Checkout[] }>("/v1/billing/checkouts"),
    ]);
    setEvents(e.events || []);
    setItems(b.checkouts || []);
    if (!eventId && e.events?.length) setEventId(e.events[0].id);
  }, [eventId]);

  useEffect(() => {
    if (!getAccessToken()) {
      window.location.href = "/login";
      return;
    }
    let active = true;
    void (async () => {
      try {
        await load();
      } catch (e) {
        if (active) setErr(String(e));
      }
    })();
    return () => {
      active = false;
    };
  }, [load]);

  async function createCheckout() {
    setErr(null);
    try {
      await apiFetch("/v1/billing/checkout", { method: "POST", json: { tier, event_id: eventId } });
      await load();
    } catch (e) {
      setErr(String(e));
    }
  }

  return (
    <main className="mx-auto max-w-4xl space-y-6 px-6 py-10 text-zinc-100">
      <Link href="/dashboard" className="text-sm text-zinc-400 hover:text-white">
        ← Dashboard
      </Link>
      <h1 className="text-2xl font-semibold text-white">Billing</h1>
      {err ? <p className="text-sm text-red-400">{err}</p> : null}

      <section className="rounded-xl border border-zinc-800 bg-zinc-900/40 p-4">
        <h2 className="text-sm font-medium text-zinc-300">Create checkout</h2>
        <div className="mt-3 grid gap-2 sm:grid-cols-2">
          <select className="rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm" value={eventId} onChange={(e) => setEventId(e.target.value)}>
            <option value="">Select event</option>
            {events.map((ev) => (
              <option key={ev.id} value={ev.id}>{ev.title}</option>
            ))}
          </select>
          <select className="rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm" value={tier} onChange={(e) => setTier(e.target.value)}>
            <option value="pro">Pro</option>
            <option value="elite">Elite</option>
          </select>
        </div>
        <button type="button" onClick={() => void createCheckout()} className="mt-3 rounded bg-amber-500 px-4 py-2 text-sm font-semibold text-black">
          Create checkout order
        </button>
      </section>

      <section className="rounded-xl border border-zinc-800 bg-zinc-900/40 p-4">
        <h2 className="text-sm font-medium text-zinc-300">Recent checkouts</h2>
        <ul className="mt-3 space-y-2 text-sm">
          {items.map((it) => (
            <li key={it.id} className="rounded border border-zinc-700 bg-zinc-950/60 p-3">
              <p className="font-medium">{it.tier.toUpperCase()} - {it.status}</p>
              <p className="text-zinc-400">order: {it.order_id}</p>
            </li>
          ))}
          {items.length === 0 ? <li className="text-zinc-500">No checkout records yet.</li> : null}
        </ul>
      </section>
    </main>
  );
}
