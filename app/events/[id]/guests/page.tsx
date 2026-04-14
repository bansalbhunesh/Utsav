"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { apiFetch } from "@/lib/api";

type Guest = { id: string; name: string; phone: string };

export default function GuestsPage() {
  const { id } = useParams();
  const eventId = String(id || "");
  const [guests, setGuests] = useState<Guest[]>([]);
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [csv, setCsv] = useState("name,phone\nPriya,+919876543210");
  const [importMsg, setImportMsg] = useState<string | null>(null);
  const [err, setErr] = useState<string | null>(null);

  const load = useCallback(async () => {
    const d = await apiFetch<{ guests: Guest[] }>(`/v1/events/${eventId}/guests`);
    setGuests(d.guests || []);
  }, [eventId]);

  useEffect(() => {
    let active = true;
    void (async () => {
      try {
        await load();
      } catch (e) {
        if (active) setErr(String(e));
      }
    })();
    return () => {
      active = false;
    };
  }, [load]);

  async function addGuest() {
    setErr(null);
    try {
      await apiFetch(`/v1/events/${eventId}/guests`, {
        method: "POST",
        json: { name, phone },
      });
      setName("");
      setPhone("");
      await load();
    } catch (e) {
      setErr(String(e));
    }
  }

  async function importCsv() {
    setErr(null);
    setImportMsg(null);
    try {
      const d = await apiFetch<{ imported: number; errors: { line: number; error: string }[] }>(
        `/v1/events/${eventId}/guests/import`,
        { method: "POST", json: { csv } },
      );
      setImportMsg(`Imported ${d.imported}. Row errors: ${d.errors?.length ?? 0}.`);
      await load();
    } catch (e) {
      setErr(String(e));
    }
  }

  return (
    <main className="mx-auto max-w-3xl space-y-6 px-6 py-12">
      <Link href={`/events/${eventId}`} className="text-sm text-zinc-400">
        ← Event
      </Link>
      <h1 className="text-xl font-semibold text-white">Guests</h1>
      <section className="rounded-xl border border-zinc-800 bg-zinc-900/40 p-4">
        <h2 className="text-sm font-medium text-zinc-400">CSV import</h2>
        <p className="mt-1 text-xs text-zinc-500">
          Header row optional: <code className="text-zinc-400">name,phone,email,relationship,side</code>. Or two columns
          without header: name then phone.
        </p>
        <textarea
          className="mt-2 w-full rounded-lg border border-zinc-800 bg-zinc-950 px-3 py-2 font-mono text-sm text-white"
          rows={6}
          value={csv}
          onChange={(e) => setCsv(e.target.value)}
        />
        <div className="mt-2 flex flex-wrap gap-2">
          <button
            type="button"
            onClick={() => void importCsv()}
            className="rounded-lg border border-amber-600/60 bg-amber-500/10 px-4 py-2 text-sm font-medium text-amber-200"
          >
            Import CSV
          </button>
          <label className="cursor-pointer rounded-lg border border-zinc-700 px-4 py-2 text-sm text-zinc-300 hover:bg-zinc-800">
            Load file
            <input
              type="file"
              accept=".csv,text/csv,text/plain"
              className="hidden"
              onChange={(e) => {
                const f = e.target.files?.[0];
                if (!f) return;
                const reader = new FileReader();
                reader.onload = () => setCsv(String(reader.result || ""));
                reader.readAsText(f);
              }}
            />
          </label>
        </div>
        {importMsg && <p className="mt-2 text-sm text-emerald-400">{importMsg}</p>}
      </section>
      <div className="flex flex-wrap gap-2">
        <input
          className="min-w-[8rem] flex-1 rounded-lg border border-zinc-800 bg-zinc-900 px-3 py-2 text-white"
          placeholder="Name"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
        <input
          className="min-w-[8rem] flex-1 rounded-lg border border-zinc-800 bg-zinc-900 px-3 py-2 text-white"
          placeholder="Phone"
          value={phone}
          onChange={(e) => setPhone(e.target.value)}
        />
        <button
          type="button"
          onClick={() => void addGuest()}
          className="rounded-lg bg-amber-500 px-4 py-2 text-sm font-semibold text-black"
        >
          Add
        </button>
      </div>
      {err && <p className="text-sm text-red-400">{err}</p>}
      <ul className="space-y-2 text-sm">
        {guests.map((g) => (
          <li key={g.id} className="rounded border border-zinc-800 px-3 py-2 text-zinc-200">
            {g.name} — {g.phone}
          </li>
        ))}
      </ul>
    </main>
  );
}
