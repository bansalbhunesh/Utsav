CREATE TABLE IF NOT EXISTS guest_relationship_scores (
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    guest_id UUID NOT NULL REFERENCES guests(id) ON DELETE CASCADE,
    priority_score INT NOT NULL,
    priority_tier TEXT NOT NULL,
    priority_reasons JSONB NOT NULL DEFAULT '[]'::jsonb,
    source_version INT NOT NULL DEFAULT 1,
    computed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (event_id, guest_id)
);

CREATE INDEX IF NOT EXISTS idx_guest_relationship_scores_event_score
    ON guest_relationship_scores(event_id, priority_score DESC);

CREATE INDEX IF NOT EXISTS idx_guest_relationship_scores_event_tier_score
    ON guest_relationship_scores(event_id, priority_tier, priority_score DESC);
