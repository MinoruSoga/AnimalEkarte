ALTER TABLE billing_items
    ADD COLUMN vaccination_id bigint,
    ADD COLUMN clinic_id bigint,
    ADD CONSTRAINT chk_billing_items_vaccination_clinic_pair
        CHECK (
            (vaccination_id IS NULL AND clinic_id IS NULL)
            OR (vaccination_id IS NOT NULL AND clinic_id IS NOT NULL)
        );

ALTER TABLE vaccinations
    ADD CONSTRAINT uq_vaccinations_id_clinic UNIQUE (id, clinic_id);

ALTER TABLE billings
    ADD CONSTRAINT uq_billings_id_clinic UNIQUE (id, clinic_id);

ALTER TABLE billing_items
    ADD CONSTRAINT fk_billing_items_billing_clinic
        FOREIGN KEY (billing_id, clinic_id)
        REFERENCES billings (id, clinic_id),
    ADD CONSTRAINT fk_billing_items_vaccination_clinic
        FOREIGN KEY (vaccination_id, clinic_id)
        REFERENCES vaccinations (id, clinic_id)
        ON DELETE RESTRICT;

CREATE INDEX idx_vaccinations_clinic_pet_date_active
    ON vaccinations(clinic_id, pet_id, date, id)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uq_billing_items_vaccination_lifetime
    ON billing_items(vaccination_id)
    WHERE vaccination_id IS NOT NULL;

COMMENT ON COLUMN billing_items.vaccination_id IS
    '予防接種イベント由来の会計明細を識別するprovenance';

COMMENT ON COLUMN billing_items.clinic_id IS
    '予防接種provenanceがある明細だけに保持する内部tenant scope';
