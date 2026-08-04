-- 006_checkup_package_import.sql
-- TASK-374 / #211 / DEC-59: versioned clinic-scoped checkup package import
-- Additive only. Do not edit applied migrations. Agent does not apply this file.

BEGIN;

-- 1) Import stable keys on checkup_types
ALTER TABLE checkup_types
    ADD COLUMN IF NOT EXISTS import_namespace text,
    ADD COLUMN IF NOT EXISTS import_key text;

CREATE UNIQUE INDEX IF NOT EXISTS uq_checkup_types_clinic_import_key
    ON checkup_types (clinic_id, import_namespace, import_key)
    WHERE import_namespace IS NOT NULL
      AND import_key IS NOT NULL
      AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_checkup_types_clinic_import_namespace
    ON checkup_types (clinic_id, import_namespace)
    WHERE import_namespace IS NOT NULL
      AND deleted_at IS NULL;

-- 2) Import stable keys on checkup_type_fields
ALTER TABLE checkup_type_fields
    ADD COLUMN IF NOT EXISTS import_namespace text,
    ADD COLUMN IF NOT EXISTS import_key text;

CREATE UNIQUE INDEX IF NOT EXISTS uq_checkup_type_fields_clinic_import_key
    ON checkup_type_fields (clinic_id, import_namespace, import_key)
    WHERE import_namespace IS NOT NULL
      AND import_key IS NOT NULL
      AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_checkup_type_fields_clinic_import_namespace
    ON checkup_type_fields (clinic_id, import_namespace)
    WHERE import_namespace IS NOT NULL
      AND deleted_at IS NULL;

-- 3) Replace single-column parent FK with clinic-composite FK
-- Drop legacy parent_id FK if present (name from 001_init).
DO $$
DECLARE
    fk_name text;
BEGIN
    SELECT con.conname INTO fk_name
    FROM pg_constraint con
    JOIN pg_class rel ON rel.oid = con.conrelid
    JOIN pg_namespace nsp ON nsp.oid = rel.relnamespace
    WHERE nsp.nspname = 'public'
      AND rel.relname = 'checkup_types'
      AND con.contype = 'f'
      AND pg_get_constraintdef(con.oid) ILIKE '%parent_id%'
      AND pg_get_constraintdef(con.oid) NOT ILIKE '%clinic_id%';
    IF fk_name IS NOT NULL THEN
        EXECUTE format('ALTER TABLE checkup_types DROP CONSTRAINT %I', fk_name);
    END IF;
END $$;

ALTER TABLE checkup_types
    ADD CONSTRAINT fk_checkup_types_parent_clinic
    FOREIGN KEY (parent_id, clinic_id)
    REFERENCES checkup_types (id, clinic_id)
    ON DELETE SET NULL (parent_id);

DROP INDEX IF EXISTS idx_checkup_types_parent_id;
CREATE INDEX IF NOT EXISTS idx_checkup_types_clinic_parent
    ON checkup_types (clinic_id, parent_id)
    WHERE parent_id IS NOT NULL;

-- 4) Clinic-scoped import provenance / receipt (internal sink)
CREATE TABLE IF NOT EXISTS checkup_package_import_receipts (
    id              bigserial PRIMARY KEY,
    clinic_id       bigint NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    namespace       text NOT NULL,
    version         text NOT NULL,
    content_digest  text NOT NULL,
    status          text NOT NULL CHECK (status IN ('applied', 'noop', 'conflict', 'failed')),
    actor_id        bigint NOT NULL REFERENCES staffs(id) ON DELETE RESTRICT,
    types_created   integer NOT NULL DEFAULT 0 CHECK (types_created >= 0),
    fields_created  integer NOT NULL DEFAULT 0 CHECK (fields_created >= 0),
    resource_mapping jsonb NOT NULL DEFAULT '{}'::jsonb,
    clinical_approval_ref text NOT NULL DEFAULT '',
    created_at      timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_checkup_package_import_receipts_clinic_ns_ver
        UNIQUE (clinic_id, namespace, version)
);

CREATE INDEX IF NOT EXISTS idx_checkup_package_import_receipts_clinic_created
    ON checkup_package_import_receipts (clinic_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_checkup_package_import_receipts_clinic_digest
    ON checkup_package_import_receipts (clinic_id, content_digest);

SELECT app_private.apply_rls_policy(
    'checkup_package_import_receipts',
    'tenant_checkup_package_import_receipts_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);

COMMIT;
