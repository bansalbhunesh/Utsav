"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useState } from "react";
import { apiFetch } from "@/lib/api";
import { getUserFacingError } from "@/lib/error-messages";

export default function CashLoggerPage() {
  const { id } = useParams();
  const eventId = String(id || "");
  const [guestPhone, setGuestPhone] = useState("");
  const [amount, setAmount] = useState("5000");
  const [msg, setMsg] = useState<string | null>(null);

  async function logCash() {
    setMsg(null);
    try {
      await apiFetch(`/v1/events/${eventId}/cash-shagun`, {
        method: "POST",
        json: { guest_phone: guestPhone, amount_inr: Number(amount) },
      });
      setMsg("Logged successfully.");
      setGuestPhone("");
    } catch (e) {
      setMsg(getUserFacingError(e, "Failed to log cash entry."));
    }
  }

  return (
    <main className="mx-auto max-w-md space-y-6 px-4 py-8">
      <Link href={`/events/${eventId}`} className="text-sm font-medium text-zinc-400 hover:text-amber-500 transition-colors">
        ← Back to Event
      </Link>
      <h1 className="text-2xl font-bold text-white tracking-tight">Fast Cash Logger</h1>
      <div className="space-y-4">
        <div className="space-y-1">
          <label className="text-[10px] font-bold text-zinc-500 uppercase tracking-widest ml-1">Guest Phone</label>
          <input
            className="w-full rounded-xl border-2 border-zinc-800 bg-zinc-900 px-4 py-4 text-2xl text-white outline-none focus:border-amber-500 transition-all"
            placeholder="9876543210"
            inputMode="tel"
            value={guestPhone}
            onChange={(e) => setGuestPhone(e.target.value)}
          />
        </div>
        <div className="space-y-1">
          <label className="text-[10px] font-bold text-zinc-500 uppercase tracking-widest ml-1">Amount (₹)</label>
          <input
            className="w-full rounded-xl border-2 border-zinc-800 bg-zinc-900 px-4 py-4 text-3xl font-bold text-amber-400 outline-none focus:border-amber-500 transition-all"
            inputMode="decimal"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
          />
        </div>
      </div>
      <button
        type="button"
        onClick={() => void logCash()}
        className="w-full rounded-xl bg-amber-500 py-5 text-xl font-bold text-black active:scale-[0.98] transition-transform shadow-lg shadow-amber-900/20"
      >
        Log Cash Entry
      </button>
      {msg && <p className="text-center text-sm font-medium text-zinc-400 animate-in fade-in transition-all">{msg}</p>}
    </main>
  );
}
