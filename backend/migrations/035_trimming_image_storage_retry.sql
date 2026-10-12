ALTER TABLE trimming_image_deletions
    ADD COLUMN storage_deleted_at TIMESTAMPTZ,
    ADD COLUMN last_attempt_at TIMESTAMPTZ;

CREATE INDEX idx_trimming_image_deletions_pending
    ON trimming_image_deletions (clinic_id, last_attempt_at, object_key)
    WHERE storage_deleted_at IS NULL;
