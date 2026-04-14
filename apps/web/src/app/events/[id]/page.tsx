"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { apiFetch, getAccessToken } from "@/lib/api";
import {
  buildGuestInviteUrl,
  buildWhatsAppInviteMessage,
  whatsappSendUrl,
} from "@/lib/inviteShare";

type SubEvent = {
  id: string;
  name: string;
  sub_type: string;
  starts_at: string | null;
  venue_label: string;
  dress_code: string;
};

export default function EventDetailPage() {
  const params = useParams();
  const id = String(params.id || "");
  const [title, setTitle] = useState("");
  const [slug, setSlug] = useState("");
  const [err, setErr] = useState<string | null>(null);
  const [subs, setSubs] = useState<SubEvent[]>([]);
  const [seName, setSeName] = useState("Ceremony");
  const [seType, setSeType] = useState("ceremony");
  const [seStart, setSeStart] = useState("");
  const [origin, setOrigin] = useState("");
  const [copied, setCopied] = useState(false);
  const [shareMeta, setShareMeta] = useState<Record<string, unknown>>({});

  useEffect(() => {
    setOrigin(typeof window !== "undefined" ? window.location.origin : "");
  }, []);

  const loadSubs = useCallback(async () => {
    const d = await apiFetch<{ sub_events: SubEvent[] }>(`/v1/events/${id}/sub-events`);
    setSubs(d.sub_events || []);
  }, [id]);

  useEffect(() => {
    if (!getAccessToken()) {
      window.location.href = "/login";
      return;
    }
    void (async () => {
      try {
        const d = await apiFetch<{
          title: string;
          slug: string;
          couple_name_a?: unknown;
          couple_name_b?: unknown;
          date_start?: unknown;
          date_end?: unknown;
        }>(`/v1/events/${id}`);
        setTitle(d.title);
        setSlug(d.slug);
        setShareMeta({
          title: d.title,
          couple_name_a: d.couple_name_a,
          couple_name_b: d.couple_name_b,
          date_start: d.date_start,
          date_end: d.date_end,
        });
        await loadSubs();
      } catch (e) {
        setErr(String(e));
      }
    })();
  }, [id, loadSubs]);

  async function addSubEvent() {
    setErr(null);
    try {
      const json: Record<string, unknown> = { name: seName, sub_type: seType };
      if (seStart.trim()) json.starts_at = new Date(seStart).toISOString();
      await apiFetch(`/v1/events/${id}/sub-events`, { method: "POST", json });
      setSeName("Ceremony");
      setSeType("ceremony");
      setSeStart("");
      await loadSubs();
    } catch (e) {
      setErr(String(e));
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
