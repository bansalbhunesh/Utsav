"use client";

import Link from "next/link";
import { useState } from "react";
import { apiFetch } from "@/lib/api";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { getUserFacingError } from "@/lib/error-messages";
import {
  parseHostBillingCheckoutsResponse,
  parseHostEventsResponse,
} from "@/lib/contracts/host";

type EventRow = { id: string; title: string };
type Checkout = { id: string; tier: string; status: string; order_id: string; event_id?: string };

export default function BillingPage() {
  const [actionErr, setActionErr] = useState<string | null>(null);
  const [eventId, setEventId] = useState("");
  const [tier, setTier] = useState("pro");
  const queryClient = useQueryClient();

  const { data, error } = useQuery({
    queryKey: ["billing-overview"],
    queryFn: async () => {
      const [eRaw, bRaw] = await Promise.all([
        apiFetch<unknown>("/v1/events"),
        apiFetch<unknown>("/v1/billing/checkouts"),
      ]);
      return {
        events: parseHostEventsResponse(eRaw).events as EventRow[],
        items: parseHostBillingCheckoutsResponse(bRaw).checkouts as Checkout[],
      };
    },
  });
  const events = data?.events || [];
  const items = data?.items || [];
  const err = error ? getUserFacingError(error, "Failed to load billing data.") : actionErr;

  const selectedEventId = eventId || events[0]?.id || "";

  const createCheckoutMutation = useMutation({
    mutationFn: async () =>
      apiFetch("/v1/billing/checkout", { method: "POST", json: { tier, event_id: selectedEventId } }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["billing-overview"] });
    },
  });

  async function createCheckout() {
    setActionErr(null);
    try {
      await createCheckoutMutation.mutateAsync();
    } catch (e) {
      setActionErr(getUserFacingError(e, "Failed to create checkout."));
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
          <select className="rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm" value={selectedEventId} onChange={(e) => setEventId(e.target.value)}>
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
        <button
          type="button"
          onClick={() => void createCheckout()}
          disabled={!selectedEventId}
          className="mt-3 rounded bg-amber-500 px-4 py-2 text-sm font-semibold text-black disabled:opacity-50 disabled:cursor-not-allowed"
        >
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
