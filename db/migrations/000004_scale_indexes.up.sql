-- Scale-oriented indexes for high-traffic read/write paths
CREATE INDEX IF NOT EXISTS idx_rsvp_event_phone ON rsvp_responses(event_id, guest_phone);
CREATE INDEX IF NOT EXISTS idx_rsvp_event_updated ON rsvp_responses(event_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_phone_otp_phone_created ON phone_otp_challenges(phone, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_rsvp_otp_event_phone_created ON rsvp_otp_challenges(event_id, phone, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_events_owner_created ON events(owner_user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_guests_event_created ON guests(event_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_shagun_event_created ON shagun_entries(event_id, created_at DESC);
