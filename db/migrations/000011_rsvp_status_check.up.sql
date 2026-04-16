UPDATE rsvp_responses
SET status = CASE
    WHEN lower(trim(status)) IN ('yes', 'no', 'maybe') THEN lower(trim(status))
    ELSE 'maybe'
END;

ALTER TABLE rsvp_responses
    ADD CONSTRAINT chk_rsvp_status CHECK (status IN ('yes', 'no', 'maybe'));
