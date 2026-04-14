"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useState } from "react";
import { getGuestToken, guestApiFetch } from "@/lib/api";

export default function GuestShagunPage() {
  const { slug } = useParams();
  const s = String(slug || "");
  const [amount, setAmount] = useState("5000");
  const [blessing, setBlessing] = useState("");
  const [upi, setUpi] = useState<string | null>(null);
  const [msg, setMsg] = useState<string | null>(null);

  async function loadUpi() {
    setMsg(null);
    if (!getGuestToken()) {
      setMsg("Verify RSVP first (guest token required).");
      return;
    }
    const d = await fetch(`/v1/public/events/${s}/upi-link`, {
      headers: { Authorization: `Bearer ${getGuestToken()}` },
    }).then((r) => r.json());
    if (d.error) throw new Error(d.error);
    setUpi(String(d.upi_uri));
  }

  async function report() {
    setMsg(null);
    await guestApiFetch(`/v1/public/events/${s}/shagun/report`, {
      method: "POST",
      json: { amount_inr: Number(amount), blessing_note: blessing },
    });
    setMsg("Recorded (guest-reported).");
  }

  return (
    <main className="mx-auto max-w-lg space-y-4 px-4 py-8">
      <Link href={`/e/${s}`} className="text-sm text-zinc-400">
        ← Event
      </Link>
      <h1 className="text-xl font-semibold text-white">Shagun</h1>
      <p className="text-sm text-zinc-500">UPI is peer-to-peer; we only store metadata after you pay in your UPI app.</p>
      <button type="button" className="rounded-lg border border-zinc-600 px-4 py-2 text-sm" onClick={() => void loadUpi().catch((e) => setMsg(String(e)))}>
        Get UPI link
      </button>
      {upi && (
        <a href={upi} className="block break-all rounded-lg bg-emerald-900/40 p-3 text-sm text-emerald-200">
          {upi}
        </a>
      )}
      <input className="w-full rounded-lg border border-zinc-800 bg-zinc-900 px-3 py-2 text-white" value={amount} onChange={(e) => setAmount(e.target.value)} placeholder="Amount INR" />
      <textarea className="w-full rounded-lg border border-zinc-800 bg-zinc-900 px-3 py-2 text-white" value={blessing} onChange={(e) => setBlessing(e.target.value)} placeholder="Blessing note" rows={3} />
      <button type="button" className="w-full rounded-lg bg-amber-500 py-3 text-sm font-semibold text-black" onClick={() => void report().catch((e) => setMsg(String(e)))}>
        I paid — record blessing
      </button>
      {msg && <p className="text-sm text-zinc-400">{msg}</p>}
    </main>
  );
}
