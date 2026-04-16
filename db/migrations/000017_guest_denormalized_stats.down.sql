DROP TRIGGER IF EXISTS trg_sub_events_refresh_guests_priority ON sub_events;
DROP TRIGGER IF EXISTS trg_guests_insert_denorm ON guests;
DROP TRIGGER IF EXISTS trg_guests_update_priority_fields ON guests;
DROP TRIGGER IF EXISTS trg_shagun_entries_refresh_guest_stats ON shagun_entries;
DROP TRIGGER IF EXISTS trg_rsvp_responses_refresh_guest_stats ON rsvp_responses;

DROP FUNCTION IF EXISTS trg_sub_events_refresh_guests_priority();
DROP FUNCTION IF EXISTS trg_guests_insert_denorm();
DROP FUNCTION IF EXISTS trg_guests_update_priority_fields();
DROP FUNCTION IF EXISTS trg_shagun_entries_refresh_guest_stats();
DROP FUNCTION IF EXISTS trg_rsvp_responses_refresh_guest_stats();
DROP FUNCTION IF EXISTS guest_refresh_denormalized(uuid);

DROP INDEX IF EXISTS idx_guests_event_shagun_id;
DROP INDEX IF EXISTS idx_guests_event_rsvp_yes_id;
DROP INDEX IF EXISTS idx_guests_event_priority_score_id;

ALTER TABLE guests
  DROP COLUMN IF EXISTS priority_score,
  DROP COLUMN IF EXISTS latest_rsvp_at,
  DROP COLUMN IF EXISTS total_shagun_paise,
  DROP COLUMN IF EXISTS rsvp_total_count,
  DROP COLUMN IF EXISTS rsvp_yes_count;
