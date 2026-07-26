ALTER TABLE billing_items
    ADD COLUMN other_reason text,
    ADD COLUMN created_by bigint,
    ADD CONSTRAINT fk_billing_items_created_by
        FOREIGN KEY (created_by) REFERENCES staffs(id) ON DELETE RESTRICT;

CREATE INDEX idx_billing_items_created_by
    ON billing_items (created_by)
    WHERE created_by IS NOT NULL;
