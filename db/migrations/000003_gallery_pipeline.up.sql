ALTER TABLE gallery_assets
    ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS mime_type TEXT,
    ADD COLUMN IF NOT EXISTS bytes BIGINT;

CREATE INDEX IF NOT EXISTS idx_gallery_assets_event_status_created
    ON gallery_assets(event_id, status, created_at DESC);
