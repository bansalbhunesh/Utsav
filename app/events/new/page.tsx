<<<<<<< HEAD
import { EventWizard } from "@/components/event/EventWizard";
import Link from "next/link";
import { ChevronLeft } from "lucide-react";

export default function NewEventPage() {
  return (
    <div className="min-h-screen flex flex-col bg-zinc-50 lg:bg-white pb-20">
      {/* Header */}
      <header className="p-6 flex items-center justify-between max-w-7xl mx-auto w-full">
        <Link 
          href="/dashboard" 
          className="inline-flex items-center text-sm font-medium text-zinc-500 hover:text-orange-600 transition-colors"
        >
          <ChevronLeft className="mr-1 h-4 w-4" />
          Back to Dashboard
        </Link>
        <div className="flex items-center gap-2">
           <div className="w-8 h-8 bg-orange-600 rounded-lg flex items-center justify-center text-white font-bold">
              U
           </div>
           <span className="font-bold text-zinc-900 tracking-tight">UTSAV</span>
        </div>
      </header>

      {/* Main Content */}
      <main className="flex-1 px-6 space-y-10">
        <div className="max-w-2xl mx-auto text-center space-y-3">
          <h1 className="text-4xl font-bold font-heading tracking-tight text-zinc-900 leading-tight">
            Create New Event
          </h1>
          <p className="text-zinc-500 max-w-md mx-auto">
            Set up your event details in just a few minutes. You can always edit 
            these later from your dashboard.
          </p>
        </div>

        <EventWizard />
      </main>
    </div>
=======
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
>>>>>>> f7494df (feat: Architectural Level Up - Go-Authoritative Backend, RSVP OTP Flow, and Frontend Consolidation (v1.5 Final))
  );
}
