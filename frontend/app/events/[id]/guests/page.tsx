"use client";

import Link from "next/link";
import { usePathname, useParams, useRouter, useSearchParams } from "next/navigation";
import { useState } from "react";
import { apiFetch } from "@/lib/api";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { getUserFacingError } from "@/lib/error-messages";
import {
  parseHostGuestsImportResponse,
  parseHostRelationshipPriorityOverview,
  parseHostGuestsResponse,
} from "@/lib/contracts/host";

const GUEST_CURSOR_STACK_PREFIX = "utsav:guestCursorStack:";

function guestCursorStackKey(eventId: string, sort: string, limit: number) {
  return `${GUEST_CURSOR_STACK_PREFIX}${eventId}:${sort}:${limit}`;
}

function readGuestCursorStack(key: string): string[] {
  if (typeof window === "undefined") return [];
  try {
    const raw = sessionStorage.getItem(key);
    if (!raw) return [];
    const v = JSON.parse(raw) as unknown;
    return Array.isArray(v) ? v.map((x) => String(x)) : [];
  } catch {
    return [];
  }
}

function writeGuestCursorStack(key: string, stack: string[]) {
  if (typeof window === "undefined") return;
  sessionStorage.setItem(key, JSON.stringify(stack));
}

function clearGuestCursorStack(key: string) {
  if (typeof window === "undefined") return;
  sessionStorage.removeItem(key);
}

/** Clear all cursor stacks for an event (sort/limit changes, new data). */
function clearAllGuestCursorStacksForEvent(eventId: string) {
  if (typeof window === "undefined") return;
  const prefix = `${GUEST_CURSOR_STACK_PREFIX}${eventId}:`;
  for (let i = sessionStorage.length - 1; i >= 0; i--) {
    const k = sessionStorage.key(i);
    if (k?.startsWith(prefix)) sessionStorage.removeItem(k);
  }
}

type Guest = {
  id: string;
  name: string;
  phone: string;
  priority_score?: number;
  priority_tier?: string;
  priority_reasons?: string[];
};

export default function GuestsPage() {
  const { id } = useParams();
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const eventId = String(id || "");
  const sort = searchParams.get("sort") || "priority_desc";
  const priorityTier = searchParams.get("priority_tier") || "";
  const limitRaw = Number(searchParams.get("limit") || "20");
  const offsetRaw = Number(searchParams.get("offset") || "0");
  const limit = [20, 50, 100].includes(limitRaw) ? limitRaw : 20;
  const offset = Number.isFinite(offsetRaw) && offsetRaw >= 0 ? offsetRaw : 0;
  const cursorParam = searchParams.get("cursor") ?? "";
  const supportsCursor =
    sort !== "priority_desc" && sort !== "priority_asc";
  const stackKey = guestCursorStackKey(eventId, sort, limit);
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [csv, setCsv] = useState("name,phone\nPriya,+919876543210");
  const [importMsg, setImportMsg] = useState<string | null>(null);
  const [actionErr, setActionErr] = useState<string | null>(null);
  const [copyMsg, setCopyMsg] = useState<string | null>(null);
  const queryClient = useQueryClient();

  function updateQuery(updates: Record<string, string | number | null>) {
    const params = new URLSearchParams(searchParams.toString());
    Object.entries(updates).forEach(([key, value]) => {
      if (value === null || value === "") {
        params.delete(key);
      } else {
        params.set(key, String(value));
      }
    });
    const next = params.toString();
    router.replace(next ? `${pathname}?${next}` : pathname);
  }

  async function copyViewLink() {
    try {
      const query = searchParams.toString();
      const relative = query ? `${pathname}?${query}` : pathname;
      const absolute = typeof window !== "undefined" ? new URL(relative, window.location.origin).toString() : relative;
      await navigator.clipboard.writeText(absolute);
      setCopyMsg("View link copied");
    } catch {
      setCopyMsg("Failed to copy link");
    }
    setTimeout(() => setCopyMsg(null), 1500);
  }

  const { data, error } = useQuery({
    queryKey: supportsCursor
      ? ["event-guests", eventId, sort, priorityTier, limit, cursorParam]
      : ["event-guests", eventId, sort, priorityTier, limit, offset],
    queryFn: async () => {
      const params = new URLSearchParams();
      params.set("sort", sort);
      params.set("limit", String(limit));
      if (supportsCursor && cursorParam) {
        params.set("cursor", cursorParam);
      } else {
        params.set("offset", String(offset));
      }
      if (priorityTier) params.set("priority_tier", priorityTier);
      const raw = await apiFetch<unknown>(
        `/v1/events/${eventId}/guests?${params.toString()}`,
      );
      return parseHostGuestsResponse(raw);
    },
  });
  const { data: relationshipOverview } = useQuery({
    queryKey: ["relationship-priority-overview", eventId],
    queryFn: async () => {
      const raw = await apiFetch<unknown>(
        `/v1/events/${eventId}/intelligence/relationship-priority`,
      );
      return parseHostRelationshipPriorityOverview(raw);
    },
  });
  const guests: Guest[] = data?.guests || [];
  const pageOffset = data?.offset ?? offset;
  const pageLimit = data?.limit ?? limit;
  const stackDepth = supportsCursor ? readGuestCursorStack(stackKey).length : 0;
  const displayStart =
    supportsCursor && cursorParam
      ? stackDepth * pageLimit + 1
      : guests.length === 0
        ? 0
        : pageOffset + 1;
  const displayEnd =
    supportsCursor && cursorParam
      ? stackDepth * pageLimit + guests.length
      : pageOffset + guests.length;
  const hasPrev = supportsCursor
    ? Boolean(cursorParam)
    : pageOffset > 0;
  const hasNext = supportsCursor
    ? Boolean(data?.next_cursor)
    : guests.length >= pageLimit;
  const err = error ? getUserFacingError(error, "Failed to load guests.") : actionErr;

  const tierCard = relationshipOverview?.tier_counts ?? {
    critical: 0,
    important: 0,
    optional: 0,
  };

  function tierClass(tier: string | undefined) {
    switch ((tier || "").toLowerCase()) {
      case "critical":
        return "border-rose-500/40 bg-rose-500/10 text-rose-200";
      case "important":
        return "border-amber-500/40 bg-amber-500/10 text-amber-200";
      case "normal":
        return "border-sky-500/40 bg-sky-500/10 text-sky-200";
      case "optional":
      default:
        return "border-zinc-600/40 bg-zinc-700/20 text-zinc-300";
    }
  }

  const addGuestMutation = useMutation({
    mutationFn: async () =>
      apiFetch(`/v1/events/${eventId}/guests`, {
        method: "POST",
        json: { name, phone },
      }),
    onSuccess: async () => {
      setName("");
      setPhone("");
      clearAllGuestCursorStacksForEvent(eventId);
      await queryClient.invalidateQueries({ queryKey: ["event-guests", eventId] });
    },
  });

  const importGuestsMutation = useMutation({
    mutationFn: async () => {
      const raw = await apiFetch<unknown>(
        `/v1/events/${eventId}/guests/import`,
        { method: "POST", json: { csv } },
      );
      return parseHostGuestsImportResponse(raw);
    },
    onSuccess: async (d) => {
      setImportMsg(`Imported ${d.imported}. Row errors: ${d.errors?.length ?? 0}.`);
      clearAllGuestCursorStacksForEvent(eventId);
      await queryClient.invalidateQueries({ queryKey: ["event-guests", eventId] });
    },
  });

  async function addGuest() {
    setActionErr(null);
    try {
      await addGuestMutation.mutateAsync();
    } catch (e) {
      setActionErr(getUserFacingError(e, "Failed to add guest."));
    }
  }

  async function importCsv() {
    setActionErr(null);
    setImportMsg(null);
    try {
      await importGuestsMutation.mutateAsync();
    } catch (e) {
      setActionErr(getUserFacingError(e, "Failed to import guests CSV."));
    }
  }

  return (
    <main className="mx-auto max-w-3xl space-y-6 px-6 py-12">
      <Link href={`/events/${eventId}`} className="text-sm text-zinc-400">
        ← Event
      </Link>
      <h1 className="text-xl font-semibold text-white">Guests</h1>
      <section className="rounded-xl border border-zinc-800 bg-zinc-900/40 p-4">
        <h2 className="text-sm font-medium text-zinc-400">Relationship Priority Score</h2>
        <div className="mt-3 grid grid-cols-2 gap-2 text-xs sm:grid-cols-4">
          <div className="rounded-lg border border-rose-500/30 bg-rose-500/10 px-3 py-2 text-rose-200">
            Critical: {tierCard.critical}
          </div>
          <div className="rounded-lg border border-amber-500/30 bg-amber-500/10 px-3 py-2 text-amber-200">
            Important: {tierCard.important}
          </div>
          <div className="rounded-lg border border-zinc-600/40 bg-zinc-700/20 px-3 py-2 text-zinc-300">
            Optional: {tierCard.optional}
          </div>
        </div>
        <div className="mt-3 rounded-lg border border-zinc-800 bg-zinc-950/60 p-3">
          <p className="text-xs font-medium text-zinc-400">Top 15 guests to personally call</p>
          <ul className="mt-2 space-y-2 text-sm">
            {(relationshipOverview?.ranked_guests || []).slice(0, 15).map((g, i) => (
              <li key={g.id} className="flex items-center justify-between rounded-md border border-zinc-800 px-2 py-1.5">
                <span className="text-zinc-200">
                  #{i + 1} {g.name}
                </span>
                <span className={`rounded-full border px-2 py-0.5 text-xs ${tierClass(g.priority_tier)}`}>
                  {g.priority_tier || "Low"} · {g.priority_score ?? 0}
                </span>
              </li>
            ))}
          </ul>
          <p className="mt-4 text-xs font-medium text-zinc-400">Guests needing attention</p>
          <ul className="mt-2 space-y-2 text-sm">
            {(relationshipOverview?.guests_needing_attention || []).slice(0, 10).map((g) => (
              <li key={`attention-${g.id}`} className="flex items-center justify-between rounded-md border border-zinc-800 px-2 py-1.5">
                <span className="text-zinc-200">{g.name}</span>
                <span className={`rounded-full border px-2 py-0.5 text-xs ${tierClass(g.priority_tier)}`}>
                  {g.priority_tier || "Optional"} · {g.priority_score ?? 0}
                </span>
              </li>
            ))}
          </ul>
          {relationshipOverview?.coming_next?.length ? (
            <p className="mt-3 text-xs text-zinc-500">
              Coming: {relationshipOverview.coming_next.join(", ")}
            </p>
          ) : null}
          <div className="mt-3 rounded-md border border-zinc-800 p-2 text-xs text-zinc-500">
            <p>Coming next: RSVP Risk Predictor (T-7 nudge, T-3 reminder, T-1 call suggestion).</p>
            <p className="mt-1">Coming next: Shagun Signal Intelligence (median +/- IQR expected range, outlier flags).</p>
          </div>
        </div>
      </section>
      <section className="rounded-xl border border-zinc-800 bg-zinc-900/40 p-4">
        <h2 className="text-sm font-medium text-zinc-400">Priority sorting</h2>
        <div className="mt-2 flex flex-wrap gap-2">
          <select
            value={sort}
            onChange={(e) => {
              clearAllGuestCursorStacksForEvent(eventId);
              updateQuery({ sort: e.target.value, offset: 0, cursor: null });
            }}
            className="rounded-lg border border-zinc-800 bg-zinc-950 px-3 py-2 text-sm text-white"
          >
            <option value="priority_desc">Priority (high to low)</option>
            <option value="priority_asc">Priority (low to high)</option>
            <option value="name_asc">Name (A to Z)</option>
            <option value="name_desc">Name (Z to A)</option>
            <option value="rsvp_desc">RSVP commitment</option>
            <option value="shagun_desc">Shagun signal</option>
          </select>
          <select
            value={priorityTier}
            onChange={(e) => {
              clearAllGuestCursorStacksForEvent(eventId);
              updateQuery({ priority_tier: e.target.value, offset: 0, cursor: null });
            }}
            className="rounded-lg border border-zinc-800 bg-zinc-950 px-3 py-2 text-sm text-white"
          >
            <option value="">All tiers</option>
            <option value="critical">Critical</option>
            <option value="important">Important</option>
            <option value="optional">Optional</option>
          </select>
          <select
            value={limit}
            onChange={(e) => {
              clearAllGuestCursorStacksForEvent(eventId);
              updateQuery({ limit: Number(e.target.value), offset: 0, cursor: null });
            }}
            className="rounded-lg border border-zinc-800 bg-zinc-950 px-3 py-2 text-sm text-white"
          >
            <option value={20}>20 / page</option>
            <option value={50}>50 / page</option>
            <option value={100}>100 / page</option>
          </select>
          <button
            type="button"
            onClick={() => void copyViewLink()}
            className="rounded-lg border border-zinc-700 px-3 py-2 text-sm text-zinc-200 hover:bg-zinc-800"
          >
            Copy view link
          </button>
        </div>
        {copyMsg && <p className="mt-2 text-xs text-emerald-400">{copyMsg}</p>}
      </section>
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
      <div className="flex items-center justify-between text-xs text-zinc-400">
        <p>
          Showing {displayStart}-{displayEnd}
          {supportsCursor && cursorParam ? (
            <span className="text-zinc-500"> · keyset pages</span>
          ) : null}
        </p>
        <div className="flex gap-2">
          <button
            type="button"
            disabled={!hasPrev}
            onClick={() => {
              if (supportsCursor && cursorParam) {
                const st = readGuestCursorStack(stackKey);
                if (st.length === 0) {
                  clearGuestCursorStack(stackKey);
                  updateQuery({ cursor: null, offset: 0 });
                  return;
                }
                const prev = st.pop()!;
                writeGuestCursorStack(stackKey, st);
                updateQuery({ cursor: prev === "" ? null : prev, offset: 0 });
                return;
              }
              updateQuery({ offset: Math.max(0, pageOffset - pageLimit) });
            }}
            className="rounded-lg border border-zinc-700 px-3 py-1 disabled:cursor-not-allowed disabled:opacity-50"
          >
            Previous
          </button>
          <button
            type="button"
            disabled={!hasNext}
            onClick={() => {
              if (supportsCursor && data?.next_cursor) {
                const st = readGuestCursorStack(stackKey);
                st.push(cursorParam);
                writeGuestCursorStack(stackKey, st);
                updateQuery({ cursor: data.next_cursor, offset: null });
                return;
              }
              updateQuery({ offset: pageOffset + pageLimit });
            }}
            className="rounded-lg border border-zinc-700 px-3 py-1 disabled:cursor-not-allowed disabled:opacity-50"
          >
            Next
          </button>
        </div>
      </div>
      <ul className="space-y-2 text-sm">
        {guests.map((g) => (
          <li key={g.id} className="rounded border border-zinc-800 px-3 py-2 text-zinc-200">
            <div className="flex items-center justify-between gap-3">
              <div>
                {g.name} — {g.phone}
              </div>
              <div className={`rounded-full border px-2 py-0.5 text-xs ${tierClass(g.priority_tier)}`}>
                {g.priority_tier || "Optional"} · {g.priority_score ?? 0}
              </div>
            </div>
            {g.priority_reasons && g.priority_reasons.length > 0 && (
              <p className="mt-1 text-xs text-zinc-400">
                {g.priority_reasons.slice(0, 2).join(" | ")}
              </p>
            )}
          </li>
        ))}
      </ul>
    </main>
  );
}
