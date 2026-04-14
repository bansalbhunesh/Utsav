"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { apiFetch, getAccessToken } from "@/lib/api";

export default function NewEventPage() {
  const [slug, setSlug] = useState("demo-wedding");
  const [title, setTitle] = useState("Demo Wedding");
  const [vpa, setVpa] = useState("host@upi");
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    if (!getAccessToken()) window.location.href = "/login";
  }, []);

  async function create() {
    setErr(null);
    try {
      const res = await apiFetch<{ id: string }>("/v1/events", {
        method: "POST",
        json: { slug, title, host_upi_vpa: vpa, privacy: "public" },
      });
      window.location.href = `/events/${res.id}`;
    } catch (e) {
      setErr(String(e));
    }
  }

  return (
    <main className="mx-auto max-w-lg space-y-6 px-6 py-12">
      <Link href="/dashboard" className="text-sm text-zinc-400 hover:text-white">
        ← Dashboard
      </Link>
      <h1 className="text-2xl font-semibold text-white">Create event</h1>
      <label className="block text-sm text-zinc-400">
        URL slug
        <input
          className="mt-1 w-full rounded-lg border border-zinc-800 bg-zinc-900 px-3 py-2 text-white"
          value={slug}
          onChange={(e) => setSlug(e.target.value)}
        />
      </label>
      <label className="block text-sm text-zinc-400">
        Title
        <input
          className="mt-1 w-full rounded-lg border border-zinc-800 bg-zinc-900 px-3 py-2 text-white"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
        />
      </label>
      <label className="block text-sm text-zinc-400">
        Host UPI VPA
        <input
          className="mt-1 w-full rounded-lg border border-zinc-800 bg-zinc-900 px-3 py-2 text-white"
          value={vpa}
          onChange={(e) => setVpa(e.target.value)}
        />
      </label>
      <button
        type="button"
        onClick={() => void create()}
        className="rounded-lg bg-amber-500 px-4 py-2 text-sm font-semibold text-black"
      >
        Create
      </button>
      {err && <p className="text-sm text-red-400">{err}</p>}
    </main>
  );
}
