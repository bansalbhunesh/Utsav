"use client";

import Link from "next/link";
import { useState } from "react";

export default function Home() {
  const [note, setNote] = useState("");
  return (
    <main className="mx-auto flex min-h-screen max-w-3xl flex-col gap-10 px-6 py-16">
      <header className="space-y-3">
        <p className="text-sm uppercase tracking-[0.2em] text-amber-400">UTSAV · उत्सव</p>
        <h1 className="text-4xl font-semibold text-white sm:text-5xl">
          India&apos;s event operating system
        </h1>
        <p className="text-lg text-zinc-300">
          Replace WhatsApp chaos, notebook shagun tracking, and scattered vendor notes with one calm,
          mobile-first platform — built for real weddings first, then every celebration.
        </p>
      </header>
      <section className="grid gap-4 sm:grid-cols-2">
        <button
          type="button"
          className="rounded-xl border border-zinc-800 bg-zinc-900 px-4 py-3 text-center text-sm font-medium text-white hover:border-amber-500/60"
          onClick={() =>
            void (async () => {
              try {
                const r = await fetch("/v1/healthz");
                setNote(`API health: ${r.status}`);
              } catch (e) {
                setNote(String(e));
              }
            })()
          }
        >
          Ping API (/v1/healthz)
        </button>
        <div className="grid gap-2">
          <Link
            href="/login"
            className="rounded-xl border border-zinc-800 bg-zinc-900 px-4 py-3 text-center text-sm font-medium text-white hover:border-amber-500/60"
          >
            Host login
          </Link>
          <Link
            href="/dashboard"
            className="rounded-xl border border-zinc-800 px-4 py-3 text-center text-sm font-medium text-zinc-200 hover:border-zinc-600"
          >
            Dashboard
          </Link>
        </div>
      </section>
      {note && <p className="text-sm text-zinc-400">{note}</p>}
      <section className="rounded-2xl border border-zinc-800 bg-zinc-900/40 p-6 text-sm text-zinc-300">
        <p className="font-medium text-white">Architecture in this repo</p>
        <ul className="mt-3 list-disc space-y-2 pl-5">
          <li>Next.js App Router + Tailwind in `apps/web`</li>
          <li>Go + Gin REST API in `services/api`</li>
          <li>PostgreSQL migrations in `db/migrations`</li>
          <li>Local Postgres via `infra/docker/compose.yml`</li>
        </ul>
        <p className="mt-4">
          <Link href="/e/demo-wedding" className="text-amber-400 hover:underline">
            Guest preview
          </Link>{" "}
          (replace slug after you create an event)
        </p>
      </section>
    </main>
  );
}
