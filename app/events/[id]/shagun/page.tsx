"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { apiFetch } from "@/lib/api";

type Row = { id: string; channel: string; amount_paise: number | null; blessing_note: string; status: string };

export default function ShagunHostPage() {
  const { id } = useParams();
  const eventId = String(id || "");
  const [rows, setRows] = useState<Row[]>([]);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    void (async () => {
      try {
        const d = await apiFetch<{ shagun: Row[] }>(`/v1/events/${eventId}/shagun`);
        setRows(d.shagun || []);
      } catch (e) {
        setErr(String(e));
      }
    })();
  }, [eventId]);

  return (
    <main className="mx-auto max-w-3xl space-y-6 px-6 py-12">
      <Link href={`/events/${eventId}`} className="text-sm text-zinc-400">
        ← Event
      </Link>
      <h1 className="text-xl font-semibold text-white">Shagun (host view)</h1>
      {err && <p className="text-sm text-red-400">{err}</p>}
      <ul className="space-y-2 text-sm">
        {rows.map((r) => (
          <li key={r.id} className="rounded border border-zinc-800 px-3 py-2 text-zinc-200">
            {r.channel} — ₹{r.amount_paise != null ? (r.amount_paise / 100).toFixed(2) : "—"} — {r.status}
            {r.blessing_note ? <span className="block text-zinc-500">&ldquo;{r.blessing_note}&rdquo;</span> : null}
          </li>
        ))}
      </ul>
    </main>
  );
}
