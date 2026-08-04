-- 008_lab_import_revert_compensation.sql
-- TASK-032: lab import compensating revert の複合 FK・receipt・retraction・RLS。
-- 適用: USER が make migrate を手動実行すること（エージェントは auto-apply しない）。
-- CASCADE DELETE 禁止。既適用 migration / seed は編集しない。
-- ファイル内に BEGIN;/COMMIT; を書かない（cmd/migrate が各ファイルを自前 tx で包む）。
--
-- 前提: 007 が 'reverted' enum 値を別 transaction で追加済みであること。
-- FK-supporting index は soft-deleted child 行を除外しない（RI は物理行を対象にする）。

-- ---------------------------------------------------------------------------
-- 0. fail-closed preflight（複合 FK 追加前）
-- ---------------------------------------------------------------------------
DO $$
BEGIN
    -- orphan exams.job_id
    IF EXISTS (
        SELECT 1
        FROM exams e
        WHERE e.job_id IS NOT NULL
          AND NOT EXISTS (
              SELECT 1 FROM lab_import_jobs j WHERE j.id = e.job_id
          )
    ) THEN
        RAISE EXCEPTION
            '008_lab_import_revert_compensation: exams.job_id orphan rows exist; resolve before migrate';
    END IF;

    -- clinic mismatch exams ↔ jobs
    IF EXISTS (
        SELECT 1
        FROM exams e
        JOIN lab_import_jobs j ON j.id = e.job_id
        WHERE e.job_id IS NOT NULL
          AND e.clinic_id <> j.clinic_id
    ) THEN
        RAISE EXCEPTION
            '008_lab_import_revert_compensation: exams.job_id clinic mismatch with lab_import_jobs; resolve before migrate';
    END IF;

    -- orphan events.job_id
    IF EXISTS (
        SELECT 1
        FROM lab_import_events e
        WHERE NOT EXISTS (
            SELECT 1 FROM lab_import_jobs j WHERE j.id = e.job_id
        )
    ) THEN
        RAISE EXCEPTION
            '008_lab_import_revert_compensation: lab_import_events.job_id orphan rows exist; resolve before migrate';
    END IF;

    -- clinic mismatch events ↔ jobs
    IF EXISTS (
        SELECT 1
        FROM lab_import_events e
        JOIN lab_import_jobs j ON j.id = e.job_id
        WHERE e.clinic_id <> j.clinic_id
    ) THEN
        RAISE EXCEPTION
            '008_lab_import_revert_compensation: lab_import_events.job_id clinic mismatch; resolve before migrate';
    END IF;
END $$;

-- ---------------------------------------------------------------------------
-- 1. candidate keys for composite FKs
-- ---------------------------------------------------------------------------
-- jobs UNIQUE(clinic_id, id) — composite job FK の参照側
CREATE UNIQUE INDEX IF NOT EXISTS uq_lab_import_jobs_clinic_id
    ON lab_import_jobs (clinic_id, id);

-- exams UNIQUE(clinic_id, id, job_id) — exam+job 複合 FK の参照側
-- job_id が NULL の手作成 exam は PostgreSQL UNIQUE で複数許可（NULL は distinct）
CREATE UNIQUE INDEX IF NOT EXISTS uq_exams_clinic_id_job_id
    ON exams (clinic_id, id, job_id);

-- ---------------------------------------------------------------------------
-- 2. Replace exams.job_id single-col SET NULL FK with composite RESTRICT
-- ---------------------------------------------------------------------------
ALTER TABLE exams
    DROP CONSTRAINT IF EXISTS exams_job_id_fkey;

-- FK-X1 supporting index: soft-deleted linked exams を含む（deleted_at 述語なし）
DROP INDEX IF EXISTS idx_exams_clinic_job;
CREATE INDEX idx_exams_clinic_job_fk
    ON exams (clinic_id, job_id)
    WHERE job_id IS NOT NULL;

-- query/active 用（FK 検査用ではない）
CREATE INDEX IF NOT EXISTS idx_exams_clinic_job_active
    ON exams (clinic_id, job_id, id)
    WHERE deleted_at IS NULL AND job_id IS NOT NULL;

ALTER TABLE exams
    ADD CONSTRAINT fk_exams_clinic_job
    FOREIGN KEY (clinic_id, job_id)
    REFERENCES lab_import_jobs (clinic_id, id)
    ON DELETE RESTRICT;

COMMENT ON COLUMN exams.job_id IS
    'lab_import_jobs.id FK — NULL for hand-created exams. Composite (clinic_id, job_id) ON DELETE RESTRICT (TASK-032).';

-- ---------------------------------------------------------------------------
-- 3. events composite FK (FK-E1) + supporting index
-- ---------------------------------------------------------------------------
ALTER TABLE lab_import_events
    DROP CONSTRAINT IF EXISTS lab_import_events_job_id_fkey;

-- FK-E1 supporting: non-partial (events are append-only, no deleted_at)
CREATE INDEX IF NOT EXISTS idx_lab_import_events_clinic_job_created
    ON lab_import_events (clinic_id, job_id, created_at, id);

ALTER TABLE lab_import_events
    ADD CONSTRAINT fk_lab_import_events_clinic_job
    FOREIGN KEY (clinic_id, job_id)
    REFERENCES lab_import_jobs (clinic_id, id)
    ON DELETE RESTRICT;

-- ---------------------------------------------------------------------------
-- 4. lab_import_exam_retractions
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS lab_import_exam_retractions (
    id              BIGSERIAL    PRIMARY KEY,
    clinic_id       BIGINT       NOT NULL,
    job_id          UUID         NOT NULL,
    exam_id         BIGINT       NOT NULL,
    actor_id        BIGINT,
    reason          TEXT         NOT NULL,
    parent_snapshot JSONB        NOT NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uq_lab_import_exam_retractions_clinic_id_job_exam
        UNIQUE (clinic_id, id, job_id, exam_id)
);

-- FK-R1 supporting: leading (clinic_id, job_id)
CREATE INDEX idx_lab_import_exam_retractions_clinic_job
    ON lab_import_exam_retractions (clinic_id, job_id, exam_id, id);

-- FK-R2 supporting: FK column order (clinic_id, exam_id, job_id)
CREATE INDEX idx_lab_import_exam_retractions_clinic_exam_job
    ON lab_import_exam_retractions (clinic_id, exam_id, job_id);

ALTER TABLE lab_import_exam_retractions
    ADD CONSTRAINT fk_lab_import_exam_retractions_clinic_job
    FOREIGN KEY (clinic_id, job_id)
    REFERENCES lab_import_jobs (clinic_id, id)
    ON DELETE RESTRICT;

ALTER TABLE lab_import_exam_retractions
    ADD CONSTRAINT fk_lab_import_exam_retractions_clinic_exam_job
    FOREIGN KEY (clinic_id, exam_id, job_id)
    REFERENCES exams (clinic_id, id, job_id)
    ON DELETE RESTRICT;

COMMENT ON TABLE lab_import_exam_retractions IS
    'TASK-032: immutable parent snapshot when a lab-import exam is soft-deleted by compensating revert. Append-only.';

-- ---------------------------------------------------------------------------
-- 5. lab_import_exam_retraction_items
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS lab_import_exam_retraction_items (
    id              BIGSERIAL    PRIMARY KEY,
    clinic_id       BIGINT       NOT NULL,
    retraction_id   BIGINT       NOT NULL,
    job_id          UUID         NOT NULL,
    exam_id         BIGINT       NOT NULL,
    item_snapshot   JSONB        NOT NULL,
    sort_order      INT          NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- FK-RI1 supporting: full composite FK columns
CREATE INDEX idx_lab_import_exam_retraction_items_fk
    ON lab_import_exam_retraction_items (clinic_id, retraction_id, job_id, exam_id);

CREATE INDEX idx_lab_import_exam_retraction_items_list
    ON lab_import_exam_retraction_items (clinic_id, retraction_id, id);

ALTER TABLE lab_import_exam_retraction_items
    ADD CONSTRAINT fk_lab_import_exam_retraction_items_parent
    FOREIGN KEY (clinic_id, retraction_id, job_id, exam_id)
    REFERENCES lab_import_exam_retractions (clinic_id, id, job_id, exam_id)
    ON DELETE RESTRICT;

COMMENT ON TABLE lab_import_exam_retraction_items IS
    'TASK-032: immutable item snapshots for a lab-import exam retraction. Append-only. Child results are not hard-deleted.';

-- ---------------------------------------------------------------------------
-- 6. lab_import_usage_receipts (append-only downstream use ledger)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS lab_import_usage_receipts (
    id              BIGSERIAL    PRIMARY KEY,
    clinic_id       BIGINT       NOT NULL,
    job_id          UUID         NOT NULL,
    exam_id         BIGINT       NOT NULL,
    use_kind        TEXT         NOT NULL,
    actor_id        BIGINT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- FK-U1 supporting: leading (clinic_id, job_id)
CREATE INDEX idx_lab_import_usage_receipts_clinic_job
    ON lab_import_usage_receipts (clinic_id, job_id, exam_id, created_at, id);

-- FK-U2 supporting: FK column order (clinic_id, exam_id, job_id)
CREATE INDEX idx_lab_import_usage_receipts_clinic_exam_job
    ON lab_import_usage_receipts (clinic_id, exam_id, job_id);

ALTER TABLE lab_import_usage_receipts
    ADD CONSTRAINT fk_lab_import_usage_receipts_clinic_job
    FOREIGN KEY (clinic_id, job_id)
    REFERENCES lab_import_jobs (clinic_id, id)
    ON DELETE RESTRICT;

ALTER TABLE lab_import_usage_receipts
    ADD CONSTRAINT fk_lab_import_usage_receipts_clinic_exam_job
    FOREIGN KEY (clinic_id, exam_id, job_id)
    REFERENCES exams (clinic_id, id, job_id)
    ON DELETE RESTRICT;

COMMENT ON TABLE lab_import_usage_receipts IS
    'TASK-032: append-only clinical downstream-use receipts for import-linked exams. Sink is separate from audit_logs.';

-- ---------------------------------------------------------------------------
-- 7. lab_import_revert_receipts (idempotency)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS lab_import_revert_receipts (
    id                   BIGSERIAL    PRIMARY KEY,
    clinic_id            BIGINT       NOT NULL,
    job_id               UUID         NOT NULL,
    idempotency_key      UUID         NOT NULL,
    request_hash         TEXT         NOT NULL,
    reason               TEXT         NOT NULL,
    actor_id             BIGINT,
    result_status        TEXT         NOT NULL,
    retracted_exam_ids   JSONB        NOT NULL DEFAULT '[]'::jsonb,
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uq_lab_import_revert_receipts_clinic_idempotency
        UNIQUE (clinic_id, idempotency_key)
);

-- FK-V1 supporting
CREATE INDEX idx_lab_import_revert_receipts_clinic_job
    ON lab_import_revert_receipts (clinic_id, job_id, created_at, id);

ALTER TABLE lab_import_revert_receipts
    ADD CONSTRAINT fk_lab_import_revert_receipts_clinic_job
    FOREIGN KEY (clinic_id, job_id)
    REFERENCES lab_import_jobs (clinic_id, id)
    ON DELETE RESTRICT;

COMMENT ON TABLE lab_import_revert_receipts IS
    'TASK-032: append-only idempotent receipts for POST /lab-imports/:id/revert. UNIQUE(clinic_id, idempotency_key).';

-- ---------------------------------------------------------------------------
-- 8. append-only rejection triggers
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION app_private.prevent_lab_import_retraction_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'lab_import_exam_retractions is append-only; UPDATE/DELETE are not allowed'
        USING ERRCODE = 'check_violation';
END;
$$;

DROP TRIGGER IF EXISTS trg_lab_import_exam_retractions_immutable ON lab_import_exam_retractions;
CREATE TRIGGER trg_lab_import_exam_retractions_immutable
    BEFORE UPDATE OR DELETE ON lab_import_exam_retractions
    FOR EACH ROW
    EXECUTE FUNCTION app_private.prevent_lab_import_retraction_mutation();

CREATE OR REPLACE FUNCTION app_private.prevent_lab_import_retraction_item_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'lab_import_exam_retraction_items is append-only; UPDATE/DELETE are not allowed'
        USING ERRCODE = 'check_violation';
END;
$$;

DROP TRIGGER IF EXISTS trg_lab_import_exam_retraction_items_immutable ON lab_import_exam_retraction_items;
CREATE TRIGGER trg_lab_import_exam_retraction_items_immutable
    BEFORE UPDATE OR DELETE ON lab_import_exam_retraction_items
    FOR EACH ROW
    EXECUTE FUNCTION app_private.prevent_lab_import_retraction_item_mutation();

CREATE OR REPLACE FUNCTION app_private.prevent_lab_import_usage_receipt_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'lab_import_usage_receipts is append-only; UPDATE/DELETE are not allowed'
        USING ERRCODE = 'check_violation';
END;
$$;

DROP TRIGGER IF EXISTS trg_lab_import_usage_receipts_immutable ON lab_import_usage_receipts;
CREATE TRIGGER trg_lab_import_usage_receipts_immutable
    BEFORE UPDATE OR DELETE ON lab_import_usage_receipts
    FOR EACH ROW
    EXECUTE FUNCTION app_private.prevent_lab_import_usage_receipt_mutation();

CREATE OR REPLACE FUNCTION app_private.prevent_lab_import_revert_receipt_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'lab_import_revert_receipts is append-only; UPDATE/DELETE are not allowed'
        USING ERRCODE = 'check_violation';
END;
$$;

DROP TRIGGER IF EXISTS trg_lab_import_revert_receipts_immutable ON lab_import_revert_receipts;
CREATE TRIGGER trg_lab_import_revert_receipts_immutable
    BEFORE UPDATE OR DELETE ON lab_import_revert_receipts
    FOR EACH ROW
    EXECUTE FUNCTION app_private.prevent_lab_import_revert_receipt_mutation();

-- ---------------------------------------------------------------------------
-- 9. RLS (ENABLE + clinic predicate; FORCE deferred per project posture)
-- ---------------------------------------------------------------------------
SELECT app_private.apply_rls_policy(
    'lab_import_jobs'::regclass,
    'tenant_clinic_id_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);

SELECT app_private.apply_rls_policy(
    'lab_import_events'::regclass,
    'tenant_clinic_id_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);

SELECT app_private.apply_rls_policy(
    'lab_import_exam_retractions'::regclass,
    'tenant_clinic_id_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);

SELECT app_private.apply_rls_policy(
    'lab_import_exam_retraction_items'::regclass,
    'tenant_clinic_id_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);

SELECT app_private.apply_rls_policy(
    'lab_import_usage_receipts'::regclass,
    'tenant_clinic_id_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);

SELECT app_private.apply_rls_policy(
    'lab_import_revert_receipts'::regclass,
    'tenant_clinic_id_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);
