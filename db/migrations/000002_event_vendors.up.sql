-- Vendor Quick-Assign (per-event), distinct from organiser_vendors
CREATE TABLE event_vendors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    category TEXT,
    phone TEXT,
    email TEXT,
    advance_paise BIGINT NOT NULL DEFAULT 0,
    total_paise BIGINT,
    balance_paise BIGINT,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE event_vendor_sub_events (
    event_vendor_id UUID NOT NULL REFERENCES event_vendors(id) ON DELETE CASCADE,
    sub_event_id UUID NOT NULL REFERENCES sub_events(id) ON DELETE CASCADE,
    PRIMARY KEY (event_vendor_id, sub_event_id)
);

CREATE INDEX idx_event_vendors_event ON event_vendors(event_id);
