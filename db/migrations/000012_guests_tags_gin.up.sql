CREATE INDEX IF NOT EXISTS idx_guests_tags_gin ON guests USING GIN (tags);
