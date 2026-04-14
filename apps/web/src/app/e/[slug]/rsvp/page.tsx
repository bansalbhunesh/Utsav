"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { guestApiFetch, setGuestToken } from "@/lib/api";

type SubEvent = { id: string; name: string };

export default function GuestRSVPPage() {
  const { slug } = useParams();
  const s = String(slug || "");
  const [phone, setPhone] = useState("+919876543210");
  const [code, setCode] = useState("123456");
  const [subs, setSubs] = useState<SubEvent[]>([]);
  const [statusBySub, setStatusBySub] = useState<Record<string, string>>({});
  const [msg, setMsg] = useState<string | null>(null);

  useEffect(() => {
    void (async () => {
      try {
        const sc = await fetch(`/v1/public/events/${s}/schedule`).then((r) => r.json());
        const list = (sc as { sub_events?: SubEvent[] }).sub_events || [];
        setSubs(list);
        const init: Record<string, string> = {};
        for (const se of list) init[se.id] = "yes";
        setStatusBySub(init);
      } catch {
        /* ignore */
      }
    })();
  }, [s]);

  async function verify() {
    setMsg(null);
    const d = await fetch(`/v1/public/events/${s}/rsvp/otp/verify`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ phone, code }),
    }).then((r) => r.json());
    if (!d.guest_access_token) throw new Error(JSON.stringify(d));
    setGuestToken(d.guest_access_token);
    setMsg("Verified. Submit RSVP below.");
  }

  async function submit() {
    setMsg(null);
    const items = subs.map((se) => ({
      sub_event_id: se.id,
      status: statusBySub[se.id] || "yes",
    }));
    await guestApiFetch(`/v1/public/events/${s}/rsvp`, { method: "POST", json: { items } });
    setMsg("RSVP saved.");
  }

  return (
    <main className="mx-auto max-w-lg space-y-4 px-4 py-8">
      <Link href={`/e/${s}`} className="text-sm text-zinc-400">
        ← Event
      </Link>
      <h1 className="text-xl font-semibold text-white">RSVP</h1>
      <input className="w-full rounded-lg border border-zinc-800 bg-zinc-900 px-3 py-2 text-white" value={phone} onChange={(e) => setPhone(e.target.value)} />
      <button type="button" className="rounded-lg bg-zinc-800 px-4 py-2 text-sm" onClick={() => void fetch(`/v1/public/events/${s}/rsvp/otp/request`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ phone }) }).then(() => setMsg("OTP sent (dev: 123456)"))}>
        Request OTP
      </button>
      <input className="w-full rounded-lg border border-zinc-800 bg-zinc-900 px-3 py-2 text-white" value={code} onChange={(e) => setCode(e.target.value)} />
      <button type="button" className="rounded-lg border border-zinc-600 px-4 py-2 text-sm" onClick={() => void verify().catch((e) => setMsg(String(e)))}>
        Verify OTP
      </button>
      {subs.map((se) => (
        <label key={se.id} className="flex items-center justify-between gap-2 text-sm text-zinc-300">
          {se.name}
          <select
            className="rounded border border-zinc-700 bg-zinc-900 px-2 py-1 text-white"
            value={statusBySub[se.id] || "yes"}
            onChange={(e) => setStatusBySub((m) => ({ ...m, [se.id]: e.target.value }))}
          >
            <option value="yes">Yes</option>
            <option value="no">No</option>
            <option value="maybe">Maybe</option>
          </select>
        </label>
      ))}
      <button type="button" className="w-full rounded-lg bg-amber-500 py-3 text-sm font-semibold text-black" onClick={() => void submit().catch((e) => setMsg(String(e)))}>
        Submit RSVP
      </button>
      {msg && <p className="text-sm text-zinc-400">{msg}</p>}
    </main>
  );
}
