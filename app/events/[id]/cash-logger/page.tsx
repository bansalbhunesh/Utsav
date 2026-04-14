"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useState } from "react";
import { apiFetch, getAccessToken } from "@/lib/api";

export default function CashLoggerPage() {
  const { id } = useParams();
  const eventId = String(id || "");
  const [guestPhone, setGuestPhone] = useState("");
  const [amount, setAmount] = useState("5000");
  const [msg, setMsg] = useState<string | null>(null);

  if (typeof window !== "undefined" && !getAccessToken()) {
    window.location.href = "/login";
  }

  async function logCash() {
    setMsg(null);
    try {
      await apiFetch(`/v1/events/${eventId}/cash-shagun`, {
        method: "POST",
        json: { guest_phone: guestPhone, amount_inr: Number(amount) },
      });
      setMsg("Logged.");
      setGuestPhone("");
    } catch (e) {
      setMsg(String(e));
    }
  }

  return (
    <main className="mx-auto max-w-md space-y-6 px-4 py-8">
      <Link href={`/events/${eventId}`} className="text-sm text-zinc-400">
        ← Event
      </Link>
      <h1 className="text-2xl font-semibold text-white">Fast cash logger</h1>
      <input
        className="w-full rounded-xl border-2 border-zinc-700 bg-zinc-900 px-4 py-4 text-2xl text-white"
        placeholder="Guest phone"
        inputMode="tel"
        value={guestPhone}
        onChange={(e) => setGuestPhone(e.target.value)}
      />
      <input
        className="w-full rounded-xl border-2 border-zinc-700 bg-zinc-900 px-4 py-4 text-3xl font-bold text-amber-400"
        inputMode="decimal"
        value={amount}
        onChange={(e) => setAmount(e.target.value)}
      />
      <button
        type="button"
        onClick={() => void logCash()}
        className="w-full rounded-xl bg-amber-500 py-5 text-xl font-semibold text-black active:scale-[0.99]"
      >
        Log cash shagun
      </button>
      {msg && <p className="text-center text-sm text-zinc-400">{msg}</p>}
    </main>
  );
}
