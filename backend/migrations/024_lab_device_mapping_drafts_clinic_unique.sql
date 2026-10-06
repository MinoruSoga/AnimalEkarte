ALTER TABLE lab_device_mapping_draft_items
    DROP CONSTRAINT fk_lab_device_mapping_draft_items_draft_clinic;

ALTER TABLE lab_device_mapping_specs
    DROP CONSTRAINT fk_lab_device_mapping_specs_draft_clinic;

ALTER TABLE lab_device_mapping_drafts
    DROP CONSTRAINT uq_lab_device_mapping_drafts_id_clinic;

ALTER TABLE lab_device_mapping_drafts
    ADD CONSTRAINT uq_lab_device_mapping_drafts_id_clinic
    UNIQUE (clinic_id, id);

ALTER TABLE lab_device_mapping_draft_items
    ADD CONSTRAINT fk_lab_device_mapping_draft_items_draft_clinic
    FOREIGN KEY (clinic_id, draft_id)
    REFERENCES lab_device_mapping_drafts (clinic_id, id)
    ON DELETE RESTRICT;

ALTER TABLE lab_device_mapping_specs
    ADD CONSTRAINT fk_lab_device_mapping_specs_draft_clinic
    FOREIGN KEY (clinic_id, confirmed_draft_id)
    REFERENCES lab_device_mapping_drafts (clinic_id, id)
    ON DELETE RESTRICT;
