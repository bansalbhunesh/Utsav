"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";

export default function PublicEventPage() {
  const { slug } = useParams();
  const s = String(slug || "");
  const [ev, setEv] = useState<Record<string, unknown> | null>(null);
  const [sched, setSched] = useState<unknown[]>([]);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    void (async () => {
      try {
        const e = await fetch(`/v1/public/events/${s}`).then((r) => r.json());
        const sc = await fetch(`/v1/public/events/${s}/schedule`).then((r) => r.json());
        setEv(e as Record<string, unknown>);
        setSched((sc as { sub_events?: unknown[] }).sub_events || []);
      } catch (e) {
        setErr(String(e));
      }
    })();
  }, [s]);

  return (
    <main className="mx-auto max-w-lg space-y-6 px-4 py-8 text-zinc-100">
      {err && <p className="text-sm text-red-400">{err}</p>}
      {ev && (
        <>
          <h1 className="text-2xl font-semibold text-white">{String(ev.title)}</h1>
          <p className="text-sm text-zinc-400">You&apos;re viewing the guest PWA shell.</p>
          <p className="text-sm">
            <Link href={`/e/${s}/invite`} className="text-amber-400 underline-offset-2 hover:underline">
              View digital invite
            </Link>
          </p>
        </>
      )}
      <section className="space-y-2">
        <h2 className="text-sm font-medium text-zinc-400">Schedule</h2>
        {sched.map((row: unknown, i: number) => {
          const r = row as Record<string, unknown>;
          return (
            <div key={i} className="rounded-lg border border-zinc-800 bg-zinc-900/50 p-3 text-sm">
              <div className="font-medium text-white">{String(r.name)}</div>
              {r.happening_now ? (
                <span className="text-xs font-semibold text-amber-400">Happening now</span>
              ) : null}
            </div>
          );
        })}
      </section>
      <div className="flex gap-3">
        <Link href={`/e/${s}/rsvp`} className="flex-1 rounded-lg bg-amber-500 py-3 text-center text-sm font-semibold text-black">
          RSVP
        </Link>
        <Link href={`/e/${s}/shagun`} className="flex-1 rounded-lg border border-zinc-600 py-3 text-center text-sm text-white">
          Shagun
        </Link>
      </div>
    </main>
  );
}
