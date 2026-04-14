"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useState } from "react";
import { apiFetch } from "@/lib/api";

type GenerateResp = {
  slug: string;
  public_api_path: string;
  payload: Record<string, unknown>;
  export_pdf_available: boolean;
};

export default function EventMemoryBookPage() {
  const params = useParams();
  const id = String(params.id || "");
  const [err, setErr] = useState<string | null>(null);
  const [data, setData] = useState<GenerateResp | null>(null);
  const [exportMsg, setExportMsg] = useState<string>("");

  async function generate() {
    setErr(null);
    setExportMsg("");
    try {
      const d = await apiFetch<GenerateResp>(`/v1/events/${id}/memory-book/generate`, {
        method: "POST",
      });
      setData(d);
    } catch (e) {
      setErr(String(e));
    }
  }

  async function exportPdf() {
    setErr(null);
    setExportMsg("");
    try {
      const d = await apiFetch<{ status?: string; next_step?: string }>(
        `/v1/events/${id}/memory-book/export`,
        { method: "POST" },
      );
      setExportMsg(d.next_step || d.status || "Export requested");
    } catch (e) {
      setErr(String(e));
    }
  }

  return (
    <main className="mx-auto max-w-4xl space-y-6 px-6 py-10 text-zinc-100">
      <Link href={`/events/${id}`} className="text-sm text-zinc-400 hover:text-white">
        ← Back to event
      </Link>
      <div>
        <h1 className="text-2xl font-semibold text-white">Memory Book</h1>
        <p className="text-sm text-zinc-400">Generate an aggregate memory payload from event activity.</p>
      </div>

      <div className="flex flex-wrap gap-3">
        <button
          type="button"
          onClick={() => void generate()}
          className="rounded-lg bg-amber-500 px-4 py-2 text-sm font-semibold text-black"
        >
          Generate payload
        </button>
        <button
          type="button"
          onClick={() => void exportPdf()}
          className="rounded-lg border border-zinc-600 px-4 py-2 text-sm text-white"
        >
          Export PDF (tier-gated)
        </button>
      </div>

      {err ? <p className="text-sm text-red-400">{err}</p> : null}
      {exportMsg ? <p className="text-sm text-emerald-400">{exportMsg}</p> : null}

      {data ? (
        <section className="rounded-xl border border-zinc-800 bg-zinc-900/40 p-4">
          <p className="text-xs text-zinc-500">Slug: {data.slug}</p>
          <p className="mt-1 text-xs text-zinc-500">Public API: {data.public_api_path}</p>
          <p className="mt-1 text-xs text-zinc-500">
            PDF export available: {data.export_pdf_available ? "yes" : "no (upgrade tier)"}
          </p>
          <pre className="mt-4 overflow-auto rounded-lg border border-zinc-800 bg-zinc-950 p-3 text-xs text-zinc-200">
            {JSON.stringify(data.payload, null, 2)}
          </pre>
        </section>
      ) : null}
    </main>
  );
}
