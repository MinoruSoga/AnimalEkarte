ALTER TABLE lstep_next_visit_sync_retries
    ADD COLUMN cleanup_line_user_id VARCHAR(64);

ALTER TABLE lstep_next_visit_sync_retries
    ADD CONSTRAINT chk_lstep_next_visit_sync_retries_cleanup_line_user_id
    CHECK (cleanup_line_user_id IS NULL OR cleanup_line_user_id ~ '^[A-Za-z0-9_-]{1,64}$');
