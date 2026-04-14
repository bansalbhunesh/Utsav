DROP INDEX IF EXISTS idx_gallery_assets_event_status_created;

ALTER TABLE gallery_assets
    DROP COLUMN IF EXISTS bytes,
    DROP COLUMN IF EXISTS mime_type,
    DROP COLUMN IF EXISTS status;
