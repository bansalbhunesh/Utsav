/** Guest-facing URLs and WhatsApp share text for digital invites. */

export function buildGuestEventUrl(origin: string, slug: string): string {
  const o = origin.replace(/\/$/, "");
  return `${o}/e/${encodeURI(slug)}`;
}

export function buildGuestInviteUrl(origin: string, slug: string): string {
  const o = origin.replace(/\/$/, "");
  return `${o}/e/${encodeURI(slug)}/invite`;
}

function nz(v: unknown): string | null {
  if (v == null) return null;
  const s = String(v).trim();
  return s.length ? s : null;
}

export function formatCoupleLine(ev: {
  couple_name_a?: unknown;
  couple_name_b?: unknown;
  title?: unknown;
}): string {
  const a = nz(ev.couple_name_a);
  const b = nz(ev.couple_name_b);
  if (a && b) return `${a} & ${b}`;
  if (a) return a;
  if (b) return b;
  return nz(ev.title) || "Our celebration";
}

export function formatDateRange(ds: unknown, de: unknown): string | null {
  const fmt = (x: unknown) => {
    if (!x) return null;
    const d = new Date(String(x));
    if (Number.isNaN(d.getTime())) return null;
    return d.toLocaleDateString(undefined, { day: "numeric", month: "short", year: "numeric" });
  };
  const s = fmt(ds);
  const e = fmt(de);
  if (s && e && s !== e) return `${s} – ${e}`;
  return s || e;
}

export function buildWhatsAppInviteMessage(opts: {
  ev: Record<string, unknown>;
  inviteUrl: string;
}): string {
  const names = formatCoupleLine(opts.ev);
  const dr = formatDateRange(opts.ev.date_start, opts.ev.date_end);
  const lines = ["You're invited!", "", names];
  if (dr) lines.push(dr);
  lines.push("", "Open our invite:", opts.inviteUrl);
  return lines.join("\n");
}

export function whatsappSendUrl(text: string): string {
  return `https://wa.me/?text=${encodeURIComponent(text)}`;
}
