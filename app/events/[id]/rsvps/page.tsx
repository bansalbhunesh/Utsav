"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useState } from "react";
import { apiFetch } from "@/lib/api";
import { useQuery } from "@tanstack/react-query";
import { getUserFacingError } from "@/lib/error-messages";
import { parseHostRSVPResponse } from "@/lib/contracts/host";

type Row = {
  id: string;
  guest_phone: string;
  sub_event_id: string;
  status: string;
};

export default function RSVPsHostPage() {
  const { id } = useParams();
  const eventId = String(id || "");
  const [actionErr] = useState<string | null>(null);
  const { data, error } = useQuery({
    queryKey: ["event-rsvps", eventId],
    queryFn: async () => {
      const raw = await apiFetch<unknown>(`/v1/events/${eventId}/rsvps`);
      return parseHostRSVPResponse(raw);
    },
  });
  const rows: Row[] = data?.rsvps || [];
  const err = error ? getUserFacingError(error, "Failed to load RSVPs.") : actionErr;

  return (
    <main className="mx-auto max-w-4xl space-y-6 px-6 py-12">
      <Link href={`/events/${eventId}`} className="text-sm text-zinc-400">
        ← Event
      </Link>
      <h1 className="text-xl font-semibold text-white">RSVP responses</h1>
      {err && <p className="text-sm text-red-400">{err}</p>}
      <div className="overflow-x-auto rounded-lg border border-zinc-800">
        <table className="w-full text-left text-sm text-zinc-200">
          <thead className="border-b border-zinc-800 bg-zinc-900">
            <tr>
              <th className="p-3">Phone</th>
              <th className="p-3">Sub-event</th>
              <th className="p-3">Status</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((r) => (
              <tr key={r.id} className="border-b border-zinc-800/80">
                <td className="p-3">{r.guest_phone}</td>
                <td className="p-3 font-mono text-xs">{r.sub_event_id}</td>
                <td className="p-3">{r.status}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </main>
  );
}
