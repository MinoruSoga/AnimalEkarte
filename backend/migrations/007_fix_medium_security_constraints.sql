-- #113 DB-M2/M3: FK の ON DELETE 規則を明示し、clinic/owner 由来の業務データ削除を RESTRICT に統一する。

ALTER TABLE prescriptions
    DROP CONSTRAINT IF EXISTS prescriptions_clinic_id_fkey,
    ADD CONSTRAINT prescriptions_clinic_id_fkey
        FOREIGN KEY (clinic_id) REFERENCES clinics(id) ON DELETE RESTRICT;

ALTER TABLE prescriptions
    DROP CONSTRAINT IF EXISTS prescriptions_owner_id_fkey,
    ADD CONSTRAINT prescriptions_owner_id_fkey
        FOREIGN KEY (owner_id) REFERENCES owners(id) ON DELETE RESTRICT;

ALTER TABLE prescriptions
    DROP CONSTRAINT IF EXISTS prescriptions_pet_id_fkey,
    ADD CONSTRAINT prescriptions_pet_id_fkey
        FOREIGN KEY (pet_id) REFERENCES pets(id) ON DELETE RESTRICT;

ALTER TABLE prescriptions
    DROP CONSTRAINT IF EXISTS prescriptions_medical_record_id_fkey,
    ADD CONSTRAINT prescriptions_medical_record_id_fkey
        FOREIGN KEY (medical_record_id) REFERENCES medical_records(id) ON DELETE RESTRICT;

ALTER TABLE lstep_tag_cache
    DROP CONSTRAINT IF EXISTS lstep_tag_cache_owner_id_fkey,
    ADD CONSTRAINT lstep_tag_cache_owner_id_fkey
        FOREIGN KEY (owner_id) REFERENCES owners(id) ON DELETE RESTRICT;

ALTER TABLE line_link_tokens
    DROP CONSTRAINT IF EXISTS line_link_tokens_clinic_id_fkey,
    ADD CONSTRAINT line_link_tokens_clinic_id_fkey
        FOREIGN KEY (clinic_id) REFERENCES clinics(id) ON DELETE RESTRICT;

ALTER TABLE line_link_tokens
    DROP CONSTRAINT IF EXISTS line_link_tokens_owner_id_fkey,
    ADD CONSTRAINT line_link_tokens_owner_id_fkey
        FOREIGN KEY (owner_id) REFERENCES owners(id) ON DELETE RESTRICT;

ALTER TABLE lstep_migration_progress
    DROP CONSTRAINT IF EXISTS lstep_migration_progress_clinic_id_fkey,
    ADD CONSTRAINT lstep_migration_progress_clinic_id_fkey
        FOREIGN KEY (clinic_id) REFERENCES clinics(id) ON DELETE RESTRICT;

ALTER TABLE lstep_migration_progress
    DROP CONSTRAINT IF EXISTS lstep_migration_progress_owner_id_fkey,
    ADD CONSTRAINT lstep_migration_progress_owner_id_fkey
        FOREIGN KEY (owner_id) REFERENCES owners(id) ON DELETE RESTRICT;

ALTER TABLE lstep_trigger_priorities
    DROP CONSTRAINT IF EXISTS lstep_trigger_priorities_clinic_id_fkey,
    ADD CONSTRAINT lstep_trigger_priorities_clinic_id_fkey
        FOREIGN KEY (clinic_id) REFERENCES clinics(id) ON DELETE RESTRICT;
