"use client";

import Link from "next/link";
import Image from "next/image";
import { useParams } from "next/navigation";
import { useState } from "react";
import { apiFetch } from "@/lib/api";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { getUserFacingError } from "@/lib/error-messages";
import {
  parseHostGalleryAssetsResponse,
  parseHostGalleryPresignResponse,
} from "@/lib/contracts/host";

type Asset = {
  id: string;
  section: string;
  object_key: string;
  status: "pending" | "approved" | "rejected";
  url?: string;
  created_at?: string;
};

export default function EventGalleryPage() {
  const params = useParams();
  const id = String(params.id || "");
  const [actionErr, setActionErr] = useState<string | null>(null);
  const [status, setStatus] = useState<"pending" | "approved" | "rejected" | "all">("pending");
  const [section, setSection] = useState("moments");
  const [busy, setBusy] = useState(false);
  const queryClient = useQueryClient();

  const { data, error } = useQuery({
    queryKey: ["event-gallery-assets", id, status],
    queryFn: async () => {
      const q = status === "all" ? "" : `?status=${status}`;
      const raw = await apiFetch<unknown>(`/v1/events/${id}/gallery/assets${q}`);
      return parseHostGalleryAssetsResponse(raw);
    },
  });
  const assets: Asset[] = data?.assets || [];
  const err = error ? getUserFacingError(error, "Failed to load gallery assets.") : actionErr;

  const registerAssetMutation = useMutation({
    mutationFn: async (payload: {
      section: string;
      object_key: string;
      mime_type: string;
      bytes: number;
      status: "pending";
    }) =>
      apiFetch(`/v1/events/${id}/gallery/assets`, {
        method: "POST",
        json: payload,
      }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["event-gallery-assets", id] });
    },
  });

  const moderateAssetMutation = useMutation({
    mutationFn: async ({ assetId, next }: { assetId: string; next: "approved" | "rejected" | "pending" }) =>
      apiFetch(`/v1/events/${id}/gallery/assets/${assetId}`, {
        method: "PATCH",
        json: { status: next },
      }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["event-gallery-assets", id] });
    },
  });

  async function onUploadFile(file: File) {
    setActionErr(null);
    setBusy(true);
    try {
      const presignRaw = await apiFetch<unknown>(`/v1/events/${id}/gallery/presign`, {
        method: "POST",
        json: {
          section,
          file_name: file.name,
          content_type: file.type || "application/octet-stream",
        },
      });
      const p = parseHostGalleryPresignResponse(presignRaw);
      const upload = p.upload;
      const put = await fetch(upload.url, {
        method: upload.method || "PUT",
        headers: upload.headers,
        body: file,
      });
      if (!put.ok) throw new Error(`upload_failed_${put.status}`);

      await registerAssetMutation.mutateAsync({
        section,
        object_key: upload.object_key,
        mime_type: file.type || "application/octet-stream",
        bytes: file.size,
        status: "pending",
      });
    } catch (e) {
      setActionErr(getUserFacingError(e, "Failed to upload gallery asset."));
    } finally {
      setBusy(false);
    }
  }

  async function moderate(assetId: string, next: "approved" | "rejected" | "pending") {
    setActionErr(null);
    try {
      await moderateAssetMutation.mutateAsync({ assetId, next });
    } catch (e) {
      setActionErr(getUserFacingError(e, "Failed to update moderation status."));
    }
  }

  return (
    <main className="mx-auto max-w-5xl space-y-6 px-6 py-10 text-zinc-100">
      <Link href={`/events/${id}`} className="text-sm text-zinc-400 hover:text-white">
        ← Back to event
      </Link>
      <div>
        <h1 className="text-2xl font-semibold text-white">Gallery moderation</h1>
        <p className="text-sm text-zinc-400">Upload to object store, then approve assets for public feed.</p>
      </div>
      {err ? <p className="text-sm text-red-400">{err}</p> : null}

      <section className="rounded-xl border border-zinc-800 bg-zinc-900/40 p-4">
        <h2 className="text-sm font-medium text-zinc-300">Upload new media</h2>
        <div className="mt-3 flex flex-wrap items-center gap-3">
          <select
            className="rounded-lg border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm"
            value={section}
            onChange={(e) => setSection(e.target.value)}
          >
            <option value="moments">Moments</option>
            <option value="ceremony">Ceremony</option>
            <option value="family">Family</option>
          </select>
          <input
            type="file"
            className="text-sm"
            disabled={busy}
            onChange={(e) => {
              const f = e.target.files?.[0];
              if (f) void onUploadFile(f);
            }}
          />
        </div>
      </section>

      <section className="rounded-xl border border-zinc-800 bg-zinc-900/40 p-4">
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-sm font-medium text-zinc-300">Moderation queue</h2>
          <select
            className="rounded-lg border border-zinc-700 bg-zinc-950 px-2 py-1 text-xs"
            value={status}
            onChange={(e) => setStatus(e.target.value as "pending" | "approved" | "rejected" | "all")}
          >
            <option value="pending">Pending</option>
            <option value="approved">Approved</option>
            <option value="rejected">Rejected</option>
            <option value="all">All</option>
          </select>
        </div>
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {assets.map((a) => (
            <article key={a.id} className="rounded-lg border border-zinc-700 bg-zinc-950/50 p-3">
              {a.url ? (
                <Image
                  src={a.url}
                  alt={a.object_key}
                  width={640}
                  height={360}
                  className="h-36 w-full rounded-md object-cover"
                  unoptimized
                />
              ) : null}
              <p className="mt-2 truncate text-xs text-zinc-500">{a.object_key}</p>
              <p className="mt-1 text-xs uppercase tracking-wide text-zinc-400">{a.status}</p>
              <div className="mt-3 flex gap-2">
                <button
                  type="button"
                  className="rounded bg-emerald-700 px-2 py-1 text-xs text-white"
                  onClick={() => void moderate(a.id, "approved")}
                >
                  Approve
                </button>
                <button
                  type="button"
                  className="rounded bg-rose-700 px-2 py-1 text-xs text-white"
                  onClick={() => void moderate(a.id, "rejected")}
                >
                  Reject
                </button>
              </div>
            </article>
          ))}
          {assets.length === 0 ? <p className="text-sm text-zinc-500">No assets in this queue.</p> : null}
        </div>
      </section>
    </main>
  );
}
