"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useState } from "react";
import { apiFetch } from "@/lib/api";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { getUserFacingError } from "@/lib/error-messages";
import {
  buildGuestInviteUrl,
  buildWhatsAppInviteMessage,
  whatsappSendUrl,
} from "@/lib/inviteShare";
import {
  parseHostEventDetail,
  parseHostSubEventsResponse,
} from "@/lib/contracts/host";

type SubEvent = {
  id: string;
  name: string;
  sub_type?: string;
  starts_at?: string | null;
  venue_label?: string;
  dress_code?: string;
};

export default function EventDetailPage() {
  const params = useParams();
  const id = String(params.id || "");
  const [actionErr, setActionErr] = useState<string | null>(null);
  const [seName, setSeName] = useState("Ceremony");
  const [seType, setSeType] = useState("ceremony");
  const [seStart, setSeStart] = useState("");
  const origin = typeof window !== "undefined" ? window.location.origin : "";
  const [copied, setCopied] = useState(false);
  const queryClient = useQueryClient();

  const { data, error } = useQuery({
    queryKey: ["event-detail", id],
    queryFn: async () => {
      const [eventRaw, subsRaw] = await Promise.all([
        apiFetch<unknown>(`/v1/events/${id}`),
        apiFetch<unknown>(`/v1/events/${id}/sub-events`),
      ]);
      const event = parseHostEventDetail(eventRaw);
      const subs = parseHostSubEventsResponse(subsRaw).sub_events as SubEvent[];
      return { event, subs };
    },
  });

  const title = data?.event.title || "";
  const slug = data?.event.slug || "";
  const subs = data?.subs || [];
  const shareMeta = data?.event || {};
  const err = error ? getUserFacingError(error, "Failed to load event details.") : actionErr;

  const addSubEventMutation = useMutation({
    mutationFn: async () => {
      const json: Record<string, unknown> = { name: seName, sub_type: seType };
      if (seStart.trim()) json.starts_at = new Date(seStart).toISOString();
      return apiFetch(`/v1/events/${id}/sub-events`, { method: "POST", json });
    },
    onSuccess: async () => {
      setSeName("Ceremony");
      setSeType("ceremony");
      setSeStart("");
      await queryClient.invalidateQueries({ queryKey: ["event-detail", id] });
    },
  });

  async function addSubEvent() {
    setActionErr(null);
    try {
      await addSubEventMutation.mutateAsync();
    } catch (e) {
      setActionErr(getUserFacingError(e, "Failed to add sub-event."));
    }
  }

  const links: [string, string][] = [
    ["Guests", `/events/${id}/guests`],
    ["Gallery", `/events/${id}/gallery`],
    ["Broadcasts", `/events/${id}/broadcasts`],
    ["Memory book", `/events/${id}/memory-book`],
    ["Fast cash logger", `/events/${id}/cash-logger`],
    ["RSVPs", `/events/${id}/rsvps`],
    ["Shagun", `/events/${id}/shagun`],
  ];
  if (slug) links.push(["Guest link", `/e/${slug}`]);
  if (slug) links.push(["Digital invite", `/e/${slug}/invite`]);

  const inviteUrl = slug && origin ? buildGuestInviteUrl(origin, slug) : "";
  const hostWaUrl =
    inviteUrl && Object.keys(shareMeta).length
      ? whatsappSendUrl(buildWhatsAppInviteMessage({ ev: shareMeta, inviteUrl }))
      : "";

  return (
    <main className="mx-auto max-w-3xl space-y-8 px-6 py-12">
      <Link href="/dashboard" className="text-sm text-zinc-400 hover:text-white">
        ← Dashboard
      </Link>
      {err && <p className="text-sm text-red-400">{err}</p>}
      <h1 className="text-2xl font-semibold text-white">{title || "Event"}</h1>
      {slug && <p className="text-sm text-zinc-500">Public slug: {slug}</p>}

      {inviteUrl ? (
        <section className="rounded-xl border border-amber-900/40 bg-amber-950/15 p-4">
          <h2 className="text-sm font-medium text-amber-200">Invite guests</h2>
          <p className="mt-1 break-all font-mono text-xs text-zinc-400">{inviteUrl}</p>
          <div className="mt-3 flex flex-wrap gap-2">
            <button
              type="button"
              className="rounded-lg border border-amber-700/50 bg-zinc-950 px-3 py-2 text-sm text-amber-100 hover:border-amber-500/60"
              onClick={() => {
                void navigator.clipboard.writeText(inviteUrl).then(() => {
                  setCopied(true);
                  window.setTimeout(() => setCopied(false), 2000);
                });
              }}
            >
              {copied ? "Copied" : "Copy invite link"}
            </button>
            {hostWaUrl ? (
              <a
                href={hostWaUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="rounded-lg bg-emerald-700 px-3 py-2 text-sm font-medium text-white hover:bg-emerald-600"
              >
                Share on WhatsApp
              </a>
            ) : null}
          </div>
        </section>
      ) : null}

      <section className="rounded-xl border border-zinc-800 bg-zinc-900/40 p-4">
        <h2 className="text-sm font-medium text-zinc-300">Sub-events</h2>
        <ul className="mt-2 space-y-1 text-sm text-zinc-200">
          {subs.map((s) => (
            <li key={s.id} className="flex flex-wrap justify-between gap-2 border-b border-zinc-800/80 py-2">
              <span className="font-medium text-white">{s.name}</span>
              <span className="text-zinc-500">{s.sub_type}</span>
            </li>
          ))}
          {subs.length === 0 && <li className="text-zinc-500">None yet — add one below.</li>}
        </ul>
        <div className="mt-4 grid gap-2 sm:grid-cols-3">
          <input
            className="rounded-lg border border-zinc-800 bg-zinc-950 px-3 py-2 text-sm text-white"
            placeholder="Name"
            value={seName}
            onChange={(e) => setSeName(e.target.value)}
          />
          <input
            className="rounded-lg border border-zinc-800 bg-zinc-950 px-3 py-2 text-sm text-white"
            placeholder="Type"
            value={seType}
            onChange={(e) => setSeType(e.target.value)}
          />
          <input
            type="datetime-local"
            className="rounded-lg border border-zinc-800 bg-zinc-950 px-3 py-2 text-sm text-white"
            value={seStart}
            onChange={(e) => setSeStart(e.target.value)}
          />
        </div>
        <button
          type="button"
          onClick={() => void addSubEvent()}
          className="mt-3 rounded-lg bg-amber-500 px-4 py-2 text-sm font-semibold text-black"
        >
          Add sub-event
        </button>
      </section>

      <nav className="grid gap-2 sm:grid-cols-2">
        {links.map(([label, href]) => (
          <Link
            key={href}
            href={href}
            className="rounded-lg border border-zinc-800 bg-zinc-900/50 px-4 py-3 text-sm text-white hover:border-amber-500/50"
          >
            {label}
          </Link>
        ))}
      </nav>
    </main>
  );
}
