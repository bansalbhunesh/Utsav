"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useState } from "react";
import { apiFetch } from "@/lib/api";
import { useQuery } from "@tanstack/react-query";
import { getUserFacingError } from "@/lib/error-messages";
import { parseHostShagunResponse } from "@/lib/contracts/host";

type Row = { id: string; channel: string; amount_paise: number | null; blessing_note: string; status: string };

export default function ShagunHostPage() {
  const { id } = useParams();
  const eventId = String(id || "");
  const [actionErr] = useState<string | null>(null);
  const { data, error } = useQuery({
    queryKey: ["event-shagun", eventId],
    queryFn: async () => {
      const raw = await apiFetch<unknown>(`/v1/events/${eventId}/shagun`);
      return parseHostShagunResponse(raw);
    },
  });
  const rows: Row[] = data?.shagun || [];
  const err = error ? getUserFacingError(error, "Failed to load shagun entries.") : actionErr;

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
