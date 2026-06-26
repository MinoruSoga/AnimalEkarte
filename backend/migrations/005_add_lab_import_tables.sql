-- 005_add_lab_import_tables.sql
-- Dr.Wan / 外部検査連携: lab_import_jobs + lab_import_events (Phase 0 scaffold)
-- 外部接続・MDB・機器通信は Phase BLOCKED。このマイグレーションはローカル write のみ。
--
-- State machine (lab_import_job_status):
--   received → validated, failed
--   validated → mapped, needs_review, failed
--   mapped → persisted, duplicate, needs_review, failed
--   persisted → (terminal)
--   duplicate → (terminal)
--   needs_review → validated, failed
--   failed → received
--
-- Source types (lab_import_source_type):
--   fixture  : テスト・開発用フィクスチャ入力 (Phase 0 で使用可能)
--   drwan    : Dr.Wan MDB アダプタ (Phase BLOCKED — MDB スキーマ未確認)
--   manual   : 手動 CSV/JSON アップロード (Phase 2+ 予定)

-- ------------------------------------
-- ENUM types
-- ------------------------------------
CREATE TYPE lab_import_job_status AS ENUM (
    'received',
    'validated',
    'mapped',
    'persisted',
    'duplicate',
    'needs_review',
    'failed'
);

CREATE TYPE lab_import_source_type AS ENUM (
    'fixture',
    'drwan',
    'manual'
);

-- ------------------------------------
-- lab_import_jobs
-- ------------------------------------
CREATE TABLE lab_import_jobs (
    id                  uuid            PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id           bigint          NOT NULL REFERENCES clinics(id)     ON DELETE RESTRICT,
    source_type         lab_import_source_type NOT NULL DEFAULT 'fixture',
    source_fingerprint  varchar(255)    NOT NULL DEFAULT '',
    status              lab_import_job_status  NOT NULL DEFAULT 'received',
    row_count           int             NOT NULL DEFAULT 0,
    persisted_count     int             NOT NULL DEFAULT 0,
    duplicate_count     int             NOT NULL DEFAULT 0,
    needs_review_count  int             NOT NULL DEFAULT 0,
    failed_count        int             NOT NULL DEFAULT 0,
    error_code          varchar(50),
    error_message       varchar(1000),
    started_at          timestamptz,
    finished_at         timestamptz,
    created_at          timestamptz     NOT NULL DEFAULT now(),
    updated_at          timestamptz     NOT NULL DEFAULT now()
);

CREATE INDEX idx_lab_import_jobs_clinic_created
    ON lab_import_jobs (clinic_id, created_at DESC);

CREATE INDEX idx_lab_import_jobs_clinic_status
    ON lab_import_jobs (clinic_id, status);

COMMENT ON TABLE lab_import_jobs IS 'Dr.Wan / 外部検査連携インポートジョブ状態管理 (Phase 0 scaffold)';
COMMENT ON COLUMN lab_import_jobs.source_fingerprint IS '入力バッチの冪等キー (ハッシュ等)。raw 接続文字列や認証情報は格納しない';
COMMENT ON COLUMN lab_import_jobs.error_code IS 'lab_error_taxonomy のコード (source_unavailable 等)。スタックトレース不可';
COMMENT ON COLUMN lab_import_jobs.error_message IS '安全なエラーメッセージのみ。生デバイスペイロード・PHI 不可';

-- ------------------------------------
-- lab_import_events (監査ログ)
-- ------------------------------------
CREATE TABLE lab_import_events (
    id                  bigserial       PRIMARY KEY,
    clinic_id           bigint          NOT NULL REFERENCES clinics(id)         ON DELETE RESTRICT,
    job_id              uuid            NOT NULL REFERENCES lab_import_jobs(id)  ON DELETE RESTRICT,
    event_type          varchar(50)     NOT NULL,
    from_status         lab_import_job_status,
    to_status           lab_import_job_status,
    row_count           int             NOT NULL DEFAULT 0,
    persisted_count     int             NOT NULL DEFAULT 0,
    duplicate_count     int             NOT NULL DEFAULT 0,
    needs_review_count  int             NOT NULL DEFAULT 0,
    error_code          varchar(50),
    created_at          timestamptz     NOT NULL DEFAULT now()
);

CREATE INDEX idx_lab_import_events_job
    ON lab_import_events (job_id, created_at ASC);

CREATE INDEX idx_lab_import_events_clinic_created
    ON lab_import_events (clinic_id, created_at DESC);

COMMENT ON TABLE lab_import_events IS '検査インポートジョブ監査イベント。PHI・raw デバイスペイロード・接続情報不可';
COMMENT ON COLUMN lab_import_events.event_type IS 'status_transition | validation_result | mapping_result | persistence_result | retry_requested';
COMMENT ON COLUMN lab_import_events.error_code IS 'lab_error_taxonomy のコードのみ。スタックトレース不可';
