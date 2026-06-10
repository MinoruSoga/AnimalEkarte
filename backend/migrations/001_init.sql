-- =============================================================================
-- Animal Ekarte - 統合スキーマ定義 v22.1 (consolidated)
-- PostgreSQL 18
-- テーブル数: 96 (旧 001–021 + mig-005〜mig-013 + 取扱説明書テーブル を統合)
-- 統合内容:
--   002: マスタシードデータ
--   003: デモシードデータ
--   004: ステージングシードデータ
--   005: clinic_integrations テーブル
--   006: shared_files テーブル
--   007: owners.line_user_id カラム + インデックス
--   008: lstep_tag_cache テーブル
--   009: pets.deceased_at/deceased_reason, owners.lstep_opt_out* カラム
--   010: medical_records.next_visit_recommended_date カラム
--   011: prescriptions テーブル
--   012: pet_chronic_conditions テーブル
--   013: line_send_logs テーブル
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
CREATE TYPE payment_method AS ENUM ('cash', 'credit_card', 'electronic_money');
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
    owner_id    bigint       NOT NULL REFERENCES owners(id)  ON DELETE CASCADE,
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
    clinic_id   bigint       NOT NULL REFERENCES clinics(id),
    owner_id    bigint       NOT NULL REFERENCES owners(id),
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
-- 7c. lstep_migration_progress（Lステップ一括同期進捗: 017 統合）
-- ------------------------------------
CREATE TABLE lstep_migration_progress (
    id            BIGSERIAL    PRIMARY KEY,
    clinic_id     bigint       NOT NULL REFERENCES clinics(id),
    owner_id      bigint       NOT NULL REFERENCES owners(id),
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
-- 7d. lstep_settings（Lステップ同期設定: ext-007 統合）
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
-- 7e. lstep_sync_error_counters（Lステップ同期エラーカウンター: ext-008 統合）
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
-- 7f. lstep_tag_code_mappings（Lステップタグコードマッピング: ext-010 統合）
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
-- 7f-2. lstep_trigger_priorities（Q23 配信衝突優先順位: mig-008 統合）
-- ------------------------------------
CREATE TABLE lstep_trigger_priorities (
    id           BIGSERIAL PRIMARY KEY,
    clinic_id    BIGINT NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
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
-- 7g. lstep_delivery_trigger_log（Lステップ配信トリガーログ: ext-012 統合）
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
-- 7h. lstep_csv_imports（Lステップ CSV インポート: ext-017 統合）
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
-- 7i. lstep_friend_attribute_snapshots（Lステップ友だち属性スナップショット: ext-018 統合）
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
-- 7j. lstep_auto_managed_prefixes（自動管理タグプレフィックス: mig-010 統合）
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
-- 7k. lstep_condition_tag_mappings（慢性疾患コード→タグ名マッピング: mig-010 統合）
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
-- 7l. lstep_send_purpose_tag_prefixes（LINE送信目的→タグプレフィックスマッピング: mig-010 統合）
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
    reservation_day_option   text                        NOT NULL DEFAULT 'none',
    is_internal              boolean                     NOT NULL DEFAULT false,
    category                 reservation_type_category   NOT NULL DEFAULT 'general',
    created_at               timestamptz                 NOT NULL DEFAULT now(),
    updated_at               timestamptz                 NOT NULL DEFAULT now(),
    deleted_at               timestamptz
);

CREATE INDEX idx_reservation_types_group_id ON reservation_types(group_id);

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
    created_at  timestamptz     NOT NULL DEFAULT now(),
    updated_at  timestamptz     NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

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
    created_at        timestamptz     NOT NULL DEFAULT now(),
    updated_at        timestamptz     NOT NULL DEFAULT now(),
    deleted_at        timestamptz
);

-- 009: ペット死亡記録インデックス
CREATE INDEX idx_pets_deceased ON pets (clinic_id, deceased_at)
    WHERE deceased_at IS NOT NULL;

COMMENT ON COLUMN pets.deceased_at     IS 'ペット死亡日。NULL = 生存中。';
COMMENT ON COLUMN pets.deceased_reason IS 'ペット死亡理由（任意記録）。';

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
    clinic_id         bigint       NOT NULL REFERENCES clinics(id),
    owner_id          bigint       NOT NULL REFERENCES owners(id),
    pet_id            bigint                REFERENCES pets(id),
    medical_record_id bigint                REFERENCES medical_records(id),
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
    display_order  integer      NOT NULL DEFAULT 0,
    is_active      boolean      NOT NULL DEFAULT true,
    created_at     timestamptz  NOT NULL DEFAULT now(),
    updated_at     timestamptz  NOT NULL DEFAULT now(),
    deleted_at     timestamptz
);

CREATE UNIQUE INDEX idx_payment_methods_clinic_name ON payment_methods(clinic_id, name) WHERE deleted_at IS NULL;
CREATE INDEX idx_payment_methods_clinic_order ON payment_methods(clinic_id, display_order) WHERE deleted_at IS NULL;

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
    created_at        timestamptz  NOT NULL DEFAULT now()
);

CREATE INDEX idx_payment_splits_clinic_billing ON payment_splits(clinic_id, billing_id);

-- ------------------------------------
-- 59. billing_refunds（返金レコード）
-- ------------------------------------
CREATE TABLE billing_refunds (
    id           BIGSERIAL   PRIMARY KEY,
    clinic_id    bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    billing_id   bigint      NOT NULL REFERENCES billings(id) ON DELETE CASCADE,
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
    period                  varchar(2)   NOT NULL CHECK (period IN ('am', 'pm')),
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

-- シフト: 1スタッフ1日1シフト
CREATE UNIQUE INDEX idx_shift_entries_staff_date ON shift_entries(staff_id, date);


-- 飼主: clinic内でemail重複不可（論理削除を除く・空文字除く）
CREATE UNIQUE INDEX uk_owners_clinic_email ON owners(clinic_id, email) WHERE deleted_at IS NULL AND email IS NOT NULL AND email != '';

-- billings: medical_record_idがある場合は1対1
CREATE UNIQUE INDEX idx_billings_medical_record_id_unique ON billings(medical_record_id) WHERE medical_record_id IS NOT NULL;

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
CREATE INDEX idx_vital_records_deleted_at ON vital_records (deleted_at) WHERE deleted_at IS NULL;
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

-- hospitalization 子テーブル FK インデックス
CREATE INDEX idx_hospitalizations_pet_id ON hospitalizations(pet_id);
CREATE INDEX idx_hospitalizations_owner_id ON hospitalizations(owner_id);
CREATE INDEX idx_hospitalizations_cage_id ON hospitalizations(cage_id);
CREATE INDEX idx_care_plan_items_hospitalization_id ON care_plan_items(hospitalization_id);

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

-- billing_confirmations インデックス
CREATE INDEX idx_billing_confirmations_status ON billing_confirmations(status);

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

-- ペット一覧（飼主別）
CREATE INDEX idx_pets_owner_id
  ON pets(owner_id)
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
COMMENT ON TABLE lstep_csv_imports                    IS 'Lステップ友だち属性CSVのインポート履歴（ext-017 統合）';
COMMENT ON TABLE lstep_friend_attribute_snapshots     IS 'Lステップ友だちの属性スナップショット（ext-018 統合）';
COMMENT ON TABLE lstep_migration_progress             IS '既存飼い主データ一括同期の進捗管理テーブル（017 統合）';

-- ------------------------------------
-- 62. audit_logs（権限変更・認証操作の監査ログ）
-- ------------------------------------
CREATE TABLE audit_logs (
    id           BIGSERIAL    PRIMARY KEY,
    clinic_id    bigint       NULL,
    actor_id     bigint       NULL,
    actor_type   varchar(30)  NOT NULL,
    action       varchar(50)  NOT NULL,
    resource     varchar(50)  NOT NULL,
    resource_id  bigint       NULL,
    old_value    jsonb        NULL,
    new_value    jsonb        NULL,
    ip_address   inet         NULL,
    user_agent   text         NULL,
    metadata     jsonb        NULL,                  -- ext-005: 追加コンテキスト情報
    created_at   timestamptz  NOT NULL DEFAULT now()
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
    INSERT INTO payment_methods (clinic_id, name, display_order, is_active)
    VALUES
        (NEW.id, '現金',            1, true),
        (NEW.id, 'クレジットカード', 2, true),
        (NEW.id, '電子マネー',       3, true),
        (NEW.id, '銀行振込',         4, true);
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
