-- EMR-454/455: master editing CAS, independent of source-adoption versions.
-- Apply only through the normal, human-run migration procedure.
-- Do not backfill is_non_insurance or infer historical source adoption.
ALTER TABLE consultations
    ADD COLUMN version bigint NOT NULL DEFAULT 1,
    ADD CONSTRAINT ck_consultations_positive_version CHECK (version > 0);
ALTER TABLE procedures
    ADD COLUMN version bigint NOT NULL DEFAULT 1,
    ADD CONSTRAINT ck_procedures_positive_version CHECK (version > 0);
ALTER TABLE vaccines
    ADD COLUMN version bigint NOT NULL DEFAULT 1,
    ADD CONSTRAINT ck_vaccines_positive_version CHECK (version > 0);

-- The adopted guard advances the version only when stored business values change.
-- Writers compare the editing-start version in WHERE and never SET version.
CREATE TRIGGER consultations_storage_version
    BEFORE UPDATE ON consultations FOR EACH ROW
    EXECUTE FUNCTION app_private.entity_version_guard();
CREATE TRIGGER procedures_storage_version
    BEFORE UPDATE ON procedures FOR EACH ROW
    EXECUTE FUNCTION app_private.entity_version_guard();
CREATE TRIGGER vaccines_storage_version
    BEFORE UPDATE ON vaccines FOR EACH ROW
    EXECUTE FUNCTION app_private.entity_version_guard();
