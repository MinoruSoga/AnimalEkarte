-- EMR-446/466 (A01): whole-bill discount exclusion flag on the 7 master item
-- tables, plus storage-managed CAS versions on the 3 tables that lacked them.
-- Apply only through the normal, human-run migration procedure.
-- Does not backfill is_bill_discount_excluded and installs no DB default:
-- existing rows and rows inserted without the column keep raw NULL
-- (contract A.2/A.3; NULL is preserved, not rewritten to false).

ALTER TABLE consultations
    ADD COLUMN is_bill_discount_excluded boolean;
COMMENT ON COLUMN consultations.is_bill_discount_excluded IS 'A01 whole-bill discount exclusion: NULL=not explicitly saved (applied as false, raw NULL kept), true=excluded, false=included. Independent of insurance, tax and line discounts.';
ALTER TABLE exam_types
    ADD COLUMN is_bill_discount_excluded boolean;
COMMENT ON COLUMN exam_types.is_bill_discount_excluded IS 'A01 whole-bill discount exclusion: NULL=not explicitly saved (applied as false, raw NULL kept), true=excluded, false=included. Independent of insurance, tax and line discounts.';
ALTER TABLE procedures
    ADD COLUMN is_bill_discount_excluded boolean;
COMMENT ON COLUMN procedures.is_bill_discount_excluded IS 'A01 whole-bill discount exclusion: NULL=not explicitly saved (applied as false, raw NULL kept), true=excluded, false=included. Independent of insurance, tax and line discounts.';
ALTER TABLE vaccines
    ADD COLUMN is_bill_discount_excluded boolean;
COMMENT ON COLUMN vaccines.is_bill_discount_excluded IS 'A01 whole-bill discount exclusion: NULL=not explicitly saved (applied as false, raw NULL kept), true=excluded, false=included. Independent of insurance, tax and line discounts.';
ALTER TABLE checkup_types
    ADD COLUMN is_bill_discount_excluded boolean;
COMMENT ON COLUMN checkup_types.is_bill_discount_excluded IS 'A01 whole-bill discount exclusion: NULL=not explicitly saved (applied as false, raw NULL kept), true=excluded, false=included. Independent of insurance, tax and line discounts.';
ALTER TABLE medicines
    ADD COLUMN is_bill_discount_excluded boolean;
COMMENT ON COLUMN medicines.is_bill_discount_excluded IS 'A01 whole-bill discount exclusion: NULL=not explicitly saved (applied as false, raw NULL kept), true=excluded, false=included. Independent of insurance, tax and line discounts.';
ALTER TABLE merchandise_items
    ADD COLUMN is_bill_discount_excluded boolean;
COMMENT ON COLUMN merchandise_items.is_bill_discount_excluded IS 'A01 whole-bill discount exclusion: NULL=not explicitly saved (applied as false, raw NULL kept), true=excluded, false=included. Independent of insurance, tax and line discounts.';

-- Storage-managed CAS versions on the 3 tables that lacked them, same shape
-- as 004_master_insurance_versions.sql (consultations/procedures/vaccines)
-- and the existing checkup_types trigger in 001. Initial 1 is the CAS origin
-- for future edits, not a restoration of past versions. The guard advances
-- the version only when stored business values change; writers compare the
-- editing-start version in WHERE and never SET version.
ALTER TABLE exam_types
    ADD COLUMN version bigint NOT NULL DEFAULT 1,
    ADD CONSTRAINT ck_exam_types_positive_version CHECK (version > 0);
ALTER TABLE medicines
    ADD COLUMN version bigint NOT NULL DEFAULT 1,
    ADD CONSTRAINT ck_medicines_positive_version CHECK (version > 0);
ALTER TABLE merchandise_items
    ADD COLUMN version bigint NOT NULL DEFAULT 1,
    ADD CONSTRAINT ck_merchandise_items_positive_version CHECK (version > 0);

CREATE TRIGGER exam_types_storage_version
    BEFORE UPDATE ON exam_types FOR EACH ROW
    EXECUTE FUNCTION app_private.entity_version_guard();
CREATE TRIGGER medicines_storage_version
    BEFORE UPDATE ON medicines FOR EACH ROW
    EXECUTE FUNCTION app_private.entity_version_guard();
CREATE TRIGGER merchandise_items_storage_version
    BEFORE UPDATE ON merchandise_items FOR EACH ROW
    EXECUTE FUNCTION app_private.entity_version_guard();
