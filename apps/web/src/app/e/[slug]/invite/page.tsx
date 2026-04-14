"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import {
  buildGuestInviteUrl,
  buildWhatsAppInviteMessage,
  formatCoupleLine,
  formatDateRange,
  whatsappSendUrl,
} from "@/lib/inviteShare";

export default function PublicInvitePage() {
  const { slug } = useParams();
  const s = String(slug || "");
  const [ev, setEv] = useState<Record<string, unknown> | null>(null);
  const [err, setErr] = useState<string | null>(null);
  const [opened, setOpened] = useState(false);
  const [origin, setOrigin] = useState("");

  useEffect(() => {
    setOrigin(typeof window !== "undefined" ? window.location.origin : "");
  }, []);

  useEffect(() => {
    void (async () => {
      try {
        const r = await fetch(`/v1/public/events/${s}`);
        if (!r.ok) {
          setErr("Event not found.");
          return;
        }
        const e = (await r.json()) as Record<string, unknown>;
        setEv(e);
      } catch (e) {
        setErr(String(e));
      }
    })();
  }, [s]);

  const inviteUrl = origin ? buildGuestInviteUrl(origin, s) : "";
  const waUrl =
    ev && inviteUrl ? whatsappSendUrl(buildWhatsAppInviteMessage({ ev, inviteUrl })) : "";

  const names = ev ? formatCoupleLine(ev) : "";
  const when = ev ? formatDateRange(ev.date_start, ev.date_end) : null;

  return (
    <div className="min-h-[100dvh] bg-gradient-to-b from-zinc-950 via-zinc-900 to-black text-zinc-100">
      <div className="mx-auto flex min-h-[100dvh] max-w-lg flex-col items-center justify-center px-4 py-10">
        {err && <p className="mb-6 text-center text-sm text-red-400">{err}</p>}

        {!opened ? (
          <button
            type="button"
            onClick={() => setOpened(true)}
            className="utsav-invite-shell w-full max-w-sm focus:outline-none focus-visible:ring-2 focus-visible:ring-amber-400/80"
            aria-label="Open invitation"
          >
            <div className="utsav-invite-closed rounded-2xl border border-amber-600/40 bg-gradient-to-br from-amber-950/60 to-zinc-950 px-8 py-14 text-center shadow-lg">
              <p className="text-xs font-medium uppercase tracking-[0.2em] text-amber-200/90">Invitation</p>
              <p className="mt-4 text-lg font-semibold text-white">{names || "Tap to open"}</p>
              <p className="mt-6 text-sm text-amber-100/80">Tap to open</p>
            </div>
          </button>
        ) : (
          <div className="utsav-invite-shell w-full max-w-sm">
            <div className="utsav-invite-reveal rounded-2xl border border-amber-500/35 bg-zinc-900/90 p-8 text-center shadow-2xl shadow-amber-900/20 backdrop-blur-sm">
              <p className="text-xs font-medium uppercase tracking-[0.25em] text-amber-400/90">With love</p>
              <h1 className="mt-3 font-serif text-3xl font-semibold leading-tight text-white">
                {names || String(ev?.title || "You're invited")}
              </h1>
              {when ? <p className="mt-4 text-sm text-zinc-400">{when}</p> : null}
              <p className="mt-6 text-sm leading-relaxed text-zinc-300">
                We would be honoured to celebrate with you. RSVP and schedule are just a tap away.
              </p>
              <div className="mt-8 flex flex-col gap-3 sm:flex-row sm:justify-center">
                <Link
                  href={`/e/${s}`}
                  className="rounded-lg bg-amber-500 px-5 py-3 text-center text-sm font-semibold text-black hover:bg-amber-400"
                >
                  View event
                </Link>
                {waUrl ? (
                  <a
                    href={waUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="rounded-lg border border-zinc-600 px-5 py-3 text-center text-sm font-medium text-white hover:border-amber-500/50"
                  >
                    Share on WhatsApp
                  </a>
                ) : null}
              </div>
              <div className="mt-6 flex flex-wrap justify-center gap-4 text-xs text-zinc-500">
                <Link href={`/e/${s}/rsvp`} className="hover:text-amber-400">
                  RSVP
                </Link>
                <span aria-hidden>·</span>
                <Link href={`/e/${s}/shagun`} className="hover:text-amber-400">
                  Shagun
                </Link>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
