-- =============================================================================
-- Animal Ekarte - 統合スキーマ定義 v23.0 (consolidated)
-- PostgreSQL 18
-- テーブル数: 108 (旧 001–021 + mig-005〜mig-013 + 取扱説明書テーブル + #81キャンペーンテーブル + 新 005-007 + 増分 005-012 を統合)
-- 統合内容:
--   002: マスタシードデータ
--   003: デモシードデータ
--   004: ステージングシードデータ
--   005: clinic_integrations テーブル
--   006: shared_files テーブル
--   006b: token_blacklist テーブル
--   007: owners.line_user_id カラム + インデックス
--   008: lstep_tag_cache テーブル
--   009: pets.deceased_at/deceased_reason, owners.lstep_opt_out* カラム
--   010: medical_records.next_visit_recommended_date カラム
--   011: prescriptions テーブル
--   012: pet_chronic_conditions テーブル
--   013: line_send_logs テーブル
--   007a: payment_splits.clinic_id FK
--   014: reservation_status ENUM に no_show 追加
--   015: owners.line_followed_at/line_blocked_at カラム
--   016: line_link_tokens テーブル
--   017: lstep_migration_progress テーブル
-- --- 外部マイグレーション統合 (ext-005〜ext-021) ---
--   ext-005: audit_logs.metadata カラム
--   ext-006: hospitalizations.insurance_* カラム
--   ext-007: lstep_settings テーブル
--   ext-008: lstep_sync_error_counters テーブル
--   ext-009: owners.line_id_confirmed_at/delivery_excluded/is_transferred 等カラム
--   ext-010: lstep_tag_code_mappings テーブル
--   ext-011: lstep_tag_code_mappings デフォルトシード → 002 へ
--   ext-012: lstep_delivery_trigger_log テーブル
--   ext-013: lstep_tag_cache.reason カラム
--   ext-014: owners.delivery_caution/* カラム
--   ext-015: medical_records.recommendation_reason カラム
--   ext-016: appointments.reservation_route/actual_reservation_at カラム
--   ext-017: lstep_csv_imports テーブル
--   ext-018: lstep_friend_attribute_snapshots テーブル
--   ext-019: permission_group_rules (lstep-csv-import/analytics) シード → 003 へ (group_id FK は 003 生成)
--   ext-021: medical_record_addenda テーブル
--   mig-005: clinic_settings.cpm_version カラム
--   mig-006: clinic_settings.dormant_prevention_* カラム
--   mig-007: owners.line_id_confirmed_by カラム + インデックス
--   mig-008: lstep_trigger_priorities テーブル
--   mig-009: lstep_delivery_trigger_log.suppressed_by_priority / suppression_reason カラム + インデックス
--   mig-010: lstep_auto_managed_prefixes / lstep_condition_tag_mappings / lstep_send_purpose_tag_prefixes テーブル (旧 006) + seed → 002 へ
--   mig-011: clinic_settings.cpm_v2_*_threshold カラム (旧 007)
--   mig-012: clinic_settings.cpm_v1_* カラム (旧 008)
--   mig-013: clinic_settings.health_prevention_lookback_days / vaccine_deadline_days (旧 009)
-- --- 後続マイグレーション統合 (新 005–007) ---
--   新 005: RLS ポリシー (app_private スキーマ) + FK 強化 (billing_refunds/audit_logs/prescriptions/staffs/vital_records 等)
--           ※ ALTER TABLE / ADD COLUMN は 001 の CREATE TABLE に統合済み
--   新 006: 冗長インデックス削除 (idx_vital_records_deleted_at / idx_billing_confirmations_status)
--   新 007: グローバル一意制約削除 (idx_shift_entries_staff_date)
-- --- 増分マイグレーション統合 (旧 005〜012 / 2026-07-04, 本ファイル末尾セクション7に原文を番号順追記) ---
--   005: lab_import_jobs / lab_import_events テーブル (Dr.Wan / 外部検査連携 Phase 0 scaffold)
--   006: idx_exam_results_exam_id インデックス
--   007: idx_exams_clinic_exam_type_date インデックス
--   008: exams.job_id カラム + idx_exams_clinic_job インデックス
--   009: medicine_dose_params テーブル + medicines/treatments 計算パラメータカラム (#201)
--   010: checkup_type_fields / checkup_field_results テーブル (#211。歯科検診暫定 seed は 003_seed_demo.sql へ)
--   011: clinic_settings.closing_am_start カラム (#215)
--   012: checkup_field_results の (checkup_type_field_id, clinic_id) 複合FK (BE-refactor R3-7/D13)
--   013: checkup_type_fields → checkup_types の (checkup_type_id, clinic_id) 複合FK (#211 A6・旧 002_checkup_field_clinic_composite_fk.sql)
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. 拡張機能
-- -----------------------------------------------------------------------------
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- -----------------------------------------------------------------------------
-- 2. ENUM型定義
-- -----------------------------------------------------------------------------

-- ペット関連
CREATE TYPE pet_status AS ENUM ('alive', 'deceased');
CREATE TYPE pet_gender AS ENUM ('male', 'female', 'unknown');
CREATE TYPE acquisition_type AS ENUM ('purchased', 'transferred', 'rescued', 'other');
CREATE TYPE danger_level AS ENUM ('low', 'medium', 'high');
CREATE TYPE membership_type AS ENUM ('non_member', 'member', 'deceased', 'transferred');

-- マスタ共通
CREATE TYPE inventory_category AS ENUM ('medicine', 'consumable', 'food', 'other');
CREATE TYPE inventory_status AS ENUM ('sufficient', 'low', 'out_of_stock');
CREATE TYPE dosage_form AS ENUM ('tablet', 'liquid', 'injection', 'topical', 'powder');
CREATE TYPE medicine_unit AS ENUM ('per_tablet', 'per_ml', 'per_dose', 'per_gram');
CREATE TYPE cage_type AS ENUM ('icu', 'dog', 'cat', 'general');
CREATE TYPE cage_size AS ENUM ('small', 'medium', 'large');
CREATE TYPE body_size AS ENUM ('small', 'medium', 'large');
CREATE TYPE billing_unit AS ENUM ('per_day', 'per_night');
CREATE TYPE target_size AS ENUM ('small', 'medium', 'large', 'cat');
CREATE TYPE anesthesia_type AS ENUM ('none', 'local', 'sedation', 'general');
CREATE TYPE vaccine_species AS ENUM ('dog', 'cat', 'both');

-- 電子カルテ関連
CREATE TYPE medical_record_status AS ENUM ('draft', 'finalized');
CREATE TYPE treatment_item_type AS ENUM ('consultation', 'procedure', 'medicine', 'other');
CREATE TYPE treatment_status AS ENUM ('pending', 'completed', 'not_applicable');
CREATE TYPE exam_status AS ENUM ('pending', 'in_progress', 'result_entered', 'completed', 'confirmed');
CREATE TYPE exam_result_status AS ENUM ('normal', 'high', 'low');
CREATE TYPE next_schedule_type AS ENUM ('3weeks', '4weeks', '1year', 'other');
CREATE TYPE appetite_level AS ENUM ('normal', 'increased', 'decreased', 'none');
CREATE TYPE water_intake_level AS ENUM ('normal', 'increased', 'decreased', 'none');
CREATE TYPE medical_image_type AS ENUM ('xray', 'echo', 'photo', 'endoscope', 'ct', 'mri', 'microscope', 'other');
CREATE TYPE estimate_status AS ENUM ('draft', 'sent', 'approved', 'rejected');
CREATE TYPE confirmation_status AS ENUM ('pending', 'confirmed', 'returned');
CREATE TYPE item_category AS ENUM ('examination', 'test', 'procedure', 'surgery', 'medicine', 'food', 'goods', 'other', 'vaccine', 'trimming', 'hotel', 'training');
CREATE TYPE item_source AS ENUM ('medical_record', 'manual', 'hospitalization', 'trimming');
CREATE TYPE campaign_discount_type AS ENUM ('rate', 'amount');

-- 予約・会計・入院関連
CREATE TYPE visit_type AS ENUM ('first', 'revisit');
CREATE TYPE reservation_status AS ENUM (
    'confirmed', 'pending', 'cancelled', 'checked_in',
    'in_consultation', 'accounting', 'completed', 'no_show'
);
CREATE TYPE staff_type AS ENUM ('doctor', 'nurse', 'trimmer', 'resource');
CREATE TYPE reservation_source AS ENUM ('manual', 'line');
CREATE TYPE billing_status AS ENUM ('waiting', 'completed', 'cancelled', 'pending');
CREATE TYPE hospitalization_type AS ENUM ('hospitalization', 'hotel');
CREATE TYPE hospitalization_status AS ENUM ('admitted', 'discharged', 'reserved');
CREATE TYPE care_plan_type AS ENUM ('food', 'medicine', 'treatment', 'instruction', 'item');
CREATE TYPE care_plan_status AS ENUM ('active', 'completed', 'discontinued');
CREATE TYPE care_log_type AS ENUM ('food', 'excretion', 'medicine', 'treatment', 'other');
CREATE TYPE care_log_status AS ENUM ('completed', 'partial', 'skipped');
CREATE TYPE plan_timing AS ENUM ('morning', 'noon', 'night');
CREATE TYPE body_weight_unit AS ENUM ('Kg', 'g');

-- トリミング・シフト関連
CREATE TYPE reservation_type_category AS ENUM ('general', 'trimming');
CREATE TYPE payment_method AS ENUM ('cash', 'credit_card', 'electronic_money', 'bank_transfer');
CREATE TYPE shift_type AS ENUM ('full', 'morning', 'afternoon', 'off', 'paid_leave');
CREATE TYPE tax_type AS ENUM ('included', 'excluded', 'exempt'); -- 内税, 外税, 非課税

-- -----------------------------------------------------------------------------
-- 3. テーブル定義（依存関係順）
-- -----------------------------------------------------------------------------

-- ==========================================================================
-- レイヤー1: 依存なし
-- ==========================================================================

-- ------------------------------------
-- 1. companies（シングルトン: 本部情報）
-- ------------------------------------
CREATE TABLE companies (
    id                  BIGSERIAL   PRIMARY KEY,
    name                text        NOT NULL,
    postal_code         text        NOT NULL DEFAULT '',
    address             text        NOT NULL DEFAULT '',
    phone_number        text        NOT NULL DEFAULT '',
    fax_number          text        NOT NULL DEFAULT '',
    registration_number text        NOT NULL DEFAULT '',
    invoice_registration_number text NOT NULL DEFAULT '',
    director_name       text        NOT NULL DEFAULT '',
    email               text        NOT NULL DEFAULT '',
    website             text        NOT NULL DEFAULT '',
    logo_url            text        NOT NULL DEFAULT '',
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);

-- ------------------------------------
-- 2. clinics（クリニック情報）
-- ------------------------------------
CREATE TABLE clinics (
    id                  BIGSERIAL   PRIMARY KEY,
    company_id          bigint      NOT NULL REFERENCES companies(id),
    name                text        NOT NULL,
    postal_code         text        NOT NULL DEFAULT '',
    address             text        NOT NULL DEFAULT '',
    phone_number        text        NOT NULL DEFAULT '',
    fax_number          text        NOT NULL DEFAULT '',
    registration_number text        NOT NULL DEFAULT '',
    director_name       text        NOT NULL DEFAULT '',
    email               text        NOT NULL DEFAULT '',
    website             text        NOT NULL DEFAULT '',
    logo_url            text        NOT NULL DEFAULT '',
    is_active           boolean     NOT NULL DEFAULT true,
    standard_tax_rate   numeric     NOT NULL DEFAULT 0.10,
    reduced_tax_rate    numeric     NOT NULL DEFAULT 0.08,
    -- 008: 帳票レイアウト設定（Issue #179）
    accounting_document_show_logo                boolean NOT NULL DEFAULT false,
    accounting_document_show_registration_warning boolean NOT NULL DEFAULT true,
    accounting_document_show_item_category        boolean NOT NULL DEFAULT true,
    accounting_document_footer_note               text    NOT NULL DEFAULT '',
    -- 010: 帳票セクション設定（Issue #190）
    accounting_document_show_clinic_header        boolean NOT NULL DEFAULT true,
    accounting_document_show_owner_pet_info       boolean NOT NULL DEFAULT true,
    accounting_document_show_items_table          boolean NOT NULL DEFAULT true,
    accounting_document_show_payment_summary      boolean NOT NULL DEFAULT true,
    accounting_document_section_order             text[]  NOT NULL DEFAULT '{}',
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);

-- ------------------------------------
-- 2a. clinic_integrations（Lステップ/LINE連携設定: 005 統合）
-- ------------------------------------
CREATE TABLE clinic_integrations (
    id          BIGSERIAL    PRIMARY KEY,
    clinic_id   bigint       NOT NULL REFERENCES clinics(id),
    service     varchar(50)  NOT NULL,
    key_name    varchar(100) NOT NULL,
    key_value   text         NOT NULL,
    created_at  timestamptz  NOT NULL DEFAULT now(),
    updated_at  timestamptz  NOT NULL DEFAULT now(),
    UNIQUE (clinic_id, service, key_name)
);

CREATE INDEX idx_clinic_integrations_clinic_service
    ON clinic_integrations (clinic_id, service);

COMMENT ON TABLE clinic_integrations IS 'Lステップ/LINE連携設定保存テーブル（005 統合）';

-- ==========================================================================
-- レイヤー2: clinics依存
-- ==========================================================================

-- ------------------------------------
-- 3. animal_species（ペット種類マスタ: システム共通）
-- ------------------------------------
CREATE TABLE animal_species (
    id         BIGSERIAL     PRIMARY KEY,
    name       text          NOT NULL,
    is_active  boolean       NOT NULL DEFAULT true,
    sort_order integer                DEFAULT 0,
    created_at timestamptz   NOT NULL DEFAULT now(),
    updated_at timestamptz   NOT NULL DEFAULT now()
);

-- ------------------------------------
-- 4. occupations（職種マスタ）
-- ------------------------------------
CREATE TABLE occupations (
    id          BIGSERIAL   PRIMARY KEY,
    clinic_id   bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name        text        NOT NULL DEFAULT '',
    description text        NOT NULL DEFAULT '',
    sort_order  integer     NOT NULL DEFAULT 0,
    is_active   boolean     NOT NULL DEFAULT true,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

-- ------------------------------------
-- 5. accounts（認証用アカウント）
-- ------------------------------------
CREATE TABLE accounts (
    id             BIGSERIAL   PRIMARY KEY,
    email          text        NOT NULL UNIQUE,
    password_hash  text        NOT NULL,
    is_active        boolean     NOT NULL DEFAULT true,
    is_system_admin  boolean     NOT NULL DEFAULT false,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    deleted_at       timestamptz
);

CREATE INDEX idx_accounts_email ON accounts(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounts_system_admin ON accounts(is_system_admin) WHERE is_system_admin = true AND deleted_at IS NULL;

-- ------------------------------------
-- 6. staffs（スタッフマスタ）
-- ------------------------------------
CREATE TABLE staffs (
    id                    BIGSERIAL   PRIMARY KEY,
    clinic_id             bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    account_id            bigint               REFERENCES accounts(id) ON DELETE SET NULL,
    name                  text        NOT NULL,
    is_active             boolean     NOT NULL DEFAULT true,
    license_number        text        NOT NULL DEFAULT '',
    occupation_id         bigint               REFERENCES occupations(id) ON DELETE SET NULL,
    sort_order            integer              DEFAULT 0,
    staff_type                 staff_type  NOT NULL DEFAULT 'doctor',
    reservation_display_name   text        NOT NULL DEFAULT '',
    reservation_visible        boolean     NOT NULL DEFAULT true,
    reservation_comment        text        NOT NULL DEFAULT '',
    reservation_image_url      text        NOT NULL DEFAULT '',
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now(),
    deleted_at            timestamptz
);

CREATE INDEX idx_staffs_account ON staffs(account_id);
CREATE INDEX idx_staffs_clinic ON staffs(clinic_id) WHERE deleted_at IS NULL;

-- ------------------------------------
-- 7. owners（飼主情報）
-- ------------------------------------
CREATE TABLE owners (
    id               BIGSERIAL       PRIMARY KEY,
    clinic_id        bigint          NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name             text            NOT NULL,
    name_kana        text            NOT NULL DEFAULT '' CHECK (name_kana !~ '[ァ-ヶ]'),
    birth_date       date,
    company          text            NOT NULL DEFAULT '',
    postal_code      text            NOT NULL DEFAULT '',
    address1         text            NOT NULL DEFAULT '',
    address2         text            NOT NULL DEFAULT '',
    home_postal_code text            NOT NULL DEFAULT '',
    home_address1    text            NOT NULL DEFAULT '',
    home_address2    text            NOT NULL DEFAULT '',
    phone            text            NOT NULL DEFAULT '',
    company_phone    text            NOT NULL DEFAULT '',
    email            text            NOT NULL DEFAULT '',
    remarks          text            NOT NULL DEFAULT '',
    is_dangerous     boolean         NOT NULL DEFAULT false,
    discount_rate    numeric(5,2)    NOT NULL DEFAULT 0,
    membership_type  membership_type NOT NULL DEFAULT 'non_member',
    -- 007: LINE連携
    line_user_id     text,                                           -- LINE User ID（Lステップ連携・LINE通知用）。NULL = 未連携。
    -- 009: Lステップオプトアウト
    lstep_opt_out        boolean     NOT NULL DEFAULT false,         -- Lステップ配信オプトアウトフラグ。true = すべてのタグ付与をスキップ。
    lstep_opt_out_at     timestamptz NULL,                           -- オプトアウト設定日時。
    lstep_opt_out_reason text        NULL,                           -- オプトアウト理由（監査ログ用）。
    -- 015: LINEフォロー・ブロック
    line_followed_at     timestamptz,                                -- LINE フォロー日時（最終フォロー時刻）。Webhook follow イベントで更新。
    line_blocked_at      timestamptz,                                -- LINE ブロック日時。Webhook unfollow イベントで更新。再フォロー時に NULL にリセット。
    -- ext-009: LINE確認・配信停止・転院フィールド
    line_id_confirmed_by    bigint      REFERENCES staffs(id) ON DELETE SET NULL,
    line_id_confirmed_at    timestamptz,
    delivery_excluded       boolean     NOT NULL DEFAULT false,
    delivery_excluded_reason varchar(100),
    is_transferred          boolean     NOT NULL DEFAULT false,
    transfer_at             timestamptz,
    -- ext-014: 配信注意フラグ
    delivery_caution        boolean     NOT NULL DEFAULT false,
    delivery_caution_reason varchar(100),
    -- 006: DM送付希望（Issue #158）
    dm_preference           boolean     NULL,                           -- DM（ダイレクトメール）送付希望。NULL=未設定 / true=必要 / false=不要。
    created_at       timestamptz     NOT NULL DEFAULT now(),
    updated_at       timestamptz     NOT NULL DEFAULT now(),
    deleted_at       timestamptz
);

-- 007: owners.line_user_id インデックス
-- 同一クリニック内で line_user_id の重複を防ぐ（NULL は一意性制約の対象外）
CREATE UNIQUE INDEX uk_owners_clinic_line_user_id
    ON owners(clinic_id, line_user_id)
    WHERE line_user_id IS NOT NULL AND deleted_at IS NULL;

-- 検索用インデックス（line_user_id 単体での参照も想定）
CREATE INDEX idx_owners_line_user_id
    ON owners(line_user_id)
    WHERE line_user_id IS NOT NULL AND deleted_at IS NULL;

COMMENT ON COLUMN owners.line_user_id       IS 'LINE User ID（Lステップ連携・LINE通知用）。NULL = 未連携。';
COMMENT ON COLUMN owners.line_id_confirmed_by IS 'LINE ID 紐付け確認者 (staff_id)。NULL = 未確認。';
COMMENT ON COLUMN owners.lstep_opt_out      IS 'Lステップ配信オプトアウトフラグ。true = すべてのタグ付与をスキップ。';
COMMENT ON COLUMN owners.lstep_opt_out_at   IS 'オプトアウト設定日時。';
COMMENT ON COLUMN owners.lstep_opt_out_reason IS 'オプトアウト理由（監査ログ用）。';
COMMENT ON COLUMN owners.line_followed_at   IS 'LINE フォロー日時（最終フォロー時刻）。Webhook follow イベントで更新。';
COMMENT ON COLUMN owners.line_blocked_at    IS 'LINE ブロック日時。Webhook unfollow イベントで更新。再フォロー時に NULL にリセット。';
COMMENT ON COLUMN owners.dm_preference      IS 'DM（ダイレクトメール）送付希望。NULL=未設定 / true=必要 / false=不要。';

-- ext-009: owners 配信・転院フィールドインデックス
CREATE INDEX idx_owners_delivery_excluded
    ON owners (clinic_id, delivery_excluded)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_owners_is_transferred
    ON owners (clinic_id, is_transferred)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_owners_line_id_confirmed
    ON owners (clinic_id, line_id_confirmed_at)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_owners_line_id_confirmed_by
    ON owners (line_id_confirmed_by)
    WHERE line_id_confirmed_by IS NOT NULL;

-- ------------------------------------
-- 7a. lstep_tag_cache（Lステップタグキャッシュ: 008 統合）
-- ------------------------------------
CREATE TABLE lstep_tag_cache (
    id          BIGSERIAL    PRIMARY KEY,
    clinic_id   bigint       NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    owner_id    bigint       NOT NULL REFERENCES owners(id)  ON DELETE RESTRICT,
    tag_name    varchar(100) NOT NULL,
    category    varchar(20)  NOT NULL DEFAULT 'auto'
                CHECK (category IN ('auto', 'manual')),
    synced_at   timestamptz  NOT NULL DEFAULT now(),
    reason      text,                                         -- ext-013: タグ付与理由（任意）
    UNIQUE (clinic_id, owner_id, tag_name)
);

-- タグ名での集計・検索用（集計API）
CREATE INDEX idx_lstep_tag_cache_clinic_tag ON lstep_tag_cache (clinic_id, tag_name);
-- 飼い主のタグ一覧取得用
CREATE INDEX idx_lstep_tag_cache_owner ON lstep_tag_cache (clinic_id, owner_id);

COMMENT ON TABLE lstep_tag_cache IS 'Lステップタグのカルテ側キャッシュ。タグ操作ごとにupsert/deleteして同期を保つ。（008 統合）';
COMMENT ON COLUMN lstep_tag_cache.category IS 'auto=各Sync*メソッドが自動付与, manual=スタッフが手動付与';

-- ------------------------------------
-- 7b. line_link_tokens（LINE User ID 紐付けトークン: 016 統合）
-- ------------------------------------
CREATE TABLE line_link_tokens (
    id          BIGSERIAL    PRIMARY KEY,
    clinic_id   bigint       NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    owner_id    bigint       NOT NULL REFERENCES owners(id) ON DELETE RESTRICT,
    token       varchar(64)  NOT NULL UNIQUE,
    expires_at  timestamptz  NOT NULL,
    used_at     timestamptz  NULL,
    created_at  timestamptz  NOT NULL DEFAULT now()
);

-- トークン検索用（未使用のみを高速検索）
CREATE INDEX idx_line_link_tokens_token ON line_link_tokens (token)
    WHERE used_at IS NULL;

-- 飼い主ごとの発行履歴確認用
CREATE INDEX idx_line_link_tokens_owner ON line_link_tokens (clinic_id, owner_id);

COMMENT ON TABLE line_link_tokens IS 'LINE User ID 紐付け用の一時トークン（24時間有効、1回限り使用）。（016 統合）';

-- ------------------------------------
-- 7c. token_blacklist（refresh_token JTI ブラックリスト: 006b 統合）
-- ------------------------------------
CREATE TABLE token_blacklist (
    jti        TEXT        NOT NULL PRIMARY KEY,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE token_blacklist IS
    'ログアウト・失効済み refresh_token の JTI ブラックリスト';
COMMENT ON COLUMN token_blacklist.jti IS
    'JWT ID クレーム (uuid v4)。PRIMARY KEY なので一意性は保証される';
COMMENT ON COLUMN token_blacklist.expires_at IS
    '元 refresh_token の有効期限。これ以降は照合対象から除外してよい（バッチ削除の目安）';

CREATE INDEX idx_token_blacklist_expires_at ON token_blacklist(expires_at);

-- ------------------------------------
-- 7d. lstep_migration_progress（Lステップ一括同期進捗: 017 統合）
-- ------------------------------------
CREATE TABLE lstep_migration_progress (
    id            BIGSERIAL    PRIMARY KEY,
    clinic_id     bigint       NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    owner_id      bigint       NOT NULL REFERENCES owners(id) ON DELETE RESTRICT,
    status        varchar(20)  NOT NULL DEFAULT 'pending',  -- pending | success | partial | failed | skipped
    tags_added    int          NOT NULL DEFAULT 0,
    tags_failed   int          NOT NULL DEFAULT 0,
    error_message text,
    started_at    timestamptz,
    completed_at  timestamptz,
    UNIQUE (clinic_id, owner_id)
);

CREATE INDEX idx_lstep_migration_progress_clinic_id
    ON lstep_migration_progress (clinic_id);

COMMENT ON TABLE lstep_migration_progress IS '既存飼い主データ一括同期の進捗管理テーブル（017 統合）';

-- ------------------------------------
-- 7e. lstep_settings（Lステップ同期設定: ext-007 統合）
-- ------------------------------------
CREATE TABLE lstep_settings (
    id               BIGSERIAL    PRIMARY KEY,
    clinic_id        bigint       NOT NULL UNIQUE REFERENCES clinics(id) ON DELETE RESTRICT,
    is_sync_enabled  boolean      NOT NULL DEFAULT false,
    sync_enabled_at  timestamptz  NULL,
    created_at       timestamptz  NOT NULL DEFAULT now(),
    updated_at       timestamptz  NOT NULL DEFAULT now()
);

CREATE INDEX idx_lstep_settings_clinic_id ON lstep_settings (clinic_id);

COMMENT ON TABLE lstep_settings IS 'クリニックごとのLステップ同期設定（ext-007 統合）';

-- ------------------------------------
-- 7f. lstep_sync_error_counters（Lステップ同期エラーカウンター: ext-008 統合）
-- ------------------------------------
CREATE TABLE lstep_sync_error_counters (
    id            BIGSERIAL    PRIMARY KEY,
    clinic_id     bigint       NOT NULL REFERENCES clinics(id)  ON DELETE RESTRICT,
    owner_id      bigint       NOT NULL REFERENCES owners(id)   ON DELETE RESTRICT,
    failure_count int          NOT NULL DEFAULT 0,
    created_at    timestamptz  NOT NULL DEFAULT now(),
    updated_at    timestamptz  NOT NULL DEFAULT now(),
    UNIQUE (clinic_id, owner_id)
);

CREATE INDEX idx_lstep_sync_error_counters_clinic_owner ON lstep_sync_error_counters (clinic_id, owner_id);

COMMENT ON TABLE lstep_sync_error_counters IS 'Lステップ同期APIの連続失敗回数を記録するカウンター（ext-008 統合）';

-- ------------------------------------
-- 7g. lstep_tag_code_mappings（Lステップタグコードマッピング: ext-010 統合）
-- ------------------------------------
CREATE TABLE lstep_tag_code_mappings (
    id            BIGSERIAL    PRIMARY KEY,
    clinic_id     bigint       NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    tag_name      text         NOT NULL,
    code_type     text         NOT NULL,
    codes         text[]       NOT NULL DEFAULT '{}',
    species_scope text,
    age_min       int,
    deleted_at    timestamptz,
    created_at    timestamptz  NOT NULL DEFAULT now(),
    updated_at    timestamptz  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_lstep_tag_code_mappings_clinic_tag_type
    ON lstep_tag_code_mappings (clinic_id, tag_name, code_type)
    WHERE deleted_at IS NULL;

COMMENT ON TABLE lstep_tag_code_mappings IS 'Lステップタグ → 診療コード の対応マスタ（ext-010 統合）';

-- ------------------------------------
-- 7g-2. lstep_trigger_priorities（Q23 配信衝突優先順位: mig-008 統合）
-- ------------------------------------
CREATE TABLE lstep_trigger_priorities (
    id           BIGSERIAL PRIMARY KEY,
    clinic_id    BIGINT NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    trigger_type VARCHAR(64) NOT NULL,
    priority     INTEGER NOT NULL CHECK (priority >= 1),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (clinic_id, trigger_type)
);

CREATE INDEX idx_lstep_trigger_priorities_clinic
    ON lstep_trigger_priorities(clinic_id);

COMMENT ON TABLE lstep_trigger_priorities IS 'Q23 配信トリガー優先順位 (clinic単位カスタマイズ可)';
COMMENT ON COLUMN lstep_trigger_priorities.priority IS '小さいほど優先 (1=最優先)。同日複数トリガー発火時、MIN(priority) のみ実配信';

-- ------------------------------------
-- 7h. lstep_delivery_trigger_log（Lステップ配信トリガーログ: ext-012 統合）
-- ------------------------------------
CREATE TABLE lstep_delivery_trigger_log (
    id               BIGSERIAL    PRIMARY KEY,
    owner_id         bigint       NOT NULL REFERENCES owners(id)   ON DELETE RESTRICT,
    clinic_id        bigint       NOT NULL REFERENCES clinics(id)  ON DELETE RESTRICT,
    trigger_type     varchar(50)  NOT NULL,
    scheduled_at     timestamptz  NOT NULL,
    status           varchar(20)  NOT NULL DEFAULT 'scheduled',  -- scheduled | fired | excluded | cancelled
    fired_at         timestamptz,
    excluded_reason  varchar(100),
    suppressed_by_priority BOOLEAN   NOT NULL DEFAULT FALSE,
    suppression_reason VARCHAR(255),
    created_at       timestamptz  NOT NULL DEFAULT now(),
    updated_at       timestamptz  NOT NULL DEFAULT now()
);

CREATE INDEX idx_lstep_delivery_trigger_log_lookup
    ON lstep_delivery_trigger_log (clinic_id, owner_id, trigger_type, scheduled_at);
CREATE INDEX idx_lstep_delivery_trigger_log_clinic_date
    ON lstep_delivery_trigger_log (clinic_id, scheduled_at);
CREATE INDEX idx_lstep_delivery_trigger_log_suppressed
    ON lstep_delivery_trigger_log(clinic_id, suppressed_by_priority)
    WHERE suppressed_by_priority = TRUE;

COMMENT ON TABLE lstep_delivery_trigger_log IS 'Lステップ自動配信トリガーの実行ログ（ext-012 統合）';
COMMENT ON COLUMN lstep_delivery_trigger_log.suppressed_by_priority IS 'Q23 優先順位により抑制されたか (FALSE=実配信 / TRUE=ログのみ)';
COMMENT ON COLUMN lstep_delivery_trigger_log.suppression_reason IS '抑制理由 (例: "owner_id=42 already triggered by dormant_365d at 2026-05-11")';

-- ------------------------------------
-- 7i. lstep_csv_imports（Lステップ CSV インポート: ext-017 統合）
-- ------------------------------------
CREATE TABLE lstep_csv_imports (
    id                  uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id           bigint       NOT NULL REFERENCES clinics(id)    ON DELETE RESTRICT,
    csv_type            varchar(50)  NOT NULL,
    file_name           varchar(255) NOT NULL,
    uploaded_by_user_id bigint       NOT NULL REFERENCES accounts(id)   ON DELETE RESTRICT,
    row_count           int          NOT NULL DEFAULT 0,
    success_count       int          NOT NULL DEFAULT 0,
    error_count         int          NOT NULL DEFAULT 0,
    status              varchar(20)  NOT NULL DEFAULT 'pending',  -- pending | processing | completed | failed
    error_log           jsonb,
    imported_at         timestamptz,
    created_at          timestamptz  NOT NULL DEFAULT now()
);

CREATE INDEX idx_lstep_csv_imports_clinic_imported
    ON lstep_csv_imports (clinic_id, imported_at DESC);

COMMENT ON TABLE lstep_csv_imports IS 'Lステップ友だち属性CSVのインポート履歴（ext-017 統合）';

-- ------------------------------------
-- 7j. lstep_friend_attribute_snapshots（Lステップ友だち属性スナップショット: ext-018 統合）
-- ------------------------------------
CREATE TABLE lstep_friend_attribute_snapshots (
    id                BIGSERIAL    PRIMARY KEY,
    clinic_id         bigint       NOT NULL REFERENCES clinics(id)       ON DELETE RESTRICT,
    line_user_id      varchar(50)  NOT NULL,
    display_name      varchar(255),
    registered_at     timestamptz,
    tags              jsonb,
    scenarios         jsonb,
    traffic_source    varchar(100),
    block_status      varchar(20),
    last_message_at   timestamptz,
    snapshot_taken_at timestamptz  NOT NULL,
    csv_import_id     uuid         REFERENCES lstep_csv_imports(id)      ON DELETE RESTRICT,
    created_at        timestamptz  NOT NULL DEFAULT now(),
    updated_at        timestamptz  NOT NULL DEFAULT now(),
    UNIQUE (clinic_id, line_user_id, snapshot_taken_at)
);

CREATE INDEX idx_lstep_friend_attribute_snapshots_clinic_user
    ON lstep_friend_attribute_snapshots (clinic_id, line_user_id);
CREATE INDEX idx_lstep_friend_attribute_snapshots_clinic_taken
    ON lstep_friend_attribute_snapshots (clinic_id, snapshot_taken_at DESC);

COMMENT ON TABLE lstep_friend_attribute_snapshots IS 'Lステップ友だちの属性スナップショット（CSVインポート経由、ext-018 統合）';

-- ------------------------------------
-- 7k. lstep_auto_managed_prefixes（自動管理タグプレフィックス: mig-010 統合）
-- B / C1 / C2 / C3 カテゴリのプレフィックスをコード固定から DB 管理へ移行
-- ------------------------------------
CREATE TABLE lstep_auto_managed_prefixes (
    id          BIGSERIAL    PRIMARY KEY,
    prefix      VARCHAR(100) NOT NULL UNIQUE,
    category    VARCHAR(20)  NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_lstep_auto_managed_prefixes_category ON lstep_auto_managed_prefixes (category);

COMMENT ON TABLE lstep_auto_managed_prefixes IS 'Lステップ自動管理タグプレフィックス (B / C1 / C2 / C3、mig-010 統合)';

-- ------------------------------------
-- 7l. lstep_condition_tag_mappings（慢性疾患コード→タグ名マッピング: mig-010 統合）
-- ------------------------------------
CREATE TABLE lstep_condition_tag_mappings (
    id             BIGSERIAL    PRIMARY KEY,
    condition_code VARCHAR(50)  NOT NULL UNIQUE,
    tag_name       VARCHAR(100) NOT NULL,
    description    TEXT,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE lstep_condition_tag_mappings IS '慢性疾患コード→Lステップタグ名マッピング (mig-010 統合)';

-- ------------------------------------
-- 7m. lstep_send_purpose_tag_prefixes（LINE送信目的→タグプレフィックスマッピング: mig-010 統合）
-- ------------------------------------
CREATE TABLE lstep_send_purpose_tag_prefixes (
    id          BIGSERIAL    PRIMARY KEY,
    purpose     VARCHAR(100) NOT NULL UNIQUE,
    tag_prefix  VARCHAR(100) NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE lstep_send_purpose_tag_prefixes IS 'LINE送信目的→Lステップタグプレフィックスマッピング (mig-010 統合)';

-- ------------------------------------
-- 8. inventory_items（在庫管理）
-- ------------------------------------
CREATE TABLE inventory_items (
    id              BIGSERIAL          PRIMARY KEY,
    clinic_id       bigint             NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name            text               NOT NULL,
    category        inventory_category NOT NULL,
    quantity        integer                     DEFAULT 0,
    unit            text               NOT NULL DEFAULT '',
    min_stock_level integer                     DEFAULT 0,
    location        text               NOT NULL DEFAULT '',
    expiry_date     date,
    supplier        text               NOT NULL DEFAULT '',
    last_restocked  date,
    status          inventory_status            DEFAULT 'sufficient',
    created_at      timestamptz        NOT NULL DEFAULT now(),
    updated_at      timestamptz        NOT NULL DEFAULT now(),
    deleted_at      timestamptz
);

-- ------------------------------------
-- 9. exam_types（検査種別マスタ）
-- ------------------------------------
CREATE TABLE exam_types (
    id               BIGSERIAL   PRIMARY KEY,
    clinic_id        bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name             text        NOT NULL,
    price            bigint,
    is_active        boolean     NOT NULL DEFAULT true,
    description      text        NOT NULL DEFAULT '',
    parent_id        bigint               REFERENCES exam_types(id) ON DELETE SET NULL,
    sort_order       integer              DEFAULT 0,
    is_non_insurance boolean     NOT NULL DEFAULT false,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    deleted_at       timestamptz
);

-- ------------------------------------
-- 10. exam_type_fields（検査項目定義）
-- ------------------------------------
CREATE TABLE exam_type_fields (
    id               BIGSERIAL   PRIMARY KEY,
    exam_type_id     bigint      NOT NULL REFERENCES exam_types(id) ON DELETE CASCADE,
    name             text        NOT NULL,
    inspection_value text        NOT NULL DEFAULT '',
    normal_value     text        NOT NULL DEFAULT '',
    unit             text        NOT NULL DEFAULT '',
    sort_order       integer              DEFAULT 0,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now()
);

-- ------------------------------------
-- 11. vaccines（ワクチンマスタ）
-- ------------------------------------
CREATE TABLE vaccines (
    id           BIGSERIAL       PRIMARY KEY,
    clinic_id    bigint          NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name         text            NOT NULL,
    price        bigint,
    is_active    boolean         NOT NULL DEFAULT true,
    description  text            NOT NULL DEFAULT '',
    species      vaccine_species,
    interval     text            NOT NULL DEFAULT '',
    inventory_id bigint                   REFERENCES inventory_items(id) ON DELETE SET NULL,
    parent_id    bigint                   REFERENCES vaccines(id) ON DELETE SET NULL,
    sort_order   integer                  DEFAULT 0,
    created_at   timestamptz     NOT NULL DEFAULT now(),
    updated_at   timestamptz     NOT NULL DEFAULT now(),
    deleted_at   timestamptz
);

-- ------------------------------------
-- 12. medicines（薬剤マスタ）
-- ------------------------------------
CREATE TABLE medicines (
    id               BIGSERIAL     PRIMARY KEY,
    clinic_id        bigint        NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name             text          NOT NULL,
    price            bigint,
    is_active        boolean       NOT NULL DEFAULT true,
    description      text          NOT NULL DEFAULT '',
    parent_id        bigint                 REFERENCES medicines(id) ON DELETE SET NULL,
    dosage_form      dosage_form,
    medicine_unit    medicine_unit,
    inventory_id     bigint                 REFERENCES inventory_items(id) ON DELETE SET NULL,
    default_quantity numeric(10,1)          DEFAULT 1,
    tax_type         tax_type      NOT NULL DEFAULT 'excluded',
    tax_rate         numeric       NOT NULL DEFAULT 0.10,
    sort_order       integer                DEFAULT 0,
    is_non_insurance boolean       NOT NULL DEFAULT false,
    created_at       timestamptz   NOT NULL DEFAULT now(),
    updated_at       timestamptz   NOT NULL DEFAULT now(),
    deleted_at       timestamptz
);

-- ------------------------------------
-- 13. insurances（保険マスタ）
-- ------------------------------------
CREATE TABLE insurances (
    id            BIGSERIAL   PRIMARY KEY,
    clinic_id     bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name          text        NOT NULL,
    is_active     boolean     NOT NULL DEFAULT true,
    description   text        NOT NULL DEFAULT '',
    coverage_rate integer     NOT NULL DEFAULT 0 CHECK (coverage_rate >= 0 AND coverage_rate <= 100),
    contact_phone text        NOT NULL DEFAULT '',
    sort_order    integer              DEFAULT 0,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    deleted_at    timestamptz
);

-- ------------------------------------
-- 14. cages（ケージマスタ）
-- ------------------------------------
CREATE TABLE cages (
    id          BIGSERIAL   PRIMARY KEY,
    clinic_id   bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name        text        NOT NULL,
    price       bigint,
    is_active   boolean     NOT NULL DEFAULT true,
    description text        NOT NULL DEFAULT '',
    cage_type   cage_type   NOT NULL,
    cage_size   cage_size   NOT NULL,
    sort_order  integer              DEFAULT 0,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

-- ------------------------------------
-- 15. reservation_type_groups（予約区分グループマスタ）
-- ------------------------------------
CREATE TABLE reservation_type_groups (
    id         BIGSERIAL   PRIMARY KEY,
    clinic_id  bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name       text        NOT NULL,
    color      text        NOT NULL DEFAULT '#3B82F6',
    sort_order integer              DEFAULT 0,
    is_active  boolean     NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE INDEX idx_rtg_clinic ON reservation_type_groups(clinic_id);

-- ------------------------------------
-- 16. reservation_types（予約区分マスタ）
-- ------------------------------------
CREATE TABLE reservation_types (
    id                       BIGSERIAL   PRIMARY KEY,
    clinic_id                bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name                     text        NOT NULL,
    is_active                boolean     NOT NULL DEFAULT true,
    description              text        NOT NULL DEFAULT '',
    color                    text        NOT NULL DEFAULT '#3B82F6',
    sort_order               integer              DEFAULT 0,
    group_id                 bigint               REFERENCES reservation_type_groups(id) ON DELETE SET NULL,
    reservation_display_name text        NOT NULL DEFAULT '',
    duration_minutes         int         NOT NULL DEFAULT 15,
    short_name               text        NOT NULL DEFAULT '',
    show_short_name          boolean     NOT NULL DEFAULT false,
    reservation_visible      boolean     NOT NULL DEFAULT true,
    reservation_comment      text        NOT NULL DEFAULT '',
    reservation_image_url    text        NOT NULL DEFAULT '',
    parent_id               bigint               REFERENCES reservation_types(id) ON DELETE RESTRICT,
    max_concurrent          integer              CHECK (max_concurrent > 0),
    reservation_day_option   text                        NOT NULL DEFAULT 'none',
    is_internal              boolean                     NOT NULL DEFAULT false,
    category                 reservation_type_category   NOT NULL DEFAULT 'general',
    created_at               timestamptz                 NOT NULL DEFAULT now(),
    updated_at               timestamptz                 NOT NULL DEFAULT now(),
    deleted_at               timestamptz
);

CREATE INDEX idx_reservation_types_group_id ON reservation_types(group_id);
COMMENT ON COLUMN reservation_types.parent_id IS
  '親予約区分ID。NULL はルートノード（トップレベル）';
COMMENT ON COLUMN reservation_types.max_concurrent IS
  '同一開始時刻の最大同時受付件数。NULL は無制限（従来動作）';
CREATE INDEX idx_reservation_types_parent
  ON reservation_types(parent_id)
  WHERE parent_id IS NOT NULL AND deleted_at IS NULL;

-- ------------------------------------
-- 16b. reservation_type_available_slots（予約区分予約可能開始時刻）
-- ------------------------------------
CREATE TABLE reservation_type_available_slots (
    id                    BIGSERIAL   PRIMARY KEY,
    clinic_id             bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    reservation_type_id   bigint      NOT NULL REFERENCES reservation_types(id) ON DELETE CASCADE,
    available_type        text        NOT NULL CHECK (available_type IN ('weekly', 'specific')),
    day_of_week           smallint    CHECK (day_of_week >= 0 AND day_of_week <= 6),
    specific_date         date,
    start_time            varchar(5)  NOT NULL CHECK (start_time ~ '^[0-2][0-9]:[0-5][0-9]$'),
    is_active             boolean     NOT NULL DEFAULT true,
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_rtype_available_slot_clinic_type ON reservation_type_available_slots(clinic_id, reservation_type_id);
CREATE UNIQUE INDEX idx_rtype_available_slot_weekly_unique   ON reservation_type_available_slots(reservation_type_id, day_of_week, start_time) WHERE available_type = 'weekly';
CREATE UNIQUE INDEX idx_rtype_available_slot_specific_unique ON reservation_type_available_slots(reservation_type_id, specific_date, start_time) WHERE available_type = 'specific';

-- ------------------------------------
-- 16c. staff_reservation_capabilities（スタッフ対応可能予約区分）
-- ------------------------------------
CREATE TABLE staff_reservation_capabilities (
    id                    BIGSERIAL   PRIMARY KEY,
    clinic_id             bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    staff_id              bigint      NOT NULL REFERENCES staffs(id) ON DELETE CASCADE,
    reservation_type_id   bigint      NOT NULL REFERENCES reservation_types(id) ON DELETE CASCADE,
    created_at            timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uq_staff_reservation_capability UNIQUE (clinic_id, staff_id, reservation_type_id)
);

CREATE INDEX idx_staff_reservation_capabilities_clinic_staff ON staff_reservation_capabilities(clinic_id, staff_id);
CREATE INDEX idx_staff_reservation_capabilities_clinic_type  ON staff_reservation_capabilities(clinic_id, reservation_type_id);

-- ------------------------------------
-- 17. consultations（診察項目マスタ）
-- ------------------------------------
CREATE TABLE consultations (
    id             BIGSERIAL   PRIMARY KEY,
    clinic_id      bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name           text        NOT NULL,
    price          bigint,
    is_active      boolean     NOT NULL DEFAULT true,
    description    text        NOT NULL DEFAULT '',
    time_condition text        NOT NULL DEFAULT '',
    duration       integer,
    parent_id      bigint               REFERENCES consultations(id) ON DELETE SET NULL,
    tax_type       tax_type    NOT NULL DEFAULT 'excluded',
    tax_rate       numeric     NOT NULL DEFAULT 0.10,
    sort_order     integer              DEFAULT 0,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    deleted_at     timestamptz
);

-- ------------------------------------
-- 18. procedures（処置項目マスタ）
-- ------------------------------------
CREATE TABLE procedures (
    id          BIGSERIAL       PRIMARY KEY,
    clinic_id   bigint          NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name        text            NOT NULL,
    price       bigint,
    is_active   boolean         NOT NULL DEFAULT true,
    description text            NOT NULL DEFAULT '',
    duration    integer,
    anesthesia  anesthesia_type          DEFAULT 'none',
    parent_id   bigint                   REFERENCES procedures(id) ON DELETE SET NULL,
    tax_type    tax_type        NOT NULL DEFAULT 'excluded',
    tax_rate    numeric         NOT NULL DEFAULT 0.10,
    sort_order  integer                  DEFAULT 0,
    -- 012: 手術処置フラグ（Issue #159）
    is_surgery  boolean         NOT NULL DEFAULT false,
    created_at  timestamptz     NOT NULL DEFAULT now(),
    updated_at  timestamptz     NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

CREATE INDEX IF NOT EXISTS idx_procedures_clinic_is_surgery
    ON procedures(clinic_id, is_surgery)
    WHERE is_surgery = true;

-- ------------------------------------
-- 19. hospitalization_plans（入院プランマスタ）
-- ------------------------------------
CREATE TABLE hospitalization_plans (
    id           BIGSERIAL    PRIMARY KEY,
    clinic_id    bigint       NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name         text         NOT NULL,
    price        bigint,
    is_active    boolean      NOT NULL DEFAULT true,
    description  text         NOT NULL DEFAULT '',
    body_size    body_size,
    billing_unit billing_unit          DEFAULT 'per_day',
    tax_type     tax_type     NOT NULL DEFAULT 'excluded',
    tax_rate     numeric      NOT NULL DEFAULT 0.10,
    sort_order   integer               DEFAULT 0,
    created_at   timestamptz  NOT NULL DEFAULT now(),
    updated_at   timestamptz  NOT NULL DEFAULT now(),
    deleted_at   timestamptz
);

-- ------------------------------------
-- 19b. trimming_course_types（トリミングコース種別マスタ）
-- ------------------------------------
CREATE TABLE trimming_course_types (
    id          BIGSERIAL   PRIMARY KEY,
    clinic_id   bigint      NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    name        varchar(50) NOT NULL,
    sort_order  integer              DEFAULT 0,
    is_active   boolean     NOT NULL DEFAULT true,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

CREATE UNIQUE INDEX idx_trimming_course_types_clinic_name ON trimming_course_types(clinic_id, name) WHERE deleted_at IS NULL;
CREATE INDEX idx_trimming_course_types_clinic_order ON trimming_course_types(clinic_id, sort_order) WHERE deleted_at IS NULL;

-- ------------------------------------
-- 20. trimming_courses（トリミングコースマスタ）
-- ------------------------------------
CREATE TABLE trimming_courses (
    id             BIGSERIAL   PRIMARY KEY,
    clinic_id      bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name           text        NOT NULL,
    price          bigint,
    is_active      boolean     NOT NULL DEFAULT true,
    description    text        NOT NULL DEFAULT '',
    target_size    target_size,
    duration       integer,
    sort_order     integer              DEFAULT 0,
    course_type_id bigint               REFERENCES trimming_course_types(id) ON DELETE SET NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    deleted_at     timestamptz
);

CREATE INDEX idx_trimming_courses_course_type_id ON trimming_courses(course_type_id) WHERE course_type_id IS NOT NULL;

-- ------------------------------------
-- 21. trimming_options（トリミングオプションマスタ）
-- ------------------------------------
CREATE TABLE trimming_options (
    id          BIGSERIAL   PRIMARY KEY,
    clinic_id   bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name        text        NOT NULL,
    price       bigint,
    is_active   boolean     NOT NULL DEFAULT true,
    description text        NOT NULL DEFAULT '',
    duration    integer,
    is_combinable  boolean     NOT NULL DEFAULT true,
    sort_order  integer              DEFAULT 0,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

-- ------------------------------------
-- 22. diagnosis_types（診断カテゴリマスタ）
-- ------------------------------------
CREATE TABLE diagnosis_types (
    id          BIGSERIAL   PRIMARY KEY,
    clinic_id   bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name        text        NOT NULL,
    is_active   boolean     NOT NULL DEFAULT true,
    description text        NOT NULL DEFAULT '',
    sort_order  integer              DEFAULT 0,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

-- ------------------------------------
-- 23. diagnosis_names（診断病名マスタ）
-- ------------------------------------
CREATE TABLE diagnosis_names (
    id                    BIGSERIAL   PRIMARY KEY,
    clinic_id             bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name                  text        NOT NULL,
    is_active             boolean     NOT NULL DEFAULT true,
    description           text        NOT NULL DEFAULT '',
    diagnosis_type_id     bigint      NOT NULL REFERENCES diagnosis_types(id) ON DELETE CASCADE,
    sort_order            integer              DEFAULT 0,
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now(),
    deleted_at            timestamptz
);

-- ------------------------------------
-- 24. checkup_types（健診種別マスタ）
-- ------------------------------------
CREATE TABLE checkup_types (
    id          BIGSERIAL   PRIMARY KEY,
    clinic_id   bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name        text        NOT NULL,
    price       bigint,
    is_active   boolean     NOT NULL DEFAULT true,
    description text        NOT NULL DEFAULT '',
    interval    text        NOT NULL DEFAULT '',
    target_age  text        NOT NULL DEFAULT '',
    parent_id   bigint               REFERENCES checkup_types(id) ON DELETE SET NULL,
    sort_order  integer              DEFAULT 0,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

-- ------------------------------------
-- 25. chief_complaint_types（主訴区分マスタ）
-- ------------------------------------
CREATE TABLE chief_complaint_types (
    id          BIGSERIAL   PRIMARY KEY,
    clinic_id   bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name        text        NOT NULL,
    description text        NOT NULL DEFAULT '',
    is_active   boolean     NOT NULL DEFAULT true,
    sort_order  integer              DEFAULT 0,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

-- ------------------------------------
-- 26. inquiry_templates（問診定型文マスタ）
-- ------------------------------------
CREATE TABLE inquiry_templates (
    id         BIGSERIAL   PRIMARY KEY,
    clinic_id  bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    category   text        NOT NULL DEFAULT '',
    title      text        NOT NULL,
    content    text        NOT NULL DEFAULT '',
    is_active  boolean     NOT NULL DEFAULT true,
    sort_order integer              DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

-- ==========================================================================
-- レイヤー3: owners/staffs/animal_species等依存
-- ==========================================================================

-- ------------------------------------
-- 27. pets（ペット情報）
-- ------------------------------------
CREATE TABLE pets (
    id                BIGSERIAL       PRIMARY KEY,
    clinic_id         bigint          NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    owner_id          bigint          NOT NULL REFERENCES owners(id) ON DELETE RESTRICT,
    pet_number        text            NOT NULL DEFAULT '',
    name              text            NOT NULL,
    name_kana         text            NOT NULL DEFAULT '' CHECK (name_kana !~ '[ァ-ヶ]'),
    animal_species_id bigint          NOT NULL REFERENCES animal_species(id) ON DELETE RESTRICT,
    gender            pet_gender      NOT NULL DEFAULT 'unknown',
    status            pet_status      NOT NULL DEFAULT 'alive',
    birth_date        date,
    breed             text            NOT NULL DEFAULT '',
    color             text            NOT NULL DEFAULT '',
    weight            numeric(6,2),
    neutered_date     date,
    acquisition_type  acquisition_type,
    danger_level      danger_level    NOT NULL DEFAULT 'low',
    food              text            NOT NULL DEFAULT '',
    environment       text            NOT NULL DEFAULT '',
    phone             text            NOT NULL DEFAULT '',
    last_visit        date,
    insurance_id      bigint                   REFERENCES insurances(id) ON DELETE SET NULL,
    remarks           text            NOT NULL DEFAULT '',
    -- 009: ペット死亡記録
    deceased_at       timestamptz     NULL,                          -- ペット死亡日。NULL = 生存中。
    deceased_reason   text            NULL,                          -- ペット死亡理由（任意記録）。
    -- 006: レガシーEMR準拠の飼主レポート項目（Issue #158）
    blood_type        text            NULL,                          -- ペット血液型。NULL=未記録。
    microchip_number  text            NULL,                          -- マイクロチップ番号。NULL=未記録。
    created_at        timestamptz     NOT NULL DEFAULT now(),
    updated_at        timestamptz     NOT NULL DEFAULT now(),
    deleted_at        timestamptz
);

-- 009: ペット死亡記録インデックス
CREATE INDEX idx_pets_deceased ON pets (clinic_id, deceased_at)
    WHERE deceased_at IS NOT NULL;

COMMENT ON COLUMN pets.deceased_at       IS 'ペット死亡日。NULL = 生存中。';
COMMENT ON COLUMN pets.deceased_reason   IS 'ペット死亡理由（任意記録）。';
COMMENT ON COLUMN pets.blood_type        IS 'ペット血液型。NULL=未記録。';
COMMENT ON COLUMN pets.microchip_number  IS 'マイクロチップ番号。NULL=未記録。';

-- ------------------------------------
-- 27a. pet_chronic_conditions（慢性疾患フラグ管理: 012 統合）
-- ------------------------------------
CREATE TABLE pet_chronic_conditions (
    id             BIGSERIAL    PRIMARY KEY,
    clinic_id      bigint       NOT NULL REFERENCES clinics(id),
    pet_id         bigint       NOT NULL REFERENCES pets(id),
    condition_code varchar(50)  NOT NULL,
    condition_name varchar(100) NOT NULL,
    diagnosed_at   date         NOT NULL,
    notes          text,
    is_active      boolean      NOT NULL DEFAULT true,
    created_at     timestamptz  NOT NULL DEFAULT now(),
    updated_at     timestamptz  NOT NULL DEFAULT now(),
    deleted_at     timestamptz  NULL
);

CREATE INDEX idx_pet_chronic_conditions_clinic_pet
    ON pet_chronic_conditions (clinic_id, pet_id)
    WHERE deleted_at IS NULL;

COMMENT ON TABLE pet_chronic_conditions IS '慢性疾患フラグ管理テーブル（012 統合）';

-- ------------------------------------
-- 28. staff_clinic_assignments（スタッフ-クリニック中間テーブル）
-- ------------------------------------
CREATE TABLE staff_clinic_assignments (
    id             BIGSERIAL   PRIMARY KEY,
    staff_id       bigint      NOT NULL REFERENCES staffs(id) ON DELETE CASCADE,
    clinic_id      bigint      NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    is_main        boolean     NOT NULL DEFAULT false,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    deleted_at     timestamptz,
    CONSTRAINT uk_staff_clinic UNIQUE (staff_id, clinic_id)
);

CREATE INDEX idx_staff_clinic_staff ON staff_clinic_assignments(staff_id);
CREATE INDEX idx_staff_clinic_clinic ON staff_clinic_assignments(clinic_id);
CREATE INDEX idx_staff_clinic_main ON staff_clinic_assignments(staff_id, is_main);

-- ------------------------------------
-- 29. permission_groups（権限グループマスタ）
-- ------------------------------------
CREATE TABLE permission_groups (
    id          BIGSERIAL    PRIMARY KEY,
    clinic_id   bigint       NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name        varchar(100) NOT NULL,
    description text         NOT NULL DEFAULT '',
    color       varchar(7)   NOT NULL DEFAULT '#6B7280',
    is_active   boolean      NOT NULL DEFAULT true,
    sort_order  int          NOT NULL DEFAULT 0,
    created_at  timestamptz  NOT NULL DEFAULT now(),
    updated_at  timestamptz  NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

CREATE UNIQUE INDEX uk_permission_groups ON permission_groups(clinic_id, name) WHERE deleted_at IS NULL;

-- ------------------------------------
-- 30. permission_group_rules（権限グループ-リソース×CRUD権限）
-- ------------------------------------
CREATE TABLE permission_group_rules (
    id         BIGSERIAL   PRIMARY KEY,
    group_id   bigint      NOT NULL REFERENCES permission_groups(id) ON DELETE CASCADE,
    resource   varchar(50) NOT NULL,
    can_view   boolean     NOT NULL DEFAULT false,
    can_create boolean     NOT NULL DEFAULT false,
    can_edit   boolean     NOT NULL DEFAULT false,
    can_delete boolean     NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT uk_permission_group_rules UNIQUE (group_id, resource)
);

CREATE INDEX idx_permission_group_rules_group ON permission_group_rules(group_id);
CREATE INDEX idx_permission_group_rules_deleted_at ON permission_group_rules(deleted_at);

-- ------------------------------------
-- 31. staff_permission_groups（スタッフ-権限グループ中間テーブル）
-- ------------------------------------
CREATE TABLE staff_permission_groups (
    staff_id  bigint NOT NULL REFERENCES staffs(id) ON DELETE CASCADE,
    group_id  bigint NOT NULL REFERENCES permission_groups(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (staff_id, group_id)
);

CREATE INDEX idx_staff_permission_groups_staff ON staff_permission_groups(staff_id);
CREATE INDEX idx_staff_permission_groups_group ON staff_permission_groups(group_id);

-- ------------------------------------
-- 32. line_customers（LINE予約顧客）
-- ------------------------------------
CREATE TABLE line_customers (
    id                BIGSERIAL   PRIMARY KEY,
    clinic_id         bigint      NOT NULL REFERENCES clinics(id),
    line_user_id      text        NOT NULL,
    display_name      text        NOT NULL DEFAULT '',
    real_name         text        NOT NULL DEFAULT '',
    additional_fields jsonb       NOT NULL DEFAULT '{}',
    owner_id          bigint               REFERENCES owners(id) ON DELETE SET NULL,
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now(),
    UNIQUE(clinic_id, line_user_id)
);
CREATE INDEX idx_line_customers_owner
    ON line_customers(owner_id) WHERE owner_id IS NOT NULL;

-- ------------------------------------
-- 32a. shared_files（LINE個別送信用ファイルストレージ: 006 統合）
-- ------------------------------------
CREATE TABLE shared_files (
    id          BIGSERIAL    PRIMARY KEY,
    clinic_id   bigint       NOT NULL REFERENCES clinics(id),
    owner_id    bigint                REFERENCES owners(id),
    uploaded_by bigint       NOT NULL REFERENCES staffs(id),
    file_type   varchar(50)  NOT NULL,
    file_name   varchar(255) NOT NULL,
    file_key    varchar(500) NOT NULL,
    file_size   bigint       NOT NULL,
    purpose     varchar(50)  NOT NULL,
    expires_at  timestamptz,
    created_at  timestamptz  NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

CREATE INDEX idx_shared_files_clinic_owner
    ON shared_files (clinic_id, owner_id);

CREATE INDEX idx_shared_files_expires_at
    ON shared_files (expires_at)
    WHERE deleted_at IS NULL;

COMMENT ON TABLE shared_files IS 'LINE個別送信用ファイルストレージ（006 統合）';

-- ------------------------------------
-- 32b. line_send_logs（LINE送信ログ: 013 統合）
-- ------------------------------------
CREATE TABLE line_send_logs (
    id                BIGSERIAL    PRIMARY KEY,
    clinic_id         bigint       NOT NULL REFERENCES clinics(id),
    owner_id          bigint       NOT NULL REFERENCES owners(id),
    sent_by_user_id   bigint       NOT NULL REFERENCES staffs(id),
    message_type      varchar(20)  NOT NULL,
    content_summary   text         NOT NULL,
    line_message_id   varchar(100),
    status            varchar(20)  NOT NULL,
    error_message     text,
    sent_at           timestamptz  NOT NULL DEFAULT now()
);

CREATE INDEX idx_line_send_logs_clinic_owner ON line_send_logs (clinic_id, owner_id, sent_at DESC);

COMMENT ON TABLE line_send_logs IS 'LINE送信ログ（013 統合）';

-- ==========================================================================
-- レイヤー4: pets依存
-- ==========================================================================

-- ------------------------------------
-- 33. appointments（予約）
-- ------------------------------------
CREATE TABLE appointments (
    id                 BIGSERIAL            PRIMARY KEY,
    clinic_id          bigint               NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    start_time         timestamptz          NOT NULL,
    end_time           timestamptz          NOT NULL,
    owner_id           bigint                        REFERENCES owners(id) ON DELETE SET NULL,
    pet_id             bigint                        REFERENCES pets(id) ON DELETE SET NULL,
    visit_type         visit_type           NOT NULL DEFAULT 'revisit',
    reservation_type_id        bigint               NOT NULL REFERENCES reservation_types(id) ON DELETE RESTRICT,
    doctor_id          bigint                        REFERENCES staffs(id) ON DELETE SET NULL,
    is_designated      boolean                       DEFAULT false,
    status             reservation_status            DEFAULT 'pending',
    notes              text                 NOT NULL DEFAULT '',
    source             reservation_source   NOT NULL DEFAULT 'manual',
    created_by         bigint                        REFERENCES staffs(id),
    is_staff_delegated boolean              NOT NULL DEFAULT false,
    customer_fields    jsonb                NOT NULL DEFAULT '{}',
    reservation_route  varchar(20),                                    -- ext-016: 予約経路（phone/web/walk_in等）
    actual_reservation_at timestamptz,                                -- ext-016: 実際の予約受付日時
    created_at         timestamptz          NOT NULL DEFAULT now(),
    updated_at         timestamptz          NOT NULL DEFAULT now(),
    deleted_at         timestamptz,
    line_customer_id   bigint                        REFERENCES line_customers(id) ON DELETE SET NULL,
    checked_in_at      timestamptz,                                    -- 受付ヘッダー テレメトリ(change-ui.md Phase 2): updated_at は autoUpdateTime のため予約編集全般でリセットされ待ち時間算出に流用できない。checked_in ステータス遷移時刻専用カラム
    CONSTRAINT chk_reservation_times CHECK (end_time >= start_time)
);

-- ------------------------------------
-- 34. hospitalizations（入院/ホテル管理）
-- ------------------------------------
CREATE TABLE hospitalizations (
    id                   BIGSERIAL              PRIMARY KEY,
    clinic_id            bigint                 NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    owner_id             bigint                 NOT NULL REFERENCES owners(id) ON DELETE RESTRICT,
    pet_id               bigint                 NOT NULL REFERENCES pets(id) ON DELETE RESTRICT,
    hospitalization_type hospitalization_type   NOT NULL,
    start_date           date                   NOT NULL,
    end_date             date                   NOT NULL,
    status               hospitalization_status          DEFAULT 'reserved',
    cage_id              bigint                          REFERENCES cages(id) ON DELETE SET NULL,
    doctor_id            bigint                          REFERENCES staffs(id) ON DELETE SET NULL,
    memo                 text                   NOT NULL DEFAULT '',
    owner_request        text                   NOT NULL DEFAULT '',
    staff_notes          text                   NOT NULL DEFAULT '',
    insurance_company_name varchar(100)         NULL,                 -- ext-006: 保険会社名
    insurance_number       varchar(50)          NULL,                 -- ext-006: 保険証番号
    created_at           timestamptz            NOT NULL DEFAULT now(),
    updated_at           timestamptz            NOT NULL DEFAULT now(),
    deleted_at           timestamptz,
    CONSTRAINT chk_hospitalizations_dates CHECK (end_date >= start_date)
);

-- ------------------------------------
-- 35. appointment_trimming_details（トリミング詳細: appointments の1:1拡張）
-- ------------------------------------
CREATE TABLE appointment_trimming_details (
    id               BIGSERIAL        PRIMARY KEY,
    clinic_id        bigint           NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    appointment_id   bigint           NOT NULL UNIQUE REFERENCES appointments(id) ON DELETE CASCADE,
    course_id        bigint                    REFERENCES trimming_courses(id) ON DELETE SET NULL,
    style_request    text             NOT NULL DEFAULT '',
    body_weight      numeric(6,2),
    bw_unit          body_weight_unit          DEFAULT 'Kg',
    body_temperature numeric(4,1),
    used_shampoo     text             NOT NULL DEFAULT '',
    used_ribbon      text             NOT NULL DEFAULT '',
    remarks          text             NOT NULL DEFAULT '',
    style_image      text             NOT NULL DEFAULT '',
    completed_image  text             NOT NULL DEFAULT '',
    created_at       timestamptz      NOT NULL DEFAULT now(),
    updated_at       timestamptz      NOT NULL DEFAULT now()
);

CREATE INDEX idx_appt_trimming_clinic_appointment ON appointment_trimming_details(clinic_id, appointment_id);

-- ------------------------------------
-- 36. appointment_trimming_options（トリミングオプション M:N）
-- ------------------------------------
CREATE TABLE appointment_trimming_options (
    id             BIGSERIAL   PRIMARY KEY,
    appointment_id bigint      NOT NULL REFERENCES appointments(id) ON DELETE CASCADE,
    option_id      bigint      NOT NULL REFERENCES trimming_options(id) ON DELETE RESTRICT,
    sort_order     integer              DEFAULT 0,
    created_at     timestamptz NOT NULL DEFAULT now(),
    UNIQUE (appointment_id, option_id)
);

CREATE INDEX idx_appt_trimming_options_appointment ON appointment_trimming_options(appointment_id);

-- ------------------------------------
-- 37. medical_records（電子カルテ）
-- ------------------------------------
CREATE TABLE medical_records (
    id                             BIGSERIAL             PRIMARY KEY,
    clinic_id                      bigint                NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    record_no                      text                  NOT NULL,
    date                           date                  NOT NULL,
    owner_id                       bigint                         REFERENCES owners(id) ON DELETE RESTRICT,
    pet_id                         bigint                         REFERENCES pets(id) ON DELETE RESTRICT,
    doctor_id                      bigint                         REFERENCES staffs(id) ON DELETE SET NULL,
    appointment_id                 bigint                         REFERENCES appointments(id) ON DELETE SET NULL,
    status                         medical_record_status          DEFAULT 'draft',
    version                        INTEGER               NOT NULL DEFAULT 1,
    entered_by                     bigint                         REFERENCES staffs(id),
    -- 010: 次回来院推奨日
    next_visit_recommended_date    date                  NULL,
    recommendation_reason          varchar(100),                  -- ext-015: 次回来院推奨理由
    visit_type                     visit_type            NULL,    -- 初診/再診（Path B 自動生成カルテで設定）
    created_at                     timestamptz           NOT NULL DEFAULT now(),
    updated_at                     timestamptz           NOT NULL DEFAULT now(),
    deleted_at                     timestamptz
);

-- ------------------------------------
-- 37a. prescriptions（処方薬記録: 011 統合）
-- ------------------------------------
CREATE TABLE prescriptions (
    id                BIGSERIAL    PRIMARY KEY,
    clinic_id         bigint       NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    owner_id          bigint       NOT NULL REFERENCES owners(id) ON DELETE RESTRICT,
    pet_id            bigint                REFERENCES pets(id) ON DELETE RESTRICT,
    medical_record_id bigint                REFERENCES medical_records(id) ON DELETE RESTRICT,
    prescribed_at     date         NOT NULL,
    duration_days     int          NOT NULL DEFAULT 0,
    deleted_at        timestamptz,
    created_at        timestamptz  NOT NULL DEFAULT now(),
    updated_at        timestamptz  NOT NULL DEFAULT now()
);

CREATE INDEX idx_prescriptions_clinic_owner    ON prescriptions(clinic_id, owner_id)       WHERE deleted_at IS NULL;
CREATE INDEX idx_prescriptions_medical_record  ON prescriptions(medical_record_id)          WHERE deleted_at IS NULL;

COMMENT ON TABLE prescriptions IS '処方薬記録テーブル（011 統合）';

-- ------------------------------------
-- 37b. medical_record_addenda（カルテ修正記録: ext-021 統合）
-- ------------------------------------
CREATE TABLE medical_record_addenda (
    id                BIGSERIAL    PRIMARY KEY,
    medical_record_id BIGINT       NOT NULL REFERENCES medical_records(id) ON DELETE RESTRICT,
    clinic_id         BIGINT       NOT NULL REFERENCES clinics(id)         ON DELETE RESTRICT,
    author_user_id    BIGINT       NOT NULL REFERENCES staffs(id)          ON DELETE RESTRICT,
    before_text       TEXT         NOT NULL DEFAULT '',
    after_text        TEXT         NOT NULL,
    reason            TEXT         NOT NULL,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_medical_record_addenda_medical_record_id ON medical_record_addenda (medical_record_id);
CREATE INDEX idx_medical_record_addenda_clinic_id         ON medical_record_addenda (clinic_id);
CREATE INDEX idx_medical_record_addenda_created_at        ON medical_record_addenda (created_at);

COMMENT ON TABLE medical_record_addenda IS 'カルテテキスト修正の追記記録。修正前後テキストと理由を保持する（ext-021 統合）';

-- ------------------------------------
-- 38. vaccinations（予防接種記録）
-- ------------------------------------
CREATE TABLE vaccinations (
    id                 BIGSERIAL          PRIMARY KEY,
    clinic_id          bigint             NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    medical_record_id  bigint                      REFERENCES medical_records(id) ON DELETE CASCADE,
    pet_id             bigint                      REFERENCES pets(id) ON DELETE RESTRICT,
    vaccine_id         bigint             NOT NULL REFERENCES vaccines(id) ON DELETE RESTRICT,
    date               date               NOT NULL,
    next_date          date,
    next_schedule_type next_schedule_type,
    doctor_id          bigint                      REFERENCES staffs(id) ON DELETE SET NULL,
    supplemental       text               NOT NULL DEFAULT '',
    lot1               text               NOT NULL DEFAULT '',
    lot2               text               NOT NULL DEFAULT '',
    lot3               text               NOT NULL DEFAULT '',
    lot4               text               NOT NULL DEFAULT '',
    remarks            text               NOT NULL DEFAULT '',
    created_at         timestamptz        NOT NULL DEFAULT now(),
    updated_at         timestamptz        NOT NULL DEFAULT now(),
    deleted_at         timestamptz
);

-- ------------------------------------
-- 39. checkups（定期健診記録）
-- ------------------------------------
CREATE TABLE checkups (
    id                BIGSERIAL     PRIMARY KEY,
    medical_record_id bigint        NOT NULL REFERENCES medical_records(id) ON DELETE CASCADE,
    clinic_id         bigint        NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    pet_id            bigint                 REFERENCES pets(id) ON DELETE RESTRICT,
    checkup_type_id   bigint        NOT NULL REFERENCES checkup_types(id) ON DELETE RESTRICT,
    date              date          NOT NULL,
    next_date         date,
    doctor_id         bigint                 REFERENCES staffs(id) ON DELETE SET NULL,
    result            text          NOT NULL DEFAULT '',
    created_at        timestamptz   NOT NULL DEFAULT now(),
    updated_at        timestamptz   NOT NULL DEFAULT now(),
    deleted_at        timestamptz
);

-- ------------------------------------
-- 40. exams（検査記録）
-- ------------------------------------
CREATE TABLE exams (
    id                BIGSERIAL          PRIMARY KEY,
    medical_record_id bigint                      REFERENCES medical_records(id) ON DELETE CASCADE,
    clinic_id         bigint             NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    pet_id            bigint                      REFERENCES pets(id) ON DELETE RESTRICT,
    date              date               NOT NULL,
    exam_type_id      bigint             NOT NULL REFERENCES exam_types(id) ON DELETE RESTRICT,
    doctor_id         bigint                      REFERENCES staffs(id) ON DELETE SET NULL,
    status            exam_status                 DEFAULT 'pending',
    result_summary    text               NOT NULL DEFAULT '',
    machine           text               NOT NULL DEFAULT '',
    created_at        timestamptz        NOT NULL DEFAULT now(),
    updated_at        timestamptz        NOT NULL DEFAULT now(),
    deleted_at        timestamptz
);

-- ==========================================================================
-- レイヤー5: medical_records/hospitalizations依存
-- ==========================================================================

-- ------------------------------------
-- 41. inquiries（問診タブ: medical_recordsと1:1）
-- ------------------------------------
CREATE TABLE inquiries (
    id                          BIGSERIAL          PRIMARY KEY,
    medical_record_id           bigint             NOT NULL UNIQUE REFERENCES medical_records(id) ON DELETE CASCADE,
    chief_complaint_type_id     bigint                      REFERENCES chief_complaint_types(id) ON DELETE SET NULL,
    chief_complaint             text               NOT NULL DEFAULT '',
    history                     text               NOT NULL DEFAULT '',
    current_medications         text               NOT NULL DEFAULT '',
    allergy_info                text               NOT NULL DEFAULT '',
    last_meal                   text               NOT NULL DEFAULT '',
    last_defecation             text               NOT NULL DEFAULT '',
    last_urination              text               NOT NULL DEFAULT '',
    appetite                    appetite_level,
    water_intake                water_intake_level,
    owner_observations          text               NOT NULL DEFAULT '',
    notes                       text               NOT NULL DEFAULT '',
    staff_id                    bigint                      REFERENCES staffs(id) ON DELETE SET NULL,
    created_at                  timestamptz        NOT NULL DEFAULT now(),
    updated_at                  timestamptz        NOT NULL DEFAULT now()
);

-- ------------------------------------
-- 42. clinical_plans（診察/治療タブ: medical_recordsと1:1）
-- ------------------------------------
CREATE TABLE clinical_plans (
    id                    BIGSERIAL   PRIMARY KEY,
    medical_record_id     bigint      NOT NULL UNIQUE REFERENCES medical_records(id) ON DELETE CASCADE,
    physical_exam         text        NOT NULL DEFAULT '',
    diagnosis_type_id     bigint               REFERENCES diagnosis_types(id) ON DELETE SET NULL,
    diagnosis_name_id     bigint               REFERENCES diagnosis_names(id) ON DELETE SET NULL,
    diagnosis_2_type_id   bigint               REFERENCES diagnosis_types(id) ON DELETE SET NULL,
    diagnosis_2_name_id   bigint               REFERENCES diagnosis_names(id) ON DELETE SET NULL,
    diagnosis_details     text        NOT NULL DEFAULT '',
    treatment_policy      text        NOT NULL DEFAULT '',
    version               INTEGER     NOT NULL DEFAULT 1,
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now(),
    deleted_at            timestamptz
);

-- ------------------------------------
-- Note: vital_records は daily_records 依存のため 50 の後に定義

-- ------------------------------------
-- 43. treatments（治療明細）
-- ------------------------------------
CREATE TABLE treatments (
    id                BIGSERIAL           PRIMARY KEY,
    medical_record_id bigint              NOT NULL REFERENCES medical_records(id) ON DELETE CASCADE,
    item_type         treatment_item_type NOT NULL DEFAULT 'other',
    consultation_id   bigint                       REFERENCES consultations(id) ON DELETE SET NULL,
    procedure_id      bigint                       REFERENCES procedures(id) ON DELETE SET NULL,
    medicine_id       bigint                       REFERENCES medicines(id) ON DELETE SET NULL,
    is_selected       boolean                      DEFAULT false,
    status            treatment_status             DEFAULT 'pending',
    content           text                NOT NULL DEFAULT '',
    memo              text                NOT NULL DEFAULT '',
    admin_route       varchar(50)         NOT NULL DEFAULT '',
    is_insurance      boolean                      DEFAULT false,
    unit_price        bigint                       DEFAULT 0,
    quantity          numeric(10,1)                DEFAULT 1,
    discount_rate     numeric(5,2)                 DEFAULT 0,
    discount_amount   bigint                       DEFAULT 0,
    inventory_id      bigint                       REFERENCES inventory_items(id) ON DELETE SET NULL,
    sort_order        integer                      DEFAULT 0,
    created_at        timestamptz         NOT NULL DEFAULT now(),
    updated_at        timestamptz         NOT NULL DEFAULT now(),
    deleted_at        timestamptz,
    CONSTRAINT chk_treatment_item_ref CHECK (
        (item_type = 'consultation' AND procedure_id IS NULL AND medicine_id IS NULL) OR
        (item_type = 'procedure'    AND consultation_id IS NULL AND medicine_id IS NULL) OR
        (item_type = 'medicine'     AND consultation_id IS NULL AND procedure_id IS NULL) OR
        (item_type = 'other'        AND consultation_id IS NULL AND procedure_id IS NULL AND medicine_id IS NULL)
    ),
    CONSTRAINT chk_treatment_quantity CHECK (quantity > 0)
);

-- ------------------------------------
-- 44. treatment_plans（治療プラン: 外来・入院共用）
-- ------------------------------------
CREATE TABLE treatment_plans (
    id                 BIGSERIAL   PRIMARY KEY,
    clinic_id          bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    medical_record_id  bigint               REFERENCES medical_records(id) ON DELETE CASCADE,
    hospitalization_id bigint               REFERENCES hospitalizations(id) ON DELETE CASCADE,
    treatment_content  text        NOT NULL DEFAULT '',
    memo               text        NOT NULL DEFAULT '',
    is_insurance       boolean              DEFAULT false,
    unit_price         bigint               DEFAULT 0,
    quantity           numeric(10,1)        DEFAULT 1,
    discount_rate      numeric(5,2)         DEFAULT 0,
    discount_amount    bigint               DEFAULT 0,
    subtotal           bigint               DEFAULT 0,
    sort_order         integer              DEFAULT 0,
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now(),
    deleted_at         timestamptz,
    CONSTRAINT chk_treatment_plans_ref CHECK (
        (medical_record_id IS NOT NULL AND hospitalization_id IS NULL) OR
        (medical_record_id IS NULL AND hospitalization_id IS NOT NULL)
    )
);

-- ------------------------------------
-- 45. medical_record_images（画像タブ）
-- ------------------------------------
CREATE TABLE medical_record_images (
    id                BIGSERIAL          PRIMARY KEY,
    medical_record_id bigint             NOT NULL REFERENCES medical_records(id) ON DELETE CASCADE,
    image_url         text               NOT NULL DEFAULT '',
    thumbnail_url     text               NOT NULL DEFAULT '',
    file_name         text               NOT NULL DEFAULT '',
    file_size         bigint                      DEFAULT 0,
    mime_type         text               NOT NULL DEFAULT '',
    image_type        medical_image_type NOT NULL DEFAULT 'other',
    description       text               NOT NULL DEFAULT '',
    taken_at          timestamptz,
    exam_id           bigint                      REFERENCES exams(id) ON DELETE SET NULL,
    staff_id          bigint                      REFERENCES staffs(id) ON DELETE SET NULL,
    sort_order        integer                     DEFAULT 0,
    created_at        timestamptz        NOT NULL DEFAULT now(),
    updated_at        timestamptz        NOT NULL DEFAULT now()
);

-- ------------------------------------
-- 46. billing_confirmations（会計医師確認タブ: medical_recordsと1:1）
-- ------------------------------------
CREATE TABLE billing_confirmations (
    id                BIGSERIAL          PRIMARY KEY,
    medical_record_id bigint             NOT NULL UNIQUE REFERENCES medical_records(id) ON DELETE CASCADE,
    status            confirmation_status         DEFAULT 'pending',
    confirmed_by      bigint                         REFERENCES staffs(id) ON DELETE SET NULL,
    confirmed_at      timestamptz,
    returned_by       bigint                         REFERENCES staffs(id) ON DELETE SET NULL,
    returned_at       timestamptz,
    return_reason     text                  NOT NULL DEFAULT '',
    memo              text                  NOT NULL DEFAULT '',
    created_at        timestamptz           NOT NULL DEFAULT now(),
    updated_at        timestamptz           NOT NULL DEFAULT now()
);

-- ------------------------------------
-- 47. estimates（見積書）
-- ------------------------------------
CREATE TABLE estimates (
    id                BIGSERIAL       PRIMARY KEY,
    clinic_id         bigint          NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    estimate_no       text            NOT NULL,
    medical_record_id bigint                   REFERENCES medical_records(id) ON DELETE RESTRICT,
    title             text            NOT NULL DEFAULT '',
    owner_id          bigint                   REFERENCES owners(id) ON DELETE SET NULL,
    status            estimate_status          DEFAULT 'draft',
    subtotal          bigint          NOT NULL DEFAULT 0,
    tax_total         bigint          NOT NULL DEFAULT 0,
    total_amount      bigint          NOT NULL DEFAULT 0,
    insurance_amount  bigint                   DEFAULT 0,
    discount_amount   bigint                   DEFAULT 0,
    valid_until       date,
    comment           text            NOT NULL DEFAULT '',
    notes             text            NOT NULL DEFAULT '',
    created_by        bigint                   REFERENCES staffs(id) ON DELETE SET NULL,
    created_at        timestamptz     NOT NULL DEFAULT now(),
    updated_at        timestamptz     NOT NULL DEFAULT now(),
    deleted_at        timestamptz
);

-- ------------------------------------
-- 48. exam_results（検査結果明細）
-- ------------------------------------
CREATE TABLE exam_results (
    id                BIGSERIAL                  PRIMARY KEY,
    exam_id           bigint                     NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    exam_type_field_id bigint                             REFERENCES exam_type_fields(id) ON DELETE SET NULL,
    name              text                       NOT NULL DEFAULT '',
    inspection_value  text                       NOT NULL DEFAULT '',
    normal_value      text                       NOT NULL DEFAULT '',
    result            text                       NOT NULL DEFAULT '',
    unit              text                       NOT NULL DEFAULT '',
    reference_value   text                       NOT NULL DEFAULT '',
    ref_min           decimal(10,4),
    ref_max           decimal(10,4),
    is_abnormal       boolean                             DEFAULT false,
    status            exam_result_status                  DEFAULT 'normal',
    sort_order        integer                             DEFAULT 0,
    created_at        timestamptz                NOT NULL DEFAULT now(),
    updated_at        timestamptz                NOT NULL DEFAULT now()
);

-- ------------------------------------
-- 49. daily_records（入院日次記録）
-- ------------------------------------
CREATE TABLE daily_records (
    id                 BIGSERIAL   PRIMARY KEY,
    hospitalization_id bigint      NOT NULL REFERENCES hospitalizations(id) ON DELETE CASCADE,
    clinic_id          bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    date               date        NOT NULL,
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now()
);

-- ------------------------------------
-- ------------------------------------
-- 50. vital_records（バイタル記録: 外来・入院統合）
--      daily_records 依存のためここに定義
-- ------------------------------------
CREATE TABLE vital_records (
    id                BIGSERIAL   PRIMARY KEY,
    clinic_id         bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    pet_id            bigint      NOT NULL REFERENCES pets(id) ON DELETE RESTRICT,
    medical_record_id bigint               REFERENCES medical_records(id) ON DELETE CASCADE,  -- 外来時
    daily_record_id   bigint               REFERENCES daily_records(id) ON DELETE CASCADE,    -- 入院時
    recorded_at       timestamptz NOT NULL DEFAULT now(),
    staff_id          bigint               REFERENCES staffs(id) ON DELETE SET NULL,
    temperature       numeric,
    heart_rate        integer,
    respiration_rate  integer,
    weight            numeric,
    weight_unit       body_weight_unit     DEFAULT 'Kg',
    notes             text        NOT NULL DEFAULT '',
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now(),
    deleted_at        timestamptz,
    CONSTRAINT chk_vital_records_context CHECK (
        (medical_record_id IS NOT NULL) OR (daily_record_id IS NOT NULL)
    ),
    CONSTRAINT chk_vital_temperature CHECK (temperature IS NULL OR (temperature >= 30.0 AND temperature <= 50.0)),
    CONSTRAINT chk_vital_heart_rate CHECK (heart_rate IS NULL OR (heart_rate > 0 AND heart_rate < 500)),
    CONSTRAINT chk_vital_respiration CHECK (respiration_rate IS NULL OR (respiration_rate > 0 AND respiration_rate < 200)),
    CONSTRAINT chk_vital_weight CHECK (weight IS NULL OR weight > 0)
);

-- ------------------------------------
-- 51. care_plan_items（ケアプラン項目）
-- ------------------------------------
CREATE TABLE care_plan_items (
    id                      BIGSERIAL        PRIMARY KEY,
    hospitalization_id      bigint           NOT NULL REFERENCES hospitalizations(id) ON DELETE CASCADE,
    type                    care_plan_type   NOT NULL,
    name                    text             NOT NULL DEFAULT '',
    description             text             NOT NULL DEFAULT '',
    timing                  plan_timing[]             DEFAULT '{}',
    status                  care_plan_status          DEFAULT 'active',
    notes                   text             NOT NULL DEFAULT '',
    medicine_id             bigint                    REFERENCES medicines(id) ON DELETE SET NULL,
    procedure_id            bigint                    REFERENCES procedures(id) ON DELETE SET NULL,
    hospitalization_plan_id bigint                    REFERENCES hospitalization_plans(id) ON DELETE SET NULL,
    unit_price              bigint                    DEFAULT 0,
    category                text             NOT NULL DEFAULT '',
    sort_order              integer                   DEFAULT 0,
    created_at              timestamptz      NOT NULL DEFAULT now(),
    updated_at              timestamptz      NOT NULL DEFAULT now(),
    CONSTRAINT chk_care_plan_item_ref CHECK (
        (type = 'medicine'    AND medicine_id IS NOT NULL) OR
        (type = 'treatment'   AND procedure_id IS NOT NULL) OR
        (type = 'item'        AND hospitalization_plan_id IS NOT NULL) OR
        (type IN ('food', 'instruction'))
    )
);

-- ==========================================================================
-- レイヤー6: estimates/treatments等依存
-- ==========================================================================

-- ------------------------------------
-- 52. merchandise_items（物販・フード・その他マスタ）
-- ------------------------------------
CREATE TABLE merchandise_items (
    id          BIGSERIAL     PRIMARY KEY,
    clinic_id   bigint        NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name        text          NOT NULL DEFAULT '',
    category    item_category NOT NULL DEFAULT 'goods',
    unit_price  bigint        NOT NULL DEFAULT 0,
    tax_type    tax_type      NOT NULL DEFAULT 'excluded',
    tax_rate    numeric       NOT NULL DEFAULT 0.10,
    is_active   boolean       NOT NULL DEFAULT true,
    sort_order  integer       NOT NULL DEFAULT 0,
    created_at  timestamptz   NOT NULL DEFAULT now(),
    updated_at  timestamptz   NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

-- ------------------------------------
-- 53. estimate_items（見積明細）
-- ------------------------------------
CREATE TABLE estimate_items (
    id                      BIGSERIAL     PRIMARY KEY,
    estimate_id             bigint        NOT NULL REFERENCES estimates(id) ON DELETE CASCADE,
    name                    text          NOT NULL DEFAULT '',
    category                item_category NOT NULL,
    unit_price              bigint        NOT NULL DEFAULT 0,
    quantity                numeric(10,1) NOT NULL DEFAULT 1,
    tax_type                tax_type               NOT NULL DEFAULT 'excluded',
    tax_rate                numeric(3,2)           DEFAULT 0.10,
    discount_rate           numeric(5,2)           DEFAULT 0,
    discount_amount         bigint                 DEFAULT 0,
    is_insurance_applicable boolean                DEFAULT false,
    consultation_id         bigint                 REFERENCES consultations(id) ON DELETE SET NULL,
    procedure_id            bigint                 REFERENCES procedures(id) ON DELETE SET NULL,
    medicine_id             bigint                 REFERENCES medicines(id) ON DELETE SET NULL,
    merchandise_item_id     bigint                 CONSTRAINT fk_estimate_items_merchandise REFERENCES merchandise_items(id) ON DELETE SET NULL,
    sort_order              integer                DEFAULT 0,
    created_at              timestamptz   NOT NULL DEFAULT now(),
    updated_at              timestamptz   NOT NULL DEFAULT now(),
    deleted_at              timestamptz,
    CONSTRAINT chk_estimate_item_quantity CHECK (quantity > 0)
);

-- ------------------------------------
-- 54. care_logs（ケアログ）
-- ------------------------------------
CREATE TABLE care_logs (
    id              BIGSERIAL       PRIMARY KEY,
    clinic_id       bigint          NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    daily_record_id bigint          NOT NULL REFERENCES daily_records(id) ON DELETE CASCADE,
    time            time            NOT NULL,
    type            care_log_type   NOT NULL,
    status          care_log_status NOT NULL DEFAULT 'completed',
    value           text            NOT NULL DEFAULT '',
    staff_id        bigint                   REFERENCES staffs(id) ON DELETE SET NULL,
    notes           text            NOT NULL DEFAULT '',
    created_at      timestamptz     NOT NULL DEFAULT now(),
    updated_at      timestamptz     NOT NULL DEFAULT now()
);

-- ------------------------------------
-- Note: vital_records は 50 に統合済み

-- ------------------------------------
-- 55. staff_notes（スタッフノート）
-- ------------------------------------
CREATE TABLE staff_notes (
    id              BIGSERIAL   PRIMARY KEY,
    daily_record_id bigint      NOT NULL REFERENCES daily_records(id) ON DELETE CASCADE,
    time            time        NOT NULL,
    content         text        NOT NULL DEFAULT '',
    staff_id        bigint               REFERENCES staffs(id) ON DELETE SET NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

-- ==========================================================================
-- レイヤー7: billings
-- ==========================================================================

-- ------------------------------------
-- 56. billings（会計）
-- ------------------------------------
CREATE TABLE billings (
    id                 BIGSERIAL      PRIMARY KEY,
    clinic_id          bigint         NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    medical_record_id  bigint                  REFERENCES medical_records(id) ON DELETE SET NULL,
    hospitalization_id bigint                  REFERENCES hospitalizations(id) ON DELETE SET NULL,
    owner_id           bigint                  REFERENCES owners(id) ON DELETE SET NULL,
    pet_id             bigint                  REFERENCES pets(id) ON DELETE SET NULL,
    subtotal           bigint         NOT NULL DEFAULT 0,
    tax_total          bigint         NOT NULL DEFAULT 0,
    total_amount       bigint         NOT NULL DEFAULT 0,
    has_insurance      boolean        NOT NULL DEFAULT false,
    status             billing_status          DEFAULT 'waiting',
    scheduled_date     date           NOT NULL,
    completed_at       timestamptz,
    memo               text           NOT NULL DEFAULT '',
    created_at         timestamptz    NOT NULL DEFAULT now(),
    updated_at         timestamptz    NOT NULL DEFAULT now(),
    deleted_at         timestamptz,
    CONSTRAINT chk_billings_amounts CHECK (subtotal >= 0 AND tax_total >= 0 AND total_amount >= 0)
);

-- ------------------------------------
-- 57. billing_items（会計明細）
-- ------------------------------------
CREATE TABLE billing_items (
    id                      BIGSERIAL     PRIMARY KEY,
    billing_id              bigint        NOT NULL REFERENCES billings(id) ON DELETE CASCADE,
    category                item_category NOT NULL,
    name                    text          NOT NULL DEFAULT '',
    unit_price              bigint        NOT NULL DEFAULT 0,
    quantity                numeric(10,1) NOT NULL DEFAULT 1,
    tax_type                tax_type               NOT NULL DEFAULT 'excluded',
    tax_rate                numeric(3,2)           DEFAULT 0.10,
    is_insurance_applicable boolean                DEFAULT false,
    source                  item_source            DEFAULT 'manual',
    merchandise_item_id     bigint                 CONSTRAINT fk_billing_items_merchandise REFERENCES merchandise_items(id) ON DELETE SET NULL,
    treatment_id            bigint                 REFERENCES treatments(id) ON DELETE SET NULL,
    appointment_id          bigint                 REFERENCES appointments(id) ON DELETE SET NULL,
    trimming_course_id      bigint                 REFERENCES trimming_courses(id) ON DELETE SET NULL,
    trimming_option_id      bigint                 REFERENCES trimming_options(id) ON DELETE SET NULL,
    discount_rate           numeric(5,2)           NOT NULL DEFAULT 0,
    discount_amount         bigint                 NOT NULL DEFAULT 0,
    sort_order              integer                DEFAULT 0,
    created_at              timestamptz   NOT NULL DEFAULT now(),
    updated_at              timestamptz   NOT NULL DEFAULT now(),
    deleted_at              timestamptz,
    CONSTRAINT chk_billing_item_quantity CHECK (quantity > 0)
);

-- ------------------------------------
-- 62c. payment_methods（支払方法マスタ — 旧 payment_method enum のマスタ化）
-- ------------------------------------
CREATE TABLE payment_methods (
    id             BIGSERIAL    PRIMARY KEY,
    clinic_id      bigint       NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    name           varchar(50)  NOT NULL,
    -- 009: 安定識別子（Issue #197）
    system_key     varchar(50),
    display_order  integer      NOT NULL DEFAULT 0,
    is_active      boolean      NOT NULL DEFAULT true,
    created_at     timestamptz  NOT NULL DEFAULT now(),
    updated_at     timestamptz  NOT NULL DEFAULT now(),
    deleted_at     timestamptz
);

CREATE UNIQUE INDEX idx_payment_methods_clinic_name ON payment_methods(clinic_id, name) WHERE deleted_at IS NULL;
CREATE INDEX idx_payment_methods_clinic_order ON payment_methods(clinic_id, display_order) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_payment_methods_clinic_system_key
    ON payment_methods (clinic_id, system_key)
    WHERE system_key IS NOT NULL AND deleted_at IS NULL;

COMMENT ON TABLE payment_methods IS '支払方法マスタ（FEAT-368: payment_method enum のマスタ化）';

-- ------------------------------------
-- 58. payments（支払い: billingsと1:1）
-- ------------------------------------
CREATE TABLE payments (
    id               BIGSERIAL      PRIMARY KEY,
    billing_id       bigint         NOT NULL UNIQUE REFERENCES billings(id) ON DELETE CASCADE,
    subtotal         bigint         NOT NULL DEFAULT 0,
    tax_total        bigint         NOT NULL DEFAULT 0,
    total_amount     bigint         NOT NULL DEFAULT 0,
    insurance_name   text           NOT NULL DEFAULT '',
    insurance_ratio  numeric(3,2)            DEFAULT 0,  -- 保険比率は小数（例: 0.7）
    insurance_amount bigint                  DEFAULT 0,
    discount_amount  bigint                  DEFAULT 0,
    billing_amount   bigint         NOT NULL DEFAULT 0,
    received_amount  bigint                  DEFAULT 0,
    change_amount    bigint                  DEFAULT 0,
    method           payment_method          DEFAULT 'cash',
    payment_method_id bigint                 REFERENCES payment_methods(id),
    paid_by          bigint         REFERENCES staffs(id),
    created_at       timestamptz    NOT NULL DEFAULT now(),
    updated_at       timestamptz    NOT NULL DEFAULT now(),
    deleted_at       timestamptz
);

-- ------------------------------------
-- 58b. payment_splits（混在支払い明細）
-- ------------------------------------
CREATE TABLE payment_splits (
    id                bigserial    PRIMARY KEY,
    clinic_id         bigint       NOT NULL,
    billing_id        bigint       NOT NULL REFERENCES billings(id) ON DELETE RESTRICT,
    method            payment_method NOT NULL,
    payment_method_id bigint       REFERENCES payment_methods(id),
    amount            bigint       NOT NULL DEFAULT 0,
    received_amount   bigint       NOT NULL DEFAULT 0,
    change_amount     bigint       NOT NULL DEFAULT 0,
    paid_by           bigint       REFERENCES staffs(id),
    created_at        timestamptz  NOT NULL DEFAULT now(),
    CONSTRAINT fk_payment_splits_clinic_id
    FOREIGN KEY (clinic_id)
    REFERENCES clinics (id)
    ON DELETE RESTRICT
);

CREATE INDEX idx_payment_splits_clinic_billing ON payment_splits(clinic_id, billing_id);

-- ------------------------------------
-- 59. billing_refunds（返金レコード）
-- ------------------------------------
CREATE TABLE billing_refunds (
    id           BIGSERIAL   PRIMARY KEY,
    clinic_id    bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    billing_id   bigint      NOT NULL REFERENCES billings(id) ON DELETE RESTRICT,
    amount       bigint      NOT NULL CHECK (amount > 0),
    reason       text        NOT NULL DEFAULT '',
    refunded_by     bigint          REFERENCES staffs(id),
    refunded_at     timestamptz     NOT NULL DEFAULT now(),
    payment_method  payment_method,
    created_at      timestamptz     NOT NULL DEFAULT now()
);

-- ------------------------------------
-- 60. shift_entries（シフト管理）
-- ------------------------------------
CREATE TABLE shift_entries (
    id         BIGSERIAL   PRIMARY KEY,
    clinic_id  bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    staff_id   bigint      NOT NULL REFERENCES staffs(id) ON DELETE RESTRICT,
    date       date        NOT NULL,
    shift_type shift_type  NOT NULL,
    start_time time,
    end_time   time,
    notes      text        NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uk_shift_staff_date UNIQUE (clinic_id, staff_id, date)
);

-- ------------------------------------
-- 61. clinic_holidays（個別休診日）
-- ------------------------------------
CREATE TABLE clinic_holidays (
    id         BIGSERIAL   PRIMARY KEY,
    clinic_id  bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    date       date        NOT NULL,
    reason     text        NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uk_clinic_holidays_clinic_date UNIQUE (clinic_id, date)
);
CREATE INDEX idx_clinic_holidays_clinic_date ON clinic_holidays(clinic_id, date);

-- ------------------------------------
-- 62a. clinic_settings（医院締め時間設定 — clinic_id PK で 1:1）
-- ------------------------------------
CREATE TABLE clinic_settings (
    clinic_id              bigint       PRIMARY KEY REFERENCES clinics(id) ON DELETE CASCADE,
    closing_am_pm_boundary time         NOT NULL DEFAULT '14:00',
    closing_weekday_end    time         NOT NULL DEFAULT '18:30',
    closing_sunday_end     time         NOT NULL DEFAULT '17:30',
    closed_weekdays        smallint[]   NOT NULL DEFAULT '{}',
    cpm_version            varchar(8)   NOT NULL DEFAULT 'v1'
                           CHECK (cpm_version IN ('v1', 'v2')),
    dormant_prevention_180_days integer  NOT NULL DEFAULT 180,
    dormant_prevention_210_days integer  NOT NULL DEFAULT 210,
    dormant_prevention_240_days integer  NOT NULL DEFAULT 240,
    dormant_prevention_365_days integer  NOT NULL DEFAULT 365,
    -- mig-011: CPM V2 来院回数閾値 (clinic 単位調整可能)
    cpm_v2_coming_threshold  INT NOT NULL DEFAULT 2  CHECK (cpm_v2_coming_threshold  >= 1),
    cpm_v2_good_threshold    INT NOT NULL DEFAULT 4  CHECK (cpm_v2_good_threshold    >= 1),
    cpm_v2_family_threshold  INT NOT NULL DEFAULT 8  CHECK (cpm_v2_family_threshold  >= 1),
    cpm_v2_noah_threshold    INT NOT NULL DEFAULT 13 CHECK (cpm_v2_noah_threshold    >= 1),
    -- mig-012: CPM V1 判定閾値 (clinic 単位調整可能)
    cpm_v1_dormant_days       INT     NOT NULL DEFAULT 240    CHECK (cpm_v1_dormant_days       >= 1),
    cpm_v1_noah_days          INT     NOT NULL DEFAULT 365    CHECK (cpm_v1_noah_days          >= 1),
    cpm_v1_noah_annual_visits INT     NOT NULL DEFAULT 3      CHECK (cpm_v1_noah_annual_visits >= 1),
    cpm_v1_noah_ltv           BIGINT  NOT NULL DEFAULT 80000  CHECK (cpm_v1_noah_ltv           >= 0),
    cpm_v1_core_days          INT     NOT NULL DEFAULT 180    CHECK (cpm_v1_core_days          >= 1),
    cpm_v1_core_annual_visits INT     NOT NULL DEFAULT 2      CHECK (cpm_v1_core_annual_visits >= 1),
    cpm_v1_core_ltv           BIGINT  NOT NULL DEFAULT 50000  CHECK (cpm_v1_core_ltv           >= 0),
    cpm_v1_spot_min_amount    BIGINT  NOT NULL DEFAULT 30000  CHECK (cpm_v1_spot_min_amount    >= 0),
    cpm_v1_spot_inactive_days INT     NOT NULL DEFAULT 90     CHECK (cpm_v1_spot_inactive_days >= 1),
    cpm_v1_growing_max_days   INT     NOT NULL DEFAULT 90     CHECK (cpm_v1_growing_max_days   >= 1),
    cpm_v1_growing_min_visits INT     NOT NULL DEFAULT 2      CHECK (cpm_v1_growing_min_visits >= 1),
    cpm_v1_growing_max_visits INT     NOT NULL DEFAULT 3      CHECK (cpm_v1_growing_max_visits >= 1),
    cpm_v1_ltv_break_low      BIGINT  NOT NULL DEFAULT 20000  CHECK (cpm_v1_ltv_break_low      >= 0),
    -- mig-013: 健診・予防タグ判定閾値 (clinic 単位調整可能)
    health_prevention_lookback_days INT NOT NULL DEFAULT 365,
    vaccine_deadline_days           INT NOT NULL DEFAULT 60,
    created_at             timestamptz  NOT NULL DEFAULT now(),
    updated_at             timestamptz  NOT NULL DEFAULT now()
);

COMMENT ON TABLE clinic_settings IS '医院締め時間・休診曜日設定（FEAT-368）';
COMMENT ON COLUMN clinic_settings.cpm_version IS 'CPM 判定方式 (v1: 既存 6-stage / v2: Q19 来院回数 5-stage, 2026-05-08 確定)';
COMMENT ON COLUMN clinic_settings.dormant_prevention_180_days IS 'dormant_prevention_1st 配信トリガー閾値日数 (Q21、デフォルト 180)';
COMMENT ON COLUMN clinic_settings.dormant_prevention_210_days IS 'dormant_prevention_2nd 配信トリガー閾値日数 (Q21、デフォルト 210)';
COMMENT ON COLUMN clinic_settings.dormant_prevention_240_days IS 'dormant_prevention_3rd 配信トリガー閾値日数 (Q21、デフォルト 240)';
COMMENT ON COLUMN clinic_settings.dormant_prevention_365_days IS 'dormant_prevention_4th 配信トリガー閾値日数 (Q21、デフォルト 365)';
COMMENT ON COLUMN clinic_settings.cpm_v2_coming_threshold  IS 'CPM V2 これから ステージ開始来院回数 (デフォルト 2)';
COMMENT ON COLUMN clinic_settings.cpm_v2_good_threshold    IS 'CPM V2 いいかんじ ステージ開始来院回数 (デフォルト 4)';
COMMENT ON COLUMN clinic_settings.cpm_v2_family_threshold  IS 'CPM V2 ファミリー ステージ開始来院回数 (デフォルト 8)';
COMMENT ON COLUMN clinic_settings.cpm_v2_noah_threshold    IS 'CPM V2 ノア ステージ開始来院回数 (デフォルト 13)';
COMMENT ON COLUMN clinic_settings.cpm_v1_dormant_days       IS 'CPM V1 cpm_dormant: 最終来院からの経過日数 >= この値で dormant 判定 (デフォルト 240)';
COMMENT ON COLUMN clinic_settings.cpm_v1_noah_days          IS 'CPM V1 cpm_noah: 初来院からの経過日数 >= この値 (デフォルト 365)';
COMMENT ON COLUMN clinic_settings.cpm_v1_noah_annual_visits IS 'CPM V1 cpm_noah: 年間来院回数 >= この値 (デフォルト 3)';
COMMENT ON COLUMN clinic_settings.cpm_v1_noah_ltv           IS 'CPM V1 cpm_noah: 累計金額 >= この値 (デフォルト 80000)';
COMMENT ON COLUMN clinic_settings.cpm_v1_core_days          IS 'CPM V1 cpm_core: 初来院からの経過日数 >= この値 (デフォルト 180)';
COMMENT ON COLUMN clinic_settings.cpm_v1_core_annual_visits IS 'CPM V1 cpm_core: 年間来院回数 >= この値 (デフォルト 2)';
COMMENT ON COLUMN clinic_settings.cpm_v1_core_ltv           IS 'CPM V1 cpm_core: 累計金額 >= この値; growing 上限にも兼用 (デフォルト 50000)';
COMMENT ON COLUMN clinic_settings.cpm_v1_spot_min_amount    IS 'CPM V1 cpm_spot: 単回最大金額 >= この値 (デフォルト 30000)';
COMMENT ON COLUMN clinic_settings.cpm_v1_spot_inactive_days IS 'CPM V1 cpm_spot: 最終来院からの経過日数 > この値 (デフォルト 90)';
COMMENT ON COLUMN clinic_settings.cpm_v1_growing_max_days   IS 'CPM V1 cpm_growing: 初来院からの経過日数 <= この値 (デフォルト 90)';
COMMENT ON COLUMN clinic_settings.cpm_v1_growing_min_visits IS 'CPM V1 cpm_growing: 総来院回数 >= この値 (デフォルト 2)';
COMMENT ON COLUMN clinic_settings.cpm_v1_growing_max_visits IS 'CPM V1 cpm_growing: 総来院回数 <= この値 (デフォルト 3)';
COMMENT ON COLUMN clinic_settings.cpm_v1_ltv_break_low      IS 'CPM V1 growing 下限 / encounter 上限境界 (デフォルト 20000)';
COMMENT ON COLUMN clinic_settings.health_prevention_lookback_days IS '健診・予防履歴の参照期間日数 (mig-013、デフォルト 365)';
COMMENT ON COLUMN clinic_settings.vaccine_deadline_days           IS 'ワクチン期限間近とみなす残日数 (mig-013、デフォルト 60)';

-- ------------------------------------
-- 62b. closing_special_periods（特別期間: 年末年始・お盆等）
-- ------------------------------------
CREATE TABLE closing_special_periods (
    id             BIGSERIAL    PRIMARY KEY,
    clinic_id      bigint       NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    start_date     date         NOT NULL,
    end_date       date         NOT NULL,
    am_pm_boundary time         NOT NULL,
    pm_end         time         NOT NULL,
    note           varchar(100) NOT NULL DEFAULT '',
    created_at     timestamptz  NOT NULL DEFAULT now(),
    updated_at     timestamptz  NOT NULL DEFAULT now(),
    CONSTRAINT chk_closing_special_periods_date_range CHECK (start_date <= end_date),
    CONSTRAINT chk_closing_special_periods_time_order CHECK (am_pm_boundary < pm_end)
);

CREATE INDEX idx_closing_special_periods_clinic ON closing_special_periods(clinic_id, start_date, end_date);

COMMENT ON TABLE closing_special_periods IS '特別診療時間設定（年末年始・お盆等, FEAT-368）';

-- ------------------------------------
-- 62d. cash_register_closes（レジ締めレコード）
-- ------------------------------------
CREATE TABLE cash_register_closes (
    id                      BIGSERIAL    PRIMARY KEY,
    clinic_id               bigint       NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    close_date              date         NOT NULL,
    period                  varchar(3)   NOT NULL CHECK (period IN ('am', 'pm', 'emg')),
    theoretical_cash        bigint       NOT NULL DEFAULT 0,
    actual_cash             bigint       NOT NULL DEFAULT 0,
    cash_difference         bigint       NOT NULL DEFAULT 0,
    category_breakdown      jsonb        NOT NULL DEFAULT '{}',
    memo                    text         NOT NULL DEFAULT '',
    closed_by               bigint       REFERENCES staffs(id),
    closed_at               timestamptz  NOT NULL DEFAULT now(),
    created_at              timestamptz  NOT NULL DEFAULT now(),
    updated_at              timestamptz  NOT NULL DEFAULT now(),
    deleted_at              timestamptz
);

CREATE INDEX idx_cash_register_closes_clinic ON cash_register_closes(clinic_id, close_date DESC);
CREATE UNIQUE INDEX uq_cash_register_closes_date_period ON cash_register_closes(clinic_id, close_date, period) WHERE deleted_at IS NULL;

COMMENT ON TABLE cash_register_closes IS 'レジ締めレコード（FEAT-368）';


CREATE INDEX idx_billing_items_merchandise_item_id  ON billing_items(merchandise_item_id)  WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_items_treatment_id         ON billing_items(treatment_id)         WHERE treatment_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_billing_items_appointment_id       ON billing_items(appointment_id)       WHERE appointment_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_billing_items_trimming_course_id   ON billing_items(trimming_course_id)   WHERE trimming_course_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_billing_items_trimming_option_id   ON billing_items(trimming_option_id)   WHERE trimming_option_id IS NOT NULL AND deleted_at IS NULL;

CREATE INDEX idx_estimate_items_merchandise_item_id ON estimate_items(merchandise_item_id);

-- =============================================================================
-- 4. インデックス定義
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 4.1 UNIQUE制約・インデックス
-- -----------------------------------------------------------------------------

-- カルテ番号の一意性（医院スコープ）
CREATE UNIQUE INDEX idx_medical_records_clinic_record_no ON medical_records(clinic_id, record_no);

-- 見積書番号の一意性（医院スコープ）
CREATE UNIQUE INDEX idx_estimates_clinic_estimate_no ON estimates(clinic_id, estimate_no);

-- 入院日次記録: 同一入院の同一日付は1件のみ
CREATE UNIQUE INDEX idx_daily_records_hosp_date ON daily_records(hospitalization_id, date);

-- 飼主: clinic内でemail重複不可（論理削除を除く・空文字除く）
CREATE UNIQUE INDEX uk_owners_clinic_email ON owners(clinic_id, email) WHERE deleted_at IS NULL AND email IS NOT NULL AND email != '';

-- billings: medical_record_idがある場合は1対1
CREATE UNIQUE INDEX idx_billings_medical_record_id_unique ON billings(medical_record_id) WHERE medical_record_id IS NOT NULL;

-- billings: hospitalization_idがある場合は active 行で1対1（退院会計の二重永続化防止）
-- soft-delete 後の再作成を許すため deleted_at IS NULL を述語に含める（medical_record 側との意図的非対称）
CREATE UNIQUE INDEX idx_billings_hospitalization_id_unique
  ON billings(hospitalization_id)
  WHERE hospitalization_id IS NOT NULL AND deleted_at IS NULL;

-- -----------------------------------------------------------------------------
-- 4.4 基本FKインデックス
-- -----------------------------------------------------------------------------

-- マスタテーブル clinic_id
CREATE INDEX idx_occupations_clinic_id ON occupations(clinic_id);
CREATE INDEX idx_inventory_items_clinic_id ON inventory_items(clinic_id);
CREATE INDEX idx_exam_types_clinic_id ON exam_types(clinic_id);
CREATE INDEX idx_exam_types_parent_id ON exam_types(parent_id);
CREATE INDEX idx_vaccines_clinic_id ON vaccines(clinic_id);
CREATE INDEX idx_vaccines_parent_id ON vaccines(parent_id);
CREATE INDEX idx_medicines_clinic_id ON medicines(clinic_id);
CREATE INDEX idx_medicines_parent_id ON medicines(parent_id);
CREATE INDEX idx_insurances_clinic_id ON insurances(clinic_id);
CREATE INDEX idx_cages_clinic_id ON cages(clinic_id);
CREATE INDEX idx_reservation_types_clinic_id ON reservation_types(clinic_id);
CREATE INDEX idx_consultations_clinic_id ON consultations(clinic_id);
CREATE INDEX idx_consultations_parent_id ON consultations(parent_id);
CREATE INDEX idx_procedures_clinic_id ON procedures(clinic_id);
CREATE INDEX idx_procedures_parent_id ON procedures(parent_id);
CREATE INDEX idx_hospitalization_plans_clinic_id ON hospitalization_plans(clinic_id);
CREATE INDEX idx_trimming_courses_clinic_id ON trimming_courses(clinic_id);
CREATE INDEX idx_trimming_options_clinic_id ON trimming_options(clinic_id);
CREATE INDEX idx_diagnosis_types_clinic_id ON diagnosis_types(clinic_id);
CREATE INDEX idx_diagnosis_names_clinic_id ON diagnosis_names(clinic_id);
CREATE INDEX idx_checkup_types_clinic_id ON checkup_types(clinic_id);
CREATE INDEX idx_checkup_types_parent_id ON checkup_types(parent_id);
CREATE INDEX idx_checkup_types_deleted_at ON checkup_types(deleted_at);
CREATE INDEX idx_chief_complaint_types_clinic_id ON chief_complaint_types(clinic_id);
CREATE INDEX idx_inquiry_templates_clinic_id ON inquiry_templates(clinic_id);
CREATE INDEX idx_inquiry_templates_clinic_category ON inquiry_templates(clinic_id, category);

-- コアテーブル clinic_id
CREATE INDEX idx_owners_clinic_id ON owners(clinic_id);
CREATE INDEX idx_pets_clinic_id ON pets(clinic_id);

-- 診療テーブル clinic_id
CREATE INDEX idx_appointments_clinic_id ON appointments(clinic_id);
CREATE INDEX idx_hospitalizations_clinic_id ON hospitalizations(clinic_id);
CREATE INDEX idx_billings_clinic_id ON billings(clinic_id);
CREATE INDEX idx_shift_entries_clinic_id ON shift_entries(clinic_id);

-- merchandise_items インデックス
CREATE INDEX idx_merchandise_items_clinic ON merchandise_items(clinic_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_merchandise_items_category ON merchandise_items(clinic_id, category) WHERE deleted_at IS NULL;
CREATE INDEX idx_merchandise_items_sort ON merchandise_items(clinic_id, sort_order);

-- マスタ一覧取得最適化 (clinic_id, sort_order) 複合インデックス
CREATE INDEX idx_vaccines_clinic_sort          ON vaccines(clinic_id, sort_order)               WHERE deleted_at IS NULL;
CREATE INDEX idx_medicines_clinic_sort         ON medicines(clinic_id, sort_order)               WHERE deleted_at IS NULL;
CREATE INDEX idx_exam_types_clinic_sort        ON exam_types(clinic_id, sort_order)              WHERE deleted_at IS NULL;
CREATE INDEX idx_procedures_clinic_sort        ON procedures(clinic_id, sort_order)              WHERE deleted_at IS NULL;
CREATE INDEX idx_cages_clinic_sort             ON cages(clinic_id, sort_order)                   WHERE deleted_at IS NULL;
CREATE INDEX idx_checkup_types_clinic_sort     ON checkup_types(clinic_id, sort_order)           WHERE deleted_at IS NULL;
CREATE INDEX idx_chief_complaints_clinic_sort  ON chief_complaint_types(clinic_id, sort_order)   WHERE deleted_at IS NULL;
CREATE INDEX idx_diagnosis_types_clinic_sort   ON diagnosis_types(clinic_id, sort_order)         WHERE deleted_at IS NULL;
CREATE INDEX idx_diagnosis_names_clinic_sort   ON diagnosis_names(clinic_id, sort_order)         WHERE deleted_at IS NULL;
CREATE INDEX idx_trimming_courses_clinic_sort  ON trimming_courses(clinic_id, sort_order)        WHERE deleted_at IS NULL;
CREATE INDEX idx_trimming_options_clinic_sort  ON trimming_options(clinic_id, sort_order)        WHERE deleted_at IS NULL;
CREATE INDEX idx_insurances_clinic_sort        ON insurances(clinic_id, sort_order)              WHERE deleted_at IS NULL;
CREATE INDEX idx_occupations_clinic_sort       ON occupations(clinic_id, sort_order)             WHERE deleted_at IS NULL;
CREATE INDEX idx_reservation_types_clinic_sort ON reservation_types(clinic_id, sort_order)       WHERE deleted_at IS NULL;
CREATE INDEX idx_reservation_type_groups_sort  ON reservation_type_groups(clinic_id, sort_order) WHERE deleted_at IS NULL;

-- 予約 FK インデックス
CREATE INDEX idx_appointments_owner_id ON appointments(owner_id);
CREATE INDEX idx_appointments_pet_id ON appointments(pet_id);
CREATE INDEX idx_appointments_reservation_type_id ON appointments(reservation_type_id);
CREATE INDEX idx_appointments_doctor_id ON appointments(doctor_id);
CREATE INDEX idx_appointments_created_by ON appointments(created_by);

-- medical_records 子テーブル FK インデックス
CREATE INDEX idx_treatments_medical_record_id ON treatments(medical_record_id);
CREATE INDEX idx_vital_records_medical_record_id ON vital_records(medical_record_id);
CREATE INDEX idx_vital_records_daily_record_id ON vital_records(daily_record_id);
CREATE INDEX idx_vital_records_pet_id ON vital_records(pet_id);
CREATE INDEX idx_vital_records_clinic_id ON vital_records(clinic_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_exams_medical_record_id ON exams(medical_record_id);
CREATE INDEX idx_exams_pet_id ON exams(pet_id);
CREATE INDEX idx_exams_exam_type_id ON exams(exam_type_id);
CREATE INDEX idx_vaccinations_clinic_id ON vaccinations(clinic_id);
CREATE INDEX idx_vaccinations_medical_record_id ON vaccinations(medical_record_id);
CREATE INDEX idx_vaccinations_pet_id ON vaccinations(pet_id);
CREATE INDEX idx_vaccinations_vaccine_id ON vaccinations(vaccine_id);
CREATE INDEX idx_checkups_medical_record_id ON checkups(medical_record_id);
CREATE INDEX idx_checkups_pet_id ON checkups(pet_id);
CREATE INDEX idx_checkups_checkup_type_id ON checkups(checkup_type_id);
CREATE INDEX idx_clinical_plans_medical_record_id ON clinical_plans(medical_record_id);
CREATE INDEX idx_inquiries_medical_record_id ON inquiries(medical_record_id);
CREATE INDEX idx_medical_record_images_medical_record_id ON medical_record_images(medical_record_id);
CREATE INDEX idx_treatment_plans_medical_record_id ON treatment_plans(medical_record_id);
CREATE INDEX idx_treatment_plans_hospitalization_id ON treatment_plans(hospitalization_id);
CREATE INDEX idx_treatment_plans_clinic_id ON treatment_plans(clinic_id) WHERE deleted_at IS NULL;

-- hospitalization 子テーブル FK インデックス
CREATE INDEX idx_hospitalizations_pet_id ON hospitalizations(pet_id);
CREATE INDEX idx_hospitalizations_owner_id ON hospitalizations(owner_id);
CREATE INDEX idx_hospitalizations_cage_id ON hospitalizations(cage_id);
CREATE INDEX idx_care_plan_items_hospitalization_id ON care_plan_items(hospitalization_id);
CREATE INDEX idx_care_logs_clinic_id ON care_logs(clinic_id);

-- billing 子テーブル FK インデックス
CREATE INDEX idx_billing_items_billing_id ON billing_items(billing_id);
CREATE INDEX idx_billing_items_deleted_at ON billing_items(deleted_at);
CREATE INDEX idx_billings_pet_id ON billings(pet_id);
CREATE INDEX idx_billings_owner_id ON billings(owner_id);

CREATE INDEX idx_billing_refunds_billing ON billing_refunds(billing_id);
CREATE INDEX idx_billing_refunds_clinic_billing ON billing_refunds(clinic_id, billing_id);
CREATE INDEX idx_payments_staff ON payments(paid_by);
CREATE INDEX idx_billing_refunds_staff ON billing_refunds(refunded_by);

-- 担当医 FK インデックス（staffs）
CREATE INDEX idx_vital_records_staff_id ON vital_records(staff_id);

-- medical_record_images インデックス
CREATE INDEX idx_medical_record_images_image_type ON medical_record_images(image_type);
CREATE INDEX idx_medical_record_images_taken_at ON medical_record_images(taken_at DESC);
CREATE INDEX idx_medical_record_images_exam_id ON medical_record_images(exam_id) WHERE exam_id IS NOT NULL;

-- estimates インデックス
CREATE INDEX idx_estimates_medical_record_id ON estimates(medical_record_id);
CREATE INDEX idx_estimates_status ON estimates(status);
CREATE INDEX idx_estimates_owner_id ON estimates(owner_id);

-- estimate_items インデックス
CREATE INDEX idx_estimate_items_estimate_id ON estimate_items(estimate_id);

-- -----------------------------------------------------------------------------
-- 4.5 全文検索インデックス（pg_trgm GIN）
-- -----------------------------------------------------------------------------
CREATE INDEX idx_owners_name_trgm ON owners USING gin (name gin_trgm_ops) WHERE deleted_at IS NULL;
CREATE INDEX idx_pets_name_trgm ON pets USING gin (name gin_trgm_ops) WHERE deleted_at IS NULL;

-- -----------------------------------------------------------------------------
-- 4.6 パフォーマンス最適化インデックス（論理削除考慮）
-- -----------------------------------------------------------------------------

-- ダッシュボード・カレンダー（最高頻度）
CREATE INDEX idx_appointments_clinic_date
  ON appointments(clinic_id, start_time)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_appointments_clinic_status
  ON appointments(clinic_id, status)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_appointments_pet_date
  ON appointments(pet_id, start_time)
  WHERE deleted_at IS NULL;

-- カルテ一覧・検索
CREATE INDEX idx_medical_records_clinic_date
  ON medical_records(clinic_id, date DESC)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_medical_records_clinic_pet
  ON medical_records(clinic_id, pet_id)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_medical_records_clinic_owner
  ON medical_records(clinic_id, owner_id)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_medical_records_clinic_status
  ON medical_records(clinic_id, status)
  WHERE deleted_at IS NULL;

-- 健診・ワクチン一覧・期限アラート（clinic_id + date / next_date）
CREATE INDEX idx_checkups_clinic_date
  ON checkups(clinic_id, date)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_checkups_clinic_next_date
  ON checkups(clinic_id, next_date)
  WHERE deleted_at IS NULL AND next_date IS NOT NULL;

CREATE INDEX idx_vaccinations_clinic_date
  ON vaccinations(clinic_id, date)
  WHERE deleted_at IS NULL;

-- ペット一覧（飼主別）
CREATE INDEX idx_pets_owner_id
  ON pets(owner_id)
  WHERE deleted_at IS NULL;

-- 飼主別生存ペット件数バッチ集計（CountLivingByOwnerIDs）
CREATE INDEX idx_pets_clinic_owner_living
  ON pets(clinic_id, owner_id, deceased_at)
  WHERE deleted_at IS NULL;

-- 会計一覧
CREATE INDEX idx_billings_clinic_date
  ON billings(clinic_id, scheduled_date)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_billings_clinic_status
  ON billings(clinic_id, status)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_billings_has_insurance
  ON billings(clinic_id, has_insurance)
  WHERE deleted_at IS NULL;

-- PERF-01: 月次レポート・締め集計最適化
CREATE INDEX idx_billings_clinic_completed_at
  ON billings(clinic_id, completed_at)
  WHERE deleted_at IS NULL AND status = 'completed';

-- 入院管理
CREATE INDEX idx_hospitalizations_clinic_status
  ON hospitalizations(clinic_id, status)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_hospitalizations_clinic_doctor
  ON hospitalizations(clinic_id, doctor_id)
  WHERE deleted_at IS NULL;

-- BE-033: 追加インデックス（検索パフォーマンス改善）
CREATE INDEX idx_owners_phone_trgm ON owners USING gin (phone gin_trgm_ops) WHERE deleted_at IS NULL;
CREATE INDEX idx_inventory_items_category ON inventory_items(category) WHERE deleted_at IS NULL;

-- 追加FKインデックス
CREATE INDEX idx_staffs_occupation_id ON staffs(occupation_id);
CREATE INDEX idx_pets_animal_species_id ON pets(animal_species_id);
CREATE INDEX idx_pets_insurance_id ON pets(insurance_id) WHERE insurance_id IS NOT NULL;
CREATE INDEX idx_diagnosis_names_type_id ON diagnosis_names(diagnosis_type_id);
CREATE INDEX idx_medical_records_doctor_id ON medical_records(doctor_id) WHERE doctor_id IS NOT NULL;
CREATE INDEX idx_medical_records_entered_by ON medical_records(entered_by) WHERE entered_by IS NOT NULL;
CREATE INDEX idx_treatments_consultation_id ON treatments(consultation_id) WHERE consultation_id IS NOT NULL;
CREATE INDEX idx_treatments_procedure_id ON treatments(procedure_id) WHERE procedure_id IS NOT NULL;
CREATE INDEX idx_treatments_medicine_id ON treatments(medicine_id) WHERE medicine_id IS NOT NULL;
CREATE INDEX idx_treatments_inventory_id ON treatments(inventory_id) WHERE inventory_id IS NOT NULL;
CREATE INDEX idx_care_plan_items_medicine_id ON care_plan_items(medicine_id) WHERE medicine_id IS NOT NULL;
CREATE INDEX idx_care_plan_items_procedure_id ON care_plan_items(procedure_id) WHERE procedure_id IS NOT NULL;
CREATE INDEX idx_care_plan_items_plan_id ON care_plan_items(hospitalization_plan_id) WHERE hospitalization_plan_id IS NOT NULL;
CREATE INDEX idx_vaccines_inventory_id ON vaccines(inventory_id) WHERE inventory_id IS NOT NULL;
CREATE INDEX idx_medicines_inventory_id ON medicines(inventory_id) WHERE inventory_id IS NOT NULL;

-- clinic_id の新規追加分インデックス
CREATE INDEX idx_checkups_clinic_id ON checkups(clinic_id);
CREATE INDEX idx_exams_clinic_id ON exams(clinic_id);
CREATE INDEX idx_daily_records_clinic_id ON daily_records(clinic_id);

-- マスタテーブル重複登録防止（同一クリニック内で同名マスタを防ぐ）
CREATE UNIQUE INDEX idx_exam_types_clinic_name ON exam_types(clinic_id, name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_vaccines_clinic_name ON vaccines(clinic_id, name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_medicines_clinic_name ON medicines(clinic_id, name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_consultations_clinic_name ON consultations(clinic_id, name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_procedures_clinic_name ON procedures(clinic_id, name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_cages_clinic_name ON cages(clinic_id, name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_reservation_types_clinic_name ON reservation_types(clinic_id, name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_diagnosis_types_clinic_name ON diagnosis_types(clinic_id, name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_trimming_courses_clinic_name ON trimming_courses(clinic_id, name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_trimming_options_clinic_name ON trimming_options(clinic_id, name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_insurance_clinic_name ON insurances(clinic_id, name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_checkup_types_clinic_name ON checkup_types(clinic_id, name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_hospitalization_plans_clinic_name ON hospitalization_plans(clinic_id, name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_occupations_clinic_name ON occupations(clinic_id, name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_chief_complaint_types_clinic_name ON chief_complaint_types(clinic_id, name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_animal_species_name ON animal_species(name) WHERE is_active = true;
CREATE UNIQUE INDEX idx_merchandise_items_clinic_name ON merchandise_items(clinic_id, name) WHERE is_active = true AND deleted_at IS NULL;

-- マスタテーブル論理削除アクティブレコードインデックス
CREATE INDEX idx_occupations_active ON occupations(id) WHERE deleted_at IS NULL;
CREATE INDEX idx_exam_types_active ON exam_types(id) WHERE deleted_at IS NULL;
CREATE INDEX idx_vaccines_active ON vaccines(id) WHERE deleted_at IS NULL;
CREATE INDEX idx_medicines_active ON medicines(id) WHERE deleted_at IS NULL;
CREATE INDEX idx_insurances_active ON insurances(id) WHERE deleted_at IS NULL;
CREATE INDEX idx_cages_active ON cages(id) WHERE deleted_at IS NULL;
CREATE INDEX idx_reservation_type_groups_active ON reservation_type_groups(id) WHERE deleted_at IS NULL;
CREATE INDEX idx_reservation_types_active ON reservation_types(id) WHERE deleted_at IS NULL;
CREATE INDEX idx_procedures_active ON procedures(id) WHERE deleted_at IS NULL;
CREATE INDEX idx_trimming_courses_active ON trimming_courses(id) WHERE deleted_at IS NULL;
CREATE INDEX idx_trimming_options_active ON trimming_options(id) WHERE deleted_at IS NULL;
CREATE INDEX idx_diagnosis_types_active ON diagnosis_types(id) WHERE deleted_at IS NULL;
CREATE INDEX idx_diagnosis_names_active ON diagnosis_names(id) WHERE deleted_at IS NULL;
CREATE INDEX idx_chief_complaint_types_active ON chief_complaint_types(id) WHERE deleted_at IS NULL;
CREATE INDEX idx_inquiry_templates_active ON inquiry_templates(id) WHERE deleted_at IS NULL;
CREATE INDEX idx_consultations_active ON consultations(clinic_id, id) WHERE deleted_at IS NULL;
CREATE INDEX idx_hospitalization_plans_active ON hospitalization_plans(clinic_id, id) WHERE deleted_at IS NULL;

-- =============================================================================
-- 5. テーブルコメント
-- =============================================================================

COMMENT ON TABLE companies IS '法人情報（シングルトン）';
COMMENT ON TABLE clinics IS '医院情報';
COMMENT ON COLUMN clinics.accounting_document_show_logo IS '明細兼領収書に病院ロゴを表示するか。';
COMMENT ON COLUMN clinics.accounting_document_show_registration_warning IS '登録番号未設定時の帳票警告を表示するか。';
COMMENT ON COLUMN clinics.accounting_document_show_item_category IS '明細兼領収書の項目カテゴリ行を表示するか。';
COMMENT ON COLUMN clinics.accounting_document_footer_note IS '明細兼領収書のフッター備考文言。';
COMMENT ON COLUMN clinics.accounting_document_show_clinic_header   IS '明細兼領収書に病院ヘッダーセクションを表示するか。';
COMMENT ON COLUMN clinics.accounting_document_show_owner_pet_info  IS '明細兼領収書に飼主・ペット情報セクションを表示するか。';
COMMENT ON COLUMN clinics.accounting_document_show_items_table     IS '明細兼領収書に明細表セクションを表示するか。';
COMMENT ON COLUMN clinics.accounting_document_show_payment_summary IS '明細兼領収書に合計・支払セクションを表示するか。';
COMMENT ON COLUMN clinics.accounting_document_section_order        IS '明細兼領収書のセクション表示順（空配列=デフォルト順）。';
COMMENT ON TABLE animal_species IS 'ペット種類マスタ（システム共通）';
COMMENT ON TABLE accounts IS '認証用アカウント';
COMMENT ON TABLE occupations IS '職種マスタ';
COMMENT ON TABLE staffs IS 'スタッフマスタ';
COMMENT ON TABLE owners IS '飼主情報';
COMMENT ON TABLE inventory_items IS '在庫アイテム';
COMMENT ON TABLE exam_types IS '検査種別マスタ';
COMMENT ON TABLE exam_type_fields IS '検査項目定義マスタ';
COMMENT ON TABLE vaccines IS 'ワクチンマスタ';
COMMENT ON TABLE medicines IS '薬剤マスタ';
COMMENT ON TABLE insurances IS '保険マスタ';
COMMENT ON TABLE cages IS 'ケージマスタ';
COMMENT ON TABLE reservation_type_groups IS '予約区分グループマスタ';
COMMENT ON TABLE reservation_types IS '予約区分マスタ';
COMMENT ON TABLE consultations IS '診察項目マスタ';
COMMENT ON TABLE procedures IS '処置項目マスタ';
COMMENT ON TABLE hospitalization_plans IS '入院プランマスタ';
COMMENT ON TABLE trimming_courses IS 'トリミングコースマスタ';
COMMENT ON TABLE trimming_options IS 'トリミングオプションマスタ';
COMMENT ON TABLE diagnosis_types IS '診断カテゴリマスタ';
COMMENT ON TABLE diagnosis_names IS '診断病名マスタ';
COMMENT ON TABLE checkup_types IS '健診種別マスタ';
COMMENT ON TABLE chief_complaint_types IS '主訴区分マスタ';
COMMENT ON TABLE inquiry_templates IS '問診定型文マスタ';
COMMENT ON TABLE pets IS 'ペット情報';
COMMENT ON TABLE staff_clinic_assignments IS 'スタッフ-クリニック所属';
COMMENT ON TABLE permission_groups IS '権限グループマスタ';
COMMENT ON TABLE permission_group_rules IS '権限グループルール';
COMMENT ON TABLE staff_permission_groups IS 'スタッフ-権限グループ';
COMMENT ON TABLE line_customers IS 'LINE予約顧客';
COMMENT ON TABLE appointments IS '予約';
COMMENT ON TABLE hospitalizations IS '入院・ホテル管理';
COMMENT ON TABLE appointment_trimming_details IS 'トリミング予約詳細';
COMMENT ON TABLE medical_records IS '電子カルテ（診療記録）';
COMMENT ON TABLE vaccinations IS 'ワクチン接種記録';
COMMENT ON TABLE checkups IS '定期健診記録';
COMMENT ON TABLE exams IS '検査記録';
COMMENT ON TABLE inquiries IS '問診情報';
COMMENT ON TABLE clinical_plans IS '診察所見・診断・治療方針';
COMMENT ON TABLE vital_records IS 'バイタル記録（外来・入院統合）';
COMMENT ON TABLE treatments IS '治療明細（処置・診察・薬剤）';
COMMENT ON TABLE treatment_plans IS '治療プラン（外来・入院共用）';
COMMENT ON TABLE medical_record_images IS '診療画像';
COMMENT ON TABLE billing_confirmations IS '会計医師確認';
COMMENT ON TABLE estimates IS '見積書';
COMMENT ON TABLE exam_results IS '検査結果項目';
COMMENT ON TABLE daily_records IS '入院日次記録';
COMMENT ON TABLE care_plan_items IS 'ケアプラン項目';
COMMENT ON TABLE estimate_items IS '見積書明細';
COMMENT ON TABLE care_logs IS 'ケアログ';
COMMENT ON TABLE staff_notes IS 'スタッフノート';
COMMENT ON TABLE appointment_trimming_options IS 'トリミング予約オプション適用';
COMMENT ON TABLE billings IS '会計';
COMMENT ON TABLE billing_items IS '会計明細';
COMMENT ON COLUMN billing_items.discount_rate   IS '#85: 項目別割引率(%)';
COMMENT ON COLUMN billing_items.discount_amount IS '#85: 項目別割引額(円)';
COMMENT ON TABLE payments IS '支払い情報';
COMMENT ON TABLE billing_refunds IS '返金レコード（Stripe モデル）';
COMMENT ON TABLE shift_entries IS 'スタッフシフト';
COMMENT ON TABLE clinic_holidays IS '医院個別休診日';
COMMENT ON TABLE merchandise_items IS '物販・フード・その他マスタ';
COMMENT ON TABLE clinic_settings IS '医院締め時間・休診曜日設定（FEAT-368）';
COMMENT ON TABLE closing_special_periods IS '特別診療時間設定（FEAT-368）';
COMMENT ON TABLE payment_methods IS '支払方法マスタ（FEAT-368）';
COMMENT ON TABLE cash_register_closes IS 'レジ締めレコード（FEAT-368）';
-- 統合テーブルコメント（005–017）
COMMENT ON TABLE clinic_integrations                  IS 'Lステップ/LINE連携設定保存テーブル（005 統合）';
COMMENT ON TABLE shared_files                         IS 'LINE個別送信用ファイルストレージ（006 統合）';
COMMENT ON TABLE lstep_settings                       IS 'クリニックごとのLステップ同期設定（ext-007 統合）';
COMMENT ON TABLE lstep_sync_error_counters            IS 'Lステップ同期APIの連続失敗回数カウンター（ext-008 統合）';
COMMENT ON TABLE lstep_tag_cache                      IS 'Lステップタグのカルテ側キャッシュ（008 統合）';
COMMENT ON TABLE lstep_tag_code_mappings              IS 'Lステップタグ → 診療コード 対応マスタ（ext-010 統合）';
COMMENT ON TABLE lstep_delivery_trigger_log           IS 'Lステップ自動配信トリガーの実行ログ（ext-012 統合）';
COMMENT ON TABLE prescriptions                        IS '処方薬記録テーブル（011 統合）';
COMMENT ON TABLE medical_record_addenda               IS 'カルテテキスト修正の追記記録（ext-021 統合）';
COMMENT ON TABLE pet_chronic_conditions               IS '慢性疾患フラグ管理テーブル（012 統合）';
COMMENT ON TABLE line_send_logs                       IS 'LINE送信ログ（013 統合）';
COMMENT ON TABLE line_link_tokens                     IS 'LINE User ID 紐付け用の一時トークン（016 統合）';
COMMENT ON TABLE token_blacklist                      IS 'ログアウト・失効済み refresh_token の JTI ブラックリスト（006b 統合）';
COMMENT ON TABLE lstep_csv_imports                    IS 'Lステップ友だち属性CSVのインポート履歴（ext-017 統合）';
COMMENT ON TABLE lstep_friend_attribute_snapshots     IS 'Lステップ友だちの属性スナップショット（ext-018 統合）';
COMMENT ON TABLE lstep_migration_progress             IS '既存飼い主データ一括同期の進捗管理テーブル（017 統合）';

-- ------------------------------------
-- 62. audit_logs（権限変更・認証操作の監査ログ）
-- ------------------------------------
CREATE TABLE audit_logs (
    id           BIGSERIAL    PRIMARY KEY,
    clinic_id    bigint       NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    actor_id     bigint       NULL REFERENCES staffs(id) ON DELETE RESTRICT,
    actor_type   varchar(30)  NOT NULL CHECK (actor_type IN ('staff', 'system')),
    action       varchar(50)  NOT NULL,
    resource     varchar(50)  NOT NULL,
    resource_id  bigint       NULL,
    old_value    jsonb        NULL,
    new_value    jsonb        NULL,
    ip_address   inet         NULL,
    user_agent   text         NULL,
    metadata     jsonb        NULL,                  -- ext-005: 追加コンテキスト情報
    created_at   timestamptz  NOT NULL DEFAULT now(),
    CONSTRAINT audit_logs_actor_consistency_check
        CHECK (
            (actor_type = 'system' AND actor_id IS NULL)
            OR
            (actor_type = 'staff' AND actor_id IS NOT NULL)
        )
);

CREATE INDEX idx_audit_logs_clinic   ON audit_logs(clinic_id, created_at DESC);
CREATE INDEX idx_audit_logs_actor    ON audit_logs(actor_id, created_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource, resource_id, created_at DESC);

COMMENT ON TABLE audit_logs IS '権限変更・認証操作の監査ログ（削除禁止）';

-- ==========================================================================
-- レイヤー8: 監査・LINE予約・シフト追加設定
-- ==========================================================================

-- ------------------------------------
-- 63. line_reservation_settings（LINE予約基本設定 — クリニック単位 1:1）
-- ------------------------------------
CREATE TABLE line_reservation_settings (
    id                         BIGSERIAL   PRIMARY KEY,
    clinic_id                  bigint      NOT NULL UNIQUE REFERENCES clinics(id),
    status                     text        NOT NULL DEFAULT 'stopped',

    -- ページ編集（トップページ）
    header_text                text        NOT NULL DEFAULT '',
    reservation_notice         text        NOT NULL DEFAULT '',
    cancel_notice              text        NOT NULL DEFAULT '',
    privacy_policy             text        NOT NULL DEFAULT '',

    -- 基本設定
    closed_weekdays            jsonb       NOT NULL DEFAULT '[]',
    closed_dates               jsonb       NOT NULL DEFAULT '[]',
    national_holiday_closed    boolean     NOT NULL DEFAULT false,
    business_hours             jsonb       NOT NULL DEFAULT '{"start":"0900","end":"1900"}',
    business_hours_by_weekday  jsonb,
    break_hours                jsonb       NOT NULL DEFAULT '[{"start":"1200","end":"1300"}]',
    daily_limit                int                  DEFAULT 1,
    monthly_limit              int,
    booking_window_max_days    int         NOT NULL DEFAULT 30,
    booking_window_min_days    int         NOT NULL DEFAULT 2,
    calendar_months            int         NOT NULL DEFAULT 2,
    phone_number               text        NOT NULL DEFAULT '',
    notification_email         text        NOT NULL DEFAULT '',
    request_example            text        NOT NULL DEFAULT '',
    time_slot_mode             text        NOT NULL DEFAULT 'minimize_gaps',
    time_slot_interval_minutes int         NOT NULL DEFAULT 15,
    no_staff_mode              text        NOT NULL DEFAULT 'first_available',
    show_no_staff_option       boolean     NOT NULL DEFAULT true,

    -- 追加入力フィールド定義
    additional_fields          jsonb       NOT NULL DEFAULT '[
        {"key":"phone","label":"電話番号","required":true,"placeholder":"例) 090-1234-5678"},
        {"key":"owner_name","label":"飼い主名","required":true,"placeholder":""},
        {"key":"pet_info","label":"ペットの名前と種類","required":true,"placeholder":"例) ポチ（柴犬）"},
        {"key":"symptoms","label":"診察内容","required":false,"placeholder":""}
    ]',

    -- LINE連携
    line_channel_id            text        NOT NULL DEFAULT '',
    line_channel_secret        text        NOT NULL DEFAULT '',
    liff_id                    text        NOT NULL DEFAULT '',
    line_access_token          text        NOT NULL DEFAULT '',

    created_at                 timestamptz NOT NULL DEFAULT now(),
    updated_at                 timestamptz NOT NULL DEFAULT now()
);
COMMENT ON TABLE line_reservation_settings IS 'LINE予約基本設定';

-- ------------------------------------
-- 64. staff_reservation_exclusions（スタッフ × 非対応予約区分 M:N）
-- ------------------------------------
CREATE TABLE staff_reservation_exclusions (
    id              BIGSERIAL PRIMARY KEY,
    staff_id        bigint    NOT NULL REFERENCES staffs(id) ON DELETE CASCADE,
    reservation_type_id     bigint    NOT NULL REFERENCES reservation_types(id) ON DELETE CASCADE,
    UNIQUE(staff_id, reservation_type_id)
);
COMMENT ON TABLE staff_reservation_exclusions IS 'スタッフ非対応予約区分';

-- ------------------------------------
-- 65. shift_entry_breaks（シフト中断時間 — shift_entries の子テーブル）
-- ------------------------------------
CREATE TABLE shift_entry_breaks (
    id             BIGSERIAL PRIMARY KEY,
    shift_entry_id bigint    NOT NULL REFERENCES shift_entries(id) ON DELETE CASCADE,
    break_start    time      NOT NULL,
    break_end      time      NOT NULL
);
CREATE INDEX idx_shift_entry_breaks_entry ON shift_entry_breaks(shift_entry_id);
COMMENT ON TABLE shift_entry_breaks IS 'シフト中の休憩時間';

-- ------------------------------------
-- 66. shift_templates（シフトテンプレートマスタ）
-- ------------------------------------
CREATE TABLE shift_templates (
    id         BIGSERIAL    PRIMARY KEY,
    clinic_id  bigint       NOT NULL REFERENCES clinics(id),
    name       varchar(100) NOT NULL,
    shift_type shift_type   NOT NULL DEFAULT 'full',
    start_time time,
    end_time   time,
    notes      text         NOT NULL DEFAULT '',
    sort_order integer      NOT NULL DEFAULT 0,
    is_active  boolean      NOT NULL DEFAULT true,
    created_at timestamptz  NOT NULL DEFAULT now(),
    updated_at timestamptz  NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

-- 部分 UNIQUE インデックス（WHERE 句は CREATE UNIQUE INDEX で記述する必要がある）
CREATE UNIQUE INDEX uk_shift_templates_clinic_name ON shift_templates(clinic_id, name) WHERE deleted_at IS NULL;
COMMENT ON TABLE shift_templates IS 'シフトテンプレートマスタ';

-- ------------------------------------
-- 67. shift_template_breaks（シフトテンプレートの休憩時間）
-- ------------------------------------
CREATE TABLE shift_template_breaks (
    id                BIGSERIAL PRIMARY KEY,
    shift_template_id bigint    NOT NULL REFERENCES shift_templates(id) ON DELETE CASCADE,
    break_start       time      NOT NULL,
    break_end         time      NOT NULL
);

CREATE INDEX idx_shift_template_breaks_template ON shift_template_breaks(shift_template_id);
COMMENT ON TABLE shift_template_breaks IS 'シフトテンプレートの休憩時間';

CREATE INDEX idx_appointments_line_customer ON appointments(line_customer_id)
    WHERE line_customer_id IS NOT NULL AND deleted_at IS NULL;

-- ------------------------------------
-- 予約時間枠の重複防止（部分ユニークインデックス）
-- ------------------------------------
CREATE UNIQUE INDEX uk_appointment_staff_time
    ON appointments (clinic_id, doctor_id, start_time)
    WHERE deleted_at IS NULL AND status != 'cancelled';

-- =============================================
-- 68. reservation_type_unavailable_times（予約区分予約不可時間）
-- =============================================
CREATE TABLE reservation_type_unavailable_times (
    id                  BIGSERIAL   PRIMARY KEY,
    clinic_id           bigint      NOT NULL REFERENCES clinics(id),
    reservation_type_id bigint      NOT NULL REFERENCES reservation_types(id) ON DELETE CASCADE,
    unavailable_type    text        NOT NULL CHECK (unavailable_type IN ('weekly', 'specific')),
    -- weekly: 0=日曜, 1=月曜, ..., 6=土曜
    day_of_week         smallint    CHECK (
                            (unavailable_type = 'weekly' AND day_of_week BETWEEN 0 AND 6)
                            OR (unavailable_type = 'specific' AND day_of_week IS NULL)
                        ),
    specific_date       date        CHECK (
                            (unavailable_type = 'specific' AND specific_date IS NOT NULL)
                            OR (unavailable_type = 'weekly' AND specific_date IS NULL)
                        ),
    -- "HH:MM" 形式で保存（VARCHAR(5)）
    -- TIME型を使わない理由: GORMがTIME列をstringにscanすると"HH:MM:SS"形式になり、
    -- timeslot_engine の minutesSinceMidnight（4文字HHMM専用）が必ずエラーになるため
    start_time          varchar(5)  NOT NULL CHECK (start_time ~ '^\d{2}:\d{2}$'),
    end_time            varchar(5)  NOT NULL CHECK (end_time ~ '^\d{2}:\d{2}$'),
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT chk_unavailable_time_range CHECK (end_time > start_time)
    -- 論理削除なし（物理削除）: reservation_types の ON DELETE CASCADE で連動削除される
);

CREATE INDEX idx_rtype_unavailable_clinic_type
    ON reservation_type_unavailable_times(clinic_id, reservation_type_id);
CREATE INDEX idx_rtype_unavailable_weekly
    ON reservation_type_unavailable_times(reservation_type_id, day_of_week)
    WHERE unavailable_type = 'weekly';
CREATE INDEX idx_rtype_unavailable_specific
    ON reservation_type_unavailable_times(reservation_type_id, specific_date)
    WHERE unavailable_type = 'specific';
COMMENT ON TABLE reservation_type_unavailable_times IS '予約区分予約不可時間';

-- =============================================
-- 69. reservation_type_occupations（予約区分 × 職種 中間テーブル）
-- =============================================
CREATE TABLE reservation_type_occupations (
    id                  BIGSERIAL   PRIMARY KEY,
    clinic_id           bigint      NOT NULL REFERENCES clinics(id),
    reservation_type_id bigint      NOT NULL REFERENCES reservation_types(id) ON DELETE CASCADE,
    occupation_id       bigint      NOT NULL REFERENCES occupations(id) ON DELETE CASCADE,
    created_at          timestamptz NOT NULL DEFAULT now(),
    UNIQUE (reservation_type_id, occupation_id)
    -- 論理削除なし（物理削除）
);

CREATE INDEX idx_rtype_occupation_clinic
    ON reservation_type_occupations(clinic_id, reservation_type_id);
CREATE INDEX idx_rtype_occupation_occupation
    ON reservation_type_occupations(occupation_id);

-- =============================================
-- 70. password_reset_tokens（パスワードリセットトークン）
-- =============================================
CREATE TABLE password_reset_tokens (
    id          BIGSERIAL   PRIMARY KEY,
    account_id  bigint      NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    token_hash  text        NOT NULL UNIQUE,
    expires_at  timestamptz NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_prt_account ON password_reset_tokens(account_id);
COMMENT ON TABLE password_reset_tokens IS 'パスワードリセットトークン';
COMMENT ON TABLE reservation_type_occupations IS '予約区分対応職種';

-- =============================================
-- デフォルト支払方法のトリガー
-- 新しいクリニック作成時に自動でデフォルト支払方法を挿入する
-- =============================================
CREATE OR REPLACE FUNCTION create_default_payment_methods()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO payment_methods (clinic_id, name, system_key, display_order, is_active)
    VALUES
        (NEW.id, '現金',            'cash',             1, true),
        (NEW.id, 'クレジットカード', 'credit_card',      2, true),
        (NEW.id, '電子マネー',       'electronic_money', 3, true),
        (NEW.id, '銀行振込',         'bank_transfer',    4, true);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_create_default_payment_methods
    AFTER INSERT ON clinics
    FOR EACH ROW
    EXECUTE FUNCTION create_default_payment_methods();

-- =============================================
-- デフォルトトリミングコース種別のトリガー
-- 新しいクリニック作成時に自動でデフォルト種別を挿入する
-- =============================================
CREATE OR REPLACE FUNCTION create_default_trimming_course_types()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO trimming_course_types (clinic_id, name, sort_order, is_active, created_at, updated_at)
    VALUES
        (NEW.id, 'シャンプー',        0, true, now(), now()),
        (NEW.id, 'シャンプー＆カット', 1, true, now(), now());
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_create_default_trimming_course_types
    AFTER INSERT ON clinics
    FOR EACH ROW
    EXECUTE FUNCTION create_default_trimming_course_types();

-- =============================================
-- 取扱説明書（マニュアル）の DB 管理
-- =============================================
-- 設計方針:
--   - フロントエンドが MD ファイルをデフォルト（バンドル）として保持
--   - DB にはオーバーライド版を保存
--   - 読み込み時: DB に該当 slug があればそれを優先、なければ MD ファイル
--   - 編集時: 該当 slug を DB に upsert
--   - マニュアルは医院共通の情報のため clinic_id は持たない
-- =============================================

CREATE TABLE manual_articles (
    id                  bigserial      PRIMARY KEY,
    category            text           NOT NULL CHECK (category IN ('screens', 'workflows')),
    slug                text           NOT NULL,
    title               text           NOT NULL,
    order_value         numeric(10, 2) NOT NULL DEFAULT 9999,
    section             text           NOT NULL,
    body_markdown       text           NOT NULL,
    updated_by_staff_id bigint         REFERENCES staffs(id) ON DELETE SET NULL,
    created_at          timestamptz    NOT NULL DEFAULT now(),
    updated_at          timestamptz    NOT NULL DEFAULT now(),

    UNIQUE (category, slug)
);

CREATE INDEX idx_manual_articles_category_slug ON manual_articles(category, slug);
CREATE INDEX idx_manual_articles_updated_at    ON manual_articles(updated_at DESC);

COMMENT ON TABLE manual_articles IS '取扱説明書のオーバーライド版（DBに保存された編集後マニュアル）';
COMMENT ON COLUMN manual_articles.category IS 'カテゴリ: screens | workflows';
COMMENT ON COLUMN manual_articles.slug IS 'ファイル名（拡張子除く）。例: 13-cash-register';
COMMENT ON COLUMN manual_articles.order_value IS 'セクション内表示順（昇順）';
COMMENT ON COLUMN manual_articles.section IS 'サイドバーのグループ名';
COMMENT ON COLUMN manual_articles.body_markdown IS 'マニュアル本文（frontmatter 除く）';

CREATE TABLE manual_article_versions (
    id                 bigserial      PRIMARY KEY,
    article_id         bigint         NOT NULL REFERENCES manual_articles(id) ON DELETE CASCADE,
    title              text           NOT NULL,
    order_value        numeric(10, 2) NOT NULL,
    section            text           NOT NULL,
    body_markdown      text           NOT NULL,
    edited_by_staff_id bigint         REFERENCES staffs(id) ON DELETE SET NULL,
    edited_at          timestamptz    NOT NULL DEFAULT now()
);

CREATE INDEX idx_manual_article_versions_article ON manual_article_versions(article_id, edited_at DESC);

COMMENT ON TABLE manual_article_versions IS 'マニュアル編集履歴（編集ごとに過去版を保持）';

-- ------------------------------------
-- #81: キャンペーン割引マスタ
-- ------------------------------------
CREATE TABLE campaigns (
    id              BIGSERIAL              PRIMARY KEY,
    clinic_id       bigint                 NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name            text                   NOT NULL DEFAULT '',
    start_date      date                   NOT NULL,
    end_date        date                   NOT NULL,
    discount_type   campaign_discount_type NOT NULL DEFAULT 'rate',
    discount_value  numeric(12,2)          NOT NULL DEFAULT 0,
    is_active       boolean                NOT NULL DEFAULT true,
    sort_order      integer                NOT NULL DEFAULT 0,
    created_at      timestamptz            NOT NULL DEFAULT now(),
    updated_at      timestamptz            NOT NULL DEFAULT now(),
    deleted_at      timestamptz,
    CONSTRAINT chk_campaigns_period CHECK (end_date >= start_date)
);

CREATE TABLE campaign_target_categories (
    id          BIGSERIAL     PRIMARY KEY,
    campaign_id bigint        NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    category    item_category NOT NULL,
    CONSTRAINT uq_campaign_target_categories UNIQUE (campaign_id, category)
);

CREATE TABLE campaign_target_items (
    id                  BIGSERIAL PRIMARY KEY,
    campaign_id         bigint    NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    merchandise_item_id bigint    NOT NULL REFERENCES merchandise_items(id) ON DELETE CASCADE,
    CONSTRAINT uq_campaign_target_items UNIQUE (campaign_id, merchandise_item_id)
);

CREATE INDEX idx_campaigns_clinic         ON campaigns(clinic_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_campaigns_clinic_period  ON campaigns(clinic_id, start_date, end_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_campaign_target_categories_campaign ON campaign_target_categories(campaign_id);
CREATE INDEX idx_campaign_target_items_campaign      ON campaign_target_items(campaign_id);
CREATE INDEX idx_campaign_target_items_merchandise   ON campaign_target_items(merchandise_item_id);

COMMENT ON TABLE campaigns IS '#81: 割引キャンペーンマスタ(期間・割引種別/値)';
COMMENT ON TABLE campaign_target_categories IS '#81: キャンペーン対象カテゴリ(Q1=D カテゴリ単位指定)';
COMMENT ON TABLE campaign_target_items IS '#81: キャンペーン対象商品(Q1=D 個別商品指定)';

-- =============================================================================
-- 6. Row Level Security (新 005 統合 / #93)
-- =============================================================================
-- この定義は RLS を ENABLE するが FORCE はしない。
-- 理由:
--   - 現状のアプリケーション接続ユーザーは migration 実行ユーザーと同一で、テーブル owner の可能性が高い。
--   - FORCE RLS には全 repository 呼び出しを同一 transaction/context DB に統一し、
--     SET LOCAL app.current_clinic_ids を必ず流す改修が先に必要。
--   - ここでは DB 直接アクセス用の非 owner ロールに対して RLS を効かせる、破壊性の低い baseline を構築する。
--
-- 運用時は対象 DB ロールに以下のような設定を付与する:
--   ALTER ROLE clinic_reader_1 SET app.current_clinic_ids = '1';
--   ALTER ROLE clinic_reader_all SET app.bypass_rls = 'on';

CREATE SCHEMA IF NOT EXISTS app_private;

CREATE OR REPLACE FUNCTION app_private.current_clinic_ids()
RETURNS bigint[]
LANGUAGE sql
STABLE
AS $$
    SELECT COALESCE(
        string_to_array(NULLIF(current_setting('app.current_clinic_ids', true), ''), ',')::bigint[],
        ARRAY[]::bigint[]
    );
$$;

CREATE OR REPLACE FUNCTION app_private.bypass_rls()
RETURNS boolean
LANGUAGE sql
STABLE
AS $$
    SELECT COALESCE(NULLIF(current_setting('app.bypass_rls', true), '')::boolean, false);
$$;

CREATE OR REPLACE FUNCTION app_private.has_clinic_access(row_clinic_id bigint)
RETURNS boolean
LANGUAGE sql
STABLE
AS $$
    SELECT app_private.bypass_rls()
        OR row_clinic_id = ANY(app_private.current_clinic_ids());
$$;

CREATE OR REPLACE FUNCTION app_private.apply_rls_policy(
    target_table regclass,
    policy_name text,
    using_expr text,
    check_expr text
)
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    EXECUTE format('ALTER TABLE %s ENABLE ROW LEVEL SECURITY', target_table);
    EXECUTE format('DROP POLICY IF EXISTS %I ON %s', policy_name, target_table);
    EXECUTE format(
        'CREATE POLICY %I ON %s FOR ALL USING (%s) WITH CHECK (%s)',
        policy_name,
        target_table,
        using_expr,
        check_expr
    );
END;
$$;

GRANT USAGE ON SCHEMA app_private TO PUBLIC;
GRANT EXECUTE ON FUNCTION app_private.current_clinic_ids() TO PUBLIC;
GRANT EXECUTE ON FUNCTION app_private.bypass_rls() TO PUBLIC;
GRANT EXECUTE ON FUNCTION app_private.has_clinic_access(bigint) TO PUBLIC;
REVOKE ALL ON FUNCTION app_private.apply_rls_policy(regclass, text, text, text) FROM PUBLIC;

-- clinic_id を直接持つ public テーブルは同一 policy で保護する。
DO $$
DECLARE
    target_table regclass;
BEGIN
    FOR target_table IN
        SELECT c.oid::regclass
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        JOIN pg_attribute a ON a.attrelid = c.oid
        WHERE n.nspname = 'public'
          AND c.relkind = 'r'
          AND a.attname = 'clinic_id'
          AND NOT a.attisdropped
        ORDER BY c.relname
    LOOP
        PERFORM app_private.apply_rls_policy(
            target_table,
            'tenant_clinic_id_isolation',
            'app_private.has_clinic_access(clinic_id)',
            'app_private.has_clinic_access(clinic_id)'
        );
    END LOOP;
END;
$$;

-- clinics は自身の id を tenant key として扱う。
SELECT app_private.apply_rls_policy(
    'clinics',
    'tenant_clinics_isolation',
    'app_private.has_clinic_access(id)',
    'app_private.has_clinic_access(id)'
);

-- accounts は staffs.account_id 経由で tenant 境界を判定する。
SELECT app_private.apply_rls_policy(
    'accounts',
    'tenant_accounts_isolation',
    'EXISTS (SELECT 1 FROM staffs s WHERE s.account_id = accounts.id AND app_private.has_clinic_access(s.clinic_id))',
    'EXISTS (SELECT 1 FROM staffs s WHERE s.account_id = accounts.id AND app_private.has_clinic_access(s.clinic_id))'
);

-- clinic_id を直接持たない子テーブルは、親テーブルの clinic_id 経由で保護する。
SELECT app_private.apply_rls_policy(
    'exam_type_fields',
    'tenant_exam_type_fields_isolation',
    'EXISTS (SELECT 1 FROM exam_types et WHERE et.id = exam_type_fields.exam_type_id AND app_private.has_clinic_access(et.clinic_id))',
    'EXISTS (SELECT 1 FROM exam_types et WHERE et.id = exam_type_fields.exam_type_id AND app_private.has_clinic_access(et.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'permission_group_rules',
    'tenant_permission_group_rules_isolation',
    'EXISTS (SELECT 1 FROM permission_groups pg WHERE pg.id = permission_group_rules.group_id AND app_private.has_clinic_access(pg.clinic_id))',
    'EXISTS (SELECT 1 FROM permission_groups pg WHERE pg.id = permission_group_rules.group_id AND app_private.has_clinic_access(pg.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'staff_permission_groups',
    'tenant_staff_permission_groups_isolation',
    'EXISTS (SELECT 1 FROM permission_groups pg WHERE pg.id = staff_permission_groups.group_id AND app_private.has_clinic_access(pg.clinic_id))',
    'EXISTS (SELECT 1 FROM permission_groups pg WHERE pg.id = staff_permission_groups.group_id AND app_private.has_clinic_access(pg.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'appointment_trimming_options',
    'tenant_appointment_trimming_options_isolation',
    'EXISTS (SELECT 1 FROM appointments a WHERE a.id = appointment_trimming_options.appointment_id AND app_private.has_clinic_access(a.clinic_id))',
    'EXISTS (SELECT 1 FROM appointments a WHERE a.id = appointment_trimming_options.appointment_id AND app_private.has_clinic_access(a.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'inquiries',
    'tenant_inquiries_isolation',
    'EXISTS (SELECT 1 FROM medical_records mr WHERE mr.id = inquiries.medical_record_id AND app_private.has_clinic_access(mr.clinic_id))',
    'EXISTS (SELECT 1 FROM medical_records mr WHERE mr.id = inquiries.medical_record_id AND app_private.has_clinic_access(mr.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'clinical_plans',
    'tenant_clinical_plans_isolation',
    'EXISTS (SELECT 1 FROM medical_records mr WHERE mr.id = clinical_plans.medical_record_id AND app_private.has_clinic_access(mr.clinic_id))',
    'EXISTS (SELECT 1 FROM medical_records mr WHERE mr.id = clinical_plans.medical_record_id AND app_private.has_clinic_access(mr.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'treatments',
    'tenant_treatments_isolation',
    'EXISTS (SELECT 1 FROM medical_records mr WHERE mr.id = treatments.medical_record_id AND app_private.has_clinic_access(mr.clinic_id))',
    'EXISTS (SELECT 1 FROM medical_records mr WHERE mr.id = treatments.medical_record_id AND app_private.has_clinic_access(mr.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'medical_record_images',
    'tenant_medical_record_images_isolation',
    'EXISTS (SELECT 1 FROM medical_records mr WHERE mr.id = medical_record_images.medical_record_id AND app_private.has_clinic_access(mr.clinic_id))',
    'EXISTS (SELECT 1 FROM medical_records mr WHERE mr.id = medical_record_images.medical_record_id AND app_private.has_clinic_access(mr.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'billing_confirmations',
    'tenant_billing_confirmations_isolation',
    'EXISTS (SELECT 1 FROM medical_records mr WHERE mr.id = billing_confirmations.medical_record_id AND app_private.has_clinic_access(mr.clinic_id))',
    'EXISTS (SELECT 1 FROM medical_records mr WHERE mr.id = billing_confirmations.medical_record_id AND app_private.has_clinic_access(mr.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'exam_results',
    'tenant_exam_results_isolation',
    'EXISTS (SELECT 1 FROM exams e WHERE e.id = exam_results.exam_id AND app_private.has_clinic_access(e.clinic_id))',
    'EXISTS (SELECT 1 FROM exams e WHERE e.id = exam_results.exam_id AND app_private.has_clinic_access(e.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'care_plan_items',
    'tenant_care_plan_items_isolation',
    'EXISTS (SELECT 1 FROM hospitalizations h WHERE h.id = care_plan_items.hospitalization_id AND app_private.has_clinic_access(h.clinic_id))',
    'EXISTS (SELECT 1 FROM hospitalizations h WHERE h.id = care_plan_items.hospitalization_id AND app_private.has_clinic_access(h.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'estimate_items',
    'tenant_estimate_items_isolation',
    'EXISTS (SELECT 1 FROM estimates e WHERE e.id = estimate_items.estimate_id AND app_private.has_clinic_access(e.clinic_id))',
    'EXISTS (SELECT 1 FROM estimates e WHERE e.id = estimate_items.estimate_id AND app_private.has_clinic_access(e.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'staff_notes',
    'tenant_staff_notes_isolation',
    'EXISTS (SELECT 1 FROM daily_records dr WHERE dr.id = staff_notes.daily_record_id AND app_private.has_clinic_access(dr.clinic_id))',
    'EXISTS (SELECT 1 FROM daily_records dr WHERE dr.id = staff_notes.daily_record_id AND app_private.has_clinic_access(dr.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'billing_items',
    'tenant_billing_items_isolation',
    'EXISTS (SELECT 1 FROM billings b WHERE b.id = billing_items.billing_id AND app_private.has_clinic_access(b.clinic_id))',
    'EXISTS (SELECT 1 FROM billings b WHERE b.id = billing_items.billing_id AND app_private.has_clinic_access(b.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'payments',
    'tenant_payments_isolation',
    'EXISTS (SELECT 1 FROM billings b WHERE b.id = payments.billing_id AND app_private.has_clinic_access(b.clinic_id))',
    'EXISTS (SELECT 1 FROM billings b WHERE b.id = payments.billing_id AND app_private.has_clinic_access(b.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'staff_reservation_exclusions',
    'tenant_staff_reservation_exclusions_isolation',
    'EXISTS (SELECT 1 FROM reservation_types rt WHERE rt.id = staff_reservation_exclusions.reservation_type_id AND app_private.has_clinic_access(rt.clinic_id))',
    'EXISTS (SELECT 1 FROM reservation_types rt WHERE rt.id = staff_reservation_exclusions.reservation_type_id AND app_private.has_clinic_access(rt.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'shift_entry_breaks',
    'tenant_shift_entry_breaks_isolation',
    'EXISTS (SELECT 1 FROM shift_entries se WHERE se.id = shift_entry_breaks.shift_entry_id AND app_private.has_clinic_access(se.clinic_id))',
    'EXISTS (SELECT 1 FROM shift_entries se WHERE se.id = shift_entry_breaks.shift_entry_id AND app_private.has_clinic_access(se.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'shift_template_breaks',
    'tenant_shift_template_breaks_isolation',
    'EXISTS (SELECT 1 FROM shift_templates st WHERE st.id = shift_template_breaks.shift_template_id AND app_private.has_clinic_access(st.clinic_id))',
    'EXISTS (SELECT 1 FROM shift_templates st WHERE st.id = shift_template_breaks.shift_template_id AND app_private.has_clinic_access(st.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'campaign_target_categories',
    'tenant_campaign_target_categories_isolation',
    'EXISTS (SELECT 1 FROM campaigns c WHERE c.id = campaign_target_categories.campaign_id AND app_private.has_clinic_access(c.clinic_id))',
    'EXISTS (SELECT 1 FROM campaigns c WHERE c.id = campaign_target_categories.campaign_id AND app_private.has_clinic_access(c.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'campaign_target_items',
    'tenant_campaign_target_items_isolation',
    'EXISTS (SELECT 1 FROM campaigns c WHERE c.id = campaign_target_items.campaign_id AND app_private.has_clinic_access(c.clinic_id))',
    'EXISTS (SELECT 1 FROM campaigns c WHERE c.id = campaign_target_items.campaign_id AND app_private.has_clinic_access(c.clinic_id))'
);

-- 公開/システム共通/認証補助テーブルは clinic_id を持たないため RLS 対象外:
-- companies, animal_species, token_blacklist,
-- lstep_auto_managed_prefixes, lstep_condition_tag_mappings, lstep_send_purpose_tag_prefixes,
-- password_reset_tokens, manual_articles, manual_article_versions

-- =============================================================================
-- 7. 増分マイグレーション統合 (旧 005〜012 / 2026-07-04)
-- =============================================================================
-- 以下は独立ファイルとして管理されていた 005_add_lab_import_tables.sql 〜
-- 012_add_clinical_result_composite_fk.sql の原文を、番号順にそのまま追記したもの
-- (010 の DML 部分 = 歯科検診パッケージの暫定 seed DO ブロックのみ 003_seed_demo.sql へ移動)。
-- 各ファイルの内部 statement 順は変更していない。詳細は docs/ERD.md §4.3 を参照。
--
-- 注意（RLS 自動ループとの順序依存）: 上記セクション6の DO ブロックは「時点で clinic_id 列を
-- 持つ public テーブル」を自動的に RLS 保護するが、本セクションのテーブル (lab_import_jobs 等) は
-- そのループより後、同一トランザクション内で作成されるため自動ループの対象にならない
-- (旧ファイル分割適用時と同じ挙動: 005/009 は明示的な apply_rls_policy 呼び出しを持たないため
-- RLS 未適用のまま、010 のみ自身で apply_rls_policy を呼ぶ)。ファイル統合後もこの順序関係は
-- 保持されており、意図せず新規テーブルに RLS が有効化されることはない。

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
--   fixture        : テスト・開発用フィクスチャ入力 (Phase 0 で使用可能)
--   drwan          : Dr.Wan MDB アダプタ (製品経路では開けない)
--   manual         : 手動 CSV/JSON アップロード (Phase 2+ 予定)
--   fuji_nx600 / fuji_au10v / arkray_pu4010 : 城東3台（ADR-007。fresh 001 に含む）

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
    'failed',
    'reverted'
);

CREATE TYPE lab_import_source_type AS ENUM (
    'fixture',
    'drwan',
    'manual',
    'fuji_nx600',
    'fuji_au10v',
    'arkray_pu4010'
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

-- 006_add_exam_results_exam_id_index.sql
-- Phase 2: exam_results.exam_id index for lab import batch performance.
--
-- ReplaceItemsByExamID (used by LabImportExaminationService) runs a DELETE + INSERT
-- keyed on exam_id. Without this index each call is a full table scan on exam_results.
-- Phase 1 comment noted this migration must be applied before large-batch lab import runs.
--
-- Note: CONCURRENTLY was removed because the migration runner wraps each file in an
-- explicit transaction (cmd/migrate/main.go:tx.Begin/Exec/Commit), and PostgreSQL
-- does not allow CREATE INDEX CONCURRENTLY inside a transaction block. The tables
-- created in this phase are new (005_add_lab_import_tables.sql) with no concurrent
-- traffic at migration time, so a plain CREATE INDEX is safe.
-- IF NOT EXISTS makes the migration idempotent.
--
-- Phase 3A decision: no DB unique constraint on (clinic_id, exam_type_id, date, pet_id).
-- Local data check showed 87 duplicate groups (95 extra rows) in migrated data; these
-- are legitimate multi-visit records with distinct medical_record_ids, not true import
-- duplicates. Duplicate prevention is enforced at service level via LabImportDuplicateCheckerDB.
-- Production data must be verified before adding any DB unique constraint.
-- See: docs/lab-go/app-integration-boundary.md Phase 3A section.

CREATE INDEX IF NOT EXISTS idx_exam_results_exam_id
    ON exam_results (exam_id);

COMMENT ON INDEX idx_exam_results_exam_id
    IS 'Phase 2: supports ReplaceItemsByExamID DELETE+INSERT in lab import batches';

-- 007_add_exams_dup_check_index.sql
-- Phase 2: composite index for LabImportDuplicateCheckerDB hot path.
--
-- IsDuplicate queries exams with:
--   WHERE clinic_id = ? AND exam_type_id = ? AND date = ? AND deleted_at IS NULL
-- Without a composite index PostgreSQL performs a bitmap-AND over individual single-column
-- indexes (idx_exams_clinic_id, idx_exams_exam_type_id), which degrades toward a seq scan
-- on any table with more than a few thousand rows.
--
-- Column order: clinic_id (most selective for multi-tenant), exam_type_id, date (equality).
-- pet_id is handled by Go-side NULL branching and cannot be added to a single composite
-- index without expression tricks; the index narrows to (clinic_id, exam_type_id, date)
-- first and pet_id is applied as a recheck predicate.
--
-- Partial index on deleted_at IS NULL keeps the index smaller (soft-deleted exams are
-- excluded from import duplicate checks by design).
--
-- Note: CONCURRENTLY was removed because the migration runner wraps each file in an
-- explicit transaction (cmd/migrate/main.go:tx.Begin/Exec/Commit), and PostgreSQL
-- does not allow CREATE INDEX CONCURRENTLY inside a transaction block. The exams table
-- exists with migrated data but lab import is not yet live, so a plain CREATE INDEX is
-- safe at migration time.
--
-- Phase 3A decision: no DB unique constraint added.
-- 87 duplicate groups exist in local migrated data on the 4-column key
-- (clinic_id, exam_type_id, date, pet_id). 84/85 non-null groups have distinct
-- medical_record_ids (same pet, different karte visits on the same day) and are
-- legitimate. A DB unique constraint on this key would reject valid historical records.
-- Service-level duplicate prevention is the formal policy until production data is
-- verified and a 5-column partial unique index (adding medical_record_id) can be assessed.
-- See: docs/lab-go/app-integration-boundary.md Phase 3A section.

CREATE INDEX IF NOT EXISTS idx_exams_clinic_exam_type_date
    ON exams (clinic_id, exam_type_id, date)
    WHERE deleted_at IS NULL;

COMMENT ON INDEX idx_exams_clinic_exam_type_date
    IS 'Phase 2/3A: LabImportDuplicateCheckerDB (clinic_id, exam_type_id, date) lookup; no unique constraint — see Phase 3A decision';

-- 008_add_exams_job_id.sql
-- Phase 4B.2: exams.job_id nullable FK to lab_import_jobs
--
-- Decision (Phase 4B.1): ADD as uuid NULL with ON DELETE SET NULL so that
-- job deletion does not cascade-delete exam rows (business data must be preserved).
-- Nullable to remain backward-compatible with hand-created exams (NULL = no import job).

ALTER TABLE exams
    ADD COLUMN job_id uuid NULL
    REFERENCES lab_import_jobs(id) ON DELETE SET NULL;

-- Index for ListJobReportSummaries: "give me all exams for this job under this clinic"
-- clinic_id + job_id covers the primary query access pattern.
-- Partial index (WHERE job_id IS NOT NULL) keeps it small for hand-created exams.
CREATE INDEX idx_exams_clinic_job
    ON exams (clinic_id, job_id)
    WHERE job_id IS NOT NULL;

COMMENT ON COLUMN exams.job_id IS 'lab_import_jobs.id FK — NULL for hand-created exams. ON DELETE SET NULL preserves exam rows when job is deleted (Phase 4B.2).';

-- 009_add_medicine_dose_params.sql
-- #201 カルテ薬量（投与量）自動計算: 薬マスタ計算パラメータ + 種別子テーブル + treatments スナップショット
--
-- 方針（新規・追記のみ・additive・後方互換）:
--   - 既存 001-008 は無編集。既存薬剤の挙動は calculation_type=none（既定）で不変（手動 quantity）。
--   - per_weight 自動計算は mg/kg 線形の部分集合のみ。CRI/IU/%濃度/血中濃度/mg/head/BSA は none（手動）。
--   - 製品軸（strength）は medicines、種軸（dose_per_kg）は子テーブル medicine_dose_params に分離。
--     薬用量マニュアル実読で犬・猫の mg/kg が網羅的に異なることが判明したため、スカラー1列では破綻する。
--   - clinic_id は子テーブルに非正規化保持し clinicScope(P4) を直適用する（JOIN スコープは base.go で不可）。
--   - FK は ON DELETE RESTRICT（CASCADE DELETE 禁止方針 + 論理削除整合）。
--   - マスタ計算パラメータ変更・per_weight 有効化・著しい逸脱の上書きは audit_logs に記録（アプリ層）。

-- ------------------------------------
-- ENUM types
-- ------------------------------------

-- calculation_type は 2 値。非線形（CRI/IU/濃度/BSA 等）は将来 ENUM 拡張で名前付き計算式を追加する
-- （コードで実装した計算式の選択。自由入力式ではない）。default 'none' で default-deny。
CREATE TYPE medicine_calculation_type AS ENUM ('none', 'per_weight');

-- dose_basis は dose_per_kg が 1回量基準か 1日量基準かを区別する。
--   per_administration: dose_per_kg は 1回投与あたりの mg/kg
--   per_day:            dose_per_kg は 1日あたりの mg/kg（1回量は frequency_per_day で按分）
CREATE TYPE medicine_dose_basis AS ENUM ('per_administration', 'per_day');

-- rounding_mode は丸め方向。臨床ソースは丸め規則を定義しないため運用前提（NULL=丸めなし）。
CREATE TYPE medicine_rounding_mode AS ENUM ('up', 'down', 'nearest');

-- dose_species は計算対象の患者種。mg/kg は犬・猫で網羅的に異なり、'both' は意味を持たない
-- （vaccine_species と異なり 'both' を持たない）。マップ不能種は子行なし → 自動計算スキップ（fail-closed）。
CREATE TYPE medicine_dose_species AS ENUM ('dog', 'cat');

-- ------------------------------------
-- medicines: 製品軸の計算パラメータ
-- ------------------------------------
ALTER TABLE medicines
  ADD COLUMN calculation_type      medicine_calculation_type NOT NULL DEFAULT 'none', -- default-deny
  ADD COLUMN strength              numeric(10,4),   -- 製品含量。medicine_unit で分母解釈（per_tablet=mg/錠, per_ml=mg/mL, per_gram=mg/g）
  ADD COLUMN frequency_per_day     integer,         -- 1日投与回数（dose_basis=per_day の按分に使用）
  ADD COLUMN default_duration_days integer,         -- 既定投与日数（プリフィル補助）
  ALTER COLUMN default_quantity TYPE numeric(10,2); -- C2: 液剤 0.25 等の精度（widening・既存値は無損失）

-- per_weight 有効時は strength 必須（ゼロ除算・含量不明の自動計算を構造的に防ぐ）。
ALTER TABLE medicines
  ADD CONSTRAINT ck_medicines_per_weight_strength
    CHECK (calculation_type = 'none' OR strength IS NOT NULL);

-- per_weight 有効時は strength > 0（ゼロ除算防止。service validators と二重化）。
ALTER TABLE medicines
  ADD CONSTRAINT ck_medicines_strength_positive
    CHECK (strength IS NULL OR strength > 0);

ALTER TABLE medicines
  ADD CONSTRAINT ck_medicines_frequency_positive
    CHECK (frequency_per_day IS NULL OR frequency_per_day > 0);

ALTER TABLE medicines
  ADD CONSTRAINT ck_medicines_duration_positive
    CHECK (default_duration_days IS NULL OR default_duration_days > 0);

-- 子テーブルからの複合 FK ターゲット。id は PK で自明に一意だが、(id, clinic_id) を参照可能にする
-- ための一意制約（防御の加重: 子の非正規化 clinic_id が親と一致することを DB で保証する）。
ALTER TABLE medicines
  ADD CONSTRAINT uq_medicines_id_clinic UNIQUE (id, clinic_id);

COMMENT ON COLUMN medicines.calculation_type IS '#201 投与量計算方式。none=手動（既定・default-deny）/per_weight=mg/kg 線形自動計算';
COMMENT ON COLUMN medicines.strength IS '#201 製品含量（mg/単位）。分母は medicine_unit で解釈。per_weight 必須';

-- ------------------------------------
-- medicine_dose_params: 製品 × 種 の種軸パラメータ（1:N 子テーブル）
-- ------------------------------------
CREATE TABLE medicine_dose_params (
    id                BIGSERIAL                 PRIMARY KEY,
    -- clinic_id は親 medicines から非正規化。clinicScope(P4) を子に直適用するため。
    clinic_id         bigint                    NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    medicine_id       bigint                    NOT NULL,   -- 複合 FK(medicine_id, clinic_id) を下部で定義（CASCADE 禁止）
    species           medicine_dose_species     NOT NULL,
    dose_basis        medicine_dose_basis       NOT NULL DEFAULT 'per_administration',
    dose_per_kg       numeric(10,6)             NOT NULL,   -- mg/kg（per_weight は mg/kg 専用。CRI/μg は calculation_type=none で手動）
    min_mg_per_kg     numeric(10,6),                        -- 安全域下限（NULL=下限なし）
    max_mg_per_kg     numeric(10,6),                        -- 体重連動上限（NULL=上限なし）
    absolute_max_dose numeric(10,4),                        -- 体重非依存 mg/head 上限（NULL=上限なし）
    rounding_step     numeric(10,4),                        -- 丸め単位（NULL=丸めなし）
    rounding_mode     medicine_rounding_mode,               -- 丸め方向（NULL=丸めなし）
    notes             text                      NOT NULL DEFAULT '',
    created_at        timestamptz               NOT NULL DEFAULT now(),
    updated_at        timestamptz               NOT NULL DEFAULT now(),
    deleted_at        timestamptz,                          -- 論理削除（スナップショット再構築保全）
    CONSTRAINT ck_dose_per_kg_positive
        CHECK (dose_per_kg > 0),
    CONSTRAINT ck_dose_min_positive
        CHECK (min_mg_per_kg IS NULL OR min_mg_per_kg > 0),
    CONSTRAINT ck_dose_max_positive
        CHECK (max_mg_per_kg IS NULL OR max_mg_per_kg > 0),
    CONSTRAINT ck_dose_absolute_max_positive
        CHECK (absolute_max_dose IS NULL OR absolute_max_dose > 0),
    CONSTRAINT ck_dose_min_max
        CHECK (min_mg_per_kg IS NULL OR max_mg_per_kg IS NULL OR min_mg_per_kg <= max_mg_per_kg),
    -- 患者安全: per_weight 計算には上限が必須（丸め上げの silent 過量を防止）。service validators と二重化。
    CONSTRAINT ck_dose_upper_bound_required
        CHECK (max_mg_per_kg IS NOT NULL OR absolute_max_dose IS NOT NULL),
    CONSTRAINT ck_dose_rounding_step_positive
        CHECK (rounding_step IS NULL OR rounding_step > 0),
    -- rounding_step と rounding_mode はペアで設定/未設定（片方だけの指定を禁止）。
    CONSTRAINT ck_dose_rounding_pair
        CHECK ((rounding_step IS NULL) = (rounding_mode IS NULL)),
    -- 防御の加重: 子の clinic_id は必ず親 medicines の clinic_id と一致する（クロステナント不整合を DB で封殺）。
    -- service 層の clinic_id 設定（採用方針）と二重化。CASCADE 禁止につき ON DELETE RESTRICT。
    CONSTRAINT fk_dose_params_medicine_clinic
        FOREIGN KEY (medicine_id, clinic_id) REFERENCES medicines(id, clinic_id) ON DELETE RESTRICT
);

-- 同一 (medicine, species) の有効パラメータは 1 件（論理削除を除く）。
CREATE UNIQUE INDEX uq_dose_params_med_species
    ON medicine_dose_params (medicine_id, species)
    WHERE deleted_at IS NULL;

-- clinicScope 主クエリ（clinic_id, medicine_id）。
CREATE INDEX idx_dose_params_clinic_medicine
    ON medicine_dose_params (clinic_id, medicine_id);

COMMENT ON TABLE medicine_dose_params IS '#201 薬剤 × 種 の体重あたり投与量パラメータ（per_weight 自動計算用）';
COMMENT ON COLUMN medicine_dose_params.clinic_id IS '親 medicines から非正規化。clinicScope(P4) を子に直適用するため';
COMMENT ON COLUMN medicine_dose_params.dose_per_kg IS '体重あたり投与量 mg/kg。dose_basis で 1回量/1日量を解釈';
COMMENT ON COLUMN medicine_dose_params.absolute_max_dose IS '体重非依存の mg/head 上限。大型患者で max_mg_per_kg より binding になり得る';

-- ------------------------------------
-- treatments: C2 精度拡張 + 計算根拠スナップショット
-- ------------------------------------
ALTER TABLE treatments
  ALTER COLUMN quantity TYPE numeric(10,2),       -- C2: 液剤 0.25mL 等。既存 CHECK(quantity > 0) 維持
  ADD COLUMN dose_weight_kg      numeric(6,2),     -- 使用体重スナップショット（kg 正規化後）
  ADD COLUMN dose_weight_source  varchar(255),      -- 体重の出典（vital_records.id / 時刻 pin 等）
  ADD COLUMN dose_amount_mg      numeric(12,6),    -- 実効用量(mg)。安全域判定(C1)はこの丸め後の値
  ADD COLUMN dose_amount_unit    text,             -- 'mg' | 'ug'
  ADD COLUMN dose_param_snapshot jsonb;            -- 適用 species/dose_per_kg/strength/丸め設定/計算式版を値で固定

COMMENT ON COLUMN treatments.dose_amount_mg IS '#201 丸め後の実効用量(mg)。安全域(C1)判定に使用';
COMMENT ON COLUMN treatments.dose_param_snapshot IS '#201 計算根拠を値で固定（マスタ後変更・論理削除でも当時値を保全）';

-- =============================================================================
-- 010_add_checkup_packages.sql
-- #211 検査・健診パッケージ化 — 型付きフィールド機構（垂直スライス: 歯科検診）
--
-- 設計: examination ドメイン（exam_types → exam_type_fields → exam_results,
--       001_init.sql:692/1680）のパターンを踏襲し、checkup 用に正規化する。
--   - anchor は既存 checkup_types を「パッケージ」として拡張（新テーブルは作らない）。
--   - checkup_type_fields  : パッケージのフィールド定義（型付き）。
--   - checkup_field_results: 健診記録（checkups）に紐づく結果値。
--
-- マルチテナント: 両テーブルとも clinic_id NOT NULL を持ち、RLS は
--   tenant_clinic_id 直接ポリシーで保護する（001_init の clinic_id 自動ループは
--   既適用済みのため、後発の本マイグレーションで明示的に apply_rls_policy する）。
--
-- CASCADE 判断: migrations/CLAUDE.md の「純粋従属子行は CASCADE 許容例外」に従う。
--   - checkup_type_fields.checkup_type_id  → exam_type_fields.exam_type_id と同型（構成要素）
--   - checkup_field_results.checkup_id     → exam_results.exam_id と同型（純粋従属の結果行）。
--     RESTRICT にすると既存 medical_records → checkups CASCADE 連鎖を壊すため CASCADE 必須。
--   - checkup_field_results.checkup_type_field_id は nullable + ON DELETE SET NULL とし、
--     field_name/field_type/unit を非正規化スナップショットとして結果行に保持する
--     （exam_results.exam_type_field_id と同型。フィールド定義削除後も結果が自己記述的）。
-- =============================================================================

-- ------------------------------------
-- フィールド型 ENUM（6種）
-- ------------------------------------
CREATE TYPE checkup_field_type AS ENUM (
    'number',
    'single_select',
    'multi_select',
    'boolean',
    'checklist',
    'text'
);

-- ------------------------------------
-- checkup_type_fields（健診パッケージのフィールド定義マスタ）
-- ------------------------------------
CREATE TABLE checkup_type_fields (
    id              BIGSERIAL          PRIMARY KEY,
    clinic_id       bigint             NOT NULL REFERENCES clinics(id)       ON DELETE RESTRICT,
    checkup_type_id bigint             NOT NULL REFERENCES checkup_types(id) ON DELETE CASCADE,
    name            text               NOT NULL,
    field_type      checkup_field_type NOT NULL,
    unit            text               NOT NULL DEFAULT '',
    -- number 型の異常値判定基準（EXAM-001 と同じく min/max は任意）
    min_value       decimal(10,4),
    max_value       decimal(10,4),
    -- single_select / multi_select / checklist の選択肢定義: [{"value":"...","label":"..."}]
    options         jsonb              NOT NULL DEFAULT '[]'::jsonb,
    -- 暫定 seed フラグ（確定値は Notion 反映の別タスク）
    is_provisional  boolean            NOT NULL DEFAULT false,
    sort_order      integer            NOT NULL DEFAULT 0,
    created_at      timestamptz        NOT NULL DEFAULT now(),
    updated_at      timestamptz        NOT NULL DEFAULT now(),
    deleted_at      timestamptz
);

-- ------------------------------------
-- checkup_field_results（健診結果値: checkups の純粋従属子）
-- ------------------------------------
CREATE TABLE checkup_field_results (
    id                    BIGSERIAL          PRIMARY KEY,
    clinic_id             bigint             NOT NULL REFERENCES clinics(id)   ON DELETE RESTRICT,
    checkup_id            bigint             NOT NULL REFERENCES checkups(id)  ON DELETE CASCADE,
    checkup_type_field_id bigint                      REFERENCES checkup_type_fields(id) ON DELETE SET NULL,
    -- 非正規化スナップショット（フィールド定義削除後も結果が自己記述的であるため）
    field_name            text               NOT NULL DEFAULT '',
    field_type            checkup_field_type NOT NULL,
    unit                  text               NOT NULL DEFAULT '',
    -- 型別の値カラム（field_type に応じてサーバが該当列のみ書き込む）
    value_number          decimal(10,4),
    value_text            text               NOT NULL DEFAULT '',
    value_bool            boolean,
    value_list            text[]             NOT NULL DEFAULT '{}',
    -- number 型の異常値判定（EXAM-001 機構を再利用。exam_result_status を共用）
    ref_min               decimal(10,4),
    ref_max               decimal(10,4),
    is_abnormal           boolean            NOT NULL DEFAULT false,
    status                exam_result_status NOT NULL DEFAULT 'normal',
    sort_order            integer            NOT NULL DEFAULT 0,
    created_at            timestamptz        NOT NULL DEFAULT now(),
    updated_at            timestamptz        NOT NULL DEFAULT now()
);

-- ------------------------------------
-- インデックス（clinic_id を含む複合 + FK）
-- ------------------------------------
CREATE INDEX idx_checkup_type_fields_clinic_id        ON checkup_type_fields(clinic_id);
CREATE INDEX idx_checkup_type_fields_checkup_type_id  ON checkup_type_fields(checkup_type_id);
CREATE INDEX idx_checkup_type_fields_clinic_type_sort ON checkup_type_fields(clinic_id, checkup_type_id, sort_order) WHERE deleted_at IS NULL;

-- FindByCheckupID / ReplaceForCheckup はともに WHERE clinic_id = ? AND checkup_id = ? を発行する。
-- clinic_id 先頭（等値）+ checkup_id（等値）の複合で両クエリを単一インデックスで賄う
-- （migrations/CLAUDE.md「clinic_id を含む複合インデックス」規約）。clinic_id 単独はこの複合の前方一致で代替できるため作らない。
CREATE INDEX idx_checkup_field_results_clinic_checkup ON checkup_field_results(clinic_id, checkup_id);
-- checkup_id 単独は FindByPetID の JOIN（checkups.id = checkup_field_results.checkup_id）用に保持する。
CREATE INDEX idx_checkup_field_results_checkup_id      ON checkup_field_results(checkup_id);
-- checkup_type_field_id は ON DELETE SET NULL で NULL 行が増えるため部分インデックス。
CREATE INDEX idx_checkup_field_results_field_id        ON checkup_field_results(checkup_type_field_id) WHERE checkup_type_field_id IS NOT NULL;

-- ------------------------------------
-- RLS（clinic_id 直接ポリシー。001_init の自動ループ相当を後発で明示適用）
-- ------------------------------------
SELECT app_private.apply_rls_policy(
    'checkup_type_fields',
    'tenant_checkup_type_fields_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);

SELECT app_private.apply_rls_policy(
    'checkup_field_results',
    'tenant_checkup_field_results_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);

-- ------------------------------------
-- COMMENT
-- ------------------------------------
COMMENT ON TABLE checkup_type_fields    IS '健診パッケージの型付きフィールド定義マスタ（#211）';
COMMENT ON TABLE checkup_field_results  IS '健診結果値（checkups の純粋従属子・#211）';
COMMENT ON COLUMN checkup_type_fields.options       IS 'select/checklist の選択肢定義 [{"value","label"}]';
COMMENT ON COLUMN checkup_type_fields.is_provisional IS '暫定 seed フラグ（確定値は別タスク）';
COMMENT ON COLUMN checkup_field_results.field_type  IS 'checkup_type_fields.field_type の非正規化スナップショット';

-- 011_add_closing_am_start.sql
-- #215: AM 開始時刻（既定 09:00）を clinic_settings に追加する（additive）。
-- 締めレンジは AM=[am_start, boundary) / PM=[boundary, pm_end) / EMG=[pm_end, 翌日 am_start) になり、
-- 深夜 0:00〜am_start の緊急会計は前日の EMG に帰属する。
-- 既存の締め記録（cash_register_closes のスナップショット）は再計算しない（過去データ非破壊）。
ALTER TABLE clinic_settings
    ADD COLUMN IF NOT EXISTS closing_am_start time NOT NULL DEFAULT '09:00';

-- 012_add_clinical_result_composite_fk.sql
-- 臨床結果テーブルの DB レベル複合 FK（clinic_id 込み）で越境 INSERT/UPDATE を物理拒否する
-- （BE-refactor.md R3-7 / D13・defense-in-depth）。
--
-- 対象は checkup_field_results のみ。exam_type_fields は旧005で clinic_id 列と複合FKを獲得済みだが、
-- exam_results 自身は clinic_id 列を持たないため、(exam_type_field_id, clinic_id) の複合 FK は
-- 構造的に張れない。exam_results への同等防御は clinic_id 列の追加 + backfill という非 additive な
-- スキーマ拡張を要し、behavior-preserving リファクタの範囲外（別タスク）。
--
-- 挙動保存: migration 010 の患者結果値保護（フィールド定義削除時に結果値を残す ON DELETE SET NULL）を
-- 列指定 SET NULL（PostgreSQL 15+ 機能・本番は PG18）で維持する。親 checkup_type_fields を削除すると
-- checkup_type_field_id のみ NULL 化され（MATCH SIMPLE で FK チェックがスキップされる）、NOT NULL の
-- clinic_id と結果値スナップショットは保持される。単一列 SET NULL FK と挙動は完全に一致する。
--
-- 適用前提（手順1・必須）: 既存データに親子 clinic_id 不整合が無いことを検証してから適用すること
-- （違反行があると複合 FK 追加が失敗する）。STG 適用は db_reset 運用ルールに従う。

-- 親テーブルに複合 FK ターゲット用の UNIQUE(id, clinic_id) を追加する。
-- id は PK のため (id, clinic_id) は常に一意で、既存データに対して無条件に充足する（挙動非破壊）。
ALTER TABLE checkup_type_fields
    ADD CONSTRAINT uq_checkup_type_fields_id_clinic UNIQUE (id, clinic_id);

-- 既存の単一列 FK（010 の CREATE TABLE インライン FK・自動命名）を複合 FK に置換する。
ALTER TABLE checkup_field_results
    DROP CONSTRAINT IF EXISTS checkup_field_results_checkup_type_field_id_fkey;

ALTER TABLE checkup_field_results
    ADD CONSTRAINT fk_checkup_field_results_field_clinic
    FOREIGN KEY (checkup_type_field_id, clinic_id)
    REFERENCES checkup_type_fields (id, clinic_id)
    ON DELETE SET NULL (checkup_type_field_id);

-- 013_checkup_field_clinic_composite_fk.sql（旧 002_checkup_field_clinic_composite_fk.sql・#211 A6）
-- 健診 checkup_type_fields → checkup_types の DB レベル複合 FK（clinic_id 込み）で
-- 越境 INSERT/UPDATE を物理拒否する。012 の checkup_field_results 側（子→フィールド定義）と
-- 対をなす、フィールド定義→パッケージ親側の defense-in-depth。
--
-- 挙動保存: checkup_type_id は NOT NULL のため SET NULL は不要。既存のインライン単一列 FK が持つ
-- CASCADE 削除挙動（親 checkup_types 削除で checkup_type_fields も連動削除・構成要素として
-- 不可分な純粋従属データ、migrations/CLAUDE.md の許容例外）を複合 FK でもそのまま維持する。
--
-- インデックス: 親 checkup_types 削除時の CASCADE 検索（ソフトデリート済み行も含む子行捜索）は
-- 非パーシャル index idx_checkup_type_fields_checkup_type_id (checkup_type_id) で既にカバーされる
-- ため追加しない（WHERE deleted_at IS NULL の部分 index はプランナが述語を保証できないため
-- CASCADE 検索には使われない）。

-- 親テーブルに複合 FK ターゲット用の UNIQUE(id, clinic_id) を追加する。
-- id は PK のため (id, clinic_id) は常に一意で、既存データに対して無条件に充足する（挙動非破壊）。
ALTER TABLE checkup_types
    ADD CONSTRAINT uq_checkup_types_id_clinic UNIQUE (id, clinic_id);

-- 既存の単一列 FK（本ファイル checkup_type_fields の CREATE TABLE インライン FK・
-- PostgreSQL 自動命名。同一ファイル内で生成されるため名前は決定論的）を複合 FK に置換する。
ALTER TABLE checkup_type_fields
    DROP CONSTRAINT IF EXISTS checkup_type_fields_checkup_type_id_fkey;

ALTER TABLE checkup_type_fields
    ADD CONSTRAINT fk_checkup_type_fields_type_clinic
    FOREIGN KEY (checkup_type_id, clinic_id)
    REFERENCES checkup_types (id, clinic_id)
    ON DELETE CASCADE;

-- =============================================================================
-- 8. 増分マイグレーション統合アーカイブ (旧 002〜009 / 2026-07-27)
-- =============================================================================
-- 以下は独立ファイルとして管理されていた旧 002〜009 の原文を番号順に追記したもの。
-- 各ブロックの元コミットと SHA-256 は統合時の出典確認用に記録する。

-- Source file: 002_lstep_snapshot_import_clinic_fk.sql
-- Purpose: LSTEP friend snapshot と CSV import の clinic 所有関係を複合 FK で保証する。
-- Source commit: 4e8fb5b91
-- Source SHA-256: 10222c570054a80a5d47cf4b66e4235e92ca35643c9c23ef00a9ce8bca0086b6
-- Enforce tenant ownership across LSTEP friend snapshots and their CSV import.
-- Existing mismatches abort the migration before either constraint is changed.

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM lstep_friend_attribute_snapshots AS snapshot
        JOIN lstep_csv_imports AS csv_import
          ON csv_import.id = snapshot.csv_import_id
        WHERE snapshot.csv_import_id IS NOT NULL
          AND snapshot.clinic_id <> csv_import.clinic_id
    ) THEN
        RAISE EXCEPTION
            'cross-clinic lstep friend snapshot csv_import reference exists'
            USING ERRCODE = '23514';
    END IF;
END
$$;

ALTER TABLE lstep_csv_imports
    ADD CONSTRAINT uq_lstep_csv_imports_clinic_id_id
    UNIQUE (clinic_id, id);

ALTER TABLE lstep_friend_attribute_snapshots
    DROP CONSTRAINT IF EXISTS lstep_friend_attribute_snapshots_csv_import_id_fkey;

ALTER TABLE lstep_friend_attribute_snapshots
    ADD CONSTRAINT fk_lstep_snapshots_clinic_csv_import
    FOREIGN KEY (clinic_id, csv_import_id)
    REFERENCES lstep_csv_imports (clinic_id, id)
    ON DELETE RESTRICT;

-- Source file: 003_medical_records_appointment_id_index.sql
-- Purpose: 予約紐付きカルテの参照と cutover rollback/maintenance 確認を支える。
-- Source commit: e4e74d1fb
-- Source SHA-256: ac7250c9101d0d1b3d55958f18be547a0a21c0c2ab83f32caac3739aa536d675
-- Supports cutover rollback/maintenance checks and normal appointment-linked
-- medical-record lookups without holding avoidable long FK scans.
CREATE INDEX IF NOT EXISTS idx_medical_records_appointment_id
    ON medical_records (appointment_id)
    WHERE appointment_id IS NOT NULL;

-- Source file: 004_payment_splits_billing_id_index.sql
-- Purpose: payment graph 検証と billing 単位の集計を支える。
-- Source commit: 20e014b36
-- Source SHA-256: 647d8cf8c89377037c4a6fc07ac81d0ed3d47e5069ab4d3a5d418a56550f1886
-- Supports payment-graph verification and billing-scoped reporting without
-- scanning every clinic's payment splits. The existing
-- (clinic_id, billing_id) index remains the tenant-scoped access path.
CREATE INDEX IF NOT EXISTS idx_payment_splits_billing_id
    ON payment_splits (billing_id);

-- Source file: 005_exam_reference_ranges_and_clinic_fk.sql
-- Purpose: 検査項目の clinic 整合性と種別別基準範囲を DB 制約で保証する。
-- Source commit: b4d10e083
-- Source SHA-256: 1841ac6be05199f2666ad78bcd50db45527dadf08e0cc63549b4bf3e4fa44381
ALTER TABLE exam_types
    ADD CONSTRAINT uq_exam_types_id_clinic UNIQUE (id, clinic_id);

ALTER TABLE exam_type_fields
    ADD COLUMN clinic_id bigint;

UPDATE exam_type_fields AS field
SET clinic_id = exam_type.clinic_id
FROM exam_types AS exam_type
WHERE exam_type.id = field.exam_type_id;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM exam_type_fields AS field
        LEFT JOIN exam_types AS exam_type ON exam_type.id = field.exam_type_id
        WHERE field.clinic_id IS NULL
           OR field.clinic_id IS DISTINCT FROM exam_type.clinic_id
    ) THEN
        RAISE EXCEPTION
            'exam_type_fields clinic backfill is incomplete or inconsistent'
            USING ERRCODE = '23514';
    END IF;
END
$$;

ALTER TABLE exam_type_fields
    ALTER COLUMN clinic_id SET NOT NULL,
    ADD CONSTRAINT fk_exam_type_fields_clinic
        FOREIGN KEY (clinic_id) REFERENCES clinics(id) ON DELETE RESTRICT,
    ADD CONSTRAINT uq_exam_type_fields_id_clinic UNIQUE (id, clinic_id);

ALTER TABLE exam_type_fields
    DROP CONSTRAINT exam_type_fields_exam_type_id_fkey;

ALTER TABLE exam_type_fields
    ADD CONSTRAINT fk_exam_type_fields_type_clinic
    FOREIGN KEY (exam_type_id, clinic_id)
    REFERENCES exam_types (id, clinic_id)
    ON DELETE CASCADE;

CREATE INDEX idx_exam_type_fields_clinic_type_sort
    ON exam_type_fields (clinic_id, exam_type_id, sort_order);

CREATE TABLE exam_reference_ranges (
    id                    BIGSERIAL     PRIMARY KEY,
    clinic_id             bigint        NOT NULL,
    exam_type_field_id    bigint        NOT NULL,
    animal_species_id     bigint        NOT NULL,
    ref_min               decimal(10,4),
    ref_max               decimal(10,4),
    qualitative_min       text,
    qualitative_max       text,
    created_at            timestamptz   NOT NULL DEFAULT now(),
    updated_at            timestamptz   NOT NULL DEFAULT now(),
    CONSTRAINT fk_exam_reference_ranges_clinic
        FOREIGN KEY (clinic_id) REFERENCES clinics(id) ON DELETE RESTRICT,
    CONSTRAINT fk_exam_reference_ranges_field_clinic
        FOREIGN KEY (exam_type_field_id, clinic_id)
        REFERENCES exam_type_fields (id, clinic_id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_exam_reference_ranges_animal_species
        FOREIGN KEY (animal_species_id) REFERENCES animal_species(id) ON DELETE RESTRICT,
    CONSTRAINT uq_exam_reference_ranges_clinic_field_species
        UNIQUE (clinic_id, exam_type_field_id, animal_species_id),
    CONSTRAINT chk_exam_reference_ranges_ref_order
        CHECK (ref_min IS NULL OR ref_max IS NULL OR ref_min <= ref_max)
);

CREATE INDEX idx_exam_reference_ranges_animal_species
    ON exam_reference_ranges (animal_species_id);

-- RLS policies
SELECT app_private.apply_rls_policy(
    'exam_type_fields',
    'tenant_exam_type_fields_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);

SELECT app_private.apply_rls_policy(
    'exam_reference_ranges',
    'tenant_exam_reference_ranges_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);

-- Source file: 006_payment_splits_payment_method_clinic_fk.sql
-- Purpose: payment split と payment method の clinic 一致を複合 FK で保証する。
-- Source commit: c434c4e66
-- Source SHA-256: dfffc94e4e0f1990842a83840c7d759390291ad76679d1cc1c253433fe64e8e7
ALTER TABLE payment_methods
    ADD CONSTRAINT uq_payment_methods_id_clinic UNIQUE (id, clinic_id);

-- Adding the composite foreign key validates all existing rows and fails if a
-- non-NULL payment_method_id belongs to another clinic. Before applying this
-- migration, use the following query to identify conflicting rows:
--
-- SELECT
--     payment_split.id,
--     payment_split.clinic_id,
--     payment_split.payment_method_id,
--     payment_method.clinic_id AS payment_method_clinic_id
-- FROM payment_splits AS payment_split
-- LEFT JOIN payment_methods AS payment_method
--     ON payment_method.id = payment_split.payment_method_id
-- WHERE payment_split.payment_method_id IS NOT NULL
--   AND (
--       payment_method.id IS NULL
--       OR payment_method.clinic_id IS DISTINCT FROM payment_split.clinic_id
--   );

ALTER TABLE payment_splits
    ADD CONSTRAINT fk_payment_splits_payment_method_clinic
    FOREIGN KEY (payment_method_id, clinic_id)
    REFERENCES payment_methods (id, clinic_id)
    ON DELETE RESTRICT;

-- Source file: 007_add_pets_danger_reason.sql
-- Purpose: ペットの危険理由を記録する。
-- Source commit: 49029973b
-- Source SHA-256: 4d9afc0389285ca2b6ae9bb92ae4924e869a6e71f9051bbf774e9831dbf9047c
ALTER TABLE pets
    ADD COLUMN danger_reason text;

-- Source file: 008_add_billing_item_vaccination_provenance.sql
-- Purpose: 会計明細へ予防接種 provenance と clinic 制約を追加する。
-- Source commit: 65a0dd08d
-- Source SHA-256: daa676aed130da0ddefa30c4fd72e18b422dcc67ab919c7ea76dd0e40ac73d79
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

-- Source file: 009_add_billing_items_other_reason.sql
-- Purpose: other 分類の理由と会計明細の作成者を記録する。
-- Source commit: 7a64d9e63
-- Source SHA-256: 0923a782f31e1f736b17e3bf1935e164ec56db0de032114d7849eb6637ff2809
ALTER TABLE billing_items
    ADD COLUMN other_reason text,
    ADD COLUMN created_by bigint,
    ADD CONSTRAINT fk_billing_items_created_by
        FOREIGN KEY (created_by) REFERENCES staffs(id) ON DELETE RESTRICT;

CREATE INDEX idx_billing_items_created_by
    ON billing_items (created_by)
    WHERE created_by IS NOT NULL;

-- Source file: 002_pets_owners_clinic_composite_unique.sql
-- Purpose: pets / owners へ (clinic_id, id) の複合 UNIQUE を追加し、clinic 相関の複合 FK 参照先を用意する。
-- Source commit: a0165b1c5
-- Source SHA-256: 374f105139de1aea24253bc7adb24430245922230d5d1077f65c81c320f2cbbd
ALTER TABLE pets
    ADD CONSTRAINT uq_pets_clinic_id_id
    UNIQUE (clinic_id, id);

ALTER TABLE owners
    ADD CONSTRAINT uq_owners_clinic_id_id
    UNIQUE (clinic_id, id);

-- Source file: 003_add_pet_owners.sql
-- Purpose: ペットと飼い主の多対多を pet_owners で表現し、clinic 相関の複合 FK と RLS を適用する。
-- Source commit: b4933b50b
-- Source SHA-256: ca059c6bcc625e9e7b0f9001292653774e5c265eec4b528a0fcf57ce91a9357a
CREATE TABLE pet_owners (
    id           BIGSERIAL PRIMARY KEY,
    clinic_id    BIGINT NOT NULL REFERENCES clinics(id),
    pet_id       BIGINT NOT NULL,
    owner_id     BIGINT NOT NULL,
    relationship TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (pet_id, owner_id),
    FOREIGN KEY (clinic_id, pet_id) REFERENCES pets (clinic_id, id),
    FOREIGN KEY (clinic_id, owner_id) REFERENCES owners (clinic_id, id)
);

CREATE INDEX idx_pet_owners_clinic_pet ON pet_owners (clinic_id, pet_id);
CREATE INDEX idx_pet_owners_clinic_owner ON pet_owners (clinic_id, owner_id);

SELECT app_private.apply_rls_policy(
    'pet_owners',
    'tenant_clinic_id_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);

-- Source file: 004_add_exam_result_qualitative_bounds.sql
-- Purpose: 定性判定の境界値スナップショットを exam_results へ保持する（#249 U3）。
-- Source commit: cb3b1c448
-- Source SHA-256: 6ce59f1051132353a377bed341a6d78f6b0562fcb705ad8949cc99b4a2066997
ALTER TABLE exam_results
    ADD COLUMN qualitative_min text,
    ADD COLUMN qualitative_max text;

-- =============================================================================
-- 9. 増分マイグレーション統合アーカイブ (旧 002〜007 / 2026-07-29)
-- =============================================================================
-- 以下は独立ファイルとして管理されていた旧 002〜007 の原文を番号順に追記したもの。
-- 各ブロックの元コミットと SHA-256 は統合時の出典確認用に記録する。
-- 様式はセクション 8（2026-07-27 統合アーカイブ）に倣う。CREATE TABLE への畳み込みは行わず、
-- 増分適用時と同じ ALTER / CREATE INDEX / 関数・トリガー定義を原文のまま保持する。

-- Source file: 002_add_pets_version.sql
-- Purpose: pets.version 楽観ロック用カラムを追加する。
-- Source commit: 38e29b2ab
-- Source SHA-256: 940914030266fc5fe80db51dba2478c197167478a39fdd79cd5f85959621fcd2
ALTER TABLE pets
    ADD COLUMN version INTEGER NOT NULL DEFAULT 1;

-- Source file: 003_add_exam_results_exam_type_field_id_index.sql
-- Purpose: exam_results.exam_type_field_id 参照用の非一意 index。
-- Source commit: 09426bdec
-- Source SHA-256: de9ce47f9313219cf6b86aeea2e27a4b505aded77ae338efcce20b181ab41d23
CREATE INDEX idx_exam_results_exam_type_field_id ON exam_results (exam_type_field_id);

-- Source file: 004_add_inventory_quantity_check.sql
-- Purpose: inventory_items.quantity を非負に制約する CHECK（BUG-466）。
-- Source commit: 699c49e4a
-- Source SHA-256: e82838cd8d3d3436ba0d9d78b0ccca308eda875435b25cb8d218a9d274cfdf51
-- BUG-466: inventory_items.quantity を非負に制約し、負数在庫の永続化を DB 層でも拒否する。
ALTER TABLE inventory_items
  ADD CONSTRAINT chk_inventory_items_quantity_non_negative CHECK (quantity >= 0);

-- Source file: 005_payment_clinic_id_and_clinic_axis_composite_fks.sql
-- Purpose: payments.clinic_id と clinic 軸複合 FK を harden する（TASK-445 / BUG-454）。
-- Source commit: c5ddacb62
-- Source SHA-256: 755f3efa1d06d412488b8e4d0bfb250896b3062e3b4badf252887ed55a0c5c3a
-- TASK-445: payments.clinic_id + clinic-axis composite FK hardening
-- BUG-454: pets(clinic_id, owner_id) → owners(clinic_id, id)
--
-- Scope:
--   1. payments.clinic_id ADD + backfill from billings + NOT NULL + FK clinics
--   2. composite FK payments(billing_id, clinic_id) → billings(id, clinic_id)
--   3. composite FK payments(payment_method_id, clinic_id) → payment_methods(id, clinic_id)
--   4. BUG-454: pets(clinic_id, owner_id) → owners(clinic_id, id)
--   5. medical_records UNIQUE(id, clinic_id) + clinic-axis FKs to pets/owners
--   6. vaccinations clinic-axis FKs to pets / medical_records
--   7. billings clinic-axis FKs to pets / owners (nullable components: PG skips FK when any is NULL)
--
-- New constraints use RESTRICT (or default). Do not introduce delete cascade here.

-- =============================================================================
-- 1–3. payments.clinic_id + composite FKs
-- =============================================================================

ALTER TABLE payments
    ADD COLUMN clinic_id bigint;

-- Backfill from the owning billing row (1:1 on billing_id).
UPDATE payments AS payment
SET clinic_id = billing.clinic_id
FROM billings AS billing
WHERE billing.id = payment.billing_id
  AND payment.clinic_id IS NULL;

-- Adding NOT NULL / composite FKs validates all existing rows and fails if any
-- payment lacks a matching billing.clinic_id, or if payment_method_id (when set)
-- belongs to another clinic. Before applying this migration, use the following
-- queries to identify conflicting rows:
--
-- SELECT
--     payment.id,
--     payment.billing_id,
--     payment.clinic_id AS payment_clinic_id,
--     billing.clinic_id AS billing_clinic_id
-- FROM payments AS payment
-- LEFT JOIN billings AS billing
--     ON billing.id = payment.billing_id
-- WHERE payment.clinic_id IS NULL
--    OR billing.id IS NULL
--    OR billing.clinic_id IS DISTINCT FROM payment.clinic_id;
--
-- SELECT
--     payment.id,
--     payment.clinic_id,
--     payment.payment_method_id,
--     payment_method.clinic_id AS payment_method_clinic_id
-- FROM payments AS payment
-- LEFT JOIN payment_methods AS payment_method
--     ON payment_method.id = payment.payment_method_id
-- WHERE payment.payment_method_id IS NOT NULL
--   AND (
--       payment_method.id IS NULL
--       OR payment_method.clinic_id IS DISTINCT FROM payment.clinic_id
--   );

ALTER TABLE payments
    ALTER COLUMN clinic_id SET NOT NULL;

ALTER TABLE payments
    ADD CONSTRAINT fk_payments_clinic_id
    FOREIGN KEY (clinic_id)
    REFERENCES clinics (id)
    ON DELETE RESTRICT;

-- Parent UNIQUE already exists: uq_billings_id_clinic UNIQUE (id, clinic_id)
ALTER TABLE payments
    ADD CONSTRAINT fk_payments_billing_clinic
    FOREIGN KEY (billing_id, clinic_id)
    REFERENCES billings (id, clinic_id)
    ON DELETE RESTRICT;

-- Parent UNIQUE already exists: uq_payment_methods_id_clinic UNIQUE (id, clinic_id)
-- payment_method_id is nullable; PG skips the FK check when any component is NULL.
ALTER TABLE payments
    ADD CONSTRAINT fk_payments_payment_method_clinic
    FOREIGN KEY (payment_method_id, clinic_id)
    REFERENCES payment_methods (id, clinic_id)
    ON DELETE RESTRICT;

CREATE INDEX idx_payments_clinic_id ON payments (clinic_id);

-- =============================================================================
-- 4. BUG-454: pets owner must belong to the same clinic
-- =============================================================================

-- Parent UNIQUE already exists: uq_owners_clinic_id_id UNIQUE (clinic_id, id)
-- Before applying, identify cross-clinic pet.owner_id rows:
--
-- SELECT
--     pet.id,
--     pet.clinic_id,
--     pet.owner_id,
--     owner.clinic_id AS owner_clinic_id
-- FROM pets AS pet
-- LEFT JOIN owners AS owner
--     ON owner.id = pet.owner_id
-- WHERE owner.id IS NULL
--    OR owner.clinic_id IS DISTINCT FROM pet.clinic_id;

ALTER TABLE pets
    ADD CONSTRAINT fk_pets_clinic_owner
    FOREIGN KEY (clinic_id, owner_id)
    REFERENCES owners (clinic_id, id)
    ON DELETE RESTRICT;

-- =============================================================================
-- 5. medical_records: parent UNIQUE + clinic-axis FKs to pets/owners
-- =============================================================================

ALTER TABLE medical_records
    ADD CONSTRAINT uq_medical_records_id_clinic UNIQUE (id, clinic_id);

-- Parent UNIQUEs: uq_pets_clinic_id_id (clinic_id, id), uq_owners_clinic_id_id (clinic_id, id)
-- pet_id / owner_id are nullable; PG skips FK when any component is NULL.
-- Before applying, identify conflicting rows:
--
-- SELECT
--     medical_record.id,
--     medical_record.clinic_id,
--     medical_record.pet_id,
--     pet.clinic_id AS pet_clinic_id
-- FROM medical_records AS medical_record
-- LEFT JOIN pets AS pet
--     ON pet.id = medical_record.pet_id
-- WHERE medical_record.pet_id IS NOT NULL
--   AND (
--       pet.id IS NULL
--       OR pet.clinic_id IS DISTINCT FROM medical_record.clinic_id
--   );
--
-- SELECT
--     medical_record.id,
--     medical_record.clinic_id,
--     medical_record.owner_id,
--     owner.clinic_id AS owner_clinic_id
-- FROM medical_records AS medical_record
-- LEFT JOIN owners AS owner
--     ON owner.id = medical_record.owner_id
-- WHERE medical_record.owner_id IS NOT NULL
--   AND (
--       owner.id IS NULL
--       OR owner.clinic_id IS DISTINCT FROM medical_record.clinic_id
--   );

ALTER TABLE medical_records
    ADD CONSTRAINT fk_medical_records_clinic_pet
    FOREIGN KEY (clinic_id, pet_id)
    REFERENCES pets (clinic_id, id)
    ON DELETE RESTRICT;

ALTER TABLE medical_records
    ADD CONSTRAINT fk_medical_records_clinic_owner
    FOREIGN KEY (clinic_id, owner_id)
    REFERENCES owners (clinic_id, id)
    ON DELETE RESTRICT;

-- =============================================================================
-- 6. vaccinations: clinic-axis FKs to pets / medical_records
-- =============================================================================

-- Parent UNIQUEs: uq_pets_clinic_id_id (clinic_id, id),
--                 uq_medical_records_id_clinic (id, clinic_id) added above.
-- pet_id / medical_record_id are nullable; PG skips FK when any component is NULL.
-- Before applying, identify conflicting rows:
--
-- SELECT
--     vaccination.id,
--     vaccination.clinic_id,
--     vaccination.pet_id,
--     pet.clinic_id AS pet_clinic_id
-- FROM vaccinations AS vaccination
-- LEFT JOIN pets AS pet
--     ON pet.id = vaccination.pet_id
-- WHERE vaccination.pet_id IS NOT NULL
--   AND (
--       pet.id IS NULL
--       OR pet.clinic_id IS DISTINCT FROM vaccination.clinic_id
--   );
--
-- SELECT
--     vaccination.id,
--     vaccination.clinic_id,
--     vaccination.medical_record_id,
--     medical_record.clinic_id AS medical_record_clinic_id
-- FROM vaccinations AS vaccination
-- LEFT JOIN medical_records AS medical_record
--     ON medical_record.id = vaccination.medical_record_id
-- WHERE vaccination.medical_record_id IS NOT NULL
--   AND (
--       medical_record.id IS NULL
--       OR medical_record.clinic_id IS DISTINCT FROM vaccination.clinic_id
--   );

ALTER TABLE vaccinations
    ADD CONSTRAINT fk_vaccinations_clinic_pet
    FOREIGN KEY (clinic_id, pet_id)
    REFERENCES pets (clinic_id, id)
    ON DELETE RESTRICT;

ALTER TABLE vaccinations
    ADD CONSTRAINT fk_vaccinations_medical_record_clinic
    FOREIGN KEY (medical_record_id, clinic_id)
    REFERENCES medical_records (id, clinic_id)
    ON DELETE RESTRICT;

-- =============================================================================
-- 7. billings: clinic-axis FKs to pets / owners
-- =============================================================================

-- Parent UNIQUEs: uq_pets_clinic_id_id (clinic_id, id), uq_owners_clinic_id_id (clinic_id, id)
-- pet_id / owner_id are nullable; PG skips FK when any component is NULL.
-- Before applying, identify conflicting rows:
--
-- SELECT
--     billing.id,
--     billing.clinic_id,
--     billing.pet_id,
--     pet.clinic_id AS pet_clinic_id
-- FROM billings AS billing
-- LEFT JOIN pets AS pet
--     ON pet.id = billing.pet_id
-- WHERE billing.pet_id IS NOT NULL
--   AND (
--       pet.id IS NULL
--       OR pet.clinic_id IS DISTINCT FROM billing.clinic_id
--   );
--
-- SELECT
--     billing.id,
--     billing.clinic_id,
--     billing.owner_id,
--     owner.clinic_id AS owner_clinic_id
-- FROM billings AS billing
-- LEFT JOIN owners AS owner
--     ON owner.id = billing.owner_id
-- WHERE billing.owner_id IS NOT NULL
--   AND (
--       owner.id IS NULL
--       OR owner.clinic_id IS DISTINCT FROM billing.clinic_id
--   );

ALTER TABLE billings
    ADD CONSTRAINT fk_billings_clinic_pet
    FOREIGN KEY (clinic_id, pet_id)
    REFERENCES pets (clinic_id, id)
    ON DELETE RESTRICT;

ALTER TABLE billings
    ADD CONSTRAINT fk_billings_clinic_owner
    FOREIGN KEY (clinic_id, owner_id)
    REFERENCES owners (clinic_id, id)
    ON DELETE RESTRICT;

-- Source file: 006_payment_method_system_key_match.sql
-- Purpose: payments / payment_splits の method ⇔ payment_methods.system_key 一致を DB 境界で強制する（TASK-ADR003）。
-- Source commit: fd29b27fe
-- Source SHA-256: cecc85e1483d91227a6408a6904e17d521f2ab1a30d5778d9563f778c1646f43
-- TASK-ADR003: payments / payment_splits の method ⇔ payment_methods.system_key 一致を DB 境界で fail-closed にする。
--
-- Scope:
--   1. app_private.enforce_payment_method_system_key_match() を追加
--   2. payment_splits / payments に BEFORE INSERT OR UPDATE トリガーを付与
--
-- Rules:
--   - payment_method_id IS NULL の行はレガシー互換として許可（MATCH SIMPLE 相当）
--   - payment_method_id が非 NULL のとき、payment_methods に
--       id = payment_method_id AND system_key = method::text AND clinic_id = NEW.clinic_id
--     の行が存在することを要求する
--   - 削除時の連鎖削除制約は導入しない（RESTRICT / 既定のみ）
--
-- 前提: 005 により payments.clinic_id が存在する。
-- 既存の不整合行を洗い出す場合（適用前の任意確認）:
--
-- SELECT p.id, p.clinic_id, p.method, p.payment_method_id, pm.system_key, pm.clinic_id AS pm_clinic_id
-- FROM payments p
-- LEFT JOIN payment_methods pm ON pm.id = p.payment_method_id
-- WHERE p.payment_method_id IS NOT NULL
--   AND (
--     pm.id IS NULL
--     OR pm.system_key IS DISTINCT FROM p.method::text
--     OR pm.clinic_id IS DISTINCT FROM p.clinic_id
--   );
--
-- SELECT s.id, s.clinic_id, s.method, s.payment_method_id, pm.system_key, pm.clinic_id AS pm_clinic_id
-- FROM payment_splits s
-- LEFT JOIN payment_methods pm ON pm.id = s.payment_method_id
-- WHERE s.payment_method_id IS NOT NULL
--   AND (
--     pm.id IS NULL
--     OR pm.system_key IS DISTINCT FROM s.method::text
--     OR pm.clinic_id IS DISTINCT FROM s.clinic_id
--   );

CREATE SCHEMA IF NOT EXISTS app_private;

CREATE OR REPLACE FUNCTION app_private.enforce_payment_method_system_key_match()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    -- レガシー行: payment_method_id 未設定は method ENUM のみで運用する（MATCH SIMPLE 相当）。
    IF NEW.payment_method_id IS NULL THEN
        RETURN NEW;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM payment_methods AS pm
        WHERE pm.id = NEW.payment_method_id
          AND pm.system_key IS NOT DISTINCT FROM NEW.method::text
          AND pm.clinic_id = NEW.clinic_id
    ) THEN
        RAISE EXCEPTION
            '支払方法の不整合: payment_method_id=% の system_key が method=% と一致しません (clinic_id=%)',
            NEW.payment_method_id,
            NEW.method,
            NEW.clinic_id
            USING ERRCODE = '23514';
    END IF;

    RETURN NEW;
END;
$$;

COMMENT ON FUNCTION app_private.enforce_payment_method_system_key_match() IS
    'TASK-ADR003: payments/payment_splits の method と payment_methods.system_key の一致を fail-closed で強制する';

DROP TRIGGER IF EXISTS trg_payment_splits_method_system_key_match ON payment_splits;
CREATE TRIGGER trg_payment_splits_method_system_key_match
    BEFORE INSERT OR UPDATE ON payment_splits
    FOR EACH ROW
    EXECUTE FUNCTION app_private.enforce_payment_method_system_key_match();

DROP TRIGGER IF EXISTS trg_payments_method_system_key_match ON payments;
CREATE TRIGGER trg_payments_method_system_key_match
    BEFORE INSERT OR UPDATE ON payments
    FOR EACH ROW
    EXECUTE FUNCTION app_private.enforce_payment_method_system_key_match();

GRANT USAGE ON SCHEMA app_private TO PUBLIC;
GRANT EXECUTE ON FUNCTION app_private.enforce_payment_method_system_key_match() TO PUBLIC;

-- Source file: 007_owners_clinic_phone_unique.sql
-- Purpose: owners の clinic 内非空 phone 部分 unique index（POC-06 / U-X05-OWNER-PHONE）。
-- Source commit: ccfaaa311
-- Source SHA-256: cd73b506a4513abfd54836e362f30a8cba9766991fdabbcc1692b595634b0ec4
-- POC-06 / U-X05-OWNER-PHONE: feed-owner phone uniqueness at the DB boundary.
--
-- Mirrors uk_owners_clinic_email (partial unique on clinic_id + contact field)
-- so concurrent Create/Update cannot insert two active owners with the same
-- non-empty phone inside one clinic. Application ensureOwnerPhoneUnique remains
-- for friendly messages; this index is the fail-closed source of truth.
--
-- Empty phone ('') is excluded so legacy rows without phone can coexist.
-- Soft-deleted rows are excluded (deleted_at IS NULL).

CREATE UNIQUE INDEX IF NOT EXISTS uk_owners_clinic_phone
    ON owners (clinic_id, phone)
    WHERE deleted_at IS NULL AND phone <> '';

-- =============================================================================
-- 10. 増分マイグレーション統合アーカイブ (旧 002〜006 / 2026-07-31)
-- =============================================================================
-- 以下は独立ファイルとして管理されていた旧 002〜006 の原文を番号順に追記したもの。
-- 各ブロックの元コミットと SHA-256 は統合時の出典確認用に記録する。
-- 様式はセクション 9（2026-07-29 統合アーカイブ）に倣う。CREATE TABLE への畳み込みは行わず、
-- 増分適用時と同じ ALTER / CREATE INDEX / 関数・トリガー / CREATE TABLE 定義を原文のまま保持する。

-- Source file: 002_lstep_delivery_trigger_log_daily_unique.sql
-- Purpose: LSA-15: lstep_delivery_trigger_log の clinic/owner/type/JST-day 部分 unique index。
-- Source commit: aeb39c07487b22da74a2eb9b1ca6c673ac9b99f8
-- Source SHA-256: 3874c608ea06935e187787fcabf33b79f3526640614c102b37d95ae87122a2ae
-- LSA-15 / LANE-BE ④: day-grain uniqueness for delivery trigger double-fire defense.
-- Complements Go CreateIfAbsentToday (advisory lock). Expression uses Asia/Tokyo date
-- to match application "today" boundaries used by ExistsTodayByOwnerAndType.

CREATE UNIQUE INDEX IF NOT EXISTS uk_lstep_delivery_trigger_log_clinic_owner_type_day
    ON lstep_delivery_trigger_log (
        clinic_id,
        owner_id,
        trigger_type,
        ((scheduled_at AT TIME ZONE 'Asia/Tokyo')::date)
    );

COMMENT ON INDEX uk_lstep_delivery_trigger_log_clinic_owner_type_day IS
    'LSA-15: at most one trigger log per clinic/owner/type/JST day';

-- Source file: 003_closing_special_periods_exclude_overlap.sql
-- Purpose: POC-05: closing_special_periods の clinic+daterange EXCLUDE 制約。
-- Source commit: aeb39c07487b22da74a2eb9b1ca6c673ac9b99f8
-- Source SHA-256: d0881891184a220559f4778d2f13b2032bce894d143cf4cdcc269ccc31a6d405
-- POC-05 / LANE-BE ④: DB-level non-overlap for closing_special_periods.
-- Complements Go CreateCheckingOverlap/UpdateCheckingOverlap (clinic advisory lock).
-- btree_gist enables equality on clinic_id together with daterange overlap.

CREATE EXTENSION IF NOT EXISTS btree_gist;

ALTER TABLE closing_special_periods
    ADD CONSTRAINT excl_closing_special_periods_clinic_daterange
    EXCLUDE USING gist (
        clinic_id WITH =,
        daterange(start_date, end_date, '[]') WITH &&
    );

COMMENT ON CONSTRAINT excl_closing_special_periods_clinic_daterange ON closing_special_periods IS
    'POC-05: no overlapping special periods within the same clinic';

-- Source file: 004_add_identity_links.sql
-- Purpose: #239 Phase 1: 医院横断 owner/pet identity link 4 テーブル + 明示 RLS。
-- Source commit: fb11108c8a9faec5cf8af07f4d1bc0f3f95ab60f
-- Source SHA-256: 17c0028a0c294ab278b6e8b9512ddbfd5f5e2977229a3280a4e816fc41ecf829
-- #239 Phase 1: owner/pet identity link tables (manual link/unlink only).
-- Parents already have uq_owners_clinic_id_id / uq_pets_clinic_id_id in 001_init.
-- RLS is defense-in-depth; application runtime scope is the first boundary.
-- created_clinic_id is the group insert RLS anchor (immutable after insert).

-- ---------------------------------------------------------------------------
-- Owner identity groups
-- ---------------------------------------------------------------------------
CREATE TABLE owner_identity_groups (
    id                BIGSERIAL PRIMARY KEY,
    created_clinic_id BIGINT NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    version           BIGINT NOT NULL DEFAULT 1 CHECK (version >= 1),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ,
    UNIQUE (created_clinic_id, id)
);

CREATE TABLE owner_identity_group_members (
    id                      BIGSERIAL PRIMARY KEY,
    group_created_clinic_id BIGINT NOT NULL,
    group_id                BIGINT NOT NULL,
    clinic_id               BIGINT NOT NULL,
    owner_id                BIGINT NOT NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at              TIMESTAMPTZ,
    FOREIGN KEY (group_created_clinic_id, group_id)
        REFERENCES owner_identity_groups(created_clinic_id, id) ON DELETE RESTRICT,
    FOREIGN KEY (clinic_id, owner_id)
        REFERENCES owners(clinic_id, id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX uq_owner_identity_active_member
    ON owner_identity_group_members(clinic_id, owner_id)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uq_owner_identity_active_group_member
    ON owner_identity_group_members(group_id, clinic_id, owner_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_owner_identity_group_members_group
    ON owner_identity_group_members(group_id)
    WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- Pet identity groups (must hang under an owner identity group)
-- ---------------------------------------------------------------------------
CREATE TABLE pet_identity_groups (
    id                            BIGSERIAL PRIMARY KEY,
    created_clinic_id             BIGINT NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    owner_group_created_clinic_id BIGINT NOT NULL,
    owner_group_id                BIGINT NOT NULL,
    version                       BIGINT NOT NULL DEFAULT 1 CHECK (version >= 1),
    created_at                    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at                    TIMESTAMPTZ,
    UNIQUE (created_clinic_id, id),
    FOREIGN KEY (owner_group_created_clinic_id, owner_group_id)
        REFERENCES owner_identity_groups(created_clinic_id, id) ON DELETE RESTRICT
);

CREATE TABLE pet_identity_group_members (
    id                      BIGSERIAL PRIMARY KEY,
    group_created_clinic_id BIGINT NOT NULL,
    group_id                BIGINT NOT NULL,
    clinic_id               BIGINT NOT NULL,
    pet_id                  BIGINT NOT NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at              TIMESTAMPTZ,
    FOREIGN KEY (group_created_clinic_id, group_id)
        REFERENCES pet_identity_groups(created_clinic_id, id) ON DELETE RESTRICT,
    FOREIGN KEY (clinic_id, pet_id)
        REFERENCES pets(clinic_id, id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX uq_pet_identity_active_member
    ON pet_identity_group_members(clinic_id, pet_id)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uq_pet_identity_active_group_member
    ON pet_identity_group_members(group_id, clinic_id, pet_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_pet_identity_group_members_group
    ON pet_identity_group_members(group_id)
    WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- Immutable created_clinic_id (groups only)
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION app_private.prevent_identity_group_created_clinic_id_update()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.created_clinic_id IS DISTINCT FROM OLD.created_clinic_id THEN
        RAISE EXCEPTION 'created_clinic_id is immutable'
            USING ERRCODE = 'check_violation';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_owner_identity_groups_created_clinic_immutable
    BEFORE UPDATE ON owner_identity_groups
    FOR EACH ROW
    EXECUTE FUNCTION app_private.prevent_identity_group_created_clinic_id_update();

CREATE TRIGGER trg_pet_identity_groups_created_clinic_immutable
    BEFORE UPDATE ON pet_identity_groups
    FOR EACH ROW
    EXECUTE FUNCTION app_private.prevent_identity_group_created_clinic_id_update();

-- ---------------------------------------------------------------------------
-- RLS (ENABLE only; FORCE remains out of scope — app scope is first boundary)
-- ---------------------------------------------------------------------------
SELECT app_private.apply_rls_policy(
    'owner_identity_groups',
    'tenant_owner_identity_groups_isolation',
    'app_private.has_clinic_access(created_clinic_id)',
    'app_private.has_clinic_access(created_clinic_id)'
);

SELECT app_private.apply_rls_policy(
    'owner_identity_group_members',
    'tenant_owner_identity_group_members_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);

SELECT app_private.apply_rls_policy(
    'pet_identity_groups',
    'tenant_pet_identity_groups_isolation',
    'app_private.has_clinic_access(created_clinic_id)',
    'app_private.has_clinic_access(created_clinic_id)'
);

SELECT app_private.apply_rls_policy(
    'pet_identity_group_members',
    'tenant_pet_identity_group_members_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);

COMMENT ON TABLE owner_identity_groups IS
    '#239 owner identity link group; created_clinic_id is immutable RLS anchor; last-member unlink soft-deletes; no revive';
COMMENT ON TABLE owner_identity_group_members IS
    '#239 owner identity members; active uniqueness per (clinic_id, owner_id); soft-delete unlink';
COMMENT ON TABLE pet_identity_groups IS
    '#239 pet identity link group under owner identity group; created_clinic_id immutable';
COMMENT ON TABLE pet_identity_group_members IS
    '#239 pet identity members; active uniqueness per (clinic_id, pet_id); soft-delete unlink';

-- Source file: 005_line_webhook_bot_user_id.sql
-- Purpose: SEC-CS-F05-R1: LINE webhook destination bot user id ルーティングキー。
-- Source commit: 559b6560d0ea22f9865b07819f4eb63301009b38
-- Source SHA-256: 8c665c3b52048ffc1e5b8752ffc614ff215685221e9b9ca1643596d2b42876a7
-- SEC-CS-F05-R1: LINE webhook signature routing key.
-- destination in webhook body is the LINE Messaging API bot user ID.
-- Lookup is O(1) via this column; empty means not yet provisioned (excluded from unique index).

ALTER TABLE line_reservation_settings
    ADD COLUMN line_bot_user_id text NOT NULL DEFAULT '';

-- Only provisioned bot IDs must be unique. Unprovisioned rows share ''.
CREATE UNIQUE INDEX uq_line_reservation_settings_line_bot_user_id
    ON line_reservation_settings (line_bot_user_id)
    WHERE line_bot_user_id <> '';

COMMENT ON COLUMN line_reservation_settings.line_bot_user_id IS
    'LINE Messaging API bot user ID (webhook destination). Used for fixed-work signature routing (SEC-CS-F05-R1). Empty until provisioned.';

-- Source file: 006_medical_record_image_upload_quota.sql
-- Purpose: SEC-CS-F08-R1: medical_record_image_upload_quota 画像 upload quota lease テーブル。
-- Source commit: 7e5fa02d6de6dcde608c2ea0eca906ab614c0db4
-- Source SHA-256: 4501ebac04d3c34a262e3889495de2f69c2cd8737ae125389a27a27ffea4eac1
-- SEC-CS-F08-R1: authoritative medical-record image upload quota leases.
-- Shared across processes/replicas for concurrency, rate, and byte-budget gates.
-- Agents must not auto-apply; run `make migrate` after pull when this is present.

CREATE TABLE IF NOT EXISTS medical_record_image_upload_quota (
  id BIGSERIAL PRIMARY KEY,
  clinic_id BIGINT NOT NULL,
  staff_id BIGINT NOT NULL,
  declared_bytes BIGINT NOT NULL CHECK (declared_bytes >= 0),
  acquired_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  released_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_mri_upload_quota_clinic_acquired
  ON medical_record_image_upload_quota (clinic_id, acquired_at);

CREATE INDEX IF NOT EXISTS idx_mri_upload_quota_staff_acquired
  ON medical_record_image_upload_quota (clinic_id, staff_id, acquired_at);

CREATE INDEX IF NOT EXISTS idx_mri_upload_quota_inflight
  ON medical_record_image_upload_quota (clinic_id, staff_id)
  WHERE released_at IS NULL;

-- =============================================================================
-- 11. 増分マイグレーション統合アーカイブ (旧 002〜008 / 2026-08-04)
-- =============================================================================
-- 以下は独立ファイルとして管理されていた旧 002〜008 の原文を番号順に追記したもの。
-- 各ブロックの元コミットと SHA-256 は統合時の出典確認用に記録する。
-- 様式はセクション 10（2026-07-31 統合アーカイブ）に倣う。CREATE TABLE への畳み込みは行わず、
-- 増分適用時と同じ ALTER / CREATE INDEX / 関数・トリガー / CREATE TABLE 定義を原文のまま保持する。

-- Source file: 002_estimate_successor_and_numbering.sql
-- Purpose: TASK-012 FINAL B: 確定見積の後継ドラフト（supersedes）と見積番号採番の前提カラム。
-- Source commit: b65cf69ef56785c473ddd233624292a3c338401e
-- Source SHA-256: d93217de4b6eb5b1c264ce66b187937576852ab7b3da9cb9fb120c11e0b056c4
-- 002_estimate_successor_and_numbering.sql
-- TASK-012 FINAL B: 確定見積の後継ドラフト（supersedes）と見積番号採番の前提カラム。
-- 適用: USER が make migrate を手動実行すること（エージェントは auto-apply しない）。
-- CASCADE DELETE 禁止。001_init.sql は編集しない。

ALTER TABLE estimates
  ADD COLUMN IF NOT EXISTS supersedes_estimate_id BIGINT NULL
  REFERENCES estimates(id);

CREATE INDEX IF NOT EXISTS idx_estimates_clinic_supersedes
  ON estimates(clinic_id, supersedes_estimate_id)
  WHERE supersedes_estimate_id IS NOT NULL;

COMMENT ON COLUMN estimates.supersedes_estimate_id IS
  'Successor draft points to the locked (approved/rejected) estimate it corrects. Original row is never unlocked.';

-- Source file: 003_cash_register_close_append_only.sql
-- Purpose: W-013 FINAL B: レジ締め append-only 強化 + 締め後訂正（adjustment）テーブル。
-- Source commit: b65cf69ef56785c473ddd233624292a3c338401e
-- Source SHA-256: f07de224ba2f1987c9962252fb942dbd7a2d10ae5675f57599250641ab017f48
-- 003_cash_register_close_append_only.sql
-- W-013 FINAL B: レジ締め append-only 強化 + 締め後訂正（adjustment）テーブル。
-- 適用: USER が make migrate を手動実行すること（エージェントは auto-apply しない）。
-- CASCADE DELETE 禁止。001_init.sql は編集しない。
--
-- 方針:
--   1. cash_register_closes の soft-delete 再オープン経路を塞ぐ（partial UNIQUE → 完全 UNIQUE）
--   2. soft-deleted 行は active 行と衝突しなければ revive、衝突すれば migration を fail-closed
--   3. deleted_at 列は残すが app は soft-delete しない（UPDATE/DELETE は immutability trigger で拒否）
--   4. 締め後の会計訂正は cash_register_close_adjustments への append-only 追記で表現する

-- ---------------------------------------------------------------------------
-- 1. soft-deleted 行の整理（完全 UNIQUE 化の前提）
-- ---------------------------------------------------------------------------
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM cash_register_closes d
        WHERE d.deleted_at IS NOT NULL
          AND EXISTS (
              SELECT 1
              FROM cash_register_closes a
              WHERE a.deleted_at IS NULL
                AND a.clinic_id = d.clinic_id
                AND a.close_date = d.close_date
                AND a.period = d.period
          )
    ) THEN
        RAISE EXCEPTION
            '003_cash_register_close_append_only: soft-deleted cash_register_closes conflict with active rows for same (clinic_id, close_date, period); resolve manually before migrate';
    END IF;

    UPDATE cash_register_closes
    SET deleted_at = NULL
    WHERE deleted_at IS NOT NULL;
END $$;

-- ---------------------------------------------------------------------------
-- 2. partial UNIQUE → 完全 UNIQUE（soft-delete 再オープン不可）
-- ---------------------------------------------------------------------------
DROP INDEX IF EXISTS uq_cash_register_closes_date_period;

CREATE UNIQUE INDEX uq_cash_register_closes_date_period
    ON cash_register_closes (clinic_id, close_date, period);

COMMENT ON TABLE cash_register_closes IS
    'レジ締めレコード（FEAT-368）。append-only。更新・削除・soft-delete 再開は不可。締め後訂正は cash_register_close_adjustments へ追記する（W-013）。';

-- ---------------------------------------------------------------------------
-- 3. cash_register_close_adjustments（締め後訂正の append-only 台帳）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS cash_register_close_adjustments (
    id                   BIGSERIAL    PRIMARY KEY,
    clinic_id            BIGINT       NOT NULL REFERENCES clinics(id),
    close_id             BIGINT       NOT NULL REFERENCES cash_register_closes(id),
    billing_id           BIGINT       NOT NULL,
    accounting_delta     BIGINT       NOT NULL DEFAULT 0,
    cash_movement_amount BIGINT       NOT NULL DEFAULT 0,
    reason               TEXT         NOT NULL,
    actor_id             BIGINT,
    executed_at          TIMESTAMPTZ  NOT NULL DEFAULT now(),
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_cash_register_close_adjustments_close
    ON cash_register_close_adjustments (clinic_id, close_id);

CREATE INDEX IF NOT EXISTS idx_cash_register_close_adjustments_executed
    ON cash_register_close_adjustments (clinic_id, executed_at);

COMMENT ON TABLE cash_register_close_adjustments IS
    'レジ締め後の会計訂正台帳（W-013 append-only）。close 自体の reverse/取消は productize しない。';
COMMENT ON COLUMN cash_register_close_adjustments.accounting_delta IS
    '会計合計の増減（円）。取得できない場合は 0 でも reason は必須。';
COMMENT ON COLUMN cash_register_close_adjustments.cash_movement_amount IS
    '現金移動額（円）。会計のみの訂正では 0。';

-- ---------------------------------------------------------------------------
-- 4. immutability triggers（UPDATE/DELETE を DB 層で拒否）
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION app_private.prevent_cash_register_close_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'cash_register_closes is append-only; UPDATE/DELETE are not allowed'
        USING ERRCODE = 'check_violation';
END;
$$;

DROP TRIGGER IF EXISTS trg_cash_register_closes_immutable ON cash_register_closes;
CREATE TRIGGER trg_cash_register_closes_immutable
    BEFORE UPDATE OR DELETE ON cash_register_closes
    FOR EACH ROW
    EXECUTE FUNCTION app_private.prevent_cash_register_close_mutation();

CREATE OR REPLACE FUNCTION app_private.prevent_cash_register_close_adjustment_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'cash_register_close_adjustments is append-only; UPDATE/DELETE are not allowed'
        USING ERRCODE = 'check_violation';
END;
$$;

DROP TRIGGER IF EXISTS trg_cash_register_close_adjustments_immutable ON cash_register_close_adjustments;
CREATE TRIGGER trg_cash_register_close_adjustments_immutable
    BEFORE UPDATE OR DELETE ON cash_register_close_adjustments
    FOR EACH ROW
    EXECUTE FUNCTION app_private.prevent_cash_register_close_adjustment_mutation();

-- Source file: 004_examination_revisions.sql
-- Purpose: TASK-027 Slice A: append-only examination revision foundation and first-confirm pointer CAS.
-- Source commit: 046615f4bc923869f189c4e104e27d0539d8c88d
-- Source SHA-256: 935466024cd36d98e0bf991b14080930712a0602a8f5b9afeb4199c7ec6f8037
-- 004_examination_revisions.sql
-- TASK-027 Slice A: append-only examination revision foundation and first-confirm pointer CAS.
-- Apply only through the USER-operated migration workflow. This file is additive; 001-003 remain unchanged.

ALTER TABLE exams
    ADD COLUMN current_revision_version BIGINT NULL
        CHECK (current_revision_version >= 1),
    ADD CONSTRAINT uq_exams_clinic_id_id
        UNIQUE (clinic_id, id);

CREATE TABLE examination_revisions (
    id                      BIGSERIAL   PRIMARY KEY,
    clinic_id               BIGINT      NOT NULL,
    examination_id          BIGINT      NOT NULL,
    version                 BIGINT      NOT NULL CHECK (version >= 1),
    kind                    TEXT        NOT NULL CHECK (kind IN ('working', 'official')),
    status                  exam_status NOT NULL,
    medical_record_id       BIGINT,
    pet_id                  BIGINT,
    medical_record_owner_id BIGINT,
    pet_owner_id            BIGINT,
    animal_species_id       BIGINT,
    exam_type_id            BIGINT      NOT NULL,
    doctor_id               BIGINT,
    job_id                  UUID,
    actor_id                BIGINT      NOT NULL,
    date                    DATE        NOT NULL,
    result_summary          TEXT        NOT NULL DEFAULT '',
    machine                 TEXT        NOT NULL DEFAULT '',
    display_snapshot        JSONB       NOT NULL,
    schema_version          SMALLINT    NOT NULL DEFAULT 1,
    change_reason           TEXT        NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_examination_revisions_clinic_exam_version
        UNIQUE (clinic_id, examination_id, version),
    CONSTRAINT ck_examination_revisions_schema_version
        CHECK (schema_version = 1),
    CONSTRAINT ck_examination_revisions_change_reason
        CHECK (change_reason IS NULL OR btrim(change_reason) <> ''),
    CONSTRAINT ck_examination_revisions_kind_status
        CHECK (
            (kind = 'official' AND status = 'confirmed') OR
            (kind = 'working' AND status IN ('pending', 'in_progress', 'result_entered', 'completed'))
        ),
    CONSTRAINT ck_examination_revisions_display_snapshot
        CHECK (
            jsonb_typeof(display_snapshot) = 'object' AND
            COALESCE(jsonb_typeof(display_snapshot -> 'medical_record_no'), '') = 'string' AND
            COALESCE(jsonb_typeof(display_snapshot -> 'pet_name'), '') = 'string' AND
            COALESCE(jsonb_typeof(display_snapshot -> 'medical_record_owner_name'), '') = 'string' AND
            COALESCE(jsonb_typeof(display_snapshot -> 'pet_owner_name'), '') = 'string' AND
            COALESCE(jsonb_typeof(display_snapshot -> 'species_name'), '') = 'string' AND
            COALESCE(jsonb_typeof(display_snapshot -> 'exam_type_name'), '') = 'string' AND
            COALESCE(jsonb_typeof(display_snapshot -> 'doctor_name'), '') = 'string'
        )
);

CREATE TABLE examination_revision_items (
    id                   BIGSERIAL          PRIMARY KEY,
    clinic_id            BIGINT             NOT NULL,
    examination_id       BIGINT             NOT NULL,
    version              BIGINT             NOT NULL CHECK (version >= 1),
    exam_type_field_id   BIGINT,
    name                 TEXT               NOT NULL DEFAULT '',
    inspection_value     TEXT               NOT NULL DEFAULT '',
    normal_value         TEXT               NOT NULL DEFAULT '',
    result               TEXT               NOT NULL DEFAULT '',
    unit                 TEXT               NOT NULL DEFAULT '',
    reference_value      TEXT               NOT NULL DEFAULT '',
    ref_min              DECIMAL(10,4),
    ref_max              DECIMAL(10,4),
    qualitative_min      TEXT,
    qualitative_max      TEXT,
    is_assessed          BOOLEAN            NOT NULL,
    is_abnormal          BOOLEAN            NOT NULL,
    status               exam_result_status NOT NULL,
    sort_order           INTEGER            NOT NULL DEFAULT 0,
    created_at           TIMESTAMPTZ        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT ck_examination_revision_items_reference_range
        CHECK (ref_min IS NULL OR ref_max IS NULL OR ref_min <= ref_max)
);

ALTER TABLE examination_revisions
    ADD CONSTRAINT fk_examination_revisions_exam
        FOREIGN KEY (clinic_id, examination_id)
        REFERENCES exams (clinic_id, id)
        ON DELETE RESTRICT NOT DEFERRABLE;

ALTER TABLE examination_revision_items
    ADD CONSTRAINT fk_examination_revision_items_revision
        FOREIGN KEY (clinic_id, examination_id, version)
        REFERENCES examination_revisions (clinic_id, examination_id, version)
        ON DELETE RESTRICT NOT DEFERRABLE;

ALTER TABLE exams
    ADD CONSTRAINT fk_exams_current_revision
        FOREIGN KEY (clinic_id, id, current_revision_version)
        REFERENCES examination_revisions (clinic_id, examination_id, version)
        ON DELETE RESTRICT NOT DEFERRABLE;

CREATE INDEX idx_examination_revisions_latest_kind
    ON examination_revisions (clinic_id, examination_id, kind, version DESC);

CREATE INDEX idx_examination_revision_items_revision_sort
    ON examination_revision_items (clinic_id, examination_id, version, sort_order, id);

CREATE INDEX idx_exams_current_revision_pointer
    ON exams (clinic_id, id, current_revision_version);

CREATE OR REPLACE FUNCTION app_private.prevent_examination_revision_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'examination_revisions is append-only; UPDATE/DELETE are not allowed'
        USING ERRCODE = 'check_violation';
END;
$$;

CREATE TRIGGER trg_examination_revisions_immutable
    BEFORE UPDATE OR DELETE ON examination_revisions
    FOR EACH ROW
    EXECUTE FUNCTION app_private.prevent_examination_revision_mutation();

CREATE OR REPLACE FUNCTION app_private.prevent_examination_revision_item_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'examination_revision_items is append-only; UPDATE/DELETE are not allowed'
        USING ERRCODE = 'check_violation';
END;
$$;

CREATE TRIGGER trg_examination_revision_items_immutable
    BEFORE UPDATE OR DELETE ON examination_revision_items
    FOR EACH ROW
    EXECUTE FUNCTION app_private.prevent_examination_revision_item_mutation();

CREATE OR REPLACE FUNCTION app_private.enforce_examination_revision_item_insert_version()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    selected_version BIGINT;
BEGIN
    SELECT current_revision_version
    INTO selected_version
    FROM exams
    WHERE clinic_id = NEW.clinic_id
      AND id = NEW.examination_id
    FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION
            'examination revision item has no matching examination: clinic_id=%, examination_id=%',
            NEW.clinic_id,
            NEW.examination_id
            USING ERRCODE = 'check_violation';
    END IF;

    IF NEW.version <> COALESCE(selected_version, 0) + 1 THEN
        RAISE EXCEPTION
            'examination revision item version must be the next unselected version: selected=%, attempted=%',
            selected_version,
            NEW.version
            USING ERRCODE = 'check_violation';
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_examination_revision_items_insert_version
    BEFORE INSERT ON examination_revision_items
    FOR EACH ROW
    EXECUTE FUNCTION app_private.enforce_examination_revision_item_insert_version();

SELECT app_private.apply_rls_policy(
    'examination_revisions',
    'tenant_examination_revisions_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);

SELECT app_private.apply_rls_policy(
    'examination_revision_items',
    'tenant_examination_revision_items_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);

-- Source file: 005_accounting_completion_idempotency.sql
-- Purpose: BUG-018: atomic accounting completion idempotency keys on billings.
-- Source commit: 75f8912fc2a8d3b2fb5c0290a766bb0f5fc12ac5
-- Source SHA-256: 777e3694ecace1dbb5ae8683cabeaa3b2a2dc888dfa1e464fe5b9e69085d181a
-- 005_accounting_completion_idempotency.sql
-- BUG-018: atomic accounting completion idempotency keys on billings.
-- 適用: USER が make migrate を手動実行すること（エージェントは auto-apply しない）。
-- CASCADE DELETE 禁止。既存 migration (001-004) は編集しない。
--
-- 方針:
--   1. completion_request_id / completion_request_hash を billings に追加
--   2. clinic-first UNIQUE で同一キーの再利用を拒否（soft-delete 後も再利用不可）
--   3. 部分 UNIQUE にしない（deleted_at を WHERE から外し、論理削除後の key 再利用を塞ぐ）

ALTER TABLE billings
    ADD COLUMN IF NOT EXISTS completion_request_id   UUID,
    ADD COLUMN IF NOT EXISTS completion_request_hash TEXT;

-- clinic-first composite unique（prompt 契約: UNIQUE(clinic_id, completion_request_id)）
-- NULL completion_request_id は PostgreSQL の UNIQUE で複数許可（legacy create 経路）
CREATE UNIQUE INDEX IF NOT EXISTS uq_billings_clinic_completion_request_id
    ON billings (clinic_id, completion_request_id);

COMMENT ON COLUMN billings.completion_request_id IS
    'BUG-018: POST /accountings/complete の Idempotency-Key（UUID）。soft-delete 後も再利用不可。';
COMMENT ON COLUMN billings.completion_request_hash IS
    'BUG-018: complete request の正規化 digest（hex）。同一 key で異 digest は 409。';

-- Source file: 006_checkup_package_import.sql
-- Purpose: TASK-374 / #211 / DEC-59: versioned clinic-scoped checkup package import.
-- Source commit: a27ea399d787f4e22af8fffacc0bfd4770129b74
-- Source SHA-256: 9a7f4b4ec66f7a97f59d06c437bcc0fc3086c6928f403c81d8b1894fcc15cbf1
-- 006_checkup_package_import.sql
-- TASK-374 / #211 / DEC-59: versioned clinic-scoped checkup package import
-- Additive only. Do not edit applied migrations. Agent does not apply this file.
-- No BEGIN/COMMIT here: cmd/migrate wraps each file in its own transaction, so a
-- COMMIT in the file ends that transaction early and the runner's commit then fails
-- with "unexpected transaction status idle". Every statement stays re-run safe.

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

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint con
        JOIN pg_class rel ON rel.oid = con.conrelid
        JOIN pg_namespace nsp ON nsp.oid = rel.relnamespace
        WHERE nsp.nspname = 'public'
          AND rel.relname = 'checkup_types'
          AND con.conname = 'fk_checkup_types_parent_clinic'
    ) THEN
        ALTER TABLE checkup_types
            ADD CONSTRAINT fk_checkup_types_parent_clinic
            FOREIGN KEY (parent_id, clinic_id)
            REFERENCES checkup_types (id, clinic_id)
            ON DELETE SET NULL (parent_id);
    END IF;
END $$;

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

-- Source file: 007_lab_import_job_status_reverted.sql
-- Purpose: TASK-032: lab import job 状態機械に terminal compensation 値 'reverted' を追加する。
-- Source commit: 0d54efa7e37d61ccef21b805794b9de0940bdc2d
-- Source SHA-256: eabec224ea5d18630f037676ca25ac11b243ef24276bb092fa76fc12014f72ee
-- 007_lab_import_job_status_reverted.sql
-- TASK-032: lab import job 状態機械に terminal compensation 値 'reverted' を追加する。
-- 適用: USER が make migrate を手動実行すること（エージェントは auto-apply しない）。
-- CASCADE DELETE 禁止。既適用 migration / seed は編集しない。
--
-- PostgreSQL 制約: 同一 transaction 内で ADD VALUE した新 enum 値は参照できない。
-- 本ファイルは ADD VALUE のみ。新値を使う DDL・DML は 008 以降に分離する。
-- ファイル内に BEGIN;/COMMIT; を書かない（cmd/migrate が各ファイルを自前 tx で包む）。

ALTER TYPE lab_import_job_status ADD VALUE IF NOT EXISTS 'reverted';

-- Source file: 008_lab_import_revert_compensation.sql
-- Purpose: TASK-032: lab import compensating revert の複合 FK・receipt・retraction・RLS。
-- Source commit: 0d54efa7e37d61ccef21b805794b9de0940bdc2d
-- Source SHA-256: 851e9571f2a224a220009ba748e3a985d9f5079de75047f713a1b6fdf508d243
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

-- =============================================================================
-- Section 12: post-integration tenant-boundary hardening
-- =============================================================================
-- cash_register_close_adjustments was introduced after the section-6 automatic
-- RLS loop. Preserve the archived source blocks above byte-for-byte and apply the
-- missing clinic-correlated FKs and explicit RLS policy here.

ALTER TABLE cash_register_closes
    ADD CONSTRAINT uq_cash_register_closes_id_clinic
        UNIQUE (id, clinic_id);

ALTER TABLE staffs
    ADD CONSTRAINT uq_staffs_id_clinic
        UNIQUE (id, clinic_id);

ALTER TABLE cash_register_close_adjustments
    DROP CONSTRAINT IF EXISTS cash_register_close_adjustments_close_id_fkey,
    ADD CONSTRAINT fk_cash_register_close_adjustments_close_clinic
        FOREIGN KEY (close_id, clinic_id)
        REFERENCES cash_register_closes (id, clinic_id)
        ON DELETE RESTRICT,
    ADD CONSTRAINT fk_cash_register_close_adjustments_billing_clinic
        FOREIGN KEY (billing_id, clinic_id)
        REFERENCES billings (id, clinic_id)
        ON DELETE RESTRICT,
    ADD CONSTRAINT fk_cash_register_close_adjustments_actor_clinic
        FOREIGN KEY (actor_id, clinic_id)
        REFERENCES staffs (id, clinic_id)
        ON DELETE RESTRICT;

SELECT app_private.apply_rls_policy(
    'cash_register_close_adjustments',
    'tenant_cash_register_close_adjustments_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);

-- =============================================================================
-- 13. 増分マイグレーション統合 (003 / 004 / 005 / 006 / 2026-08-19)
-- =============================================================================
-- 本番運用前の DB リセット前提で独立ファイルを 001 へ畳み込む。
-- 005 の ADD VALUE は CREATE TYPE へ取り込んだためここでは繰り返さない。
-- Source SHA-256:
--   003_add_estimates_pet_id.sql          94795f5b76fef336e18fed758527ececd126c23df44ade31f386c227595d9604
--   004_lab_device_item_masters.sql       5db554b103c787d20fc9181fcbff593a72468003400517d7ddb0133201586abf
--   005_lab_import_source_type_device.sql 82c2bb2cf268fa42f47a00a07c520c6fdc56c32d4ed25b3522ef0ee5667e0145
--   006_lab_device_receive.sql            1254627bf033fa57602b14d2958f4e7ce185894fb13c4cec196d8f892fc66349

-- Source file: 003_add_estimates_pet_id.sql
-- Purpose: BUG-009 見積 pet_id。CREATE TABLE 途中へ畳み込むと seed COPY の列順が崩れる。
-- estimates.csv は supersedes_estimate_id の後ろに pet_id がある（SELECT * 順）。
ALTER TABLE estimates
  ADD COLUMN IF NOT EXISTS pet_id bigint REFERENCES pets(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_estimates_clinic_pet
  ON estimates (clinic_id, pet_id)
  WHERE pet_id IS NOT NULL AND deleted_at IS NULL;

COMMENT ON COLUMN estimates.pet_id IS
    'BUG-009: /estimates/new?petId= の永続紐付け。clinic 所有は service で検証する';

-- Source file: 004_lab_device_item_masters.sql
CREATE TABLE lab_device_item_masters (
    id                  bigserial       PRIMARY KEY,
    clinic_id           bigint          NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    source_type         varchar(32)     NOT NULL,
    device_item_code    varchar(64)     NOT NULL,
    display_name        varchar(100)    NOT NULL,
    unit                varchar(32)     NOT NULL DEFAULT '',
    value_shape         varchar(32)     NOT NULL,
    exam_type_field_id  bigint,
    sort_order          integer         NOT NULL DEFAULT 0,
    is_active           boolean         NOT NULL DEFAULT true,
    created_at          timestamptz     NOT NULL DEFAULT now(),
    updated_at          timestamptz     NOT NULL DEFAULT now(),
    CONSTRAINT chk_lab_device_item_masters_source_type
        CHECK (source_type IN ('fuji_nx600', 'fuji_au10v', 'arkray_pu4010')),
    CONSTRAINT chk_lab_device_item_masters_value_shape
        CHECK (value_shape IN ('numeric', 'inequality', 'qual_and_num', 'dash', 'text')),
    CONSTRAINT uq_lab_device_item_masters_clinic_source_code
        UNIQUE (clinic_id, source_type, device_item_code)
);

ALTER TABLE lab_device_item_masters
    ADD CONSTRAINT fk_lab_device_item_masters_field_clinic
    FOREIGN KEY (exam_type_field_id, clinic_id)
    REFERENCES exam_type_fields (id, clinic_id)
    ON DELETE RESTRICT;

CREATE INDEX idx_lab_device_item_masters_clinic_source
    ON lab_device_item_masters (clinic_id, source_type, sort_order);

COMMENT ON TABLE lab_device_item_masters IS
    '検査機器コード→検査項目フィールド。初期25行は埋め込みカタログ。CSVアップロードは製品経路にしない';
COMMENT ON COLUMN lab_device_item_masters.device_item_code IS
    '電文コードそのもの。Na-P の -P を削らない';
COMMENT ON COLUMN lab_device_item_masters.exam_type_field_id IS
    '未設定なら persist しない（needs_review）';

-- Source file: 006_lab_device_receive.sql
ALTER TABLE lab_import_jobs
    ADD COLUMN IF NOT EXISTS pet_id bigint,
    ADD COLUMN IF NOT EXISTS measured_at timestamptz,
    ADD COLUMN IF NOT EXISTS received_at timestamptz,
    ADD COLUMN IF NOT EXISTS device_hint varchar(32) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS specimen_id_raw varchar(64) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS unmapped_item_count int NOT NULL DEFAULT 0;

ALTER TABLE lab_import_jobs
    ADD CONSTRAINT fk_lab_import_jobs_pet_clinic
    FOREIGN KEY (clinic_id, pet_id)
    REFERENCES pets (clinic_id, id)
    ON DELETE RESTRICT;

CREATE UNIQUE INDEX IF NOT EXISTS uq_lab_import_jobs_clinic_source_fingerprint
    ON lab_import_jobs (clinic_id, source_type, source_fingerprint)
    WHERE source_fingerprint <> '';

CREATE INDEX IF NOT EXISTS idx_lab_import_jobs_clinic_unlinked
    ON lab_import_jobs (clinic_id, received_at DESC)
    WHERE pet_id IS NULL
      AND source_type IN ('fuji_nx600', 'fuji_au10v', 'arkray_pu4010');

COMMENT ON COLUMN lab_import_jobs.pet_id IS 'device 行のみ。未紐付けは NULL。検体IDでは埋めない';
COMMENT ON COLUMN lab_import_jobs.measured_at IS '電文日時。検査日の正';
COMMENT ON COLUMN lab_import_jobs.received_at IS 'デコード成功時刻。検査日にしない';
COMMENT ON COLUMN lab_import_jobs.specimen_id_raw IS '表示専用。紐付けキーにしない。ログに出さない';

CREATE TABLE lab_import_job_items (
    id                  bigserial       PRIMARY KEY,
    clinic_id           bigint          NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    job_id              uuid            NOT NULL,
    device_item_code    varchar(64)     NOT NULL,
    value_raw           varchar(64)     NOT NULL DEFAULT '',
    unit                varchar(32)     NOT NULL DEFAULT '',
    flag                varchar(32)     NOT NULL DEFAULT '',
    exam_type_field_id  bigint,
    needs_review        boolean         NOT NULL DEFAULT false,
    sort_order          integer         NOT NULL DEFAULT 0,
    created_at          timestamptz     NOT NULL DEFAULT now(),
    CONSTRAINT fk_lab_import_job_items_job_clinic
        FOREIGN KEY (clinic_id, job_id)
        REFERENCES lab_import_jobs (clinic_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_lab_import_job_items_field_clinic
        FOREIGN KEY (exam_type_field_id, clinic_id)
        REFERENCES exam_type_fields (id, clinic_id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_lab_import_job_items_clinic_job
    ON lab_import_job_items (clinic_id, job_id, sort_order);

COMMENT ON TABLE lab_import_job_items IS '1受信フレームの項目。生バイトは持たない';
COMMENT ON COLUMN lab_import_job_items.device_item_code IS '電文コードそのもの。Na-P の -P を削らない';

CREATE TABLE lab_device_waits (
    id          bigserial       PRIMARY KEY,
    clinic_id   bigint          NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    pet_id      bigint          NOT NULL,
    staff_id    bigint          NOT NULL,
    expires_at  timestamptz     NOT NULL,
    cleared_at  timestamptz,
    created_at  timestamptz     NOT NULL DEFAULT now(),
    updated_at  timestamptz     NOT NULL DEFAULT now(),
    CONSTRAINT fk_lab_device_waits_pet_clinic
        FOREIGN KEY (clinic_id, pet_id)
        REFERENCES pets (clinic_id, id)
        ON DELETE RESTRICT
);

CREATE UNIQUE INDEX uq_lab_device_waits_clinic_active
    ON lab_device_waits (clinic_id)
    WHERE cleared_at IS NULL;

CREATE INDEX idx_lab_device_waits_clinic_expires
    ON lab_device_waits (clinic_id, expires_at);

COMMENT ON TABLE lab_device_waits IS '医院あたり有効待機は1件。期限切れは未紐付けへ落とす';

CREATE TABLE lab_device_station_settings (
    clinic_id           bigint          PRIMARY KEY REFERENCES clinics(id) ON DELETE RESTRICT,
    wait_ttl_seconds    integer         NOT NULL DEFAULT 1800,
    slots_json          jsonb           NOT NULL DEFAULT '[]'::jsonb,
    created_at          timestamptz     NOT NULL DEFAULT now(),
    updated_at          timestamptz     NOT NULL DEFAULT now(),
    CONSTRAINT chk_lab_device_station_settings_ttl
        CHECK (wait_ttl_seconds >= 60 AND wait_ttl_seconds <= 86400)
);

COMMENT ON TABLE lab_device_station_settings IS '待機TTLと論理スロット。clinic_settings には載せない';
COMMENT ON COLUMN lab_device_station_settings.wait_ttl_seconds IS '製品KPIにしない。数値チューニングUIは作らない';

SELECT app_private.apply_rls_policy(
    'lab_device_item_masters'::regclass,
    'tenant_clinic_id_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);
SELECT app_private.apply_rls_policy(
    'lab_import_job_items'::regclass,
    'tenant_clinic_id_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);
SELECT app_private.apply_rls_policy(
    'lab_device_waits'::regclass,
    'tenant_clinic_id_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);
SELECT app_private.apply_rls_policy(
    'lab_device_station_settings'::regclass,
    'tenant_clinic_id_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);
