"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { apiFetch, getAccessToken } from "@/lib/api";

type EventRow = { id: string; slug: string; title: string };

export default function DashboardPage() {
  const [events, setEvents] = useState<EventRow[]>([]);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    if (!getAccessToken()) {
      window.location.href = "/login";
      return;
    }
    void (async () => {
      try {
        const d = await apiFetch<{ events: EventRow[] }>("/v1/events");
        setEvents(d.events || []);
      } catch (e) {
        setErr(String(e));
      }
    })();
  }, []);

  return (
    <main className="mx-auto max-w-3xl space-y-8 px-6 py-12">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <h1 className="text-2xl font-semibold text-white">Your events</h1>
        <div className="flex gap-2">
          <Link
            href="/organiser"
            className="rounded-lg border border-zinc-700 px-3 py-2 text-sm text-zinc-100 hover:border-amber-500/60"
          >
            Organiser
          </Link>
          <Link
            href="/billing"
            className="rounded-lg border border-zinc-700 px-3 py-2 text-sm text-zinc-100 hover:border-amber-500/60"
          >
            Billing
          </Link>
          <Link
            href="/events/new"
            className="rounded-lg bg-amber-500 px-3 py-2 text-sm font-semibold text-black hover:bg-amber-400"
          >
            New event
          </Link>
        </div>
      </div>
      {err && <p className="text-sm text-red-400">{err}</p>}
      <ul className="space-y-3">
        {events.map((e) => (
          <li key={e.id} className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-4">
            <Link href={`/events/${e.id}`} className="text-lg text-white hover:underline">
              {e.title}
            </Link>
            <p className="text-sm text-zinc-400">Slug: {e.slug}</p>
          </li>
        ))}
        {events.length === 0 && !err && <p className="text-sm text-zinc-500">No events yet.</p>}
      </ul>
    </main>
  );
}
