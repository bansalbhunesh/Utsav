-- Denormalize per-guest RSVP/shagun aggregates and priority_score on guests to avoid
-- per-row LATERAL subqueries on list endpoints. Kept in sync via triggers.

ALTER TABLE guests
  ADD COLUMN IF NOT EXISTS rsvp_yes_count INT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS rsvp_total_count INT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS total_shagun_paise BIGINT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS latest_rsvp_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS priority_score INT NOT NULL DEFAULT 0;

-- Recompute one guest row from source tables (same semantics as guestrepo ListGuests).
CREATE OR REPLACE FUNCTION guest_refresh_denormalized(p_guest_id uuid)
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
  UPDATE guests g
  SET
    rsvp_yes_count = src.rsvp_yes_count,
    rsvp_total_count = src.rsvp_total_count,
    latest_rsvp_at = src.latest_rsvp_at,
    total_shagun_paise = src.total_shagun_paise,
    priority_score = src.priority_score
  FROM (
    SELECT
      g2.id,
      COALESCE(r.rsvp_yes_count, 0)::int AS rsvp_yes_count,
      COALESCE(r.rsvp_total_count, 0)::int AS rsvp_total_count,
      r.latest_rsvp_at AS latest_rsvp_at,
      COALESCE(s.total_shagun_paise, 0)::bigint AS total_shagun_paise,
      LEAST(100, GREATEST(0,
        ROUND(100.0 * (
          (0.30 * prio.rel_w + 0.20 * prio.rc + 0.15 * prio.rs + 0.15 * prio.ec + 0.10 * prio.hr + 0.10 * prio.ho)
          * prio.decay * prio.unc
        ))::double precision
      ))::int AS priority_score
    FROM guests g2
    CROSS JOIN LATERAL (
      SELECT COUNT(*)::int AS sub_event_total
      FROM sub_events WHERE event_id = g2.event_id
    ) evt
    LEFT JOIN LATERAL (
      SELECT
        COUNT(*) FILTER (WHERE rr.status = 'yes') AS rsvp_yes_count,
        COUNT(*)::int AS rsvp_total_count,
        MAX(rr.updated_at) AS latest_rsvp_at
      FROM rsvp_responses rr
      WHERE rr.event_id = g2.event_id AND rr.guest_phone = g2.phone
    ) r ON TRUE
    LEFT JOIN LATERAL (
      SELECT COALESCE(SUM(se.amount_paise), 0) AS total_shagun_paise
      FROM shagun_entries se
      WHERE se.event_id = g2.event_id
        AND (se.guest_id = g2.id OR COALESCE(se.meta->>'guest_phone', '') = g2.phone)
    ) s ON TRUE
    CROSS JOIN LATERAL (
      SELECT
        (CASE lower(trim(COALESCE(g2.relationship, '')))
          WHEN 'close_family' THEN 1.0::float8
          WHEN 'immediate_family' THEN 1.0::float8
          WHEN 'family' THEN 0.85::float8
          WHEN 'relative' THEN 0.85::float8
          WHEN 'relatives' THEN 0.85::float8
          WHEN 'friend' THEN 0.65::float8
          WHEN 'friends' THEN 0.65::float8
          WHEN 'colleague' THEN 0.45::float8
          WHEN 'coworker' THEN 0.45::float8
          ELSE CASE WHEN trim(COALESCE(g2.relationship, '')) = '' THEN 0.2::float8 ELSE 0.35::float8 END
        END) AS rel_w,
        LEAST(1.0::float8, GREATEST(0.0::float8, COALESCE(r.rsvp_yes_count, 0)::float8 / 3.0)) AS rc,
        CASE WHEN r.latest_rsvp_at IS NULL THEN 0.0::float8
          ELSE LEAST(1.0::float8, GREATEST(0.0::float8, 1.0::float8 - ((EXTRACT(EPOCH FROM (now() - r.latest_rsvp_at)) / 86400.0) / 14.0)))
        END AS rs,
        CASE WHEN COALESCE(evt.sub_event_total, 0) = 0 THEN 0.0::float8
          ELSE LEAST(1.0::float8, GREATEST(0.0::float8, COALESCE(r.rsvp_total_count, 0)::float8 / NULLIF(evt.sub_event_total, 0)::float8))
        END AS ec,
        CASE WHEN COALESCE(r.rsvp_total_count, 0) = 0 THEN 0.0::float8
          ELSE LEAST(1.0::float8, GREATEST(0.0::float8, COALESCE(r.rsvp_yes_count, 0)::float8 / NULLIF(r.rsvp_total_count, 0)::float8))
        END AS hr,
        (CASE WHEN EXISTS (
          SELECT 1 FROM unnest(COALESCE(g2.tags, ARRAY[]::text[])) AS t(tag)
          WHERE lower(trim(tag)) IN ('vip', 'priority', 'must_call')
        ) THEN 1.0::float8 ELSE 0.0::float8 END) AS ho,
        CASE WHEN r.latest_rsvp_at IS NULL THEN 0.90::float8
          ELSE 0.75::float8 + 0.25::float8 * exp(-(EXTRACT(EPOCH FROM (now() - r.latest_rsvp_at)) / 86400.0) / 30.0)
        END AS decay,
        GREATEST(0.70::float8, LEAST(1.0::float8, 1.0::float8 - 0.08::float8 * (
          (CASE WHEN COALESCE(r.rsvp_total_count, 0) = 0 THEN 3 ELSE 0 END) +
          (CASE WHEN COALESCE(evt.sub_event_total, 0) = 0 THEN 1 ELSE 0 END)
        )::float8)) AS unc
    ) prio
    WHERE g2.id = p_guest_id
  ) src
  WHERE g.id = p_guest_id AND g.id = src.id;
END;
$$;

-- Initial backfill
DO $$
DECLARE r RECORD;
BEGIN
  FOR r IN SELECT id FROM guests LOOP
    PERFORM guest_refresh_denormalized(r.id);
  END LOOP;
END $$;

CREATE INDEX IF NOT EXISTS idx_guests_event_priority_score_id
  ON guests(event_id, priority_score DESC, id);

CREATE INDEX IF NOT EXISTS idx_guests_event_rsvp_yes_id
  ON guests(event_id, rsvp_yes_count DESC, id);

CREATE INDEX IF NOT EXISTS idx_guests_event_shagun_id
  ON guests(event_id, total_shagun_paise DESC, id);

-- RSVP changes → refresh affected guest row(s)
CREATE OR REPLACE FUNCTION trg_rsvp_responses_refresh_guest_stats()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
  gid uuid;
BEGIN
  IF TG_OP = 'DELETE' THEN
    IF OLD.guest_id IS NOT NULL THEN
      PERFORM guest_refresh_denormalized(OLD.guest_id);
    ELSE
      SELECT id INTO gid FROM guests WHERE event_id = OLD.event_id AND phone = OLD.guest_phone LIMIT 1;
      IF gid IS NOT NULL THEN
        PERFORM guest_refresh_denormalized(gid);
      END IF;
    END IF;
    RETURN OLD;
  END IF;

  IF NEW.guest_id IS NOT NULL THEN
    PERFORM guest_refresh_denormalized(NEW.guest_id);
  ELSE
    SELECT id INTO gid FROM guests WHERE event_id = NEW.event_id AND phone = NEW.guest_phone LIMIT 1;
    IF gid IS NOT NULL THEN
      PERFORM guest_refresh_denormalized(gid);
    END IF;
  END IF;

  IF TG_OP = 'UPDATE' AND (
    OLD.event_id IS DISTINCT FROM NEW.event_id OR
    OLD.guest_phone IS DISTINCT FROM NEW.guest_phone OR
    OLD.guest_id IS DISTINCT FROM NEW.guest_id
  ) THEN
    IF OLD.guest_id IS NOT NULL THEN
      PERFORM guest_refresh_denormalized(OLD.guest_id);
    ELSE
      SELECT id INTO gid FROM guests WHERE event_id = OLD.event_id AND phone = OLD.guest_phone LIMIT 1;
      IF gid IS NOT NULL THEN
        PERFORM guest_refresh_denormalized(gid);
      END IF;
    END IF;
  END IF;

  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_rsvp_responses_refresh_guest_stats ON rsvp_responses;
CREATE TRIGGER trg_rsvp_responses_refresh_guest_stats
  AFTER INSERT OR UPDATE OR DELETE ON rsvp_responses
  FOR EACH ROW EXECUTE PROCEDURE trg_rsvp_responses_refresh_guest_stats();

-- Shagun changes → refresh guest(s) tied to row (old + new on update)
CREATE OR REPLACE FUNCTION trg_shagun_entries_refresh_guest_stats()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
  g_old uuid;
  g_new uuid;
BEGIN
  IF TG_OP = 'DELETE' THEN
    IF OLD.guest_id IS NOT NULL THEN
      PERFORM guest_refresh_denormalized(OLD.guest_id);
    ELSIF OLD.meta->>'guest_phone' IS NOT NULL AND OLD.meta->>'guest_phone' <> '' THEN
      SELECT id INTO g_old FROM guests WHERE event_id = OLD.event_id AND phone = OLD.meta->>'guest_phone' LIMIT 1;
      IF g_old IS NOT NULL THEN
        PERFORM guest_refresh_denormalized(g_old);
      END IF;
    END IF;
    RETURN OLD;
  END IF;

  g_new := NULL;
  IF NEW.guest_id IS NOT NULL THEN
    g_new := NEW.guest_id;
  ELSIF NEW.meta->>'guest_phone' IS NOT NULL AND NEW.meta->>'guest_phone' <> '' THEN
    SELECT id INTO g_new FROM guests WHERE event_id = NEW.event_id AND phone = NEW.meta->>'guest_phone' LIMIT 1;
  END IF;

  IF TG_OP = 'INSERT' THEN
    IF g_new IS NOT NULL THEN
      PERFORM guest_refresh_denormalized(g_new);
    END IF;
    RETURN NEW;
  END IF;

  -- UPDATE
  g_old := NULL;
  IF OLD.guest_id IS NOT NULL THEN
    g_old := OLD.guest_id;
  ELSIF OLD.meta->>'guest_phone' IS NOT NULL AND OLD.meta->>'guest_phone' <> '' THEN
    SELECT id INTO g_old FROM guests WHERE event_id = OLD.event_id AND phone = OLD.meta->>'guest_phone' LIMIT 1;
  END IF;

  IF g_old IS NOT NULL AND g_new IS NOT NULL AND g_old = g_new THEN
    PERFORM guest_refresh_denormalized(g_old);
  ELSE
    IF g_old IS NOT NULL THEN
      PERFORM guest_refresh_denormalized(g_old);
    END IF;
    IF g_new IS NOT NULL THEN
      PERFORM guest_refresh_denormalized(g_new);
    END IF;
  END IF;
  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_shagun_entries_refresh_guest_stats ON shagun_entries;
CREATE TRIGGER trg_shagun_entries_refresh_guest_stats
  AFTER INSERT OR UPDATE OR DELETE ON shagun_entries
  FOR EACH ROW EXECUTE PROCEDURE trg_shagun_entries_refresh_guest_stats();

-- Relationship / tags / side affect priority (formula); counts come from RSVP/shagun triggers
CREATE OR REPLACE FUNCTION trg_guests_update_priority_fields()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  PERFORM guest_refresh_denormalized(NEW.id);
  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_guests_update_priority_fields ON guests;
CREATE TRIGGER trg_guests_update_priority_fields
  AFTER UPDATE OF relationship, side, tags ON guests
  FOR EACH ROW EXECUTE PROCEDURE trg_guests_update_priority_fields();

-- New guest: initial denormalized values + priority
CREATE OR REPLACE FUNCTION trg_guests_insert_denorm()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  PERFORM guest_refresh_denormalized(NEW.id);
  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_guests_insert_denorm ON guests;
CREATE TRIGGER trg_guests_insert_denorm
  AFTER INSERT ON guests
  FOR EACH ROW EXECUTE PROCEDURE trg_guests_insert_denorm();

-- Sub-event count affects priority for every guest in the event
CREATE OR REPLACE FUNCTION trg_sub_events_refresh_guests_priority()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
  eid uuid;
  r RECORD;
BEGIN
  IF TG_OP = 'DELETE' THEN
    eid := OLD.event_id;
  ELSE
    eid := NEW.event_id;
  END IF;

  FOR r IN SELECT id FROM guests WHERE event_id = eid LOOP
    PERFORM guest_refresh_denormalized(r.id);
  END LOOP;

  IF TG_OP = 'DELETE' THEN
    RETURN OLD;
  END IF;
  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_sub_events_refresh_guests_priority ON sub_events;
CREATE TRIGGER trg_sub_events_refresh_guests_priority
  AFTER INSERT OR DELETE ON sub_events
  FOR EACH ROW EXECUTE PROCEDURE trg_sub_events_refresh_guests_priority();
