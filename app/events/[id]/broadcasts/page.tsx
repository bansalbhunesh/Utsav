"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";
import { apiFetch } from "@/lib/api";

type SubEvent = { id: string; name: string };
type Broadcast = {
  id: string;
  title: string;
  body: string;
  announcement_type: string;
  audience: string;
  created_at: string;
};

export default function EventBroadcastsPage() {
  const params = useParams();
  const id = String(params.id || "");
  const [err, setErr] = useState<string | null>(null);
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [announcementType, setAnnouncementType] = useState("general");
  const [tagCsv, setTagCsv] = useState("");
  const [sidesCsv, setSidesCsv] = useState("");
  const [onlyRsvpYes, setOnlyRsvpYes] = useState(false);
  const [includeSubIds, setIncludeSubIds] = useState<string[]>([]);
  const [subs, setSubs] = useState<SubEvent[]>([]);
  const [items, setItems] = useState<Broadcast[]>([]);

  const audience = useMemo(() => {
    const tags = tagCsv
      .split(",")
      .map((v) => v.trim())
      .filter(Boolean);
    const sides = sidesCsv
      .split(",")
      .map((v) => v.trim())
      .filter(Boolean);
    return {
      segment: "custom",
      tags_any: tags,
      sides_any: sides,
      only_rsvp_yes: onlyRsvpYes,
      sub_event_ids: includeSubIds,
    };
  }, [tagCsv, sidesCsv, onlyRsvpYes, includeSubIds]);

  const load = useCallback(async () => {
    const [subData, listData] = await Promise.all([
      apiFetch<{ sub_events: SubEvent[] }>(`/v1/events/${id}/sub-events`),
      apiFetch<{ broadcasts: Broadcast[] }>(`/v1/events/${id}/broadcasts`),
    ]);
    setSubs(subData.sub_events || []);
    setItems(listData.broadcasts || []);
  }, [id]);

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

  async function createBroadcast() {
    setErr(null);
    try {
      await apiFetch(`/v1/events/${id}/broadcasts`, {
        method: "POST",
        json: {
          title,
          body,
          announcement_type: announcementType,
          audience,
        },
      });
      setTitle("");
      setBody("");
      await load();
    } catch (e) {
      setErr(String(e));
    }
  }

  return (
    <main className="mx-auto max-w-5xl space-y-6 px-6 py-10 text-zinc-100">
      <Link href={`/events/${id}`} className="text-sm text-zinc-400 hover:text-white">
        ← Back to event
      </Link>

      <header>
        <h1 className="text-2xl font-semibold text-white">Broadcasts</h1>
        <p className="text-sm text-zinc-400">Build audience segments and create announcements.</p>
      </header>
      {err ? <p className="text-sm text-red-400">{err}</p> : null}

      <section className="rounded-xl border border-zinc-800 bg-zinc-900/40 p-4">
        <h2 className="text-sm font-medium text-zinc-300">Create broadcast</h2>
        <div className="mt-3 grid gap-3">
          <input
            className="rounded-lg border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm"
            placeholder="Title"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
          />
          <textarea
            className="min-h-24 rounded-lg border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm"
            placeholder="Message"
            value={body}
            onChange={(e) => setBody(e.target.value)}
          />
          <div className="grid gap-2 sm:grid-cols-3">
            <select
              className="rounded-lg border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm"
              value={announcementType}
              onChange={(e) => setAnnouncementType(e.target.value)}
            >
              <option value="general">General</option>
              <option value="update">Update</option>
              <option value="urgent">Urgent</option>
            </select>
            <input
              className="rounded-lg border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm"
              placeholder="tags_any (csv)"
              value={tagCsv}
              onChange={(e) => setTagCsv(e.target.value)}
            />
            <input
              className="rounded-lg border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm"
              placeholder="sides_any (csv)"
              value={sidesCsv}
              onChange={(e) => setSidesCsv(e.target.value)}
            />
          </div>
          <div className="rounded-lg border border-zinc-700/70 bg-zinc-950/40 p-3">
            <label className="flex items-center gap-2 text-sm text-zinc-300">
              <input
                type="checkbox"
                checked={onlyRsvpYes}
                onChange={(e) => setOnlyRsvpYes(e.target.checked)}
              />
              Only RSVP &quot;yes&quot;
            </label>
            <div className="mt-2 flex flex-wrap gap-2">
              {subs.map((s) => {
                const checked = includeSubIds.includes(s.id);
                return (
                  <label key={s.id} className="rounded border border-zinc-700 px-2 py-1 text-xs text-zinc-300">
                    <input
                      type="checkbox"
                      checked={checked}
                      onChange={(e) => {
                        if (e.target.checked) {
                          setIncludeSubIds((v) => [...v, s.id]);
                        } else {
                          setIncludeSubIds((v) => v.filter((x) => x !== s.id));
                        }
                      }}
                      className="mr-1"
                    />
                    {s.name}
                  </label>
                );
              })}
            </div>
          </div>
          <div className="rounded-lg border border-zinc-700 bg-zinc-950 p-3">
            <p className="mb-2 text-xs text-zinc-400">Audience JSON preview</p>
            <pre className="overflow-auto text-xs text-zinc-200">{JSON.stringify(audience, null, 2)}</pre>
          </div>
          <div>
            <button
              type="button"
              onClick={() => void createBroadcast()}
              className="rounded-lg bg-amber-500 px-4 py-2 text-sm font-semibold text-black"
              disabled={!title.trim() || !body.trim()}
            >
              Create broadcast
            </button>
          </div>
        </div>
      </section>

      <section className="rounded-xl border border-zinc-800 bg-zinc-900/40 p-4">
        <h2 className="text-sm font-medium text-zinc-300">Recent broadcasts</h2>
        <div className="mt-3 space-y-3">
          {items.map((b) => (
            <article key={b.id} className="rounded-lg border border-zinc-700 bg-zinc-950/60 p-3">
              <div className="flex items-center justify-between gap-3">
                <h3 className="text-sm font-semibold text-white">{b.title}</h3>
                <span className="text-xs uppercase text-zinc-500">{b.announcement_type}</span>
              </div>
              <p className="mt-2 text-sm text-zinc-300">{b.body}</p>
              <details className="mt-2">
                <summary className="cursor-pointer text-xs text-zinc-400">Audience JSON</summary>
                <pre className="mt-2 overflow-auto rounded border border-zinc-800 bg-black/40 p-2 text-xs text-zinc-300">
                  {b.audience}
                </pre>
              </details>
            </article>
          ))}
          {items.length === 0 ? <p className="text-sm text-zinc-500">No broadcasts yet.</p> : null}
        </div>
      </section>
    </main>
  );
}
