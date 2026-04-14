-- UTSAV core schema (v1 foundation + modules through billing stubs)
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone TEXT NOT NULL UNIQUE,
    display_name TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE phone_otp_challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone TEXT NOT NULL,
    code_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    attempts INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id UUID NOT NULL REFERENCES users(id),
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    event_type TEXT NOT NULL DEFAULT 'wedding',
    couple_name_a TEXT,
    couple_name_b TEXT,
    love_story TEXT,
    cover_image_url TEXT,
    date_start DATE,
    date_end DATE,
    privacy TEXT NOT NULL DEFAULT 'public',
    toggles JSONB NOT NULL DEFAULT '{}',
    branding JSONB NOT NULL DEFAULT '{}',
    host_upi_vpa TEXT,
    tier TEXT NOT NULL DEFAULT 'free',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE event_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    role TEXT NOT NULL,
    invited_phone TEXT,
    invited_email TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(event_id, user_id)
);

CREATE TABLE sub_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    sub_type TEXT,
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    venue_place_id TEXT,
    venue_label TEXT,
    dress_code TEXT,
    description TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE guest_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE guests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    group_id UUID REFERENCES guest_groups(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    phone TEXT NOT NULL,
    email TEXT,
    relationship TEXT,
    side TEXT,
    tags TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(event_id, phone)
);

CREATE INDEX idx_guests_event_phone ON guests(event_id, phone);

CREATE TABLE guest_sub_event_invites (
    guest_id UUID NOT NULL REFERENCES guests(id) ON DELETE CASCADE,
    sub_event_id UUID NOT NULL REFERENCES sub_events(id) ON DELETE CASCADE,
    PRIMARY KEY (guest_id, sub_event_id)
);

CREATE TABLE rsvp_otp_challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    phone TEXT NOT NULL,
    code_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    attempts INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE rsvp_responses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    guest_id UUID REFERENCES guests(id) ON DELETE SET NULL,
    guest_phone TEXT NOT NULL,
    sub_event_id UUID NOT NULL REFERENCES sub_events(id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    meal_pref TEXT,
    dietary TEXT,
    accommodation_needed BOOLEAN NOT NULL DEFAULT false,
    travel_mode TEXT,
    plus_one_names TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(event_id, guest_phone, sub_event_id)
);

CREATE TABLE shagun_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    guest_id UUID REFERENCES guests(id) ON DELETE SET NULL,
    reporter_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    channel TEXT NOT NULL,
    amount_paise BIGINT,
    blessing_note TEXT,
    status TEXT NOT NULL DEFAULT 'guest_reported',
    sub_event_id UUID REFERENCES sub_events(id) ON DELETE SET NULL,
    gift_description TEXT,
    gift_value_paise BIGINT,
    meta JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_shagun_event ON shagun_entries(event_id);

CREATE TABLE gallery_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    section TEXT NOT NULL,
    object_key TEXT NOT NULL,
    uploader_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    sub_event_id UUID REFERENCES sub_events(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE memory_books (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    slug TEXT NOT NULL UNIQUE,
    payload JSONB NOT NULL DEFAULT '{}',
    generated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE broadcasts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    image_url TEXT,
    audience JSONB NOT NULL DEFAULT '{}',
    announcement_type TEXT NOT NULL DEFAULT 'general',
    created_by_user_id UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE organiser_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    company_name TEXT NOT NULL,
    logo_url TEXT,
    description TEXT,
    verified BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE organiser_team_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organiser_id UUID NOT NULL REFERENCES organiser_profiles(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(organiser_id, user_id)
);

CREATE TABLE organiser_clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organiser_id UUID NOT NULL REFERENCES organiser_profiles(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    contact_email TEXT,
    contact_phone TEXT,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE organiser_client_events (
    organiser_client_id UUID NOT NULL REFERENCES organiser_clients(id) ON DELETE CASCADE,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    PRIMARY KEY(organiser_client_id, event_id)
);

CREATE TABLE organiser_vendors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organiser_id UUID NOT NULL REFERENCES organiser_profiles(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    category TEXT,
    phone TEXT,
    email TEXT,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE organiser_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organiser_id UUID NOT NULL REFERENCES organiser_profiles(id) ON DELETE CASCADE,
    event_id UUID REFERENCES events(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    due_at TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'open',
    assignee_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE budget_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    category TEXT NOT NULL,
    label TEXT NOT NULL,
    amount_paise BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE billing_checkouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_id UUID REFERENCES events(id) ON DELETE SET NULL,
    tier TEXT NOT NULL,
    razorpay_order_id TEXT,
    status TEXT NOT NULL DEFAULT 'created',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_events_owner ON events(owner_user_id);
CREATE INDEX idx_event_members_event ON event_members(event_id);
CREATE INDEX idx_sub_events_event ON sub_events(event_id);
