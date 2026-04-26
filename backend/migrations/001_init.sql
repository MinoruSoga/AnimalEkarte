-- =============================================================================
-- Animal Ekarte - 統合スキーマ定義 v20.0 (consolidated)
-- PostgreSQL 18
-- テーブル数: 79 (旧 001–017 を統合)
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
CREATE TYPE item_source AS ENUM ('medical_record', 'manual', 'hospitalization');

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
COMMENT ON COLUMN owners.lstep_opt_out      IS 'Lステップ配信オプトアウトフラグ。true = すべてのタグ付与をスキップ。';
COMMENT ON COLUMN owners.lstep_opt_out_at   IS 'オプトアウト設定日時。';
COMMENT ON COLUMN owners.lstep_opt_out_reason IS 'オプトアウト理由（監査ログ用）。';
COMMENT ON COLUMN owners.line_followed_at   IS 'LINE フォロー日時（最終フォロー時刻）。Webhook follow イベントで更新。';
COMMENT ON COLUMN owners.line_blocked_at    IS 'LINE ブロック日時。Webhook unfollow イベントで更新。再フォロー時に NULL にリセット。';

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
    id          BIGSERIAL   PRIMARY KEY,
    clinic_id   bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name        text        NOT NULL,
    price       bigint,
    is_active   boolean     NOT NULL DEFAULT true,
    description text        NOT NULL DEFAULT '',
    parent_id   bigint               REFERENCES exam_types(id) ON DELETE SET NULL,
    sort_order  integer              DEFAULT 0,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
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
-- 20. trimming_courses（トリミングコースマスタ）
-- ------------------------------------
CREATE TABLE trimming_courses (
    id          BIGSERIAL   PRIMARY KEY,
    clinic_id   bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name        text        NOT NULL,
    price       bigint,
    is_active   boolean     NOT NULL DEFAULT true,
    description text        NOT NULL DEFAULT '',
    target_size target_size,
    duration    integer,
    sort_order  integer              DEFAULT 0,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

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
    pet_id            bigint      NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
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
-- 59. billing_refunds（返金レコード）
-- ------------------------------------
CREATE TABLE billing_refunds (
    id           BIGSERIAL   PRIMARY KEY,
    clinic_id    bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    billing_id   bigint      NOT NULL REFERENCES billings(id) ON DELETE CASCADE,
    amount       bigint      NOT NULL CHECK (amount > 0),
    reason       text        NOT NULL DEFAULT '',
    refunded_by  bigint      REFERENCES staffs(id),
    refunded_at  timestamptz NOT NULL DEFAULT now(),
    created_at   timestamptz NOT NULL DEFAULT now()
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
    created_at             timestamptz  NOT NULL DEFAULT now(),
    updated_at             timestamptz  NOT NULL DEFAULT now()
);

COMMENT ON TABLE clinic_settings IS '医院締め時間・休診曜日設定（FEAT-368）';

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
    CONSTRAINT uq_cash_register_closes_date_period UNIQUE (clinic_id, close_date, period)
);

CREATE INDEX idx_cash_register_closes_clinic ON cash_register_closes(clinic_id, close_date DESC);

COMMENT ON TABLE cash_register_closes IS 'レジ締めレコード（FEAT-368）';


CREATE INDEX idx_billing_items_merchandise_item_id ON billing_items(merchandise_item_id) WHERE deleted_at IS NULL;

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
COMMENT ON TABLE clinic_integrations       IS 'Lステップ/LINE連携設定保存テーブル（005 統合）';
COMMENT ON TABLE shared_files              IS 'LINE個別送信用ファイルストレージ（006 統合）';
COMMENT ON TABLE lstep_tag_cache           IS 'Lステップタグのカルテ側キャッシュ（008 統合）';
COMMENT ON TABLE prescriptions             IS '処方薬記録テーブル（011 統合）';
COMMENT ON TABLE pet_chronic_conditions    IS '慢性疾患フラグ管理テーブル（012 統合）';
COMMENT ON TABLE line_send_logs            IS 'LINE送信ログ（013 統合）';
COMMENT ON TABLE line_link_tokens          IS 'LINE User ID 紐付け用の一時トークン（016 統合）';
COMMENT ON TABLE lstep_migration_progress  IS '既存飼い主データ一括同期の進捗管理テーブル（017 統合）';

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

-- =============================================================================
-- SEED: マスタデータ (002_seed_master)
-- =============================================================================
-- =============================================================================
-- Animal Ekarte - マスタデータ v1.0
-- PostgreSQL 18
-- 冪等性保証: ON CONFLICT DO NOTHING / DO UPDATE
-- 内容: システム共通マスタデータ（clinic_id なし・全クリニック共通）
-- 依存: 001_init.sql
-- 実行順: 001_init.sql → 002_seed_master.sql → 003_seed_demo.sql
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. companies（本部情報: 1件）
-- -----------------------------------------------------------------------------
INSERT INTO companies (id, name) VALUES
    (1, 'ノア動物病院')
ON CONFLICT (id) DO UPDATE
    SET name = EXCLUDED.name,
        updated_at = now();

SELECT setval(pg_get_serial_sequence('companies', 'id'), (SELECT MAX(id) FROM companies));

-- -----------------------------------------------------------------------------
-- 2. animal_species（ペット種類: 6件、システム共通・clinic_idなし）
-- -----------------------------------------------------------------------------
INSERT INTO animal_species (id, name, is_active, sort_order) VALUES
    (1, '犬',         true, 1),
    (2, '猫',         true, 2),
    (3, '鳥',         true, 3),
    (4, 'うさぎ',     true, 4),
    (5, 'ハムスター', true, 5),
    (6, 'その他',     true, 6)
ON CONFLICT (id) DO UPDATE
    SET name = EXCLUDED.name,
        is_active = EXCLUDED.is_active,
        sort_order = EXCLUDED.sort_order,
        updated_at = now();

SELECT setval(pg_get_serial_sequence('animal_species', 'id'), (SELECT MAX(id) FROM animal_species));

-- =============================================================================
-- SEED: デモデータ (003_seed_demo)
-- =============================================================================
-- =============================================================================
-- Animal Ekarte - デモデータ v1.0
-- PostgreSQL 18
-- 再投入耐性: 主な固定IDデータは ON CONFLICT DO UPDATE、追加専用の関連データは DO NOTHING
-- 内容: クリニック設定・スタッフ・飼主・ペット・診療記録・会計等（デモ/テスト用）
-- 依存: 001_init.sql → 002_seed_master.sql
-- 実行順: 001_init.sql → 002_seed_master.sql → 003_seed_demo.sql
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. clinics（クリニック: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO clinics (id, company_id, name) VALUES
    (1, 1, '八王子病院'),
    (2, 1, '城東センター病院'),
    (3, 1, '敷島医院')
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('clinics', 'id'), (SELECT MAX(id) FROM clinics));

-- -----------------------------------------------------------------------------
-- 4. occupations（職種: 各医院4件 × 3医院 = 12件）
-- -----------------------------------------------------------------------------
INSERT INTO occupations (id, clinic_id, name, is_active, sort_order) VALUES
    -- 八王子病院 (clinic_id=1)
    (1, 1, '獣医師',   true, 1),
    (2, 1, '看護師',   true, 2),
    (3, 1, 'トリマー', true, 3),
    (4, 1, '受付',     true, 4),
    -- 城東センター病院 (clinic_id=2)
    (5, 2, '獣医師',   true, 1),
    (6, 2, '看護師',   true, 2),
    (7, 2, 'トリマー', true, 3),
    (8, 2, '受付',     true, 4),
    -- 敷島医院 (clinic_id=3)
    (9, 3, '獣医師',   true, 1),
    (10, 3, '看護師',   true, 2),
    (11, 3, 'トリマー', true, 3),
    (12, 3, '受付',     true, 4)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('occupations', 'id'), (SELECT MAX(id) FROM occupations));

-- -----------------------------------------------------------------------------
-- 5. staffs（スタッフ: 35件）
-- 八王子病院: ID 1-15（人間11名 + リソース4件）
-- 城東センター病院: ID 16-25（人間7名 + リソース3件）
-- 敷島医院: ID 26-32（人間5名 + リソース2件）
-- account_id は accounts INSERT 後に UPDATE で設定される
-- clinic_id は staff_clinic_assignments で管理される
-- -----------------------------------------------------------------------------
INSERT INTO staffs (id, account_id, name, is_active, license_number, occupation_id, sort_order, staff_type, reservation_visible) VALUES
    -- 八王子病院 人間スタッフ (clinic_id=1)
    (1,  NULL, '林 文明',              true, 'V-10001', 1, 1,  'doctor',   true),
    (2,  NULL, '山﨑 晶子',           true, 'V-10002', 1, 2,  'doctor',   true),
    (3,  NULL, '三井 隆之',           true, 'V-10003', 1, 3,  'doctor',   true),
    (4,  NULL, 'ノア',                 true, 'V-10004', 1, 4,  'doctor',   true),
    (5,  NULL, '加藤 茉里',           true, '',        2, 5,  'nurse',    false),
    (6,  NULL, '金谷 亜美',           true, '',        2, 6,  'nurse',    false),
    (7,  NULL, '稲村 一真',           true, '',        2, 7,  'nurse',    false),
    (8,  NULL, '安田 希恵',           true, '',        2, 8,  'nurse',    false),
    (9,  NULL, '倉田 春香',           true, '',        2, 9,  'nurse',    false),
    (10, NULL, '梶原 梨夢',           true, '',        2, 10, 'nurse',    false),
    (11, NULL, '髙木 賀央里',         true, '',        2, 11, 'nurse',    false),
    -- 八王子病院 リソース（予約枠管理用）
    (12, NULL, 'お手入れ・オゾン療法', true, '',        3, 12, 'resource', true),
    (13, NULL, '健診・ワクチン・狂犬病', true, '',     3, 13, 'resource', true),
    (14, NULL, 'ドッグラン(アジリティ解放)', true, '', 4, 14, 'resource', true),
    (15, NULL, 'クイックシャンプー',   true, '',        3, 15, 'resource', true)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 6. accounts（認証用アカウント: 16件）
-- password_hash: bcrypt("password", cost=10)
-- account→staff mapping:
--   1(admin@noavet.jp)→staff 4(ノア / 院長代理), 2(hayashi)→staff 1(林 文明)
--   3(yamazaki)→staff 2(山﨑 晶子), 4(mitsui)→staff 3(三井 隆之)
--   5(admin@example.com)→staff 8(安田), 6(vet)→staff 9(倉田)
--   7(nurse)→staff 10(梶原), 8(reception)→staff 11(髙木)
--   9(system)→NULL, 10(kimura)→staff 16(木村 健太)
--   11(sasaki)→staff 17(佐々木), 12(fujiwara)→staff 26(藤原)
--   13(matsumoto)→staff 27(松本)
--   14(trimmer@example.com)→staff 33(さくら/八王子病院デモ / トリマー)
--   15(joto-vet@example.com)→新staff(城東獣医デモ)
--   16(shiki-vet@example.com)→新staff(敷島獣医デモ)
-- -----------------------------------------------------------------------------
INSERT INTO accounts (id, email, password_hash, is_active, is_system_admin) VALUES
    -- システム管理者（全院アクセス）
    (1, 'admin@noavet.jp',           '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, true),
    -- 八王子病院スタッフ（実名メール）
    (2, 'hayashi@noah-vet.co.jp',    '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, true),  -- システム管理者（林 文明）
    (3, 'yamazaki@noah-vet.co.jp',   '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false),
    (4, 'mitsui@noah-vet.co.jp',     '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false),
    -- デモアカウント（八王子病院・frontend mock-data.ts 対応）
    (5, 'admin@example.com',         '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false),
    (6, 'vet@example.com',           '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false),
    (7, 'nurse@example.com',         '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false),
    (8, 'reception@example.com',     '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false),
    (9, 'system@example.com',        '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false),
    -- 城東センター病院スタッフ
    (10, 'kimura@noah-vet.co.jp',    '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false),
    (11, 'sasaki@noah-vet.co.jp',    '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false),
    -- 敷島医院スタッフ
    (12, 'fujiwara@noah-vet.co.jp',  '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false),
    (13, 'matsumoto@noah-vet.co.jp', '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false),
    -- デモアカウント（frontend LoginForm の DEMO_ACCOUNTS に対応）
    (14, 'trimmer@example.com',      '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false),
    (15, 'joto-vet@example.com',     '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false),
    (16, 'shiki-vet@example.com',    '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('accounts', 'id'), (SELECT MAX(id) FROM accounts));

-- 城東センター病院 (clinic_id=2) スタッフ
INSERT INTO staffs (id, account_id, name, is_active, license_number, occupation_id, sort_order, staff_type, reservation_visible) VALUES
    (16, NULL, '木村 健太',       true, 'V-40001', 5, 1,  'doctor',   true),
    (17, NULL, '佐々木 美香',     true, 'V-40002', 5, 2,  'doctor',   true),
    (18, NULL, '高橋 翔太',       true, 'V-40003', 5, 3,  'doctor',   true),
    (19, NULL, '田村 由紀',       true, '',        6, 4,  'nurse',    false),
    (20, NULL, '中村 大輔',       true, '',        6, 5,  'nurse',    false),
    (21, NULL, '小林 麻衣',       true, '',        6, 6,  'nurse',    false),
    (22, NULL, '井上 拓也',       true, '',        6, 7,  'nurse',    false),
    -- 城東センター病院 リソース
    (23, NULL, '健診・ワクチン・狂犬病', true, '', 7, 8,  'resource', true),
    (24, NULL, 'トリミング',       true, '',        7, 9,  'resource', true),
    (25, NULL, 'クイックシャンプー', true, '',      7, 10, 'resource', true)
ON CONFLICT DO NOTHING;

-- 敷島医院 (clinic_id=3) スタッフ
INSERT INTO staffs (id, account_id, name, is_active, license_number, occupation_id, sort_order, staff_type, reservation_visible) VALUES
    (26, NULL, '藤原 誠一',       true, 'V-50001', 9, 1,  'doctor',   true),
    (27, NULL, '松本 さやか',     true, 'V-50002', 9, 2,  'doctor',   true),
    (28, NULL, '石田 和也',       true, 'V-50003', 9, 3,  'doctor',   true),
    (29, NULL, '岡本 菜々子',     true, '',        10, 4, 'nurse',    false),
    (30, NULL, '西村 健二',       true, '',        10, 5, 'nurse',    false),
    -- 敷島医院 リソース
    (31, NULL, '健診・ワクチン・狂犬病', true, '', 11, 6, 'resource', true),
    (32, NULL, 'トリミング',       true, '',        11, 7, 'resource', true)
ON CONFLICT DO NOTHING;

-- デモアカウント用スタッフ（frontend LoginForm の DEMO_ACCOUNTS に対応）
-- occupation_id: 3=トリマー(八王子), 5=獣医師(城東), 9=獣医師(敷島)
INSERT INTO staffs (id, account_id, name, is_active, occupation_id, staff_type) VALUES
    (33, 14, 'さくら（デモ）',    true, 3, 'trimmer'),
    (34, 15, '城東 獣医（デモ）', true, 5, 'doctor'),
    (35, 16, '敷島 獣医（デモ）', true, 9, 'doctor')
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('staffs', 'id'), (SELECT MAX(id) FROM staffs));

-- account_id → staff_id マッピング
UPDATE staffs SET account_id = 1  WHERE id = 4;  -- admin@noavet.jp → ノア (院長代理)
UPDATE staffs SET account_id = 2  WHERE id = 1;  -- hayashi@noah-vet.co.jp → 林 文明
UPDATE staffs SET account_id = 3  WHERE id = 2;  -- yamazaki@noah-vet.co.jp → 山﨑 晶子
UPDATE staffs SET account_id = 4  WHERE id = 3;  -- mitsui@noah-vet.co.jp → 三井 隆之
UPDATE staffs SET account_id = 5  WHERE id = 8;  -- admin@example.com → 安田 希恵 (デモ用)
UPDATE staffs SET account_id = 6  WHERE id = 9;  -- vet@example.com → 倉田 春香 (デモ用)
UPDATE staffs SET account_id = 7  WHERE id = 10; -- nurse@example.com → 梶原 梨夢 (デモ用)
UPDATE staffs SET account_id = 8  WHERE id = 11; -- reception@example.com → 髙木 賀央里 (デモ用)
UPDATE staffs SET account_id = 10 WHERE id = 16; -- kimura@noah-vet.co.jp → 木村 健太
UPDATE staffs SET account_id = 11 WHERE id = 17; -- sasaki@noah-vet.co.jp → 佐々木 美香
UPDATE staffs SET account_id = 12 WHERE id = 26; -- fujiwara@noah-vet.co.jp → 藤原 誠一
UPDATE staffs SET account_id = 13 WHERE id = 27; -- matsumoto@noah-vet.co.jp → 松本 さやか
-- デモアカウント（inline account_id: staffs INSERT で設定済みだが念のため）
UPDATE staffs SET account_id = 14 WHERE id = 33; -- trimmer@example.com → さくら（デモ / トリマー）
UPDATE staffs SET account_id = 15 WHERE id = 34; -- joto-vet@example.com → 城東 獣医（デモ）
UPDATE staffs SET account_id = 16 WHERE id = 35; -- shiki-vet@example.com → 敷島 獣医（デモ）

-- -----------------------------------------------------------------------------
-- 7. staff_clinic_assignments（スタッフ・クリニック割当: 37件）
-- Hachioji Clinic: staff 1-15, Noa (4) is also assigned to Joto and Shiki Clinics
-- Joto Clinic: staff 16-25
-- Shikishima Clinic: staff 26-32
-- -----------------------------------------------------------------------------
INSERT INTO staff_clinic_assignments (staff_id, clinic_id, is_main) VALUES
    -- Hachioji Clinic (clinic_id=1)
    (1, 1, true),   -- 林 文明 (hayashi@noah-vet.co.jp)
    (2, 1, true),   -- 山﨑 晶子 (yamazaki@noah-vet.co.jp)
    (3, 1, true),   -- 三井 隆之 (mitsui@noah-vet.co.jp)
    (4, 1, true),   -- ノア (admin@noavet.jp)
    (5, 1, true),   -- 加藤 茉里
    (6, 1, true),   -- 金谷 亜美
    (7, 1, true),   -- 稲村 一真
    (8, 1, true),   -- 安田 希恵 (admin@example.com デモ)
    (9, 1, true),   -- 倉田 春香 (vet@example.com デモ)
    (10, 1, true),   -- 梶原 梨夢 (nurse@example.com デモ)
    (11, 1, true),   -- 髙木 賀央里 (reception@example.com デモ)
    (12, 1, true),   -- お手入れ・オゾン療法 (resource)
    (13, 1, true),   -- 健診・ワクチン・狂犬病 (resource)
    (14, 1, true),   -- ドッグラン(アジリティ解放) (resource)
    (15, 1, true),   -- クイックシャンプー (resource)
    -- システム管理者 admin@noavet.jp は全院アクセス (staff_id=4 = ノア)
    (4, 2, false),  -- ノア - Joto Clinic (admin purpose)
    (4, 3, false),  -- ノア - Shikishima Clinic (admin purpose)
    -- Joto Clinic (clinic_id=2)
    (16, 2, true),   -- 木村 健太 (kimura@noah-vet.co.jp)
    (17, 2, true),   -- 佐々木 美香 (sasaki@noah-vet.co.jp)
    (18, 2, true),   -- 高橋 翔太
    (19, 2, true),   -- 田村 由紀
    (20, 2, true),   -- 中村 大輔
    (21, 2, true),   -- 小林 麻衣
    (22, 2, true),   -- 井上 拓也
    (23, 2, true),   -- 健診・ワクチン・狂犬病 (resource)
    (24, 2, true),   -- トリミング (resource)
    (25, 2, true),   -- クイックシャンプー (resource)
    -- 敷島医院 (clinic_id=3)
    (26, 3, true),   -- 藤原 誠一 (fujiwara@noah-vet.co.jp)
    (27, 3, true),   -- 松本 さやか (matsumoto@noah-vet.co.jp)
    (28, 3, true),   -- 石田 和也
    (29, 3, true),   -- 岡本 菜々子
    (30, 3, true),   -- 西村 健二
    (31, 3, true),   -- 健診・ワクチン・狂犬病 (resource)
    (32, 3, true),   -- トリミング (resource)
    -- デモアカウント用割当
    (33, 1, true),   -- さくら（デモ）→ 八王子病院
    (34, 2, true),   -- 城東 獣医（デモ）→ 城東センター病院
    (35, 3, true)    -- 敷島 獣医（デモ）→ 敷島医院
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 7b. permission_groups（権限グループ: 各医院2グループ × 3医院 = 6グループ）
-- 注記: CreateClinic で自動作成される「執行」「一般」グループと統一
-- -----------------------------------------------------------------------------
INSERT INTO permission_groups (id, clinic_id, name, description, color, is_active, sort_order) VALUES
    -- Hachioji Clinic (clinic_id=1)
    (1, 1, '執行', '執行権限', '#6366F1', true, 1),
    (2, 1, '一般', '一般スタッフ権限', '#10B981', true, 2),
    -- Joto Hospital (clinic_id=2)
    (3, 2, '執行', '執行権限', '#6366F1', true, 1),
    (4, 2, '一般', '一般スタッフ権限', '#10B981', true, 2),
    -- Shikishima Hospital (clinic_id=3)
    (5, 3, '執行', '執行権限', '#6366F1', true, 1),
    (6, 3, '一般', '一般スタッフ権限', '#10B981', true, 2)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('permission_groups', 'id'), (SELECT MAX(id) FROM permission_groups));

-- -----------------------------------------------------------------------------
-- 7c. permission_group_rules（権限グループルール: 23リソース × 2グループ = 46件）
-- V=View, C=Create, E=Edit, D=Delete
-- -----------------------------------------------------------------------------
INSERT INTO permission_group_rules (group_id, resource, can_view, can_create, can_edit, can_delete) VALUES
    -- 執行（group_id=1）: 全リソース全権限
    (1, 'reception',              true, true,  true,  true),
    (1, 'owners',                 true, true,  true,  true),
    (1, 'reservations',           true, true,  true,  true),
    (1, 'medical-records',        true, true,  true,  true),
    (1, 'hospitalization',        true, true,  true,  true),
    (1, 'trimming',               true, true,  true,  true),
    (1, 'examinations',           true, true,  true,  true),
    (1, 'accounting',             true, true,  true,  true),
    (1, 'vaccinations',           true, true,  true,  true),
    (1, 'checkups',               true, true,  true,  true),
    (1, 'inventory',              true, true,  true,  true),
    (1, 'estimates',              true, true,  true,  true),
    (1, 'shifts',                 true, true,  true,  true),
    -- hospital-settings (医院マスタ) の新規作成・削除は is_system_admin のみ。
    -- 執行は view+edit のみ（BUG-377/378 対応）。
    (1, 'hospital-settings',      true, false, true,  false),
    (1, 'master-animal-species',  true, true,  true,  true),
    (1, 'master-medical',         true, true,  true,  true),
    (1, 'master-reservation-type',    true, true,  true,  true),
    (1, 'master-hospitalization', true, true,  true,  true),
    (1, 'master-trimming',        true, true,  true,  true),
    (1, 'master-permission',      true, true,  true,  true),
    (1, 'master-staff',           true, true,  true,  true),
    (1, 'master-insurance',       true, true,  true,  true),
    (1, 'master-merchandise',     true, true,  true,  true),
    (1, 'discount',               true, true,  true,  true),
    -- 一般（group_id=2）: 基本業務（マスタは閲覧のみ）
    (2, 'reception',              true, false, false, false),
    (2, 'owners',                 true, true,  true,  false),
    (2, 'reservations',           true, true,  true,  false),
    (2, 'medical-records',        true, true,  true,  false),
    (2, 'hospitalization',        true, true,  true,  false),
    (2, 'trimming',               true, true,  true,  false),
    (2, 'examinations',           true, true,  true,  false),
    (2, 'accounting',             true, false, false, false),
    (2, 'vaccinations',           true, true,  true,  false),
    (2, 'checkups',               true, false, false, false),
    (2, 'inventory',              true, false, false, false),
    (2, 'estimates',              true, false, false, false),
    (2, 'shifts',                 true, true,  true,  false),
    (2, 'hospital-settings',      true, false, false, false),
    (2, 'master-animal-species',  true, false, false, false),
    (2, 'master-medical',         true, false, false, false),
    (2, 'master-reservation-type',    true, false, false, false),
    (2, 'master-hospitalization', true, false, false, false),
    (2, 'master-trimming',        true, false, false, false),
    (2, 'master-permission',      false, false, false, false),
    (2, 'master-staff',           true, false, false, false),
    (2, 'master-insurance',       true, false, false, false),
    (2, 'master-merchandise',     true, false, false, false),
    (2, 'discount',               false, false, false, false),
    -- 城東センター病院 執行（group_id=3）: 全リソース全権限
    (3, 'reception',              true, true,  true,  true),
    (3, 'owners',                 true, true,  true,  true),
    (3, 'reservations',           true, true,  true,  true),
    (3, 'medical-records',        true, true,  true,  true),
    (3, 'hospitalization',        true, true,  true,  true),
    (3, 'trimming',               true, true,  true,  true),
    (3, 'examinations',           true, true,  true,  true),
    (3, 'accounting',             true, true,  true,  true),
    (3, 'vaccinations',           true, true,  true,  true),
    (3, 'checkups',               true, true,  true,  true),
    (3, 'inventory',              true, true,  true,  true),
    (3, 'estimates',              true, true,  true,  true),
    (3, 'shifts',                 true, true,  true,  true),
    (3, 'hospital-settings',      true, false, true,  false),
    (3, 'master-animal-species',  true, true,  true,  true),
    (3, 'master-medical',         true, true,  true,  true),
    (3, 'master-reservation-type',    true, true,  true,  true),
    (3, 'master-hospitalization', true, true,  true,  true),
    (3, 'master-trimming',        true, true,  true,  true),
    (3, 'master-permission',      true, true,  true,  true),
    (3, 'master-staff',           true, true,  true,  true),
    (3, 'master-insurance',       true, true,  true,  true),
    (3, 'master-merchandise',     true, true,  true,  true),
    (3, 'discount',               true, true,  true,  true),
    -- 城東センター病院 一般（group_id=4）
    (4, 'reception',              true, false, false, false),
    (4, 'owners',                 true, true,  true,  false),
    (4, 'reservations',           true, true,  true,  false),
    (4, 'medical-records',        true, true,  true,  false),
    (4, 'hospitalization',        true, true,  true,  false),
    (4, 'trimming',               true, true,  true,  false),
    (4, 'examinations',           true, true,  true,  false),
    (4, 'accounting',             true, false, false, false),
    (4, 'vaccinations',           true, true,  true,  false),
    (4, 'checkups',               true, false, false, false),
    (4, 'inventory',              true, false, false, false),
    (4, 'estimates',              true, false, false, false),
    (4, 'shifts',                 true, true,  true,  false),
    (4, 'hospital-settings',      true, false, false, false),
    (4, 'master-animal-species',  true, false, false, false),
    (4, 'master-medical',         true, false, false, false),
    (4, 'master-reservation-type',    true, false, false, false),
    (4, 'master-hospitalization', true, false, false, false),
    (4, 'master-trimming',        true, false, false, false),
    (4, 'master-permission',      false, false, false, false),
    (4, 'master-staff',           true, false, false, false),
    (4, 'master-insurance',       true, false, false, false),
    (4, 'master-merchandise',     true, false, false, false),
    (4, 'discount',               false, false, false, false),
    -- 敷島医院 執行（group_id=5）: 全リソース全権限
    (5, 'reception',              true, true,  true,  true),
    (5, 'owners',                 true, true,  true,  true),
    (5, 'reservations',           true, true,  true,  true),
    (5, 'medical-records',        true, true,  true,  true),
    (5, 'hospitalization',        true, true,  true,  true),
    (5, 'trimming',               true, true,  true,  true),
    (5, 'examinations',           true, true,  true,  true),
    (5, 'accounting',             true, true,  true,  true),
    (5, 'vaccinations',           true, true,  true,  true),
    (5, 'checkups',               true, true,  true,  true),
    (5, 'inventory',              true, true,  true,  true),
    (5, 'estimates',              true, true,  true,  true),
    (5, 'shifts',                 true, true,  true,  true),
    (5, 'hospital-settings',      true, false, true,  false),
    (5, 'master-animal-species',  true, true,  true,  true),
    (5, 'master-medical',         true, true,  true,  true),
    (5, 'master-reservation-type',    true, true,  true,  true),
    (5, 'master-hospitalization', true, true,  true,  true),
    (5, 'master-trimming',        true, true,  true,  true),
    (5, 'master-permission',      true, true,  true,  true),
    (5, 'master-staff',           true, true,  true,  true),
    (5, 'master-insurance',       true, true,  true,  true),
    (5, 'master-merchandise',     true, true,  true,  true),
    (5, 'discount',               true, true,  true,  true),
    -- 敷島医院 一般（group_id=6）
    (6, 'reception',              true, false, false, false),
    (6, 'owners',                 true, true,  true,  false),
    (6, 'reservations',           true, true,  true,  false),
    (6, 'medical-records',        true, true,  true,  false),
    (6, 'hospitalization',        true, true,  true,  false),
    (6, 'trimming',               true, true,  true,  false),
    (6, 'examinations',           true, true,  true,  false),
    (6, 'accounting',             true, false, false, false),
    (6, 'vaccinations',           true, true,  true,  false),
    (6, 'checkups',               true, false, false, false),
    (6, 'inventory',              true, false, false, false),
    (6, 'estimates',              true, false, false, false),
    (6, 'shifts',                 true, true,  true,  false),
    (6, 'hospital-settings',      true, false, false, false),
    (6, 'master-animal-species',  true, false, false, false),
    (6, 'master-medical',         true, false, false, false),
    (6, 'master-reservation-type',    true, false, false, false),
    (6, 'master-hospitalization', true, false, false, false),
    (6, 'master-trimming',        true, false, false, false),
    (6, 'master-permission',      false, false, false, false),
    (6, 'master-staff',           true, false, false, false),
    (6, 'master-insurance',       true, false, false, false),
    (6, 'master-merchandise',     true, false, false, false),
    (6, 'discount',               false, false, false, false)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 7d. staff_permission_groups（スタッフ→権限グループ割当: 35件）
-- 執行(1): 管理系スタッフ + デモ管理者アカウント
-- 一般(2): 一般業務スタッフ
-- -----------------------------------------------------------------------------
INSERT INTO staff_permission_groups (staff_id, group_id) VALUES
    -- 八王子病院 執行グループ (group_id=1)
    (1,  1),  -- 林 文明     (hayashi@noah-vet.co.jp / システム管理者)
    (3,  1),  -- 三井 隆之   (mitsui@noah-vet.co.jp)
    (4,  1),  -- ノア        (admin@noavet.jp / 執行権限保持)
    -- 八王子病院 執行グループ (group_id=1) 追加: デモ admin アカウント
    (8,  1),  -- 安田 希恵   (admin@example.com デモ) → 執行権限（削除操作を含む機能テスト用）
    -- 八王子病院 一般グループ (group_id=2)
    (2,  2),  -- 山﨑 晶子   (yamazaki@noah-vet.co.jp)
    (5,  2),  -- 加藤 茉里
    (6,  2),  -- 金谷 亜美
    (7,  2),  -- 稲村 一真
    (9,  2),  -- 倉田 春香   (vet@example.com デモ)
    (10, 2),  -- 梶原 梨夢   (nurse@example.com デモ)
    (11, 2),  -- 髙木 賀央里 (reception@example.com デモ)
    (12, 2),  -- お手入れ・オゾン療法 (resource)
    (13, 2),  -- 健診・ワクチン・狂犬病 (resource)
    (14, 2),  -- ドッグラン (resource)
    (15, 2),  -- クイックシャンプー (resource)
    -- 城東センター病院 執行グループ (group_id=3)
    (16, 3),  -- 木村 健太   (kimura@noah-vet.co.jp)
    (17, 3),  -- 佐々木 美香 (sasaki@noah-vet.co.jp)
    -- 城東センター病院 一般グループ (group_id=4)
    (18, 4),  -- 高橋 翔太
    (19, 4),  -- 田村 由紀
    (20, 4),  -- 中村 大輔
    (21, 4),  -- 小林 麻衣
    (22, 4),  -- 井上 拓也
    (23, 4),  -- 健診・ワクチン・狂犬病 (resource)
    (24, 4),  -- トリミング (resource)
    (25, 4),  -- クイックシャンプー (resource)
    -- 敷島医院 執行グループ (group_id=5)
    (26, 5),  -- 藤原 誠一   (fujiwara@noah-vet.co.jp)
    (27, 5),  -- 松本 さやか (matsumoto@noah-vet.co.jp)
    -- 敷島医院 一般グループ (group_id=6)
    (28, 6),  -- 石田 和也
    (29, 6),  -- 岡本 菜々子
    (30, 6),  -- 西村 健二
    (31, 6),  -- 健診・ワクチン・狂犬病 (resource)
    (32, 6),  -- トリミング (resource)
    -- デモアカウント用権限グループ割当
    (33, 2),  -- さくら（デモ）→ 八王子病院 一般
    (34, 3),  -- 城東 獣医（デモ）→ 城東センター病院 執行
    (35, 5)   -- 敷島 獣医（デモ）→ 敷島医院 執行
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 8a. reservation_type_groups（予約区分グループ）
-- 八王子病院 (clinic_id=1): 7グループ (ID 1-7)
-- 城東センター病院 (clinic_id=2): 7グループ (ID 8-14)
-- 敷島医院 (clinic_id=3): 6グループ (ID 15-20)
-- -----------------------------------------------------------------------------

-- 八王子病院 (clinic_id=1)
INSERT INTO reservation_type_groups (id, clinic_id, name, color, sort_order, is_active) VALUES
    (1, 1, '診療系',           '#3B82F6', 1, true),
    (2, 1, '予防・ワクチン',   '#10B981', 2, true),
    (3, 1, '健康診断・検査',   '#8B5CF6', 3, true),
    (4, 1, 'トリミング・美容', '#F59E0B', 4, true),
    (5, 1, '手術・処置',       '#EF4444', 5, true),
    (6, 1, 'ホテル・入院',     '#06B6D4', 6, true),
    (7, 1, '管理・内部',       '#9CA3AF', 7, true)
ON CONFLICT (id) DO UPDATE
    SET name=EXCLUDED.name, color=EXCLUDED.color, sort_order=EXCLUDED.sort_order, is_active=EXCLUDED.is_active;

-- 城東センター病院 (clinic_id=2)
INSERT INTO reservation_type_groups (id, clinic_id, name, color, sort_order, is_active) VALUES
    (8, 2, '診療系',           '#3B82F6', 1, true),
    (9, 2, '予防・ワクチン',   '#10B981', 2, true),
    (10, 2, '健康診断・検査',   '#8B5CF6', 3, true),
    (11, 2, 'トリミング・美容', '#F59E0B', 4, true),
    (12, 2, '手術・処置',       '#EF4444', 5, true),
    (13, 2, 'ホテル・入院',     '#06B6D4', 6, true),
    (14, 2, '管理・内部',       '#9CA3AF', 7, true)
ON CONFLICT (id) DO UPDATE
    SET name=EXCLUDED.name, color=EXCLUDED.color, sort_order=EXCLUDED.sort_order, is_active=EXCLUDED.is_active;

-- 敷島医院 (clinic_id=3)
INSERT INTO reservation_type_groups (id, clinic_id, name, color, sort_order, is_active) VALUES
    (15, 3, '診療系',           '#3B82F6', 1, true),
    (16, 3, '予防・ワクチン',   '#10B981', 2, true),
    (17, 3, '健康診断・検査',   '#8B5CF6', 3, true),
    (18, 3, 'トリミング・美容', '#F59E0B', 4, true),
    (19, 3, '手術・処置',       '#EF4444', 5, true),
    (20, 3, '管理・内部',       '#9CA3AF', 6, true)
ON CONFLICT (id) DO UPDATE
    SET name=EXCLUDED.name, color=EXCLUDED.color, sort_order=EXCLUDED.sort_order, is_active=EXCLUDED.is_active;

SELECT setval(pg_get_serial_sequence('reservation_type_groups','id'), (SELECT MAX(id) FROM reservation_type_groups));

-- -----------------------------------------------------------------------------
-- 8b. reservation_types（予約区分）
-- 八王子病院 (clinic_id=1): 25件 (ID 1-25)
-- 城東センター病院 (clinic_id=2): 19件 (ID 26-44)
-- 敷島医院 (clinic_id=3): 14件 (ID 45-58)
-- -----------------------------------------------------------------------------

-- 八王子病院 (clinic_id=1) 公開コース (is_internal=false, reservation_visible=true)
INSERT INTO reservation_types (id, clinic_id, name, short_name, is_active, description, color, sort_order, duration_minutes, reservation_visible, reservation_comment, is_internal) VALUES
    (1, 1, '一般診察',               '診察',     true, '内科・外科・皮膚科などの一般的な診察',         '#3B82F6', 1,  15, true,  '', false),
    (2, 1, '一般診察(再診)',          '再診',     true, '継続通院の一般診察',                           '#3B82F6', 2,  15, true,  '', false),
    (3, 1, 'ワクチン接種',            'ワクチン', true, '各種ワクチン接種（予防接種）',                 '#10B981', 3,  15, true,  '', false),
    (4, 1, 'お手入れ',               'お手入れ', true, '爪切り・耳掃除・肛門腺絞りなど',               '#F59E0B', 4,  15, true,  '', false),
    (5, 1, '狂犬病',                 '狂犬病',   true, '狂犬病予防法に基づくワクチン接種',             '#10B981', 5,  15, true,  '', false),
    (6, 1, 'フィラリア予防',          'フィラリア', true, 'フィラリア予防薬投与・処方',                '#10B981', 6,  15, true,  '', false),
    (7, 1, '健康診断',               '健診',     true, '定期健康診断・フィラリア検査など',             '#8B5CF6', 7,  15, true,  '', false),
    (8, 1, '健康診断結果報告',        '結果報告', true, '健康診断結果の説明・報告',                     '#8B5CF6', 8,  15, true,  '', false),
    (9, 1, 'トリミングコース',        'トリミング', true, 'カット・シャンプー・ブロー・爪切り・耳掃除', '#F59E0B', 9,  15, true,  '', false),
    (10, 1, 'トリミング部分カットコース', '部分カット', true, '部分的なカット・トリミング',              '#F59E0B', 10, 15, true,  '', false),
    (11, 1, 'トリミングシャンプーコース', 'シャンプー', true, 'シャンプー・ブロー・ブラッシング',        '#F59E0B', 11, 15, true,  '', false),
    (12, 1, 'クイックシャンプー',      'Qシャンプー', true, '短時間シャンプー',                        '#F59E0B', 12, 15, true,  '', false),
    (13, 1, '室内ドッグラン',          'ドッグラン', true, '室内ドッグラン利用（60分）',                '#6B7280', 13, 60, true,  '', false)
ON CONFLICT DO NOTHING;

-- 八王子病院 (clinic_id=1) スタッフ専用コース (is_internal=true, reservation_visible=false)
INSERT INTO reservation_types (id, clinic_id, name, short_name, is_active, description, color, sort_order, duration_minutes, reservation_visible, reservation_comment, is_internal) VALUES
    (14, 1, '手術60',                 '手術60',   true, '手術枠（60分）',                               '#EF4444', 14, 60, false, '', true),
    (15, 1, 'ホテルお迎え',           'お迎え',   true, 'ホテルお迎え対応',                             '#6B7280', 15, 15, false, '', true),
    (16, 1, 'ホテル預かり',           '預かり',   true, 'ペットホテル預かり',                           '#6B7280', 16, 15, false, '', true),
    (17, 1, '面会',                   '面会',     true, '入院動物の面会対応',                           '#6B7280', 17, 15, false, '', true),
    (18, 1, '☎︎',                   '☎︎',      true, '電話対応枠',                                   '#6B7280', 18, 15, false, '', true),
    (19, 1, '手術30',                 '手術30',   true, '手術枠（30分）',                               '#EF4444', 19, 30, false, '', true),
    (20, 1, '休憩枠',                 '休憩',     true, '休憩・ブロック枠',                             '#6B7280', 20, 60, false, '', true),
    (21, 1, '×',                     '×',       true, '予約不可（15分）',                             '#6B7280', 21, 15, false, '', true),
    (22, 1, '電話予約60',             '電話予約', true, '電話予約枠（60分）',                           '#6B7280', 22, 60, false, '', true),
    (23, 1, '予約不可60',             '不可60',   true, '予約不可ブロック（60分）',                     '#6B7280', 23, 60, false, '', true),
    (24, 1, 'エコー枠',               'エコー',   true, '超音波検査専用枠',                             '#8B5CF6', 24, 30, false, '', true),
    (25, 1, '予約不可30',             '不可30',   true, '予約不可ブロック（30分）',                     '#6B7280', 25, 30, false, '', true)
ON CONFLICT DO NOTHING;


-- 八王子病院 (clinic_id=1) グループ紐付け
UPDATE reservation_types SET group_id=1 WHERE clinic_id=1 AND id IN (1,2);
UPDATE reservation_types SET group_id=2 WHERE clinic_id=1 AND id IN (3,5,6);
UPDATE reservation_types SET group_id=3 WHERE clinic_id=1 AND id IN (7,8,24);
UPDATE reservation_types SET group_id=4 WHERE clinic_id=1 AND id IN (4,9,10,11,12);
UPDATE reservation_types SET group_id=5 WHERE clinic_id=1 AND id IN (14,19);
UPDATE reservation_types SET group_id=6 WHERE clinic_id=1 AND id IN (13,15,16,17);
UPDATE reservation_types SET group_id=7, is_internal=true WHERE clinic_id=1 AND id IN (18,20,21,22,23,25);

-- -----------------------------------------------------------------------------
-- 9. cages（ケージ: 8件）
-- -----------------------------------------------------------------------------
INSERT INTO cages (id, clinic_id, name, price, is_active, description, cage_type, cage_size, sort_order) VALUES
    (1, 1, 'ICUケージA',     8000, true, '酸素吸入可・重症患者用',  'icu',     'medium', 1),
    (2, 1, 'ICUケージB',     8000, true, '酸素吸入可・重症患者用',  'icu',     'medium', 2),
    (3, 1, '犬用ケージ（小）', 3000, true, '小型犬・ホテル利用可',    'dog',     'small',  3),
    (4, 1, '犬用ケージ（中）', 3500, true, '中型犬・一般入院用',      'dog',     'medium', 4),
    (5, 1, '犬用ケージ（大）', 4000, true, '大型犬・術後管理用',      'dog',     'large',  5),
    (6, 1, '猫用ケージ（小）', 3000, true, '猫専用・ストレス軽減設計', 'cat',     'small',  6),
    (7, 1, '猫用ケージ（中）', 3000, true, '猫専用・ストレス軽減設計', 'cat',     'medium', 7),
    (8, 1, '汎用ケージA',     2500, true, '小動物・鳥類等対応',      'general', 'small',  8)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 10. insurances（保険: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO insurances (id, clinic_id, name, is_active, description, coverage_rate, contact_phone, sort_order) VALUES
    (1, 1, 'アニコム損保',         true, 'ペット保険大手・どうぶつ健保シリーズ', 70, '0120-025-034',  1),
    (2, 1, 'アイペット損保',       true, 'うちの子シリーズ',                     70, '0120-956-099',  2),
    (3, 1, 'ペット&ファミリー',     true, 'げんきナンバーワンシリーズ',           80, '0120-81-8505',  3),
    (4, 1, '楽天ペット保険',       true, '楽天が提供するペット保険',             70, '0120-600-810',  4),
    (5, 1, 'その他（自費）',       true, '保険未加入・全額自費',                100, '',              5)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 11. exam_types（検査種別: 5件）+ exam_type_fields（検査項目定義）
-- -----------------------------------------------------------------------------
INSERT INTO exam_types (id, clinic_id, name, price, is_active, description, sort_order) VALUES
    (1, 1, '血液検査（CBC）',     3000, true, '全血球計算（Complete Blood Count）',         1),
    (2, 1, '血液化学検査',         5000, true, '肝機能・腎機能・血糖値など生化学的検査',     2),
    (3, 1, '尿検査',               1500, true, '尿試験紙・尿沈渣検査',                       3),
    (4, 1, 'レントゲン検査',       3000, true, 'X線撮影（胸部・腹部・四肢）',                4),
    (5, 1, '超音波検査（エコー）', 5000, true, '腹部エコー・心臓エコー',                     5)
ON CONFLICT DO NOTHING;


-- exam_type_fields: 血液検査（CBC）
INSERT INTO exam_type_fields (id, exam_type_id, name, inspection_value, normal_value, sort_order) VALUES
    (1, 1, 'WBC（白血球数）',      '', '6.0-17.0 x10^3/uL', 1),
    (2, 1, 'RBC（赤血球数）',      '', '5.5-8.5 x10^6/uL',  2),
    (3, 1, 'HCT（ヘマトクリット）', '', '37-55%',            3),
    (4, 1, 'PLT（血小板数）',      '', '175-500 x10^3/uL',  4)
ON CONFLICT DO NOTHING;

-- exam_type_fields: 血液化学検査
INSERT INTO exam_type_fields (id, exam_type_id, name, inspection_value, normal_value, sort_order) VALUES
    (5, 2, 'ALT（GPT）',        '', '10-125 U/L',    1),
    (6, 2, 'BUN（尿素窒素）',   '', '7-27 mg/dL',    2),
    (7, 2, 'CRE（クレアチニン）', '', '0.5-1.8 mg/dL', 3),
    (8, 2, 'GLU（血糖値）',     '', '74-143 mg/dL',   4)
ON CONFLICT DO NOTHING;

-- exam_type_fields: 尿検査
INSERT INTO exam_type_fields (id, exam_type_id, name, inspection_value, normal_value, sort_order) VALUES
    (9, 1, '尿比重',   '', '1.015-1.045', 1),
    (10, 1, '尿pH',     '', '5.5-7.5',     2),
    (11, 1, '尿タンパク', '', '陰性',       3),
    (12, 1, '尿潜血',   '', '陰性',        4)
ON CONFLICT DO NOTHING;

-- exam_type_fields: レントゲン検査
INSERT INTO exam_type_fields (id, exam_type_id, name, inspection_value, normal_value, sort_order) VALUES
    (13, 2, '胸部正面', '', '異常なし', 1),
    (14, 2, '腹部正面', '', '異常なし', 2),
    (15, 2, '四肢',     '', '異常なし', 3)
ON CONFLICT DO NOTHING;

-- exam_type_fields: 超音波検査
INSERT INTO exam_type_fields (id, exam_type_id, name, inspection_value, normal_value, sort_order) VALUES
    (16, 3, '腹部エコー', '', '異常なし', 1),
    (17, 3, '心臓エコー', '', '異常なし', 2)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 12. vaccines（ワクチン: 10件）
-- -----------------------------------------------------------------------------
INSERT INTO vaccines (id, clinic_id, name, price, is_active, description, species, interval, sort_order) VALUES
    (1, 1, '混合ワクチン5種（犬）',      4500, true, 'ジステンパー・パルボ・アデノ1型・アデノ2型・パラインフルエンザ',         'dog', '1年',   1),
    (2, 1, '混合ワクチン6種（犬）',      5500, true, '5種＋コロナウイルス',                                                    'dog', '1年',   2),
    (3, 1, '混合ワクチン8種（犬）',      6500, true, '5種＋レプトスピラ3種',                                                   'dog', '1年',   3),
    (4, 1, '混合ワクチン10種（犬）',     8000, true, '5種＋レプトスピラ5種',                                                   'dog', '1年',   4),
    (5, 1, '混合ワクチン3種（猫）',      4000, true, '猫ウイルス性鼻気管炎・カリシウイルス・汎白血球減少症',                     'cat', '1年',   5),
    (6, 1, '混合ワクチン5種（猫）',      5500, true, '3種＋猫白血病・猫クラミジア',                                             'cat', '1年',   6),
    (7, 1, '狂犬病ワクチン',             3000, true, '狂犬病予防法に基づく接種',                                               'dog', '1年',   7),
    (8, 1, 'フィラリア予防薬（小型犬）',  900, true, '体重10kg以下犬用フィラリア予防',                                          'dog', '1ヶ月', 8),
    (9, 1, 'フィラリア予防薬（中型犬）', 1100, true, '体重11-25kg犬用フィラリア予防',                                           'dog', '1ヶ月', 9),
    (10, 1, 'フィラリア予防薬（大型犬）', 1500, true, '体重26kg以上犬用フィラリア予防',                                          'dog', '1ヶ月', 10)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 13. medicines（薬剤カテゴリ: 9件 + 薬剤: 15件）
-- カテゴリレコードは id 1001〜1009、price=NULL、parent_id=NULL
-- 薬剤レコードは parent_id でカテゴリを参照
-- -----------------------------------------------------------------------------

-- カテゴリレコード（id 1001〜1009）
INSERT INTO medicines (id, clinic_id, name, price, is_active, description, sort_order) VALUES
    (1001, 1, '抗生剤',       NULL, true, '抗生物質カテゴリ',   1),
    (1002, 1, 'ステロイド',   NULL, true, 'ステロイド剤カテゴリ', 2),
    (1003, 1, '利尿剤',       NULL, true, '利尿剤カテゴリ',     3),
    (1004, 1, '消炎剤',       NULL, true, '消炎鎮痛剤カテゴリ', 4),
    (1005, 1, '神経系薬',     NULL, true, '神経系薬カテゴリ',   5),
    (1006, 1, '制吐剤',       NULL, true, '制吐剤カテゴリ',     6),
    (1007, 1, '消化器用薬',   NULL, true, '消化器用薬カテゴリ', 7),
    (1008, 1, '駆虫剤',       NULL, true, '駆虫剤カテゴリ',     8),
    (1009, 1, '輸液',         NULL, true, '輸液カテゴリ',       9)
ON CONFLICT DO NOTHING;

-- 薬剤レコード（parent_id でカテゴリ参照）
INSERT INTO medicines (id, clinic_id, name, price, is_active, description, dosage_form, medicine_unit, default_quantity, sort_order, parent_id) VALUES
    (1, 1, 'アモキシシリン 50mg',         500,  true, '広域スペクトラム抗生物質',               'tablet',    'per_tablet', 1,   1, 1001),
    (2, 1, 'メトロニダゾール 250mg',       600,  true, '嫌気性菌・原虫感染症治療薬',             'tablet',    'per_tablet', 1,   2, 1001),
    (3, 1, 'プレドニゾロン 5mg',           400,  true, 'ステロイド系抗炎症・免疫抑制剤',         'tablet',    'per_tablet', 1,   3, 1002),
    (4, 1, 'フロセミド注射液 20mg/2ml',    800,  true, '利尿剤（心臓・腎臓病の浮腫治療）',       'injection', 'per_ml',     2,   4, 1003),
    (5, 1, 'メロキシカム経口液',           700,  true, 'NSAIDs・痛み・炎症の緩和',               'liquid',    'per_ml',     1,   5, 1004),
    (6, 1, 'ガバペンチン 100mg',           550,  true, '神経因性疼痛・てんかん補助療法',         'tablet',    'per_tablet', 1,   6, 1005),
    (7, 1, 'マロピタント 16mg',            800,  true, '制吐剤（乗り物酔い・嘔吐治療）',         'tablet',    'per_tablet', 1,   7, 1006),
    (8, 1, 'ラクツロース液',               500,  true, '便秘・肝性脳症の治療',                   'liquid',    'per_ml',     5,   8, 1007),
    (9, 1, 'ノミ・ダニ駆除薬（犬用）',     2500, true, '外部寄生虫予防・駆除（スポットオン）',   'topical',   'per_dose',   1,   9, 1008),
    (10, 1, 'ノミ・ダニ駆除薬（猫用）',     2500, true, '外部寄生虫予防・駆除（スポットオン）',   'topical',   'per_dose',   1,  10, 1008),
    (11, 1, '抗生剤点眼薬',                 600,  true, '眼科用抗菌点眼剤',                       'liquid',    'per_ml',     1,  11, 1001),
    (12, 1, 'デキサメタゾン注射液',         700,  true, '強力ステロイド・アレルギー緊急治療',     'injection', 'per_ml',     1,  12, 1002),
    (13, 1, '生理食塩水 500ml',             400,  true, '点滴・洗浄用生理食塩水',                 'liquid',    'per_ml',     500, 13, 1009),
    (14, 1, 'セファレキシン 250mg',         450,  true, '第1世代セフェム系抗生物質',              'tablet',    'per_tablet', 1,  14, 1001),
    (15, 1, 'オメプラゾール 10mg',          350,  true, 'プロトンポンプ阻害薬（胃酸抑制）',       'tablet',    'per_tablet', 1,  15, 1007)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 14. consultations（診察項目: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO consultations (id, clinic_id, name, price, is_active, description, time_condition, duration, sort_order) VALUES
    (1, 1, '初診料',       2000, true, '初めての受診または6ヶ月以上受診がない場合', 'first_visit',  30, 1),
    (2, 1, '再診料',        800, true, '継続通院の診察料',                         'revisit',      15, 2),
    (3, 1, '往診料',       5000, true, '自宅への往診料（基本料金）',               'anytime',      60, 3),
    (4, 1, '時間外診療料', 3000, true, '診療時間外・休日の緊急診察',               'after_hours',  30, 4),
    (5, 1, '電話相談料',    500, true, '電話による診察相談',                       'anytime',      15, 5)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 15. procedures（処置項目: 10件）
-- -----------------------------------------------------------------------------
INSERT INTO procedures (id, clinic_id, name, price, is_active, description, duration, anesthesia, sort_order) VALUES
    (1, 1, '去勢手術（犬）',   25000, true, '雄犬の去勢手術',                   60,  'general', 1),
    (2, 1, '避妊手術（猫）',   25000, true, '雌猫の避妊手術',                   90,  'general', 2),
    (3, 1, '歯石除去',         15000, true, '全身麻酔下での歯石除去・歯周治療', 45,  'general', 3),
    (4, 1, '耳洗浄',            2500, true, '外耳炎治療・耳道内の洗浄処置',     15,  'none',    4),
    (5, 1, '爪切り',             500, true, '爪のカット・やすりがけ',           10,  'none',    5),
    (6, 1, '皮膚縫合',          5000, true, '裂傷・切傷の縫合処置',             30,  'local',   6),
    (7, 1, '骨折整復',         80000, true, '骨折の外科的整復・固定',          120,  'general', 7),
    (8, 1, '腫瘍摘出',         20000, true, '皮膚腫瘍の外科的摘出',             60,  'local',   8),
    (9, 1, '胃洗浄',           10000, true, '異物誤飲時の胃洗浄処置',           30,  'general', 9),
    (10, 1, '点滴処置',          3000, true, '静脈内点滴（1時間）',              60,  'none',   10)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 16. hospitalization_plans（入院プラン: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO hospitalization_plans (id, clinic_id, name, price, is_active, description, body_size, billing_unit, sort_order) VALUES
    (1, 1, '一般入院（小型）', 3000, true, '体重10kg以下の入院管理料（1日）',  'small',  'per_day',   1),
    (2, 1, '一般入院（中型）', 3500, true, '体重10-25kgの入院管理料（1日）',   'medium', 'per_day',   2),
    (3, 1, '一般入院（大型）', 4500, true, '体重25kg以上の入院管理料（1日）',  'large',  'per_day',   3),
    (4, 1, 'ICU入院',          8000, true, '集中治療室管理料（1日）',          'small',  'per_day',   4),
    (5, 1, 'ホテル（小型）',   2500, true, '体重10kg以下のペットホテル（1泊）', 'small',  'per_night', 5)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 17. trimming_courses（トリミングコース: 5件）
-- ※ duration は integer (分)
-- -----------------------------------------------------------------------------
INSERT INTO trimming_courses (id, clinic_id, name, price, is_active, description, target_size, duration, sort_order) VALUES
    (1, 1, 'シャンプー&ブロー（小型）', 4000,  true, 'シャンプー・ブロー・ブラッシング',            'small',  60,  1),
    (2, 1, 'シャンプー&ブロー（中型）', 5500,  true, 'シャンプー・ブロー・ブラッシング',            'medium', 90,  2),
    (3, 1, 'フルコース（小型）',        7000,  true, 'カット・シャンプー・ブロー・爪切り・耳掃除', 'small',  120, 3),
    (4, 1, 'フルコース（中型）',        9000,  true, 'カット・シャンプー・ブロー・爪切り・耳掃除', 'medium', 150, 4),
    (5, 1, 'フルコース（大型）',        12000, true, 'カット・シャンプー・ブロー・爪切り・耳掃除', 'large',  180, 5)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 18. trimming_options（トリミングオプション: 5件）
-- ※ is_combinable は boolean, duration は integer（分単位）
-- -----------------------------------------------------------------------------
INSERT INTO trimming_options (id, clinic_id, name, price, is_active, description, duration, is_combinable, sort_order) VALUES
    (1, 1, '爪切り',     300, true, '爪のカット・やすりがけ',       10, true, 1),
    (2, 1, '耳掃除',     500, true, '外耳道の洗浄・清掃',           10, true, 2),
    (3, 1, '歯磨き',     500, true, '歯ブラシによるデンタルケア',   15, true, 3),
    (4, 1, '肛門腺絞り', 300, true, '肛門嚢の分泌液除去',            5, true, 4),
    (5, 1, 'リボン装着', 200, true, '仕上げのアクセサリー装着',      5, true, 5)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 19. diagnosis_types（診断カテゴリ: 8件）
-- -----------------------------------------------------------------------------
INSERT INTO diagnosis_types (id, clinic_id, name, is_active, description, sort_order) VALUES
    (1, 1, '消化器系',       true, '胃腸・肝臓・膵臓などの消化器系疾患',   1),
    (2, 1, '呼吸器系',       true, '肺・気管・鼻腔などの呼吸器系疾患',     2),
    (3, 1, '皮膚・被毛',     true, 'アレルギー・感染症などの皮膚疾患',     3),
    (4, 1, '泌尿器系',       true, '腎臓・膀胱・尿道などの泌尿器系疾患',   4),
    (5, 1, '神経系',         true, '脳・脊髄・末梢神経などの神経系疾患',   5),
    (6, 1, '感染症・寄生虫', true, '細菌・ウイルス・寄生虫感染症',         6),
    (7, 1, '腫瘍',           true, '良性・悪性腫瘍（がん）',               7),
    (8, 1, '外傷・骨格',     true, '骨折・咬傷・関節疾患など',             8)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 20. diagnosis_names（診断名: 各カテゴリ2-3件、計20件）
-- -----------------------------------------------------------------------------
INSERT INTO diagnosis_names (id, clinic_id, name, is_active, description, diagnosis_type_id, sort_order) VALUES
    -- 消化器系
    (1, 1, '胃腸炎',             true, '胃・腸の炎症（嘔吐・下痢）',         1, 1),
    (2, 1, '膵炎',               true, '膵臓の炎症',                         1, 2),
    (3, 1, '肝疾患',             true, '肝炎・肝不全・脂肪肝など',           1, 3),
    -- 呼吸器系
    (4, 1, '気管支炎',           true, '気管支の炎症',                       2, 1),
    (5, 1, '肺炎',               true, '肺の感染性・非感染性炎症',           2, 2),
    -- 皮膚・被毛
    (6, 1, 'アトピー性皮膚炎',   true, 'アレルゲンによるアレルギー性皮膚炎', 3, 1),
    (7, 1, '膿皮症',             true, '細菌性の皮膚感染症',                 3, 2),
    (8, 1, '真菌症',             true, '皮膚糸状菌による感染症',             3, 3),
    -- 泌尿器系
    (9, 1, '膀胱炎',             true, '細菌性・特発性膀胱炎',               4, 1),
    (10, 1, '腎不全',             true, '急性・慢性腎不全',                   4, 2),
    (11, 1, '尿路結石',           true, '腎結石・膀胱結石・尿道結石',         4, 3),
    -- 神経系
    (12, 1, 'てんかん',           true, '反復性の痙攣発作',                   5, 1),
    (13, 1, '椎間板ヘルニア',     true, '頸椎・腰椎の椎間板突出',             5, 2),
    -- 感染症・寄生虫
    (14, 1, 'パルボウイルス感染症', true, '犬パルボウイルスによる感染症',       6, 1),
    (15, 1, 'フィラリア症',       true, '犬糸状虫による心肺疾患',             6, 2),
    (16, 1, '猫風邪（FVR）',      true, '猫ウイルス性鼻気管炎',               6, 3),
    -- 腫瘍
    (17, 1, '肥満細胞腫',         true, '皮膚または内臓の肥満細胞腫瘍',       7, 1),
    (18, 1, 'リンパ腫',           true, '悪性リンパ腫',                       7, 2),
    -- 外傷・骨格
    (19, 1, '骨折',               true, '各部位の骨折',                       8, 1),
    (20, 1, '咬傷',               true, '他動物による咬傷・咬傷感染',         8, 2),
    -- 皮膚・被毛（追加）
    (41, 1, '外耳炎',             true, '外耳道の炎症・細菌/真菌/寄生虫感染', 3, 4),
    (42, 1, '膝蓋骨脱臼',         true, '膝蓋骨の内方/外方脱臼',             8, 3)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 21. checkup_types（健診種別: 4件）
-- -----------------------------------------------------------------------------
INSERT INTO checkup_types (id, clinic_id, name, price, is_active, description, interval, target_age, sort_order) VALUES
    (1, 1, '一般健診',       5000,  true, '身体検査・体重測定・問診',                     '1年',   '全年齢', 1),
    (2, 1, '老齢検診',       15000, true, '身体検査＋血液検査＋レントゲン＋超音波',         '6ヶ月', '7歳以上', 2),
    (3, 1, 'フィラリア検査', 2500,  true, 'フィラリア抗原検査（予防シーズン前）',           '1年',   '成犬',   3),
    (4, 1, '歯科検診',       3000,  true, '歯周病チェック・歯石付着度の確認',             '1年',   '成犬',   4)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 22. chief_complaint_types（主訴区分: 6件）
-- -----------------------------------------------------------------------------
INSERT INTO chief_complaint_types (id, clinic_id, name, is_active, sort_order) VALUES
    (1, 1, '食欲不振',       true, 1),
    (2, 1, '嘔吐・下痢',     true, 2),
    (3, 1, '皮膚・被毛異常', true, 3),
    (4, 1, '呼吸困難',       true, 4),
    (5, 1, '排尿・排泄異常', true, 5),
    (6, 1, '外傷・骨折',     true, 6)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 23. inquiry_templates（問診定型文: 10件）
-- ※ category は text 型: chief_complaint / history / current_medications / notes
-- -----------------------------------------------------------------------------
INSERT INTO inquiry_templates (id, clinic_id, category, title, content, is_active, sort_order) VALUES
    (1, 1, 'chief_complaint',    '食欲不振（急性）',           'いつ頃から食欲が落ちましたか？完全に食べないのか、減っているだけか確認してください。', true, 1),
    (2, 1, 'chief_complaint',    '嘔吐（回数・内容物）',       '嘔吐の回数、内容物（食物・胆汁・血液など）、嘔吐のタイミング（食後すぐ・空腹時）を確認してください。', true, 2),
    (3, 1, 'chief_complaint',    '下痢（性状・頻度）',         '便の性状（軟便・水様便・血便・粘液便）、排便頻度、いつから続いているか確認してください。', true, 3),
    (4, 1, 'chief_complaint',    '皮膚の痒み・発赤',           '痒がる部位、発症時期、季節性の有無、ノミ・ダニ予防の状況を確認してください。', true, 4),
    (5, 1, 'chief_complaint',    '排尿異常',                   '排尿回数の変化、尿の色・量、排尿時の痛みの有無、血尿の有無を確認してください。', true, 5),
    (6, 1, 'history',            '既往歴確認（手術歴）',       '過去の手術歴（去勢・避妊含む）、入院歴、重大な疾患の既往を確認してください。', true, 6),
    (7, 1, 'history',            '予防接種歴確認',             '最終ワクチン接種日、狂犬病予防接種の有無、フィラリア予防の状況を確認してください。', true, 7),
    (8, 1, 'current_medications', '現在の投薬状況',            '現在服用中の薬剤名、用量、投与期間、処方元の病院を確認してください。', true, 8),
    (9, 1, 'current_medications', 'サプリメント・フード',      '現在与えているサプリメント、療法食、おやつの種類を確認してください。', true, 9),
    (10, 1, 'notes',              '生活環境確認',               '室内飼い/外飼い、同居動物の有無、散歩の頻度・時間を確認してください。', true, 10)
ON CONFLICT DO NOTHING;

-- 城東センター病院 (clinic_id=2)
INSERT INTO inquiry_templates (id, clinic_id, category, title, content, is_active, sort_order) VALUES
    (11, 2, 'chief_complaint',    '食欲不振',                   'いつ頃から食欲が低下しましたか？完全絶食か減少かを確認してください。', true, 1),
    (12, 2, 'chief_complaint',    '嘔吐',                       '嘔吐の回数・内容物・タイミングを確認してください。', true, 2),
    (13, 2, 'chief_complaint',    '下痢',                       '便の性状・排便頻度・持続期間を確認してください。', true, 3),
    (14, 2, 'history',            '既往歴・手術歴',             '過去の手術歴・入院歴・重大な疾患の既往を確認してください。', true, 4),
    (15, 2, 'history',            '予防接種歴',                 '最終ワクチン接種日・狂犬病予防・フィラリア予防の状況を確認してください。', true, 5),
    (16, 2, 'current_medications', '投薬状況',                  '現在服用中の薬剤名・用量・処方元を確認してください。', true, 6),
    (17, 2, 'notes',              '生活環境',                   '室内飼い/外飼い・同居動物・散歩頻度を確認してください。', true, 7),
-- 敷島医院 (clinic_id=3)
    (18, 3, 'chief_complaint',    '食欲低下',                   'いつから食べなくなったか、食事内容の変化はあるか確認してください。', true, 1),
    (19, 3, 'chief_complaint',    '嘔吐・吐出',                 '嘔吐の頻度・内容物・食事との関連を確認してください。', true, 2),
    (20, 3, 'chief_complaint',    '消化器症状（下痢）',         '便の状態・頻度・血便の有無を確認してください。', true, 3),
    (21, 3, 'history',            '既往歴確認',                 '手術歴・入院歴・アレルギー歴を確認してください。', true, 4),
    (22, 3, 'history',            'ワクチン・予防歴',           'ワクチン接種歴・フィラリア予防の有無を確認してください。', true, 5),
    (23, 3, 'current_medications', '現在の治療状況',            '他院での治療中の薬・サプリメントを確認してください。', true, 6),
    (24, 3, 'notes',              '飼育環境',                   '飼育場所・同居動物・食事内容を確認してください。', true, 7)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('inquiry_templates', 'id'), (SELECT MAX(id) FROM inquiry_templates));

-- -----------------------------------------------------------------------------
-- 24. inventory_items（在庫管理: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO inventory_items (id, clinic_id, name, category, quantity, unit, min_stock_level, location, supplier, status) VALUES
    (1, 1, 'フロントライン プラス（犬用）',     'medicine',   50, '本',    10, '薬品棚A',   'メリアルジャパン',         'sufficient'),
    (2, 1, 'ネクスガード チュアブル',            'medicine',   30, '錠',    10, '薬品棚A',   'ベーリンガーインゲルハイム', 'sufficient'),
    (3, 1, 'ヒルズ i/d（犬用・消化器ケア）',     'food',       20, '袋',     5, '食品棚B',   'ヒルズ・コルゲート',       'sufficient'),
    (4, 1, 'ロイヤルカナン 消化器サポート',      'food',       15, '袋',     5, '食品棚B',   'ロイヤルカナン',           'sufficient'),
    (5, 1, '包帯・ガーゼセット',                 'consumable', 100, 'セット', 20, '消耗品棚C', '白十字',               'sufficient')
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 25. merchandise_items（物販・フード・その他: 7件）
-- -----------------------------------------------------------------------------
INSERT INTO merchandise_items (id, clinic_id, name, category, unit_price, tax_rate, sort_order) VALUES
    -- BUG-374 TC-367-04/05 検証用: 軽減税率 8% 品目（ペットフード）を含める
    (1, 1, 'ロイヤルカナン 消化器サポート 1kg', 'food', 2800, 0.08, 1),
    (2, 1, 'ヒルズ k/d 2kg', 'food', 3500, 0.08, 2),
    (3, 1, 'ペット用歯ブラシセット', 'goods', 1200, 0.10, 3),
    (4, 1, 'エリザベスカラー（S）', 'goods', 800, 0.10, 4),
    (5, 1, 'ノミ・ダニ予防首輪', 'goods', 1500, 0.10, 5),
    (6, 1, '文書料', 'other', 3000, 0.10, 6),
    (7, 1, '時間外診療費', 'other', 5000, 0.10, 7)
ON CONFLICT DO NOTHING;


-- =============================================================================
-- マスタ設定完了
-- =============================================================================


-- =============================================================================
-- デモデータ投入（飼主・ペット一覧ページ対応）
-- 内容: 飼主・ペット・取引記録（カルテ・予約・会計・入院・在庫・監査ログ等）
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. owners（飼主: 22件）
-- -----------------------------------------------------------------------------
INSERT INTO owners (id, clinic_id, name, name_kana, birth_date, company, postal_code, address1, address2, phone, company_phone, email, remarks, is_dangerous, discount_rate, membership_type) VALUES
    (1, 1, '林 文明', 'はやし ふみあき', '1980-05-15', 'サンプル株式会社', '150-0001', '東京都渋谷区神宮前1-2-3', '', '090-1111-2222', '03-1234-5678', 'hayashi@example.com', '定期検診を希望', false, 10, 'member'),
    (2, 1, '田中 花子', 'たなか はなこ', '1985-03-20', '', '160-0022', '東京都新宿区新宿1-1-1', '', '080-3333-4444', '', 'tanaka@example.com', '', false, 0, 'non_member'),
    (3, 1, '鈴木 一郎', 'すずき いちろう', '1978-11-03', '', '170-0001', '東京都豊島区西巣鴨1-3-5', '', '070-5555-6666', '', 'suzuki@example.com', '', false, 0, 'member'),
    (4, 1, '田中 美咲', 'たなか みさき', '1990-07-22', '', '153-0044', '東京都目黒区大橋2-4-6', '', '090-9999-8888', '', 'misaki.tanaka@example.com', '', false, 0, 'non_member'),
    (5, 1, '佐藤 花子', 'さとう はなこ', '1975-02-14', '', '140-0001', '東京都品川区北品川3-5-7', '', '080-2222-3333', '', 'hanako.sato@example.com', '', false, 5, 'member'),
    (6, 1, '伊藤 次郎', 'いとう じろう', '1983-09-30', '', '166-0013', '東京都杉並区堀ノ内1-7-9', '', '090-1234-5678', '', 'jiro.ito@example.com', '', false, 0, 'non_member'),
    (7, 1, '小林 さくら', 'こばやし さくら', '1992-04-05', '', '176-0012', '東京都練馬区豊玉北4-2-8', '', '080-9876-5432', '', 'sakura.kobayashi@example.com', '', false, 0, 'member'),
    (8, 1, '中村 勇気', 'なかむら ゆうき', '1987-12-18', '', '174-0041', '東京都板橋区舟渡2-6-10', '', '090-1122-3344', '', 'yuuki.nakamura@example.com', '', false, 0, 'non_member'),
    (9, 1, '加藤 恵', 'かとう めぐみ', '1995-06-25', '', '134-0083', '東京都江戸川区中葛西5-3-2', '', '080-5566-7788', '', 'megumi.kato@example.com', '', false, 10, 'member'),
    (10, 1, '山田 太郎', 'やまだ たろう', '1970-01-10', '', '144-0051', '東京都大田区西蒲田6-8-4', '', '090-2233-4455', '', 'taro.yamada@example.com', '', false, 0, 'non_member'),
    (11, 1, '高橋 由美', 'たかはし ゆみ', '1988-08-15', '', '110-0005', '東京都台東区上野5-1-3', '', '080-6677-8899', '', 'yumi.takahashi@example.com', '', false, 0, 'member'),
    (12, 1, '松本 隆', 'まつもと たかし', '1965-03-28', '', '125-0061', '東京都葛飾区亀有3-9-7', '', '090-3344-5566', '', 'takashi.matsumoto@example.com', '', false, 0, 'non_member'),
    (13, 1, '吉田 誠', 'よしだ まこと', '1982-11-05', '', '123-0845', '東京都足立区西新井7-4-6', '', '080-7788-9900', '', 'makoto.yoshida@example.com', '', false, 0, 'non_member'),
    (14, 1, '井上 京子', 'いのうえ きょうこ', '1973-05-19', '', '189-0023', '東京都東村山市美住町1-5-2', '', '090-4455-6677', '', 'kyoko.inoue@example.com', '', false, 5, 'member'),
    (15, 1, '木村 拓也', 'きむら たくや', '1991-07-14', '', '179-0081', '東京都練馬区北町3-6-9', '', '080-8899-0011', '', 'takuya.kimura@example.com', '', false, 0, 'non_member'),
    (16, 1, '佐々木 亮', 'ささき りょう', '1986-02-23', '', '207-0013', '東京都東大和市清水2-4-8', '', '090-5566-7788', '', 'ryo.sasaki@example.com', '', false, 0, 'non_member'),
    (17, 1, '山本 健太', 'やまもと けんた', '1998-09-12', '', '206-0802', '東京都稲城市東長沼2-8-3', '', '090-1234-9876', '', 'kenta.yamamoto@example.com', '', false, 0, 'non_member'),
    (18, 1, '青木 麻衣', 'あおき まい', '1993-03-10', '', '150-0002', '東京都渋谷区渋谷2-1-1', '', '090-1111-1111', '', 'mai.aoki@example.com', '', false, 0, 'non_member'),
    (19, 1, '橋本 俊介', 'はしもと しゅんすけ', '1980-07-25', '', '130-0001', '東京都墨田区吾妻橋1-3-5', '', '080-2222-2222', '', 'shunsuke.h@example.com', '', false, 0, 'member'),
    (20, 1, '福田 裕子', 'ふくだ ゆうこ', '1977-11-14', '', '145-0062', '東京都大田区北千束2-5-8', '', '090-3333-3333', '', 'yuko.fukuda@example.com', '', false, 5, 'member'),
    (21, 1, '石川 大輔', 'いしかわ だいすけ', '1989-04-02', '', '167-0041', '東京都杉並区善福寺3-2-6', '', '080-4444-4444', '', 'daisuke.ishikawa@example.com', '', false, 0, 'non_member'),
    (22, 1, '村田 奈々', 'むらた なな', '1996-09-19', '', '182-0021', '東京都調布市調布ヶ丘1-4-7', '', '090-5555-5555', '', 'nana.murata@example.com', '', false, 0, 'non_member')
ON CONFLICT (id) DO UPDATE SET
    name      = EXCLUDED.name,
    name_kana = EXCLUDED.name_kana,
    updated_at      = now();


-- -----------------------------------------------------------------------------
-- 2. pets（ペット: 28件）
-- -----------------------------------------------------------------------------
INSERT INTO pets (id, clinic_id, owner_id, pet_number, name, name_kana, animal_species_id, gender, status, birth_date, breed, color, weight, insurance_id, last_visit) VALUES
    (1, 1, 1,  '1-1', 'Iris(イリス)', 'いりす', 1, 'male',   'alive', '2015-04-14', 'ゴールデンレトリーバー',     '茶色',           26.5,  1, '2015-08-28'),
    (2, 1, 1,  '1-2', 'Max(マックス)', 'まっくす', 1, 'male', 'alive', '2018-06-20', 'ラブラドール',               'ゴールデン',     15.2,  NULL, '2024-11-15'),
    (3, 1, 2,  '2-1', 'ミケ',         'みけ',     2, 'female','alive', '2020-03-10', '三毛猫',                     '三毛',            4.20, 2, '2024-11-18'),
    (4, 1, 3,  '3-1', 'タロウ',       'たろう',   1, 'male',  'alive', '2019-05-15', '柴犬',                       'レッド',          8.3,  NULL, NULL),
    (5, 1, 3,  '3-2', 'ジロウ',       'じろう',   1, 'male',  'alive', '2021-08-10', '柴犬',                       'ブラック',        7.1,  NULL, NULL),
    (6, 1, 4,  '4-1', 'チョコ',       'ちょこ',   1, 'female','alive', '2017-11-20', 'トイプードル',               'チョコ',          3.80, 1, NULL),
    (7, 1, 5,  '5-1', 'レオ',         'れお',     2, 'male',  'alive', '2016-07-04', 'スコティッシュフォールド',   'グレー',          5.5,  NULL, NULL),
    (8, 1, 6,  '6-1', 'ハチ',         'はち',     1, 'male',  'alive', '2018-03-25', '秋田犬',                     'ホワイト',       22.0,  NULL, NULL),
    (9, 1, 7,  '7-1', 'モモ',         'もも',     2, 'female','alive', '2022-01-15', 'マンチカン',                 'キャリコ',        3.2,  2, NULL),
    (10, 1, 8,  '8-1', 'ロッキー',     'ろっきー', 1, 'male',  'alive', '2014-09-08', 'ボーダーコリー',             'ブラックホワイト',18.5,  NULL, NULL),
    (11, 1, 9,  '9-1', 'ルナ',         'るな',     2, 'female','alive', '2021-02-28', 'ペルシャ',                   'シルバー',        4.80, 1, NULL),
    (12, 1, 10, '10-1', 'ケン',        'けん',     1, 'male',  'alive', '2013-06-18', 'ジャーマンシェパード',       'ブラックタン',   32.0,  NULL, NULL),
    (13, 1, 11, '11-1', 'ソラ',        'そら',     2, 'male',  'alive', '2023-04-01', 'アメリカンショートヘア',     'タビー',          3.0,  NULL, NULL),
    (14, 1, 12, '12-1', 'ゴン',        'ごん',     1, 'male',  'alive', '2016-12-05', '紀州犬',                     'ホワイト',       19.5,  NULL, NULL),
    (15, 1, 13, '13-1', 'シロ',        'しろ',     1, 'male',  'alive', '2020-08-10', 'ミックス犬',                 'ホワイト',        6.2,  NULL, NULL),
    (16, 1, 14, '14-1', 'トラ',        'とら',     2, 'male',  'alive', '2019-10-22', 'トラ猫',                     'トラ',            5.1,  NULL, NULL),
    (17, 1, 15, '15-1', 'ベロ',        'べろ',     1, 'male',  'alive', '2018-05-03', 'ビーグル',                   'トライカラー',   13.2,  NULL, NULL),
    (18, 1, 16, '16-1', 'チビ',        'ちび',     2, 'female','alive', '2022-06-20', 'ミックス猫',                 'サビ',            3.50, NULL, NULL),
    (19, 1, 17, '17-1', 'ポチ',        'ぽち',     1, 'male',  'alive', '2017-02-14', 'ダックスフンド',             'チョコ',          7.8,  NULL, NULL),
    (20, 1, 18, '18-1', 'モカ',        'もか',     2, 'female','alive', '2022-05-10', 'ミックス猫',                 'ホワイト',        4.1,  NULL, NULL),
    (21, 1, 18, '18-2', 'クルミ',      'くるみ',   1, 'male',  'alive', '2020-08-20', 'ミックス犬',                 'ベージュ',        8.3,  NULL, NULL),
    (22, 1, 19, '19-1', 'ハル',        'はる',     1, 'male',  'alive', '2019-03-15', 'ミックス犬',                 'ブラック',       12.5,  NULL, NULL),
    (23, 1, 19, '19-2', 'ユキ',        'ゆき',     2, 'female','alive', '2021-12-01', 'ミックス猫',                 'ホワイト',        3.80, NULL, NULL),
    (24, 1, 20, '20-1', 'ピーチ',      'ぴーち',   2, 'female','alive', '2023-01-07', 'ミックス猫',                 'オレンジ',        3.2,  NULL, NULL),
    (25, 1, 21, '21-1', 'コタ',        'こた',     1, 'male',  'alive', '2018-09-23', 'ミックス犬',                 'ブラウン',       22.0,  NULL, NULL),
    (26, 1, 21, '21-2', 'アン',        'あん',     2, 'female','alive', '2020-04-11', 'ミックス猫',                 'キャリコ',        4.5,  NULL, NULL),
    (27, 1, 22, '22-1', 'ゴマ',        'ごま',     2, 'male',  'alive', '2022-11-30', 'ミックス猫',                 'グレー',          5.0,  NULL, NULL),
    (28, 1, 22, '22-2', 'マル',        'まる',     1, 'female','alive', '2021-06-18', 'ミックス犬',                 'ゴールデン',      9.7,  NULL, NULL)
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();


-- -----------------------------------------------------------------------------
-- 3. appointments（予約: 10件）
-- -----------------------------------------------------------------------------
INSERT INTO appointments (id, clinic_id, start_time, end_time, owner_id, pet_id, visit_type, reservation_type_id, doctor_id, is_designated, status, notes) VALUES
    (1, 1, '2026-03-12 09:00:00+09', '2026-03-12 09:15:00+09', 1,  1,  'revisit', 1, 1, true,  'completed',       '皮膚の経過観察'),
    (2, 1, '2026-03-12 09:15:00+09', '2026-03-12 09:30:00+09', 2,  3,  'revisit', 7, 2, false, 'accounting',      '猫の定期健診'),
    (3, 1, '2026-03-12 10:00:00+09', '2026-03-12 10:15:00+09', 3,  4,  'revisit', 1, 1, true,  'in_consultation', '足を引きずっている'),
    (4, 1, '2026-03-12 10:15:00+09', '2026-03-12 10:30:00+09', 4,  6,  'first',   3, 2, false, 'checked_in',      'ワクチン接種希望'),
    (5, 1, '2026-03-12 14:00:00+09', '2026-03-12 14:15:00+09', 6,  8,  'revisit', 1, 1, false, 'confirmed',       '食欲低下が続いている'),
    (6, 1, '2026-03-13 09:00:00+09', '2026-03-13 09:15:00+09', 7,  9,  'revisit', 1, 2, true,  'confirmed',       '耳の治療経過確認'),
    (7, 1, '2026-03-13 10:00:00+09', '2026-03-13 10:15:00+09', 8,  10, 'first',   1, 1, false, 'confirmed',       '嘔吐が続いている'),
    (8, 1, '2026-03-14 09:30:00+09', '2026-03-14 09:45:00+09', 9,  11, 'revisit', 1, 2, false, 'confirmed',       'ルナの経過観察'),
    (9, 1, '2026-03-15 11:00:00+09', '2026-03-15 11:15:00+09', 10, 12, 'first',   3, 1, false, 'confirmed',       '初回ワクチン接種'),
    (10, 1, '2026-03-16 14:00:00+09', '2026-03-16 14:15:00+09', 11, 13, 'revisit', 1, 2, true,  'confirmed',       '腎臓値の経過観察')
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();


-- -----------------------------------------------------------------------------
-- 4. medical_records（カルテ: 20件）
-- -----------------------------------------------------------------------------
INSERT INTO medical_records (id, clinic_id, record_no, date, owner_id, pet_id, doctor_id, status) VALUES
    (1, 1, 'R-2025-001', '2025-10-10', 1,  1,  1, 'finalized'),
    (2, 1, 'R-2025-002', '2025-12-15', 1,  1,  1, 'finalized'),
    (3, 1, 'R-2026-001', '2026-01-20', 1,  1,  2, 'finalized'),
    (4, 1, 'R-2025-003', '2025-11-05', 1,  2,  2, 'finalized'),
    (5, 1, 'R-2025-004', '2025-09-15', 2,  3,  2, 'finalized'),
    (6, 1, 'R-2026-002', '2026-01-06', 2,  3,  1, 'finalized'),
    (7, 1, 'R-2025-005', '2025-08-22', 3,  4,  2, 'finalized'),
    (8, 1, 'R-2025-006', '2025-10-18', 4,  6,  1, 'finalized'),
    (9, 1, 'R-2025-007', '2025-07-30', 5,  7,  2, 'finalized'),
    (10, 1, 'R-2026-003', '2026-01-15', 6,  8,  1, 'draft'),
    (11, 1, 'R-2025-008', '2025-12-01', 7,  9,  2, 'finalized'),
    (12, 1, 'R-2025-009', '2025-11-20', 8,  10, 1, 'finalized'),
    (13, 1, 'R-2026-004', '2026-02-10', 9,  11, 2, 'draft'),
    (14, 1, 'R-2025-010', '2025-06-15', 10, 12, 1, 'finalized'),
    (15, 1, 'R-2026-005', '2026-01-06', 11, 13, 2, 'finalized'),
    (16, 1, 'R-2025-011', '2025-09-08', 12, 14, 1, 'finalized'),
    (17, 1, 'R-2026-006', '2026-02-28', 13, 15, 2, 'draft'),
    (18, 1, 'R-2025-012', '2025-08-20', 14, 16, 1, 'finalized'),
    (19, 1, 'R-2026-007', '2026-01-03', 15, 17, 2, 'finalized'),
    (20, 1, 'R-2026-008', '2026-01-06', 16, 18, 1, 'finalized')
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();


-- -----------------------------------------------------------------------------
-- 5. inquiries（問診: 20件）
-- -----------------------------------------------------------------------------
INSERT INTO inquiries (id, medical_record_id, chief_complaint_type_id, chief_complaint, notes, staff_id) VALUES
    (1,  1,  NULL, '狂犬病ワクチン接種',         '体調良好。', 1),
    (2,  2,  NULL, '定期健診',                   '特に異常なし。', 2),
    (3, 4,  6,    '右足の跛行',                 '膝蓋骨脱臼を確認。', 1),
    (4, 5,  NULL, 'フィラリア予防',             '予防薬処方。', 2),
    (5, 3,  NULL, '5種混合ワクチン接種',        '体調良好。', 1),
    (6,  6,  NULL, '5種混合ワクチン接種',        '体調良好。', 2),
    (7,  7,  NULL, 'ノミダニ予防薬',             '予防薬処方。', 1),
    (8,  8,  3,    '皮膚の痒み',                 'アトピー性皮膚炎疑い。', 2),
    (9,  9,  3,    'トリミング後の皮膚チェック', '軽度の赤みあり。', 1),
    (10, 10, 1,    '食欲不振',                   '2日前から食欲減退。', 2),
    (11, 11, NULL, '耳を痒がる',                 '外耳炎疑い。', 1),
    (12, 12, NULL, '定期健診・予防接種',         '年次健診。', 2),
    (13, 13, 2,    '嘔吐・下痢',                 '昨日から嘔吐3回。', 1),
    (14, 14, NULL, '生化学検査',                 'シニア健診。', 2),
    (15, 15, NULL, '3種混合ワクチン接種（猫）',  '初回ワクチン。', 1),
    (16, 16, NULL, '血液検査',                   '異常なし。', 2),
    (17, 17, NULL, '歯石除去',                   '重度の歯石付着。', 1),
    (18, 18, NULL, '定期検診',                   '体重管理継続。', 2),
    (19, 19, 6,    '再診（右足跛行）',           '改善傾向。', 1),
    (20, 20, NULL, '5種混合ワクチン接種',        '体調良好。', 2)
ON CONFLICT (id) DO UPDATE SET
    medical_record_id = EXCLUDED.medical_record_id,
    updated_at        = now();


-- -----------------------------------------------------------------------------
-- 5b. clinical_plans（診察/治療プラン: 20件）
-- -----------------------------------------------------------------------------
INSERT INTO clinical_plans (id, medical_record_id, physical_exam, diagnosis_type_id, diagnosis_name_id, diagnosis_details, treatment_policy) VALUES
    (1,  1, '体温38.5℃。心肺音正常。', NULL, NULL, '健康状態良好。ワクチン接種可。', '5種混合ワクチン接種実施。'),
    (2,  2, '体重増加あり。他異常なし。', NULL, NULL, '維持状態良好。', '定期検診継続。'),
    (3, 4, '右後肢跛行。パテラG2。', 8, 42, '膝蓋骨脱臼。', '消炎剤処方。体重管理指導。'),
    (4, 5, '異常なし。', NULL, NULL, '予防シーズン開始。', 'フィラリア予防薬処方。'),
    (5, 3, '良好。', NULL, NULL, '年次予防。', 'ワクチン接種。'),
    (6,  6, '良好。', NULL, NULL, '年次予防。', 'ワクチン接種。'),
    (7,  7, '良好。', NULL, NULL, '外部寄生虫予防。', 'スポットオン投与。'),
    (8,  8, '全身に発赤。搔痒感強。', 3, 6, 'アトピー性皮膚炎。', '抗ヒスタミン薬処方。薬用シャンプー推奨。'),
    (9,  9, '皮膚の一部に発赤。', 3, 7, '膿皮症初期。', '洗浄と消毒。'),
    (10, 10, '腹部軽度緊張。', 1, 1, '急性胃腸炎疑い。', '絶食・皮下補液実施。'),
    (11, 11, '耳道内に分泌物。', 3, 41, '外耳炎。', '耳道洗浄・点耳薬処方。'),
    (12, 12, 'シニア期に入る。', NULL, NULL, '健康診断実施。', '結果待ち。'),
    (13, 13, '脱水傾向あり。', 1, 1, '急性胃腸炎。', '対症療法と食事療法。'),
    (14, 14, '良好。', NULL, NULL, '経過観察。', '維持。'),
    (15, 15, '良好。', NULL, NULL, '幼若期検診。', '成長記録。'),
    (16, 16, '良好。', NULL, NULL, 'スクリーニング。', '異常なし。'),
    (17, 17, '重度の歯石。', NULL, NULL, '歯周病。', '抜歯を含めた歯科処置を計画。'),
    (18, 18, '良好。', NULL, NULL, '肥満気味。', 'ダイエットフード提案。'),
    (19, 19, '跛行消失。', 8, 42, '回復期。', '運動制限解除。'),
    (20, 20, '良好。', NULL, NULL, '年次予防。', 'ワクチン接種。')
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();


-- -----------------------------------------------------------------------------
-- 6. vital_records（バイタル: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO vital_records (id, pet_id, medical_record_id, recorded_at, staff_id, temperature, heart_rate, respiration_rate, weight, weight_unit, notes) VALUES
    (1, 1,  3, '2026-01-20 09:15:00+09', 1, 38.5, 80,  20, 26.5, 'Kg', '皮膚の搔痒感あり。体重良好。'),
    (2, 1,  2, '2025-12-15 10:00:00+09', 2, 38.8, 82,  22, 26.0, 'Kg', '体重前回比-500g'),
    (3, 1,  3, '2026-01-20 09:30:00+09', 1, 38.3, 78,  20, 26.5, 'Kg', '定期検診。皮膚搔痒感 軽快傾向。'),
    (4, 2,  4, '2025-11-05 11:00:00+09', 1, 39.1, 95,  24, 15.2, 'Kg', '軽度脱水。CRT 2秒。'),
    (5, 1,  5, '2025-09-15 14:30:00+09', 2, 38.2, 160, 30, 4200, 'g',  '粘膜色やや蒼白。食欲低下継続。')
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('vital_records', 'id'), (SELECT MAX(id) FROM vital_records));

-- -----------------------------------------------------------------------------
-- 7. treatments（治療明細: 8件）
-- -----------------------------------------------------------------------------
INSERT INTO treatments (id, medical_record_id, item_type, consultation_id, procedure_id, medicine_id, inventory_id, is_selected, status, content, unit_price, quantity, sort_order) VALUES
    (1, 1, 'consultation', 2,    NULL, NULL, NULL, true, 'completed', '再診料',                    800,  1, 1),
    (2, 1, 'medicine',     NULL, NULL, 1,    NULL, true, 'completed', 'アモキシシリン 50mg x 7日分', 500,  7, 2),
    (3, 2, 'consultation', 2,    NULL, NULL, NULL, true, 'completed', '再診料',                    800,  1, 1),
    (4, 1, 'consultation', 2,    NULL, NULL, NULL, true, 'completed', '再診料',                    800,  1, 1),
    (5, 1, 'procedure',    NULL, 4,    NULL, NULL, true, 'completed', '耳道洗浄（左耳）',          2500, 1, 2),
    (6, 2, 'consultation', 1,    NULL, NULL, NULL, true, 'completed', '初診料',                    2000, 1, 1),
    (7, 2, 'medicine',     NULL, NULL, 1,    NULL, true, 'completed', 'アモキシシリン 50mg x 5日分', 500,  5, 2),
    (8, 3, 'consultation', 2,    NULL, NULL, NULL, true, 'completed', '再診料',                    800,  1, 1)
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();


-- -----------------------------------------------------------------------------
-- 8. trimming appointments（予約ベーストリミング: 8件）
-- reservation_types の category を trimming に設定
-- -----------------------------------------------------------------------------
UPDATE reservation_types SET category = 'trimming' WHERE clinic_id = 1 AND id IN (9, 10, 11, 12);

INSERT INTO appointments (id, clinic_id, start_time, end_time, owner_id, pet_id, visit_type, reservation_type_id, doctor_id, is_designated, status, notes) VALUES
    (101, 1, '2025-10-10 10:00:00+09', '2025-10-10 11:30:00+09', 1,  1,  'first', 9,  6,  false, 'completed',       'サマーカット希望'),
    (102, 1, '2025-10-15 10:00:00+09', '2025-10-15 11:30:00+09', 1,  2,  'first', 9,  12, false, 'confirmed',       'ふんわりカット'),
    (103, 1, '2025-10-12 10:00:00+09', '2025-10-12 11:30:00+09', 2,  3,  'first', 9,  6,  false, 'in_consultation', '毛玉カット'),
    (104, 1, '2026-01-06 10:00:00+09', '2026-01-06 11:00:00+09', 4,  6,  'first', 11, 6,  false, 'completed',       'シャンプーコース'),
    (105, 1, '2026-01-06 11:00:00+09', '2026-01-06 12:30:00+09', 15, 17, 'first', 9,  12, false, 'completed',       '全体カット'),
    (106, 1, '2026-01-06 13:00:00+09', '2026-01-06 14:30:00+09', 8,  10, 'first', 10, 12, false, 'confirmed',       '爪切り・ブラッシング'),
    (107, 1, '2026-01-06 14:00:00+09', '2026-01-06 15:00:00+09', 13, 15, 'first', 11, 6,  false, 'completed',       'シャンプー'),
    (108, 1, '2026-01-06 15:00:00+09', '2026-01-06 16:30:00+09', 4,  6,  'first', 9,  6,  false, 'confirmed',       'トリミング')
ON CONFLICT (id) DO UPDATE SET updated_at = now();

INSERT INTO appointment_trimming_details (appointment_id, clinic_id, course_id, body_weight, bw_unit, style_request) VALUES
    (101, 1, 5, 26.5, 'Kg', 'サマーカット希望'),
    (102, 1, 4, 15.2, 'Kg', 'ふんわりカット'),
    (103, 1, 1, 4.2,  'Kg', '毛玉カット'),
    (104, 1, 1, 3800, 'g',  'シャンプーコース'),
    (105, 1, 4, 12.0, 'Kg', '全体カット'),
    (106, 1, 2, 8.0,  'Kg', '爪切り・ブラッシング'),
    (107, 1, 1, 5.0,  'Kg', 'シャンプー'),
    (108, 1, 3, 3800, 'g',  'トリミング')
ON CONFLICT (appointment_id) DO UPDATE SET updated_at = now();

SELECT setval(pg_get_serial_sequence('appointment_trimming_details', 'id'), (SELECT MAX(id) FROM appointment_trimming_details));

-- -----------------------------------------------------------------------------
-- 9. hospitalizations（入院: 7件）
-- -----------------------------------------------------------------------------
INSERT INTO hospitalizations (id, clinic_id, owner_id, pet_id, hospitalization_type, start_date, end_date, status, cage_id, doctor_id, memo, owner_request, staff_notes) VALUES
    (1, 1, 3, 5,  'hospitalization', '2026-03-10', '2026-03-14', 'admitted',   5,    1, '急性胃腸炎による脱水治療。点滴管理中。',  '食事のアレルギーに注意してほしい（鶏肉不可）', '3/10入院開始。静脈点滴開始。3/11嘔吐1回。3/12状態改善傾向。'),
    (2, 1, 6, 8,  'hospitalization', '2026-02-25', '2026-02-28', 'discharged', 4,    1, '外耳炎重症化に伴う入院治療。',             '怖がりなので優しく接してほしい',               '耳道洗浄を毎日実施。2/28退院時、症状改善。点耳薬処方。'),
    (3, 1, 17, 19, 'hospitalization', '2026-02-10', '2026-02-20', 'discharged', NULL, 1, '骨折治療による入院。手術後経過観察。', '', '2/10手術実施。2/15抜糸。2/20退院。'),
    (4, 1, 4,  6,  'hotel',           '2026-03-15', '2026-03-18', 'reserved',   NULL, 1, '旅行中のホテル預かり。', 'フードはロイヤルカナンのみ', ''),
    (5, 1, 1,  1,  'hospitalization', '2026-03-20', '2026-03-25', 'reserved',   NULL, 1, '膝蓋骨脱臼手術予定。術前検査済み。', '怖がりなので静かな環境を希望', ''),
    (6, 1, 9,  11, 'hospitalization', '2026-03-05', '2026-03-12', 'admitted',   1,    2, '慢性腎臓病の集中治療。点滴管理中。', 'ペルシャ猫のため温度管理に注意', '3/5入院。毎日皮下補液実施。3/12現在状態安定。'),
    (7, 1, 3,  4,  'hospitalization', '2026-01-03', '2026-01-06', 'discharged', NULL, 1, '急性胃腸炎による脱水治療。', 'チキンアレルギーあり', '1/3入院。点滴開始。1/6状態改善し退院。')
ON CONFLICT (id) DO UPDATE SET
    updated_at            = now();

SELECT setval(pg_get_serial_sequence('hospitalizations', 'id'), (SELECT MAX(id) FROM hospitalizations));

-- -----------------------------------------------------------------------------
-- 10. care_plan_items（ケアプラン: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO care_plan_items (id, hospitalization_id, type, name, description, timing, status, notes, medicine_id, procedure_id, hospitalization_plan_id, unit_price, category, sort_order) VALUES
    (1, 1, 'food',        '療法食（消化器ケア）', '1日3回、少量ずつ与える', ARRAY['morning','noon','night']::plan_timing[], 'active', '鶏肉不可。', NULL, NULL, NULL, 0, '食事', 1),
    (2, 1, 'medicine',    'アモキシシリン',       '1回1錠、朝夕食後',       ARRAY['morning','night']::plan_timing[],       'active', '抗生剤。', 1,    NULL, NULL, 500, '投薬', 2),
    (3, 1, 'instruction', 'バイタルチェック',     '1日3回測定',             ARRAY['morning','noon','night']::plan_timing[], 'active', '異常値報告。', NULL, NULL, NULL, 0, '観察', 3),
    (4, 2, 'treatment',   '耳道洗浄',             '1日1回、朝に実施',       ARRAY['morning']::plan_timing[],               'completed', '左耳。', NULL, 4,    NULL, 2500, '処置', 1),
    (5, 2, 'item',        '入院管理料',           '小型犬1日分',            ARRAY['morning']::plan_timing[],               'completed', '', NULL, NULL, 1,    3000, '入院', 2)
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('care_plan_items', 'id'), (SELECT MAX(id) FROM care_plan_items));

-- -----------------------------------------------------------------------------
-- 11. inventory_items（在庫管理: 9件追加）
-- -----------------------------------------------------------------------------
INSERT INTO inventory_items (id, clinic_id, name, category, quantity, unit, min_stock_level, location, supplier, status) VALUES
    (6, 1, '5種混合ワクチン',               'medicine',   25,  'バイアル', 15, '冷蔵庫 1',    '共立製薬',                'sufficient'),
    (7, 1, '留置針 22G',                    'consumable',  0,   '本',       50, '処置室 棚D',  'テルモ',                  'out_of_stock'),
    (8, 1, 'シリンジ 5mL',                  'consumable', 300,  '本',      100, '処置室 棚D',  'テルモ',                  'sufficient'),
    (9, 1, 'メトクロプラミド注 10mg',        'medicine',    8,   'アンプル', 10, '薬品棚 A-3', '日本全薬工業',            'low'),
    (10, 1, '療法食 消化器サポート（猫用）',  'food',       10,   '袋',        5, 'フード棚 C-1','ヒルズ',                 'sufficient'),
    (11, 1, 'エリザベスカラー（S）',          'other',      15,   '個',        5, '倉庫 A',     'ペットメディカルサプライ', 'sufficient'),
    (12, 1, 'ガーゼ 滅菌 7.5cm',            'consumable',  45,   '枚',       50, '処置室 棚E',  '白十字',                  'low'),
    (13, 1, 'フィラリア予防薬（S）',          'medicine',   60,   '錠',       30, '薬品棚 B-1', 'メリアル・ジャパン',       'sufficient'),
    (14, 1, 'ノミダニ駆除薬 スポット',        'medicine',   40,   'ピペット',  20, '薬品棚 B-2', 'エランコジャパン',         'sufficient')
ON CONFLICT (id) DO NOTHING;


-- -----------------------------------------------------------------------------
-- 12. billings / billing_items / payments
-- -----------------------------------------------------------------------------
INSERT INTO billings (id, clinic_id, medical_record_id, hospitalization_id, owner_id, pet_id, subtotal, tax_total, total_amount, has_insurance, status, scheduled_date, completed_at, memo) VALUES
    (1, 1, 1,    NULL, 1,  1,  4300, 430, 4730, true, 'completed', '2026-02-15', '2026-02-15 10:30:00+09', 'アニコム保険適用'),
    -- BUG-374 TC-367-04 検証用: billing_id=2 に軽減税率 8% 品目を追加したため金額再計算
    -- 再診料 800 + 耳道洗浄 2500 (10%対象 = 3300) + ロイヤルカナン 2800 (8%対象) = 6100
    -- tax = 330 (10%) + 224 (8%) = 554 / total = 6654
    (2, 1, 3,    NULL, 1,  1,  6100, 554, 6654, true, 'completed', '2026-02-28', '2026-02-28 11:00:00+09', 'アニコム保険適用（Iris 耳炎治療）+ フード販売'),
    (3, 1, 6,    NULL, 2,  3,  800,  80,  880,  true, 'waiting',   '2026-03-12', NULL,                     'アニコム保険適用。会計待ち。')
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('billings', 'id'), (SELECT MAX(id) FROM billings));

INSERT INTO billing_items (id, billing_id, category, name, unit_price, quantity, tax_rate, is_insurance_applicable, source, sort_order) VALUES
    (1, 1, 'other',    '再診料',                    800,  1, 0.10, true, 'medical_record', 1),
    (2, 1, 'medicine', 'アモキシシリン 50mg x 7日分', 500,  7, 0.10, true, 'medical_record', 2),
    (3, 2, 'other',    '再診料',                    800,  1, 0.10, true, 'medical_record', 1),
    (4, 2, 'procedure','耳道洗浄',                  2500, 1, 0.10, true, 'medical_record', 2),
    -- BUG-374 TC-367-04 検証用: 軽減税率 8% 品目を会計 billing_id=2 に追加
    (6, 2, 'food',     'ロイヤルカナン 消化器サポート 1kg', 2800, 1, 0.08, false, 'manual', 3),
    (5, 1, 'other',    '再診料',                    800,  1, 0.10, true, 'medical_record', 1)
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('billing_items', 'id'), (SELECT MAX(id) FROM billing_items));

INSERT INTO payments (id, billing_id, subtotal, tax_total, total_amount, insurance_name, insurance_ratio, insurance_amount, discount_amount, billing_amount, received_amount, change_amount, method, paid_by) VALUES
    (1, 1, 4300, 430, 4730, 'アニコム損保', 0.70, 3311, 0, 1419, 1500, 81, 'cash', 1),
    -- BUG-374 TC-367-04: billing_id=2 は軽減税率 8% 品目を含むため payment も再計算
    -- insurance 対象は 10% 品目のみ (3300 税抜 × 1.10 = 3630) × 70% = 2541 (変更なし)
    -- billing_amount = total(6654) - insurance(2541) = 4113、受領 4200、釣り 87
    (2, 2, 6100, 554, 6654, 'アニコム損保', 0.70, 2541, 0, 4113, 4200, 87, 'credit_card', 1)
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('payments', 'id'), (SELECT MAX(id) FROM payments));

-- -----------------------------------------------------------------------------
-- 13. billing_refunds（返金デモデータ）
-- -----------------------------------------------------------------------------
INSERT INTO billing_refunds (id, clinic_id, billing_id, amount, reason, refunded_by, refunded_at) VALUES
    (1, 1, 1, 919,  '処置内容の変更に伴う部分返金',  1, '2026-02-16 10:00:00+09'),
    (2, 1, 1, 500,  '薬剤変更による差額返金',        1, '2026-02-20 14:30:00+09'),
    (3, 1, 2, 500,  '診察キャンセル分の返金',         1, '2026-03-01 09:00:00+09')
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('billing_refunds', 'id'), (SELECT MAX(id) FROM billing_refunds));

-- -----------------------------------------------------------------------------
-- 13b. 2026-04-25 デモ会計データ（レジ締めプレビュー用、clinic_id=1）
-- AM 区分: completed_at 00:00-14:00 JST / PM 区分: 14:00-18:30 JST
-- -----------------------------------------------------------------------------

-- 本日分カルテ（医師 doctor_id 1 or 2、clinic_id=1 のみ）
INSERT INTO medical_records (id, clinic_id, record_no, date, owner_id, pet_id, doctor_id, status) VALUES
    (21, 1, 'R-2026-021', '2026-04-25', 2,  3,  1, 'finalized'),
    (22, 1, 'R-2026-022', '2026-04-25', 4,  6,  2, 'finalized'),
    (23, 1, 'R-2026-023', '2026-04-25', 8,  10, 1, 'finalized'),
    (24, 1, 'R-2026-024', '2026-04-25', 9,  11, 2, 'finalized'),
    (25, 1, 'R-2026-025', '2026-04-25', 11, 13, 1, 'finalized')
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('medical_records', 'id'), (SELECT MAX(id) FROM medical_records));

-- 本日会計（AM 4件 / PM 3件）
INSERT INTO billings (id, clinic_id, medical_record_id, hospitalization_id, owner_id, pet_id, subtotal, tax_total, total_amount, has_insurance, status, scheduled_date, completed_at, memo) VALUES
    -- AM 09:30: 田中花子 / ミケ（初診 + 猫3種ワクチン）
    (4,  1, 21, NULL, 2,  3,  4500,  450,  4950,  false, 'completed', '2026-04-25', '2026-04-25 09:30:00+09', ''),
    -- AM 10:30: 田中美咲 / チョコ（再診 + 皮膚炎処置 + 外用薬）
    (5,  1, 22, NULL, 4,  6,  3500,  350,  3850,  false, 'completed', '2026-04-25', '2026-04-25 10:30:00+09', ''),
    -- AM 11:30: 佐藤花子 / レオ（トリミング + フード購入、10%/8% 混在）
    (6,  1, NULL, NULL, 5, 7,  6200,  596,  6796,  false, 'completed', '2026-04-25', '2026-04-25 11:30:00+09', 'トリミング来院'),
    -- AM 13:00: 中村勇気 / ロッキー（再診 + 血液検査、アニコム 70%）
    (7,  1, 23, NULL, 8,  10, 6800,  680,  7480,  true,  'completed', '2026-04-25', '2026-04-25 13:00:00+09', 'アニコム保険適用'),
    -- PM 14:30: 加藤恵 / ルナ（避妊手術、アニコム 70%）
    (8,  1, 24, NULL, 9,  11, 25000, 2500, 27500, true,  'completed', '2026-04-25', '2026-04-25 14:30:00+09', 'アニコム保険適用（避妊手術）'),
    -- PM 15:30: 山田太郎 / ケン（ペットホテル 2泊）
    (9,  1, NULL, NULL, 10, 12, 5000,  500,  5500,  false, 'completed', '2026-04-25', '2026-04-25 15:30:00+09', 'ペットホテル利用'),
    -- PM 16:30: 高橋由美 / ソラ（再診 + 点眼薬）
    (10, 1, 25, NULL, 11, 13, 1400,  140,  1540,  false, 'completed', '2026-04-25', '2026-04-25 16:30:00+09', '')
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('billings', 'id'), (SELECT MAX(id) FROM billings));

INSERT INTO billing_items (id, billing_id, category, name, unit_price, quantity, tax_rate, is_insurance_applicable, source, sort_order) VALUES
    -- Billing 4: ミケ（初診 + ワクチン）
    (7,  4, 'examination', '初診料',                            1500, 1, 0.10, true,  'medical_record', 1),
    (8,  4, 'vaccine',     '猫3種混合ワクチン',                  3000, 1, 0.10, true,  'medical_record', 2),
    -- Billing 5: チョコ（再診 + 処置 + 薬）
    (9,  5, 'other',       '再診料',                             800,  1, 0.10, true,  'medical_record', 1),
    (10, 5, 'procedure',   'アトピー皮膚炎処置',                 1500, 1, 0.10, true,  'medical_record', 2),
    (11, 5, 'medicine',    'ステロイド外用薬 30g',               1200, 1, 0.10, true,  'medical_record', 3),
    -- Billing 6: レオ（トリミング + フード、8% 混在）
    (12, 6, 'trimming',    'フルトリミング（猫）',               5000, 1, 0.10, false, 'manual',         1),
    (13, 6, 'food',        'ヒルズ サイエンスダイエット 猫用 400g', 1200, 1, 0.08, false, 'manual',       2),
    -- Billing 7: ロッキー（再診 + 血液検査）
    (14, 7, 'other',       '再診料',                             800,  1, 0.10, true,  'medical_record', 1),
    (15, 7, 'test',        '血液検査（CBC + 生化学）',           6000, 1, 0.10, true,  'medical_record', 2),
    -- Billing 8: ルナ（避妊手術）
    (16, 8, 'surgery',     '避妊手術（猫）',                    25000, 1, 0.10, true,  'medical_record', 1),
    -- Billing 9: ケン（ペットホテル 2泊）
    (17, 9, 'hotel',       'ペットホテル（1泊）',               2500, 2, 0.10, false, 'manual',         1),
    -- Billing 10: ソラ（再診 + 点眼薬）
    (18, 10, 'other',      '再診料',                             800,  1, 0.10, true,  'medical_record', 1),
    (19, 10, 'medicine',   '抗菌点眼薬（5ml）',                  600,  1, 0.10, true,  'medical_record', 2)
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('billing_items', 'id'), (SELECT MAX(id) FROM billing_items));

-- payments — 現金は payment_method_id=NULL（calcTheoreticalCash は NULL 行のみを現金として計算するため）
-- 非現金（クレカ/電子マネー/銀行振込）はサブクエリで name 引き（ID が環境依存のため）
INSERT INTO payments (id, billing_id, subtotal, tax_total, total_amount, insurance_name, insurance_ratio, insurance_amount, discount_amount, billing_amount, received_amount, change_amount, method, payment_method_id, paid_by) VALUES
    -- Billing 4: 現金 4950円（受取 5000、釣 50）— payment_method_id=NULL で理論現金に加算
    (3,  4,  4500,  450,  4950,  '',              0.00, 0,     0, 4950, 5000, 50,  'cash',         NULL, 1),
    -- Billing 5: クレジットカード 3850円
    (4,  5,  3500,  350,  3850,  '',              0.00, 0,     0, 3850, 3850, 0,   'credit_card',
     (SELECT id FROM payment_methods WHERE clinic_id=1 AND name='クレジットカード' AND deleted_at IS NULL LIMIT 1), 1),
    -- Billing 6: 現金 6796円（受取 7000、釣 204）— payment_method_id=NULL で理論現金に加算
    (5,  6,  6200,  596,  6796,  '',              0.00, 0,     0, 6796, 7000, 204, 'cash',         NULL, 1),
    -- Billing 7: クレジットカード（アニコム 70%、実費 2244円）
    (6,  7,  6800,  680,  7480,  'アニコム損保',  0.70, 5236,  0, 2244, 2244, 0,   'credit_card',
     (SELECT id FROM payment_methods WHERE clinic_id=1 AND name='クレジットカード' AND deleted_at IS NULL LIMIT 1), 1),
    -- Billing 8: 銀行振込（アニコム 70%、実費 8250円）
    (7,  8,  25000, 2500, 27500, 'アニコム損保',  0.70, 19250, 0, 8250, 8250, 0,   'cash',
     (SELECT id FROM payment_methods WHERE clinic_id=1 AND name='銀行振込'       AND deleted_at IS NULL LIMIT 1), 1),
    -- Billing 9: 電子マネー 5500円
    (8,  9,  5000,  500,  5500,  '',              0.00, 0,     0, 5500, 5500, 0,   'electronic_money',
     (SELECT id FROM payment_methods WHERE clinic_id=1 AND name='電子マネー'     AND deleted_at IS NULL LIMIT 1), 1),
    -- Billing 10: 現金 1540円（受取 2000、釣 460）— payment_method_id=NULL で理論現金に加算
    (9,  10, 1400,  140,  1540,  '',              0.00, 0,     0, 1540, 2000, 460, 'cash',         NULL, 1)
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('payments', 'id'), (SELECT MAX(id) FROM payments));

-- 返金デモ（billing_id=8 避妊手術の術前重複請求分、PM 区分内）
INSERT INTO billing_refunds (id, clinic_id, billing_id, amount, reason, refunded_by, refunded_at) VALUES
    (4, 1, 8, 2750, '術前検査費用の重複請求による部分返金', 1, '2026-04-25 17:00:00+09')
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('billing_refunds', 'id'), (SELECT MAX(id) FROM billing_refunds));

-- =============================================================================
-- 城東センター病院 (clinic_id=2) マスタデータ
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 城東センター病院 reservation_types（サービス種別: 19件、ID 26-44）
-- -----------------------------------------------------------------------------

-- 城東センター病院 公開コース
INSERT INTO reservation_types (id, clinic_id, name, short_name, is_active, description, color, sort_order, duration_minutes, reservation_visible, reservation_comment, is_internal) VALUES
    (26, 2, '一般診察',               '診察',     true, '内科・外科・皮膚科などの一般的な診察',         '#3B82F6', 1,  15, true,  '', false),
    (27, 2, '一般診察(再診)',          '再診',     true, '継続通院の一般診察',                           '#3B82F6', 2,  15, true,  '', false),
    (28, 2, 'ワクチン接種',            'ワクチン', true, '各種ワクチン接種（予防接種）',                 '#10B981', 3,  15, true,  '', false),
    (29, 2, '狂犬病',                 '狂犬病',   true, '狂犬病予防法に基づくワクチン接種',             '#10B981', 4,  15, true,  '', false),
    (30, 2, 'フィラリア予防',          'フィラリア', true, 'フィラリア予防薬投与・処方',                '#10B981', 5,  15, true,  '', false),
    (31, 2, '健康診断',               '健診',     true, '定期健康診断・フィラリア検査など',             '#8B5CF6', 6,  15, true,  '', false),
    (32, 2, '健康診断結果報告',        '結果報告', true, '健康診断結果の説明・報告',                     '#8B5CF6', 7,  15, true,  '', false),
    (33, 2, 'トリミングコース',        'トリミング', true, 'カット・シャンプー・ブロー・爪切り・耳掃除', '#F59E0B', 8,  15, true,  '', false),
    (34, 2, 'トリミングシャンプーコース', 'シャンプー', true, 'シャンプー・ブロー・ブラッシング',        '#F59E0B', 9,  15, true,  '', false),
    (35, 2, 'クイックシャンプー',      'Qシャンプー', true, '短時間シャンプー',                        '#F59E0B', 10, 15, true,  '', false)
ON CONFLICT DO NOTHING;

-- 城東センター病院 スタッフ専用コース
INSERT INTO reservation_types (id, clinic_id, name, short_name, is_active, description, color, sort_order, duration_minutes, reservation_visible, reservation_comment, is_internal) VALUES
    (36, 2, '手術60',                 '手術60',   true, '手術枠（60分）',                               '#EF4444', 11, 60, false, '', true),
    (37, 2, '手術30',                 '手術30',   true, '手術枠（30分）',                               '#EF4444', 12, 30, false, '', true),
    (38, 2, 'ホテルお迎え',           'お迎え',   true, 'ホテルお迎え対応',                             '#6B7280', 13, 15, false, '', true),
    (39, 2, 'ホテル預かり',           '預かり',   true, 'ペットホテル預かり',                           '#6B7280', 14, 15, false, '', true),
    (40, 2, '休憩枠',                 '休憩',     true, '休憩・ブロック枠',                             '#6B7280', 15, 60, false, '', true),
    (41, 2, '×',                     '×',       true, '予約不可（15分）',                             '#6B7280', 16, 15, false, '', true),
    (42, 2, '予約不可60',             '不可60',   true, '予約不可ブロック（60分）',                     '#6B7280', 17, 60, false, '', true),
    (43, 2, '予約不可30',             '不可30',   true, '予約不可ブロック（30分）',                     '#6B7280', 18, 30, false, '', true),
    (44, 2, 'エコー枠',               'エコー',   true, '超音波検査専用枠',                             '#8B5CF6', 19, 30, false, '', true)
ON CONFLICT DO NOTHING;


-- 城東センター病院 (clinic_id=2) グループ紐付け
UPDATE reservation_types SET group_id=8  WHERE clinic_id=2 AND id IN (26,27);
UPDATE reservation_types SET group_id=9  WHERE clinic_id=2 AND id IN (28,29,30);
UPDATE reservation_types SET group_id=10 WHERE clinic_id=2 AND id IN (31,32,44);
UPDATE reservation_types SET group_id=11 WHERE clinic_id=2 AND id IN (33,34,35);
UPDATE reservation_types SET group_id=12 WHERE clinic_id=2 AND id IN (36,37);
UPDATE reservation_types SET group_id=13 WHERE clinic_id=2 AND id IN (38,39);
UPDATE reservation_types SET group_id=14, is_internal=true WHERE clinic_id=2 AND id IN (40,41,42,43);

-- -----------------------------------------------------------------------------
-- 城東センター病院 cages（ケージ: 4件）
-- -----------------------------------------------------------------------------
INSERT INTO cages (id, clinic_id, name, price, is_active, description, cage_type, cage_size, sort_order) VALUES
    (9, 2, 'ICUケージ',       7500, true, '酸素吸入可・重症患者用',    'icu',     'medium', 1),
    (10, 2, '犬用ケージ（小）', 2800, true, '小型犬・ホテル利用可',      'dog',     'small',  2),
    (11, 2, '犬用ケージ（中）', 3200, true, '中型犬・一般入院用',        'dog',     'medium', 3),
    (12, 2, '猫用ケージ',       2800, true, '猫専用・ストレス軽減設計',  'cat',     'small',  4)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 城東センター病院 insurances（保険: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO insurances (id, clinic_id, name, is_active, description, coverage_rate, contact_phone, sort_order) VALUES
    (6, 2, 'アニコム損保',   true, 'ペット保険大手・どうぶつ健保シリーズ', 70, '0120-025-034', 1),
    (7, 2, 'アイペット損保', true, 'うちの子シリーズ',                     70, '0120-956-099', 2),
    (8, 2, 'その他（自費）', true, '保険未加入・全額自費',                100, '',             3)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 城東センター病院 exam_types（検査種別: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO exam_types (id, clinic_id, name, price, is_active, description, sort_order) VALUES
    (6, 2, '血液検査（CBC）', 3000, true, '全血球計算（Complete Blood Count）',     1),
    (7, 2, '血液化学検査',     5000, true, '肝機能・腎機能・血糖値など生化学的検査', 2),
    (8, 2, 'レントゲン検査',   3200, true, 'X線撮影（胸部・腹部・四肢）',            3)
ON CONFLICT DO NOTHING;


-- exam_type_fields for clinic 4
INSERT INTO exam_type_fields (id, exam_type_id, name, inspection_value, normal_value, sort_order) VALUES
    (18, 6, 'WBC（白血球数）', '', '6.0-17.0 x10^3/uL', 1),
    (19, 6, 'RBC（赤血球数）', '', '5.5-8.5 x10^6/uL',  2),
    (20, 6, 'HCT（ヘマトクリット）', '', '37-55%',       3),
    (21, 7, 'ALT（GPT）',      '', '10-125 U/L',         1),
    (22, 7, 'BUN（尿素窒素）', '', '7-27 mg/dL',         2),
    (23, 7, 'CRE（クレアチニン）', '', '0.5-1.8 mg/dL',  3),
    (24, 8, '胸部正面',        '', '異常なし',            1),
    (25, 8, '腹部正面',        '', '異常なし',            2)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 城東センター病院 vaccines（ワクチン: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO vaccines (id, clinic_id, name, price, is_active, description, species, interval, sort_order) VALUES
    (11, 2, '混合ワクチン5種（犬）',  4800, true, 'ジステンパー・パルボ・アデノ1型・アデノ2型・パラインフルエンザ', 'dog', '1年',   1),
    (12, 2, '混合ワクチン8種（犬）',  6800, true, '5種＋レプトスピラ3種',                                          'dog', '1年',   2),
    (13, 2, '混合ワクチン3種（猫）',  4200, true, '猫ウイルス性鼻気管炎・カリシウイルス・汎白血球減少症',           'cat', '1年',   3),
    (14, 2, '狂犬病ワクチン',         3000, true, '狂犬病予防法に基づく接種',                                      'dog', '1年',   4),
    (15, 2, 'フィラリア予防薬（小型犬）', 950, true, '体重10kg以下犬用フィラリア予防',                             'dog', '1ヶ月', 5)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 城東センター病院 medicines（薬剤: カテゴリ4件 + 薬剤10件）
-- -----------------------------------------------------------------------------
INSERT INTO medicines (id, clinic_id, name, price, is_active, description, sort_order) VALUES
    (2001, 2, '抗生剤',     NULL, true, '抗生物質カテゴリ',   1),
    (2002, 2, 'ステロイド', NULL, true, 'ステロイド剤カテゴリ', 2),
    (2003, 2, '消炎剤',     NULL, true, '消炎鎮痛剤カテゴリ', 3),
    (2004, 2, '駆虫剤',     NULL, true, '駆虫剤カテゴリ',     4),
    (2005, 2, '輸液',       NULL, true, '輸液カテゴリ',       5)
ON CONFLICT DO NOTHING;

INSERT INTO medicines (id, clinic_id, name, price, is_active, description, dosage_form, medicine_unit, default_quantity, sort_order, parent_id) VALUES
    (101, 2, 'アモキシシリン 50mg',        520,  true, '広域スペクトラム抗生物質',             'tablet',    'per_tablet', 1,    1, 2001),
    (102, 2, 'セファレキシン 250mg',        470,  true, '第1世代セフェム系抗生物質',           'tablet',    'per_tablet', 1,    2, 2001),
    (103, 2, 'メトロニダゾール 250mg',      620,  true, '嫌気性菌・原虫感染症治療薬',          'tablet',    'per_tablet', 1,    3, 2001),
    (104, 2, 'プレドニゾロン 5mg',          420,  true, 'ステロイド系抗炎症・免疫抑制剤',      'tablet',    'per_tablet', 1,    4, 2002),
    (105, 2, 'デキサメタゾン注射液',        720,  true, '強力ステロイド・アレルギー緊急治療',  'injection', 'per_ml',     1,    5, 2002),
    (106, 2, 'メロキシカム経口液',          730,  true, 'NSAIDs・痛み・炎症の緩和',            'liquid',    'per_ml',     1,    6, 2003),
    (107, 2, 'カルプロフェン 25mg',         680,  true, 'NSAIDs・術後疼痛管理',               'tablet',    'per_tablet', 1,    7, 2003),
    (108, 2, 'ノミ・ダニ駆除薬（犬用）',    2600, true, '外部寄生虫予防・駆除（スポットオン）', 'topical',   'per_dose',   1,    8, 2004),
    (109, 2, 'ノミ・ダニ駆除薬（猫用）',    2600, true, '外部寄生虫予防・駆除（スポットオン）', 'topical',   'per_dose',   1,    9, 2004),
    (110, 2, '生理食塩水 500ml',            420,  true, '点滴・洗浄用生理食塩水',              'liquid',    'per_ml',     500,  10, 2005)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 城東センター病院 consultations（診察項目: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO consultations (id, clinic_id, name, price, is_active, description, time_condition, duration, sort_order) VALUES
    (6, 2, '初診料',       2200, true, '初めての受診または6ヶ月以上受診がない場合', 'first_visit', 30, 1),
    (7, 2, '再診料',        900, true, '継続通院の診察料',                         'revisit',     15, 2),
    (8, 2, '時間外診療料', 3200, true, '診療時間外・休日の緊急診察',               'after_hours', 30, 3)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 城東センター病院 procedures（処置項目: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO procedures (id, clinic_id, name, price, is_active, description, duration, anesthesia, sort_order) VALUES
    (11, 2, '去勢手術（犬）', 26000, true, '雄犬の去勢手術',                  60, 'general', 1),
    (12, 2, '避妊手術（猫）', 26000, true, '雌猫の避妊手術',                  90, 'general', 2),
    (13, 2, '耳洗浄',          2800, true, '外耳炎治療・耳道内の洗浄処置',    15, 'none',    3),
    (14, 2, '爪切り',           500, true, '爪のカット・やすりがけ',           10, 'none',    4),
    (15, 2, '点滴処置',        3200, true, '静脈内点滴（1時間）',              60, 'none',    5)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 城東センター病院 hospitalization_plans（入院プラン: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO hospitalization_plans (id, clinic_id, name, price, is_active, description, body_size, billing_unit, sort_order) VALUES
    (6, 2, '一般入院（小型）', 3200, true, '体重10kg以下の入院管理料（1日）', 'small',  'per_day',   1),
    (7, 2, '一般入院（中型）', 3700, true, '体重10-25kgの入院管理料（1日）',  'medium', 'per_day',   2),
    (8, 2, 'ICU入院',          8500, true, '集中治療室管理料（1日）',         'small',  'per_day',   3)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 城東センター病院 trimming_courses（トリミングコース: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO trimming_courses (id, clinic_id, name, price, is_active, description, target_size, duration, sort_order) VALUES
    (6, 2, 'シャンプー&ブロー（小型）', 4200, true, 'シャンプー・ブロー・ブラッシング',            'small',  60,  1),
    (7, 2, 'フルコース（小型）',        7200, true, 'カット・シャンプー・ブロー・爪切り・耳掃除', 'small',  120, 2),
    (8, 2, 'フルコース（中型）',        9500, true, 'カット・シャンプー・ブロー・爪切り・耳掃除', 'medium', 150, 3)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 城東センター病院 trimming_options（トリミングオプション: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO trimming_options (id, clinic_id, name, price, is_active, description, duration, is_combinable, sort_order) VALUES
    (6, 2, '爪切り',     320, true, '爪のカット・やすりがけ',    10, true, 1),
    (7, 2, '耳掃除',     520, true, '外耳道の洗浄・清掃',        10, true, 2),
    (8, 2, '肛門腺絞り', 320, true, '肛門嚢の分泌液除去',         5, true, 3)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 城東センター病院 diagnosis_types（診断カテゴリ: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO diagnosis_types (id, clinic_id, name, is_active, description, sort_order) VALUES
    (9, 2, '消化器系',   true, '胃腸・肝臓・膵臓などの消化器系疾患', 1),
    (10, 2, '呼吸器系',   true, '肺・気管・鼻腔などの呼吸器系疾患',   2),
    (11, 2, '皮膚・被毛', true, 'アレルギー・感染症などの皮膚疾患',   3),
    (12, 2, '泌尿器系',   true, '腎臓・膀胱・尿道などの泌尿器系疾患', 4),
    (13, 2, '感染症',     true, '細菌・ウイルス感染症',               5)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 城東センター病院 diagnosis_names（診断名: 10件）
-- -----------------------------------------------------------------------------
INSERT INTO diagnosis_names (id, clinic_id, name, is_active, description, diagnosis_type_id, sort_order) VALUES
    (21, 2, '胃腸炎',             true, '胃・腸の炎症（嘔吐・下痢）',         9,  1),
    (22, 2, '膵炎',               true, '膵臓の炎症',                         9,  2),
    (23, 2, '気管支炎',           true, '気管支の炎症',                       10, 1),
    (24, 2, '肺炎',               true, '肺の感染性・非感染性炎症',           10, 2),
    (25, 2, 'アトピー性皮膚炎',   true, 'アレルゲンによるアレルギー性皮膚炎', 11, 1),
    (26, 2, '膿皮症',             true, '細菌性の皮膚感染症',                 11, 2),
    (27, 2, '膀胱炎',             true, '細菌性・特発性膀胱炎',               12, 1),
    (28, 2, '腎不全',             true, '急性・慢性腎不全',                   12, 2),
    (29, 2, 'パルボウイルス感染症', true, '犬パルボウイルスによる感染症',       13, 1),
    (30, 2, '猫風邪（FVR）',      true, '猫ウイルス性鼻気管炎',               13, 2)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 城東センター病院 checkup_types（健診種別: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO checkup_types (id, clinic_id, name, price, is_active, description, interval, target_age, sort_order) VALUES
    (5, 2, '一般健診',       5200,  true, '身体検査・体重測定・問診',                     '1年',   '全年齢',  1),
    (6, 2, '老齢検診',       16000, true, '身体検査＋血液検査＋レントゲン＋超音波',         '6ヶ月', '7歳以上', 2),
    (7, 2, 'フィラリア検査', 2600,  true, 'フィラリア抗原検査（予防シーズン前）',           '1年',   '成犬',    3)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 城東センター病院 chief_complaint_types（主訴区分: 4件）
-- -----------------------------------------------------------------------------
INSERT INTO chief_complaint_types (id, clinic_id, name, is_active, sort_order) VALUES
    (7, 2, '食欲不振',       true, 1),
    (8, 2, '嘔吐・下痢',     true, 2),
    (9, 2, '皮膚・被毛異常', true, 3),
    (10, 2, '排尿・排泄異常', true, 4)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 城東センター病院 merchandise_items（物販: 4件）
-- -----------------------------------------------------------------------------
INSERT INTO merchandise_items (id, clinic_id, name, category, unit_price, tax_rate, sort_order) VALUES
    (8, 2, 'ロイヤルカナン 消化器サポート 1kg', 'food',  2900, 0.10, 1),
    (9, 2, 'ヒルズ k/d 2kg',                   'food',  3600, 0.10, 2),
    (10, 2, 'ペット用歯ブラシセット',            'goods', 1250, 0.10, 3),
    (11, 2, '文書料',                            'other', 3000, 0.10, 4)
ON CONFLICT DO NOTHING;


-- =============================================================================
-- 城東センター病院 (clinic_id=2) デモデータ（飼主・ペット）
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 城東センター病院 owners（飼主: 8件、ID 23〜30）
-- -----------------------------------------------------------------------------
INSERT INTO owners (id, clinic_id, name, name_kana, birth_date, company, postal_code, address1, address2, phone, company_phone, email, remarks, is_dangerous, discount_rate, membership_type) VALUES
    (23, 2, '大野 健司',   'おおの けんじ',   '1979-06-10', '',           '136-0071', '東京都江東区亀戸3-5-8',         '', '090-6601-2233', '', 'kenji.ono@example.com',      '定期通院',   false, 10, 'member'),
    (24, 2, '松田 香織',   'まつだ かおり',   '1988-02-14', '',           '135-0044', '東京都江東区越中島2-1-4',       '', '080-7702-3344', '', 'kaori.matsuda@example.com',  '',           false, 0,  'non_member'),
    (25, 2, '渡辺 直樹',   'わたなべ なおき', '1972-10-28', '渡辺商事',   '132-0025', '東京都江戸川区松江3-7-2',       '', '090-8803-4455', '', 'naoki.watanabe@example.com', '',           false, 5,  'member'),
    (26, 2, '中島 奈緒',   'なかじま なお',   '1994-04-17', '',           '133-0065', '東京都江戸川区南篠崎町1-6-3',   '', '080-9904-5566', '', 'nao.nakajima@example.com',   '',           false, 0,  'non_member'),
    (27, 2, '斎藤 浩二',   'さいとう こうじ', '1967-12-05', '',           '131-0031', '東京都墨田区墨田1-4-9',         '', '090-1105-6677', '', 'koji.saito@example.com',     '',           false, 0,  'non_member'),
    (28, 2, '坂本 真由美', 'さかもと まゆみ', '1991-08-22', '',           '130-0022', '東京都墨田区横川4-2-7',         '', '080-2206-7788', '', 'mayumi.sakamoto@example.com','猫アレルギーに注意', false, 0,  'member'),
    (29, 2, '岡田 俊雄',   'おかだ としお',   '1983-03-30', 'オカダ工業', '132-0034', '東京都江戸川区小松川1-8-5',     '', '090-3307-8899', '', 'toshio.okada@example.com',   '',           false, 0,  'non_member'),
    (30, 2, '藤田 彩',     'ふじた あや',     '1997-11-11', '',           '136-0076', '東京都江東区南砂3-9-1',         '', '080-4408-9900', '', 'aya.fujita@example.com',     '',           false, 0,  'non_member')
ON CONFLICT (id) DO UPDATE SET
    name      = EXCLUDED.name,
    name_kana = EXCLUDED.name_kana,
    updated_at      = now();


-- -----------------------------------------------------------------------------
-- 城東センター病院 pets（ペット: 10件、ID 29〜38）
-- -----------------------------------------------------------------------------
INSERT INTO pets (id, clinic_id, owner_id, pet_number, name, name_kana, animal_species_id, gender, status, birth_date, breed, color, weight, insurance_id, last_visit) VALUES
    (29, 2, 23, '23-1', 'クロ',   'くろ',   1, 'male',   'alive', '2019-03-20', 'ラブラドール',             'ブラック',   28.0,  6,    '2025-11-10'),
    (30, 2, 23, '23-2', 'シナモン', 'しなもん', 2, 'female','alive', '2021-07-15', 'アビシニアン',           'レッド',      4.1,  NULL, NULL),
    (31, 2, 24, '24-1', 'ポポ',   'ぽぽ',   2, 'female', 'alive', '2020-05-10', 'ロシアンブルー',           'グレー',      3.8,  7,    '2025-10-20'),
    (32, 2, 25, '25-1', 'ダン',   'だん',   1, 'male',   'alive', '2017-09-05', 'ウェルシュコーギー',       'セーブルホワイト', 13.2, NULL, NULL),
    (33, 2, 26, '26-1', 'キナ',   'きな',   2, 'male',   'alive', '2022-02-28', 'ミックス猫',               'オレンジ',    4.5,  NULL, NULL),
    (34, 2, 27, '27-1', 'バロン', 'ばろん', 1, 'male',   'alive', '2015-11-18', 'シェパード',               'ブラックタン',30.5, NULL, '2025-12-01'),
    (35, 2, 28, '28-1', 'ユズ',   'ゆず',   2, 'female', 'alive', '2023-03-03', 'ミックス猫',               'キャリコ',    3.0,  NULL, NULL),
    (36, 2, 28, '28-2', 'レン',   'れん',   1, 'male',   'alive', '2021-06-14', 'ビーグル',                 'トライカラー',12.8, NULL, NULL),
    (37, 2, 29, '29-1', 'ナナ',   'なな',   2, 'female', 'alive', '2018-08-07', 'メインクーン',             'ブラウンタビー', 6.2, 6,   '2026-01-15'),
    (38, 2, 30, '30-1', 'コウ',   'こう',   1, 'male',   'alive', '2020-12-25', 'トイプードル',             'アプリコット', 3.5, NULL, NULL)
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();


-- =============================================================================
-- 敷島医院 (clinic_id=3) マスタデータ
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 敷島医院 reservation_types（サービス種別: 14件、ID 45-58）
-- -----------------------------------------------------------------------------

-- 敷島医院 公開コース
INSERT INTO reservation_types (id, clinic_id, name, short_name, is_active, description, color, sort_order, duration_minutes, reservation_visible, reservation_comment, is_internal) VALUES
    (45, 3, '一般診察',               '診察',     true, '内科・外科・皮膚科などの一般的な診察',         '#3B82F6', 1,  15, true,  '', false),
    (46, 3, '一般診察(再診)',          '再診',     true, '継続通院の一般診察',                           '#3B82F6', 2,  15, true,  '', false),
    (47, 3, 'ワクチン接種',            'ワクチン', true, '各種ワクチン接種（予防接種）',                 '#10B981', 3,  15, true,  '', false),
    (48, 3, '狂犬病',                 '狂犬病',   true, '狂犬病予防法に基づくワクチン接種',             '#10B981', 4,  15, true,  '', false),
    (49, 3, 'フィラリア予防',          'フィラリア', true, 'フィラリア予防薬投与・処方',                '#10B981', 5,  15, true,  '', false),
    (50, 3, '健康診断',               '健診',     true, '定期健康診断・フィラリア検査など',             '#8B5CF6', 6,  15, true,  '', false),
    (51, 3, '健康診断結果報告',        '結果報告', true, '健康診断結果の説明・報告',                     '#8B5CF6', 7,  15, true,  '', false),
    (52, 3, 'トリミングコース',        'トリミング', true, 'カット・シャンプー・ブロー・爪切り・耳掃除', '#F59E0B', 8,  15, true,  '', false)
ON CONFLICT DO NOTHING;

-- 敷島医院 スタッフ専用コース
INSERT INTO reservation_types (id, clinic_id, name, short_name, is_active, description, color, sort_order, duration_minutes, reservation_visible, reservation_comment, is_internal) VALUES
    (53, 3, '手術60',                 '手術60',   true, '手術枠（60分）',                               '#EF4444', 9,  60, false, '', true),
    (54, 3, '手術30',                 '手術30',   true, '手術枠（30分）',                               '#EF4444', 10, 30, false, '', true),
    (55, 3, '休憩枠',                 '休憩',     true, '休憩・ブロック枠',                             '#6B7280', 11, 60, false, '', true),
    (56, 3, '×',                     '×',       true, '予約不可（15分）',                             '#6B7280', 12, 15, false, '', true),
    (57, 3, '予約不可60',             '不可60',   true, '予約不可ブロック（60分）',                     '#6B7280', 13, 60, false, '', true),
    (58, 3, '予約不可30',             '不可30',   true, '予約不可ブロック（30分）',                     '#6B7280', 14, 30, false, '', true)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('reservation_types', 'id'), (SELECT MAX(id) FROM reservation_types));

-- 敷島医院 (clinic_id=3) グループ紐付け
UPDATE reservation_types SET group_id=15 WHERE clinic_id=3 AND id IN (45,46);
UPDATE reservation_types SET group_id=16 WHERE clinic_id=3 AND id IN (47,48,49);
UPDATE reservation_types SET group_id=17 WHERE clinic_id=3 AND id IN (50,51);
UPDATE reservation_types SET group_id=18 WHERE clinic_id=3 AND id IN (52);
UPDATE reservation_types SET group_id=19 WHERE clinic_id=3 AND id IN (53,54);
UPDATE reservation_types SET group_id=20, is_internal=true WHERE clinic_id=3 AND id IN (55,56,57,58);

-- -----------------------------------------------------------------------------
-- 敷島医院 cages（ケージ: 4件）
-- -----------------------------------------------------------------------------
INSERT INTO cages (id, clinic_id, name, price, is_active, description, cage_type, cage_size, sort_order) VALUES
    (13, 3, 'ICUケージ',       8000, true, '酸素吸入可・重症患者用',     'icu',     'medium', 1),
    (14, 3, '犬用ケージ（小）', 2900, true, '小型犬・ホテル利用可',       'dog',     'small',  2),
    (15, 3, '犬用ケージ（大）', 4200, true, '大型犬・術後管理用',         'dog',     'large',  3),
    (16, 3, '猫用ケージ',       3100, true, '猫専用・ストレス軽減設計',   'cat',     'medium', 4)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 敷島医院 insurances（保険: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO insurances (id, clinic_id, name, is_active, description, coverage_rate, contact_phone, sort_order) VALUES
    (9, 3, 'アニコム損保',         true, 'ペット保険大手・どうぶつ健保シリーズ', 70, '0120-025-034', 1),
    (10, 3, 'ペット&ファミリー',     true, 'げんきナンバーワンシリーズ',           80, '0120-81-8505', 2),
    (11, 3, 'その他（自費）',       true, '保険未加入・全額自費',                100, '',             3)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('insurances', 'id'), (SELECT MAX(id) FROM insurances));

-- -----------------------------------------------------------------------------
-- 敷島医院 exam_types（検査種別: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO exam_types (id, clinic_id, name, price, is_active, description, sort_order) VALUES
    (9, 3, '血液検査（CBC）', 2800, true, '全血球計算（Complete Blood Count）', 1),
    (10, 3, '尿検査',           1600, true, '尿試験紙・尿沈渣検査',               2),
    (11, 3, '超音波検査（エコー）', 5200, true, '腹部エコー・心臓エコー',         3)
ON CONFLICT DO NOTHING;


-- exam_type_fields for clinic 5
INSERT INTO exam_type_fields (id, exam_type_id, name, inspection_value, normal_value, sort_order) VALUES
    (26, 9,  'WBC（白血球数）',        '', '6.0-17.0 x10^3/uL', 1),
    (27, 9,  'RBC（赤血球数）',        '', '5.5-8.5 x10^6/uL',  2),
    (28, 9,  'PLT（血小板数）',        '', '175-500 x10^3/uL',  3),
    (29, 10, '尿比重',                 '', '1.015-1.045',        1),
    (30, 10, '尿pH',                   '', '5.5-7.5',            2),
    (31, 10, '尿タンパク',             '', '陰性',               3),
    (32, 11, '腹部エコー',             '', '異常なし',           1),
    (33, 11, '心臓エコー',             '', '異常なし',           2)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 敷島医院 vaccines（ワクチン: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO vaccines (id, clinic_id, name, price, is_active, description, species, interval, sort_order) VALUES
    (16, 3, '混合ワクチン5種（犬）',    4600, true, 'ジステンパー・パルボ・アデノ1型・アデノ2型・パラインフルエンザ', 'dog', '1年',   1),
    (17, 3, '混合ワクチン10種（犬）',   8200, true, '5種＋レプトスピラ5種',                                          'dog', '1年',   2),
    (18, 3, '混合ワクチン3種（猫）',    4100, true, '猫ウイルス性鼻気管炎・カリシウイルス・汎白血球減少症',           'cat', '1年',   3),
    (19, 3, '混合ワクチン5種（猫）',    5600, true, '3種＋猫白血病・猫クラミジア',                                   'cat', '1年',   4),
    (20, 3, '狂犬病ワクチン',           3000, true, '狂犬病予防法に基づく接種',                                      'dog', '1年',   5)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 敷島医院 medicines（薬剤: カテゴリ4件 + 薬剤10件）
-- -----------------------------------------------------------------------------
INSERT INTO medicines (id, clinic_id, name, price, is_active, description, sort_order) VALUES
    (3001, 3, '抗生剤',     NULL, true, '抗生物質カテゴリ',   1),
    (3002, 3, 'ステロイド', NULL, true, 'ステロイド剤カテゴリ', 2),
    (3003, 3, '制吐剤',     NULL, true, '制吐剤カテゴリ',     3),
    (3004, 3, '消化器用薬', NULL, true, '消化器用薬カテゴリ', 4)
ON CONFLICT DO NOTHING;

INSERT INTO medicines (id, clinic_id, name, price, is_active, description, dosage_form, medicine_unit, default_quantity, sort_order, parent_id) VALUES
    (201, 3, 'アモキシシリン 50mg',         510,  true, '広域スペクトラム抗生物質',              'tablet',    'per_tablet', 1,    1, 3001),
    (202, 3, 'エンロフロキサシン 15mg',      580,  true, 'フルオロキノロン系抗生物質',            'tablet',    'per_tablet', 1,    2, 3001),
    (203, 3, 'メトロニダゾール 250mg',       610,  true, '嫌気性菌・原虫感染症治療薬',            'tablet',    'per_tablet', 1,    3, 3001),
    (204, 3, 'プレドニゾロン 5mg',           410,  true, 'ステロイド系抗炎症・免疫抑制剤',        'tablet',    'per_tablet', 1,    4, 3002),
    (205, 3, 'デキサメタゾン注射液',         710,  true, '強力ステロイド・アレルギー緊急治療',    'injection', 'per_ml',     1,    5, 3002),
    (206, 3, 'マロピタント 16mg',            820,  true, '制吐剤（乗り物酔い・嘔吐治療）',        'tablet',    'per_tablet', 1,    6, 3003),
    (207, 3, 'メトクロプラミド注',           480,  true, '制吐・消化管運動促進',                  'injection', 'per_ml',     1,    7, 3003),
    (208, 3, 'ラクツロース液',               510,  true, '便秘・肝性脳症の治療',                  'liquid',    'per_ml',     5,    8, 3004),
    (209, 3, 'オメプラゾール 10mg',          360,  true, 'プロトンポンプ阻害薬（胃酸抑制）',      'tablet',    'per_tablet', 1,    9, 3004),
    (210, 3, '生理食塩水 500ml',             410,  true, '点滴・洗浄用生理食塩水',                'liquid',    'per_ml',     500,  10, NULL)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 敷島医院 consultations（診察項目: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO consultations (id, clinic_id, name, price, is_active, description, time_condition, duration, sort_order) VALUES
    (9, 3, '初診料',       1900, true, '初めての受診または6ヶ月以上受診がない場合', 'first_visit', 30, 1),
    (10, 3, '再診料',        800, true, '継続通院の診察料',                         'revisit',     15, 2),
    (11, 3, '往診料',       5500, true, '自宅への往診料（基本料金）',               'anytime',     60, 3)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 敷島医院 procedures（処置項目: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO procedures (id, clinic_id, name, price, is_active, description, duration, anesthesia, sort_order) VALUES
    (16, 3, '去勢手術（犬）', 24000, true, '雄犬の去勢手術',                  60, 'general', 1),
    (17, 3, '避妊手術（猫）', 24000, true, '雌猫の避妊手術',                  90, 'general', 2),
    (18, 3, '歯石除去',       15000, true, '全身麻酔下での歯石除去・歯周治療', 45, 'general', 3),
    (19, 3, '爪切り',            500, true, '爪のカット・やすりがけ',           10, 'none',    4),
    (20, 3, '腫瘍摘出',        22000, true, '皮膚腫瘍の外科的摘出',             60, 'local',   5)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 敷島医院 hospitalization_plans（入院プラン: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO hospitalization_plans (id, clinic_id, name, price, is_active, description, body_size, billing_unit, sort_order) VALUES
    (9, 3, '一般入院（小型）', 3100, true, '体重10kg以下の入院管理料（1日）',  'small',  'per_day',   1),
    (10, 3, '一般入院（大型）', 4800, true, '体重25kg以上の入院管理料（1日）',  'large',  'per_day',   2),
    (11, 3, 'ホテル（小型）',   2600, true, '体重10kg以下のペットホテル（1泊）', 'small',  'per_night', 3)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 敷島医院 trimming_courses（トリミングコース: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO trimming_courses (id, clinic_id, name, price, is_active, description, target_size, duration, sort_order) VALUES
    (9, 3, 'シャンプー&ブロー（小型）', 3900, true, 'シャンプー・ブロー・ブラッシング',            'small',  60,  1),
    (10, 3, 'シャンプー&ブロー（中型）', 5400, true, 'シャンプー・ブロー・ブラッシング',            'medium', 90,  2),
    (11, 3, 'フルコース（小型）',        6800, true, 'カット・シャンプー・ブロー・爪切り・耳掃除', 'small',  120, 3)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 敷島医院 trimming_options（トリミングオプション: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO trimming_options (id, clinic_id, name, price, is_active, description, duration, is_combinable, sort_order) VALUES
    (9, 3, '爪切り',   300, true, '爪のカット・やすりがけ',    10, true, 1),
    (10, 3, '耳掃除',   500, true, '外耳道の洗浄・清掃',        10, true, 2),
    (11, 3, '歯磨き',   500, true, '歯ブラシによるデンタルケア', 15, true, 3)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 敷島医院 diagnosis_types（診断カテゴリ: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO diagnosis_types (id, clinic_id, name, is_active, description, sort_order) VALUES
    (14, 3, '消化器系',       true, '胃腸・肝臓・膵臓などの消化器系疾患', 1),
    (15, 3, '皮膚・被毛',     true, 'アレルギー・感染症などの皮膚疾患',   2),
    (16, 3, '泌尿器系',       true, '腎臓・膀胱・尿道などの泌尿器系疾患', 3),
    (17, 3, '腫瘍',           true, '良性・悪性腫瘍（がん）',             4),
    (18, 3, '外傷・骨格',     true, '骨折・咬傷・関節疾患など',           5)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 敷島医院 diagnosis_names（診断名: 10件）
-- -----------------------------------------------------------------------------
INSERT INTO diagnosis_names (id, clinic_id, name, is_active, description, diagnosis_type_id, sort_order) VALUES
    (31, 3, '胃腸炎',             true, '胃・腸の炎症（嘔吐・下痢）',         14, 1),
    (32, 3, '肝疾患',             true, '肝炎・肝不全・脂肪肝など',           14, 2),
    (33, 3, 'アトピー性皮膚炎',   true, 'アレルゲンによるアレルギー性皮膚炎', 15, 1),
    (34, 3, '膿皮症',             true, '細菌性の皮膚感染症',                 15, 2),
    (35, 3, '真菌症',             true, '皮膚糸状菌による感染症',             15, 3),
    (36, 3, '膀胱炎',             true, '細菌性・特発性膀胱炎',               16, 1),
    (37, 3, '尿路結石',           true, '腎結石・膀胱結石・尿道結石',         16, 2),
    (38, 3, '肥満細胞腫',         true, '皮膚または内臓の肥満細胞腫瘍',       17, 1),
    (39, 3, '骨折',               true, '各部位の骨折',                       18, 1),
    (40, 3, '咬傷',               true, '他動物による咬傷・咬傷感染',         18, 2)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 敷島医院 checkup_types（健診種別: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO checkup_types (id, clinic_id, name, price, is_active, description, interval, target_age, sort_order) VALUES
    (8, 3, '一般健診',       4800,  true, '身体検査・体重測定・問診',                   '1年',   '全年齢',  1),
    (9, 3, '老齢検診',       14500, true, '身体検査＋血液検査＋レントゲン＋超音波',     '6ヶ月', '7歳以上', 2),
    (10, 3, '歯科検診',        3100, true, '歯周病チェック・歯石付着度の確認',           '1年',   '成犬',    3)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 敷島医院 chief_complaint_types（主訴区分: 4件）
-- -----------------------------------------------------------------------------
INSERT INTO chief_complaint_types (id, clinic_id, name, is_active, sort_order) VALUES
    (11, 3, '食欲不振',     true, 1),
    (12, 3, '嘔吐・下痢',   true, 2),
    (13, 3, '呼吸困難',     true, 3),
    (14, 3, '外傷・骨折',   true, 4)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- 敷島医院 merchandise_items（物販: 4件）
-- -----------------------------------------------------------------------------
INSERT INTO merchandise_items (id, clinic_id, name, category, unit_price, tax_rate, sort_order) VALUES
    (12, 3, 'ロイヤルカナン 消化器サポート 1kg', 'food',  2750, 0.10, 1),
    (13, 3, 'ノミ・ダニ予防首輪',               'goods', 1600, 0.10, 2),
    (14, 3, 'エリザベスカラー（M）',             'goods', 1000, 0.10, 3),
    (15, 3, '文書料',                            'other', 3000, 0.10, 4)
ON CONFLICT DO NOTHING;


-- =============================================================================
-- 敷島医院 (clinic_id=3) デモデータ（飼主・ペット）
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 敷島医院 owners（飼主: 8件、ID 31〜38）
-- -----------------------------------------------------------------------------
INSERT INTO owners (id, clinic_id, name, name_kana, birth_date, company, postal_code, address1, address2, phone, company_phone, email, remarks, is_dangerous, discount_rate, membership_type) VALUES
    (31, 3, '村上 俊平',   'むらかみ しゅんぺい', '1976-04-12', '',           '400-0031', '山梨県甲府市丸の内2-3-4',     '', '090-5501-1122', '', 'shunpei.murakami@example.com', '',           false, 10, 'member'),
    (32, 3, '長谷川 恵子', 'はせがわ けいこ',     '1989-09-07', '',           '400-0032', '山梨県甲府市中央3-5-6',       '', '080-6602-2233', '', 'keiko.hasegawa@example.com',   '',           false, 0,  'non_member'),
    (33, 3, '野口 正樹',   'のぐち まさき',       '1971-01-25', '野口設計',   '400-0035', '山梨県甲府市丸の内3-7-8',     '', '090-7703-3344', '', 'masaki.noguchi@example.com',   '',           false, 5,  'member'),
    (34, 3, '石田 沙織',   'いしだ さおり',       '1993-07-18', '',           '400-0801', '山梨県甲府市横根町5-2-1',     '', '080-8804-4455', '', 'saori.ishida@example.com',     '猫2匹飼い', false, 0,  'non_member'),
    (35, 3, '前田 修',     'まえだ おさむ',       '1968-11-02', '',           '400-0031', '山梨県甲府市丸の内1-9-3',     '', '090-9905-5566', '', 'osamu.maeda@example.com',      '',           false, 0,  'non_member'),
    (36, 3, '菊池 里奈',   'きくち りな',         '1995-06-14', '',           '400-0032', '山梨県甲府市中央5-4-7',       '', '080-1106-6677', '', 'rina.kikuchi@example.com',     '',           false, 0,  'member'),
    (37, 3, '清水 和彦',   'しみず かずひこ',     '1980-03-28', '清水工務店', '400-0034', '山梨県甲府市北口1-6-2',       '', '090-2207-7788', '', 'kazuhiko.shimizu@example.com', '',           false, 0,  'non_member'),
    (38, 3, '岩崎 美穂',   'いわさき みほ',       '1998-12-05', '',           '400-0803', '山梨県甲府市横根町2-8-9',     '', '080-3308-8899', '', 'miho.iwasaki@example.com',     '',           false, 0,  'non_member')
ON CONFLICT (id) DO UPDATE SET
    name      = EXCLUDED.name,
    name_kana = EXCLUDED.name_kana,
    updated_at      = now();

SELECT setval(pg_get_serial_sequence('owners', 'id'), (SELECT MAX(id) FROM owners));

-- -----------------------------------------------------------------------------
-- 敷島医院 pets（ペット: 10件、ID 39〜48）
-- -----------------------------------------------------------------------------
INSERT INTO pets (id, clinic_id, owner_id, pet_number, name, name_kana, animal_species_id, gender, status, birth_date, breed, color, weight, insurance_id, last_visit) VALUES
    (39, 3, 31, '31-1', 'フク',   'ふく',   1, 'male',   'alive', '2018-05-05', 'シベリアンハスキー',       'グレーホワイト', 24.0, 9,    '2025-10-05'),
    (40, 3, 32, '32-1', 'アズキ', 'あずき', 2, 'female', 'alive', '2021-11-20', 'ノルウェージャンフォレストキャット', 'ブラウンタビー', 4.8, NULL, '2025-12-18'),
    (41, 3, 33, '33-1', 'カイ',   'かい',   1, 'male',   'alive', '2016-07-10', 'ゴールデンレトリーバー',   'ゴールデン',     31.2, NULL, '2026-01-08'),
    (42, 3, 34, '34-1', 'キキ',   'きき',   2, 'female', 'alive', '2022-04-25', 'ミックス猫',               'トラ',           3.5,  NULL, NULL),
    (43, 3, 34, '34-2', 'ニコ',   'にこ',   2, 'female', 'alive', '2020-09-12', 'ミックス猫',               'サビ',           4.2,  NULL, NULL),
    (44, 3, 35, '35-1', 'ジャック', 'じゃっく', 1, 'male', 'alive', '2014-02-28', 'ジャックラッセルテリア', 'ホワイトタン',   6.8,  NULL, '2025-09-20'),
    (45, 3, 36, '36-1', 'ミル',   'みる',   2, 'female', 'alive', '2023-01-15', 'ミックス猫',               'ホワイト',       3.2,  NULL, NULL),
    (46, 3, 37, '37-1', 'ガイア', 'がいあ', 1, 'male',   'alive', '2017-08-30', 'アラスカンマラミュート',   'グレー',         36.5, NULL, '2025-11-25'),
    (47, 3, 38, '38-1', 'ハナ',   'はな',   2, 'female', 'alive', '2021-03-08', 'アメリカンショートヘア',   'シルバータビー', 4.0,  9,    '2026-02-10'),
    (48, 3, 38, '38-2', 'ソウ',   'そう',   1, 'male',   'alive', '2019-10-01', 'ミックス犬',               'ベージュ',       9.2,  NULL, NULL)
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('pets', 'id'), (SELECT MAX(id) FROM pets));

-- -----------------------------------------------------------------------------
-- ワクチン接種記録（八王子病院: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO vaccinations (id, clinic_id, medical_record_id, pet_id, vaccine_id, doctor_id, date, lot1, next_schedule_type, next_date, remarks) VALUES
    (1, 1, 1,  1,  1,  1, '2025-10-10', 'LOT-2025-A001', '1year',  '2026-10-10', '5種混合ワクチン接種。体調良好。'),
    (2, 1, 5,  3,  5,  1, '2025-09-15', 'LOT-2025-C001', '1year',  '2026-09-15', '3種混合ワクチン接種。'),
    (3, 1, 6,  3,  5,  2, '2026-01-06', 'LOT-2026-C001', '1year',  '2027-01-06', '3種混合ワクチン追加接種。'),
    (4, 1, 15, 13, 5,  1, '2026-01-06', 'LOT-2026-C003', '4weeks', '2026-02-03', '初回3種混合（猫）。4週後に2回目接種予定。'),
    (5, 1, 20, 18, 5,  1, '2026-01-06', 'LOT-2026-C002', '1year',  '2027-01-06', '3種混合ワクチン接種。体調良好。')
ON CONFLICT (id) DO NOTHING;


-- -----------------------------------------------------------------------------
-- 検査記録（八王子病院: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO exams (id, clinic_id, medical_record_id, exam_type_id, doctor_id, date, result_summary, status) VALUES
    (1, 1, 2,  1, 1, '2025-12-15', 'CBC全項目正常範囲内。', 'completed'),
    (2, 1, 14, 2, 2, '2025-06-15', 'ALT軽度上昇（145 U/L）。他正常。', 'completed'),
    (3, 1, 13, 1, 1, '2026-02-10', 'WBC上昇（19.2）。脱水を反映。', 'result_entered')
ON CONFLICT (id) DO NOTHING;


INSERT INTO exam_results (id, exam_id, exam_type_field_id, inspection_value, status) VALUES
    -- Exam 1 (CBC for MR2)
    (34, 1, 1, '9.8',  'normal'),
    (35, 1, 2, '7.2',  'normal'),
    (36, 1, 3, '45',   'normal'),
    (37, 1, 4, '320',  'normal'),
    -- Exam 2 (Chemistry for MR14)
    (38, 2, 5, '145',  'high'),
    (39, 2, 6, '18',   'normal'),
    (40, 2, 7, '1.2',  'normal'),
    (41, 2, 8, '98',   'normal'),
    -- Exam 3 (CBC for MR13)
    (42, 1, 1, '19.2', 'high'),
    (43, 1, 2, '8.1',  'normal'),
    (44, 1, 3, '52',   'normal'),
    (45, 1, 4, '280',  'normal')
ON CONFLICT (id) DO NOTHING;


-- -----------------------------------------------------------------------------
-- 城東センター病院 inventory_items（在庫管理: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO inventory_items (id, clinic_id, name, category, quantity, unit, min_stock_level, location, supplier, status) VALUES
    (15, 2, '5種混合ワクチン',               'medicine',   20,  'バイアル', 10, '冷蔵庫',     '共立製薬',               'sufficient'),
    (16, 2, 'アモキシシリン 50mg',           'medicine',   100, '錠',       30, '薬品棚A',    '日本全薬工業',           'sufficient'),
    (17, 2, 'ロイヤルカナン 消化器サポート', 'food',       12,  '袋',        5, 'フード棚',   'ロイヤルカナン',         'sufficient'),
    (18, 2, '包帯・ガーゼセット',            'consumable', 40,  'セット',   20, '処置室',     '白十字',                 'sufficient'),
    (19, 2, 'ノミダニ駆除薬 スポット（犬用）','medicine',  25,  'ピペット',  10, '薬品棚B',   'エランコジャパン',       'sufficient')
ON CONFLICT (id) DO NOTHING;

-- -----------------------------------------------------------------------------
-- 敷島医院 inventory_items（在庫管理: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO inventory_items (id, clinic_id, name, category, quantity, unit, min_stock_level, location, supplier, status) VALUES
    (20, 3, '5種混合ワクチン（犬）',         'medicine',   15,  'バイアル', 10, '冷蔵庫',     '共立製薬',               'sufficient'),
    (21, 3, 'アモキシシリン 50mg',           'medicine',   80,  '錠',       30, '薬品棚A',    '日本全薬工業',           'sufficient'),
    (22, 3, 'ヒルズ k/d',                   'food',        8,   '袋',        5, 'フード棚',   'ヒルズ・コルゲート',     'sufficient'),
    (23, 3, 'シリンジ 5mL',                 'consumable', 200,  '本',       50, '処置室',     'テルモ',                 'sufficient'),
    (24, 3, '生理食塩水 500ml',             'medicine',    30,  '本',       10, '薬品棚B',    '大塚製薬工場',           'sufficient')
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('inventory_items', 'id'), (SELECT MAX(id) FROM inventory_items));

-- -----------------------------------------------------------------------------
-- 14. audit_logs（監査ログ: 8件）
-- -----------------------------------------------------------------------------
INSERT INTO audit_logs (clinic_id, actor_id, actor_type, action, resource, resource_id, old_value, new_value, ip_address, user_agent) VALUES
    (3, 10, 'staff', 'permission_rules.update', 'permission_groups', 1, '{"can_delete": false}', '{"can_delete": true}', '192.168.1.1', 'Mozilla/5.0...'),
    (3, 10, 'staff', 'auth.login.success', 'user_accounts', 10, NULL, NULL, '192.168.1.1', 'Mozilla/5.0...'),
    (3, 11, 'staff', 'owner.update', 'owners', 1, '{"phone": "old"}', '{"phone": "090-1234-5678"}', '192.168.1.2', 'Mozilla/5.0...'),
    (3, 10, 'staff', 'pet.create', 'pets', 28, NULL, '{"name": "マル"}', '192.168.1.1', 'Mozilla/5.0...'),
    (3, 10, 'staff', 'medical_record.finalize', 'medical_records', 1, '{"status": "draft"}', '{"status": "finalized"}', '192.168.1.1', 'Mozilla/5.0...'),
    (3, 11, 'staff', 'auth.login.failure', 'user_accounts', NULL, '{"email": "exec@example.com"}', NULL, '192.168.1.5', 'PostmanRuntime/7.26.8'),
    (3, 10, 'staff', 'inventory.decrease', 'inventory_items', 1, '{"quantity": 50}', '{"quantity": 43}', '192.168.1.10', 'Mozilla/5.0...'),
    (3, 10, 'staff', 'treatment.create', 'treatments', 2, NULL, '{"content": "アモキシシリン"}', '192.168.1.10', 'Mozilla/5.0...')
ON CONFLICT DO NOTHING;

-- =============================================================================
-- LINE予約システム シードデータ
-- =============================================================================

-- A. reservation_types の予約用カラムはすべて INSERT 時に設定済み（更新不要）
-- B. staffs の予約用カラムはすべて INSERT 時に設定済み（更新不要）

-- C. line_reservation_settings 初期データ
INSERT INTO line_reservation_settings (
    clinic_id, status, business_hours, break_hours, daily_limit,
    booking_window_max_days, booking_window_min_days, calendar_months,
    phone_number, time_slot_mode, time_slot_interval_minutes,
    no_staff_mode, show_no_staff_option, national_holiday_closed, additional_fields
)
SELECT
    c.id, 'stopped',
    '{"start":"0900","end":"1900"}',
    '[{"start":"1200","end":"1300"}]',
    1, 30, 2, 2, '0552334126', 'minimize_gaps', 15,
    'first_available', true, false,
    '[
        {"key":"phone","label":"電話番号","required":true,"placeholder":"例) 090-1234-5678"},
        {"key":"owner_name","label":"飼い主名","required":true,"placeholder":""},
        {"key":"pet_info","label":"ペットの名前と種類","required":true,"placeholder":"例) ポチ（柴犬）"},
        {"key":"symptoms","label":"診察内容","required":false,"placeholder":""}
    ]'
FROM clinics c
ON CONFLICT (clinic_id) DO NOTHING;

-- C-2. テスト環境用 LINE クレデンシャル（docs/LINE-SETUP.md 参照）
-- clinic_id=1: テスト-八王子（@642hdxoh）
-- clinic_id=2: テスト-城東（@151lnsqa）
UPDATE line_reservation_settings SET
    status              = 'running',
    header_text         = 'ノア動物病院 八王子',
    line_channel_id     = '2009755544',
    line_channel_secret = '5344ef84eb7072b5894f7e087db28827',
    liff_id             = '2009755581-w5NOA3EW',
    line_access_token   = 'pwMi3yP6jhRa0xbmnR0IPEcE5l+OIp21a7ia3hmoiuFSCvqkR5Tmmfm6fLoSTB1Bt7uQjAe9NN7fZ+LBDtNKLGnrqBrjDmhTnws9PVxQKLyinomNzUAb61KADX7NJmFBfEsLQQ9VmlU+tMJcWh+zswdB04t89/1O/w1cDnyilFU='
WHERE clinic_id = 1;

UPDATE line_reservation_settings SET
    status              = 'running',
    header_text         = 'ノア動物病院 城東',
    line_channel_id     = '2009755545',
    line_channel_secret = '25e4661a8de553953a4b34c6ac7a91cb',
    liff_id             = '2009755586-nvKfG3Cp',
    line_access_token   = 'CUAtYMry8doD9ALCF/6Y0JocVqRrxC8IbOCMyRyxwDw5EJhyujJ4lQ8mVGrt7WawTi+voAxZ79mKAg+4qlsUPBVU6VMZdk7wEA42NZXQAl8gBr2tSYmZpbRzAiLfuGhxuba+koBHVk8yTuaCCjLBzAdB04t89/1O/w1cDnyilFU='
WHERE clinic_id = 2;

-- D. staff_reservation_exclusions 初期データ
-- 獣医スタッフはトリミング系コースを非対応とする（同一クリニック内のみ）
INSERT INTO staff_reservation_exclusions (staff_id, reservation_type_id)
SELECT s.id, st.id
FROM staffs s
JOIN staff_clinic_assignments sca ON sca.staff_id = s.id
JOIN reservation_types st ON sca.clinic_id = st.clinic_id
WHERE s.staff_type = 'doctor'
  AND st.name IN (
      'トリミングコース', 'トリミング部分カットコース', 'トリミングシャンプーコース',
      'クイックシャンプー', 'お手入れ', '室内ドッグラン'
  )
ON CONFLICT (staff_id, reservation_type_id) DO NOTHING;

-- E. line_customers テスト用 LINE 顧客データ
INSERT INTO line_customers (clinic_id, line_user_id, display_name, real_name, additional_fields, owner_id)
VALUES
    (1, 'U_test_hachioji_001', 'テスト 太郎', '執行 太郎', '{"phone":"090-1234-5678","owner_name":"執行 太郎","pet_name":"ポチ","pet_type":"柴犬"}', 1),
    (1, 'U_test_hachioji_002', 'テスト 花子', '一般 花子', '{"phone":"080-9876-5432","owner_name":"一般 花子","pet_name":"ミケ","pet_type":"三毛猫"}', 2),
    (2, 'U_test_joto_001', 'テスト 次郎', '城東テスト', '{"phone":"070-1111-2222","owner_name":"城東テスト","pet_name":"チョコ","pet_type":"トイプードル"}', NULL)
ON CONFLICT DO NOTHING;

-- F. 城東センター病院 (clinic_id=2) テスト予約 3件
-- ※ owners/pets が先に挿入されている必要がある
INSERT INTO appointments (id, clinic_id, start_time, end_time, owner_id, pet_id, visit_type, reservation_type_id, doctor_id, is_designated, status, notes) VALUES
    (11, 2, '2026-03-12 10:00:00+09', '2026-03-12 10:15:00+09', 23, 29, 'revisit', 26, 16, true,  'confirmed', 'クロの定期診察'),
    (12, 2, '2026-03-13 14:00:00+09', '2026-03-13 14:15:00+09', 24, 31, 'first',   28, 16, false, 'confirmed', 'ポポのワクチン接種'),
    (13, 2, '2026-03-14 11:00:00+09', '2026-03-14 11:15:00+09', 25, 32, 'revisit', 31, 16, false, 'confirmed', 'ダンの健康診断')
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('appointments', 'id'), (SELECT MAX(id) FROM appointments));

-- -----------------------------------------------------------------------------
-- G. shift_templates（シフトテンプレートマスタ）
--    各クリニック共通の標準テンプレート 5種
-- -----------------------------------------------------------------------------
INSERT INTO shift_templates (id, clinic_id, name, shift_type, start_time, end_time, notes, sort_order, is_active) VALUES
    -- 八王子病院 (clinic_id=1)
    (1, 1, '通常勤務',   'full',       '09:00', '18:00', '', 1, true),
    (2, 1, '午前勤務',   'morning',    '09:00', '13:00', '', 2, true),
    (3, 1, '午後勤務',   'afternoon',  '13:00', '18:00', '', 3, true),
    (4, 1, '休日',       'off',        NULL,    NULL,    '', 4, true),
    (5, 1, '有給休暇',   'paid_leave', NULL,    NULL,    '', 5, true),
    -- 城東センター病院 (clinic_id=2)
    (6, 2, 'Regular',   'full',       '09:00', '18:00', '', 1, true),
    (7, 2, '午前勤務',   'morning',    '09:00', '13:00', '', 2, true),
    (8, 2, '午後勤務',   'afternoon',  '13:00', '18:00', '', 3, true),
    (9, 2, '休日',       'off',        NULL,    NULL,    '', 4, true),
    (10, 2, '有給休暇',   'paid_leave', NULL,    NULL,    '', 5, true),
    -- 敷島医院 (clinic_id=3)
    (11, 3, '通常勤務',   'full',       '09:00', '18:00', '', 1, true),
    (12, 3, '午前勤務',   'morning',    '09:00', '13:00', '', 2, true),
    (13, 3, '午後勤務',   'afternoon',  '13:00', '18:00', '', 3, true),
    (14, 3, '休日',       'off',        NULL,    NULL,    '', 4, true),
    (15, 3, '有給休暇',   'paid_leave', NULL,    NULL,    '', 5, true)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('shift_templates', 'id'), (SELECT MAX(id) FROM shift_templates));

-- 通常勤務テンプレートの休憩時間（12:00〜13:00）
INSERT INTO shift_template_breaks (shift_template_id, break_start, break_end) VALUES
    (1,  '12:00', '13:00'),  -- 八王子病院: 通常勤務
    (6,  '12:00', '13:00'),  -- 城東センター病院: 通常勤務
    (11, '12:00', '13:00')   -- 敷島医院: 通常勤務
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- H. clinic_holidays（休診日マスタ — 2026年祝日）
-- -----------------------------------------------------------------------------
INSERT INTO clinic_holidays (clinic_id, date, reason)
SELECT c.id, h.date, h.reason
FROM (VALUES
    ('2026-04-29'::date, '昭和の日'),
    ('2026-05-03'::date, '憲法記念日'),
    ('2026-05-04'::date, 'みどりの日'),
    ('2026-05-05'::date, 'こどもの日'),
    ('2026-05-06'::date, '振替休日')
) h(date, reason)
CROSS JOIN (VALUES (1), (2), (3)) c(id)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- I. shift_entries（2026-04-13〜2026-06-13 の 2ヶ月分）
--    人間スタッフのみ（resource 系スタッフは除外）
--    平日: 通常勤務（full / 09:00-18:00）
--    土曜: 午前勤務（morning / 09:00-13:00）
--    日曜・祝日: 休日（off）
-- -----------------------------------------------------------------------------
WITH holidays AS (
    SELECT unnest(ARRAY[
        '2026-04-29'::date,
        '2026-05-03'::date,
        '2026-05-04'::date,
        '2026-05-05'::date,
        '2026-05-06'::date
    ]) AS hdate
),
staff_clinic(staff_id, clinic_id) AS (VALUES
    -- Hachioji Clinic (clinic_id=1): Human staff IDs 1-11, 33
    (1,  1), (2,  1), (3,  1), (4,  1), (5,  1), (6,  1),
    (7,  1), (8,  1), (9,  1), (10, 1), (11, 1), (33, 1),
    -- Joto Hospital (clinic_id=2): Human staff IDs 16-22, 34
    (16, 2), (17, 2), (18, 2), (19, 2), (20, 2), (21, 2), (22, 2), (34, 2),
    -- 敷島医院 (clinic_id=3): 人間スタッフ IDs 26-30, 35
    (26, 3), (27, 3), (28, 3), (29, 3), (30, 3), (35, 3)
)
INSERT INTO shift_entries (clinic_id, staff_id, date, shift_type, start_time, end_time, notes)
SELECT
    sc.clinic_id,
    sc.staff_id,
    d::date,
    CASE
        WHEN d::date IN (SELECT hdate FROM holidays)
          OR EXTRACT(DOW FROM d) = 0 THEN 'off'::shift_type
        WHEN EXTRACT(DOW FROM d) = 6   THEN 'morning'::shift_type
        ELSE                                'full'::shift_type
    END,
    CASE
        WHEN d::date IN (SELECT hdate FROM holidays)
          OR EXTRACT(DOW FROM d) = 0 THEN NULL
        ELSE '09:00'::time
    END,
    CASE
        WHEN d::date IN (SELECT hdate FROM holidays)
          OR EXTRACT(DOW FROM d) = 0 THEN NULL
        WHEN EXTRACT(DOW FROM d) = 6   THEN '13:00'::time
        ELSE                                '18:00'::time
    END,
    ''
FROM generate_series('2026-04-13'::date, '2026-06-13'::date, '1 day') d
CROSS JOIN staff_clinic sc
ON CONFLICT DO NOTHING;

-- =============================================================================
-- J. マスタデータ均一化 — 城東センター病院 (clinic_id=2) + 敷島医院 (clinic_id=3)
-- 八王子病院 (clinic_id=1) と同水準のマスタデータを追加
-- =============================================================================

-- -----------------------------------------------------------------------------
-- J-1. cages（ケージ追加）
-- 城東: IDs 17-20 (+4 → 計8件), 敷島: IDs 21-24 (+4 → 計8件)
-- -----------------------------------------------------------------------------
INSERT INTO cages (id, clinic_id, name, price, is_active, description, cage_type, cage_size, sort_order) VALUES
    (17, 2, '犬用ケージ（大）',   3800, true, '大型犬・術後管理用',             'dog',     'large',  5),
    (18, 2, '猫用ケージ（小）',   2800, true, '猫専用・ストレス軽減設計',       'cat',     'small',  6),
    (19, 2, '猫用ケージ（中）',   3000, true, '猫専用・中型',                   'cat',     'medium', 7),
    (20, 2, '汎用ケージ',          2500, true, '小動物・鳥類等対応',             'general', 'small',  8),
    (21, 3, 'ICUケージB',          8000, true, '酸素吸入可・重症患者用（副）',   'icu',     'medium', 5),
    (22, 3, '犬用ケージ（中）',   3600, true, '中型犬・一般入院用',             'dog',     'medium', 6),
    (23, 3, '猫用ケージ（小）',   2900, true, '猫専用・ストレス軽減設計',       'cat',     'small',  7),
    (24, 3, '汎用ケージ',          2400, true, '小動物・鳥類等対応',             'general', 'small',  8)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('cages', 'id'), (SELECT MAX(id) FROM cages));

-- -----------------------------------------------------------------------------
-- J-2. exam_types + exam_type_fields
-- 城東: IDs 12-13 (+2 → 計5件), 敷島: IDs 14-15 (+2 → 計5件)
-- -----------------------------------------------------------------------------
INSERT INTO exam_types (id, clinic_id, name, price, is_active, description, sort_order) VALUES
    (12, 2, '超音波検査（エコー）', 5200, true, '腹部エコー・心臓エコー',                 4),
    (13, 2, '尿検査',               1600, true, '尿試験紙・尿沈渣検査',                   5),
    (14, 3, '血液化学検査',          5000, true, '肝機能・腎機能・血糖値など生化学的検査', 4),
    (15, 3, 'レントゲン検査',        3200, true, 'X線撮影（胸部・腹部・四肢）',            5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('exam_types', 'id'), (SELECT MAX(id) FROM exam_types));

INSERT INTO exam_type_fields (id, exam_type_id, name, inspection_value, normal_value, sort_order) VALUES
    -- 城東 exam_type 12 (超音波)
    (34, 12, '腹部エコー', '', '異常なし', 1),
    (35, 12, '心臓エコー', '', '異常なし', 2),
    -- 城東 exam_type 13 (尿検査)
    (36, 13, '尿比重',     '', '1.015-1.045', 1),
    (37, 13, '尿pH',       '', '5.5-7.5',     2),
    (38, 13, '尿タンパク', '', '陰性',         3),
    -- 敷島 exam_type 14 (血液化学)
    (39, 14, 'ALT（GPT）',          '', '10-125 U/L',    1),
    (40, 14, 'BUN（尿素窒素）',     '', '7-27 mg/dL',    2),
    (41, 14, 'CRE（クレアチニン）', '', '0.5-1.8 mg/dL', 3),
    (42, 14, 'GLU（血糖値）',       '', '74-143 mg/dL',  4),
    -- 敷島 exam_type 15 (レントゲン)
    (43, 15, '胸部正面', '', '異常なし', 1),
    (44, 15, '腹部正面', '', '異常なし', 2)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('exam_type_fields', 'id'), (SELECT MAX(id) FROM exam_type_fields));

-- -----------------------------------------------------------------------------
-- J-3. vaccines
-- 城東: IDs 21-25 (+5 → 計10件), 敷島: IDs 26-30 (+5 → 計10件)
-- -----------------------------------------------------------------------------
INSERT INTO vaccines (id, clinic_id, name, price, is_active, description, species, interval, sort_order) VALUES
    (21, 2, '混合ワクチン10種（犬）',  8200, true, '5種＋レプトスピラ5種',                                          'dog', '1年',    6),
    (22, 2, '混合ワクチン5種（猫）',   5600, true, '3種＋猫白血病・猫クラミジア',                                   'cat', '1年',    7),
    (23, 2, 'フィラリア予防薬（中型犬）', 1100, true, '体重11-25kg犬用フィラリア予防',                              'dog', '1ヶ月', 8),
    (24, 2, 'フィラリア予防薬（大型犬）', 1500, true, '体重26kg以上犬用フィラリア予防',                             'dog', '1ヶ月', 9),
    (25, 2, '混合ワクチン6種（犬）',   5600, true, '5種＋コロナウイルス',                                          'dog', '1年',   10),
    (26, 3, '混合ワクチン6種（犬）',   5800, true, '5種＋コロナウイルス',                                          'dog', '1年',    6),
    (27, 3, 'フィラリア予防薬（小型犬）', 950, true, '体重10kg以下犬用フィラリア予防',                              'dog', '1ヶ月', 7),
    (28, 3, 'フィラリア予防薬（中型犬）', 1150, true, '体重11-25kg犬用フィラリア予防',                             'dog', '1ヶ月', 8),
    (29, 3, 'フィラリア予防薬（大型犬）', 1550, true, '体重26kg以上犬用フィラリア予防',                            'dog', '1ヶ月', 9),
    (30, 3, '混合ワクチン5種（猫）プレミアム', 5800, true, '3種＋猫白血病・猫クラミジア（高免疫原性）', 'cat', '1年', 10)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('vaccines', 'id'), (SELECT MAX(id) FROM vaccines));

-- -----------------------------------------------------------------------------
-- J-4. medicines（カテゴリ + 薬剤）
-- 城東: カテゴリ 2006-2009 (+4 → 計9件), 薬剤 111-115 (+5 → 計15件)
-- 敷島: カテゴリ 3005-3009 (+5 → 計9件), 薬剤 211-215 (+5 → 計15件)
-- -----------------------------------------------------------------------------
INSERT INTO medicines (id, clinic_id, name, price, is_active, description, sort_order) VALUES
    (2006, 2, '制吐剤',       NULL, true, '制吐剤カテゴリ',         6),
    (2007, 2, '消化器用薬',   NULL, true, '消化器用薬カテゴリ',     7),
    (2008, 2, '神経系薬',     NULL, true, '神経系薬カテゴリ',       8),
    (2009, 2, '輸液（追加）', NULL, true, '輸液・補液カテゴリ',     9),
    (3005, 3, '消炎剤',       NULL, true, '消炎鎮痛剤カテゴリ',     5),
    (3006, 3, '利尿剤',       NULL, true, '利尿剤カテゴリ',         6),
    (3007, 3, '神経系薬',     NULL, true, '神経系薬カテゴリ',       7),
    (3008, 3, '駆虫剤',       NULL, true, '駆虫剤カテゴリ',         8),
    (3009, 3, '輸液（追加）', NULL, true, '輸液・補液カテゴリ',     9)
ON CONFLICT DO NOTHING;

INSERT INTO medicines (id, clinic_id, name, price, is_active, description, dosage_form, medicine_unit, default_quantity, sort_order, parent_id) VALUES
    (111, 2, 'マロピタント 16mg',              820, true, '制吐剤（乗り物酔い・嘔吐治療）',    'tablet',    'per_tablet', 1,   11, 2006),
    (112, 2, 'オメプラゾール 10mg',            360, true, 'プロトンポンプ阻害薬（胃酸抑制）',  'tablet',    'per_tablet', 1,   12, 2007),
    (113, 2, 'ラクツロース液',                 510, true, '便秘・肝性脳症の治療',              'liquid',    'per_ml',     5,   13, 2007),
    (114, 2, 'ガバペンチン 100mg',             560, true, '神経因性疼痛・てんかん補助療法',    'tablet',    'per_tablet', 1,   14, 2008),
    (115, 2, '乳酸リンゲル液 500ml',           450, true, '補液・電解質補正',                  'liquid',    'per_ml',     500, 15, 2009),
    (211, 3, 'メロキシカム経口液',             710, true, 'NSAIDs・痛み・炎症の緩和',          'liquid',    'per_ml',     1,   11, 3005),
    (212, 3, 'カルプロフェン 25mg',            670, true, 'NSAIDs・術後疼痛管理',              'tablet',    'per_tablet', 1,   12, 3005),
    (213, 3, 'フロセミド注射液',               820, true, '利尿剤（心臓・腎臓病の浮腫治療）',  'injection', 'per_ml',     2,   13, 3006),
    (214, 3, 'ガバペンチン 100mg',             560, true, '神経因性疼痛・てんかん補助療法',    'tablet',    'per_tablet', 1,   14, 3007),
    (215, 3, 'ノミ・ダニ駆除薬（犬用）',      2550, true, '外部寄生虫予防・駆除（スポットオン）', 'topical', 'per_dose',   1,   15, 3008)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('medicines', 'id'), (SELECT MAX(id) FROM medicines));

-- -----------------------------------------------------------------------------
-- J-5. consultations
-- 城東: IDs 12-13 (+2 → 計5件), 敷島: IDs 14-15 (+2 → 計5件)
-- -----------------------------------------------------------------------------
INSERT INTO consultations (id, clinic_id, name, price, is_active, description, time_condition, duration, sort_order) VALUES
    (12, 2, '往診料',       5200, true, '自宅への往診料（基本料金）',       'anytime',    60, 4),
    (13, 2, '電話相談料',    500, true, '電話による診察相談',               'anytime',    15, 5),
    (14, 3, '時間外診療料', 3100, true, '診療時間外・休日の緊急診察',       'after_hours', 30, 4),
    (15, 3, '電話相談料',    500, true, '電話による診察相談',               'anytime',    15, 5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('consultations', 'id'), (SELECT MAX(id) FROM consultations));

-- -----------------------------------------------------------------------------
-- J-6. procedures
-- 城東: IDs 21-25 (+5 → 計10件), 敷島: IDs 26-30 (+5 → 計10件)
-- -----------------------------------------------------------------------------
INSERT INTO procedures (id, clinic_id, name, price, is_active, description, duration, anesthesia, sort_order) VALUES
    (21, 2, '歯石除去',       15500, true, '全身麻酔下での歯石除去・歯周治療',  45, 'general', 6),
    (22, 2, '皮膚縫合',        5200, true, '裂傷・切傷の縫合処置',              30, 'local',   7),
    (23, 2, '骨折整復',       82000, true, '骨折の外科的整復・固定',           120, 'general', 8),
    (24, 2, '胃洗浄',         10500, true, '異物誤飲時の胃洗浄処置',            30, 'general', 9),
    (25, 2, '腫瘍摘出',       21000, true, '皮膚腫瘍の外科的摘出',              60, 'local',   10),
    (26, 3, '歯科処置（歯石除去）', 15200, true, '全身麻酔下での歯石除去・歯周治療', 45, 'general', 6),
    (27, 3, '耳洗浄',          2600, true, '外耳炎治療・耳道内の洗浄処置',      15, 'none',    7),
    (28, 3, '皮膚縫合',        5100, true, '裂傷・切傷の縫合処置',              30, 'local',   8),
    (29, 3, '胃洗浄',         10200, true, '異物誤飲時の胃洗浄処置',            30, 'general', 9),
    (30, 3, '点滴処置',        3100, true, '静脈内点滴（1時間）',               60, 'none',    10)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('procedures', 'id'), (SELECT MAX(id) FROM procedures));

-- -----------------------------------------------------------------------------
-- J-7. hospitalization_plans
-- 城東: IDs 12-13 (+2 → 計5件), 敷島: IDs 14-15 (+2 → 計5件)
-- -----------------------------------------------------------------------------
INSERT INTO hospitalization_plans (id, clinic_id, name, price, is_active, description, body_size, billing_unit, sort_order) VALUES
    (12, 2, 'ICU入院（ハイケア）', 8800, true, '集中治療室・ハイケア管理料（1日）', 'small',  'per_day',   4),
    (13, 2, 'ホテル（小型）',     2700, true, '体重10kg以下のペットホテル（1泊）', 'small',  'per_night', 5),
    (14, 3, '一般入院（中型）',   3900, true, '体重10-25kgの入院管理料（1日）',   'medium', 'per_day',   4),
    (15, 3, 'ICU入院',            8500, true, '集中治療室管理料（1日）',           'small',  'per_day',   5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('hospitalization_plans', 'id'), (SELECT MAX(id) FROM hospitalization_plans));

-- -----------------------------------------------------------------------------
-- J-8. trimming_courses
-- 城東: IDs 12-13 (+2 → 計5件), 敷島: IDs 14-15 (+2 → 計5件)
-- -----------------------------------------------------------------------------
INSERT INTO trimming_courses (id, clinic_id, name, price, is_active, description, target_size, duration, sort_order) VALUES
    (12, 2, 'シャンプー&ブロー（中型）', 5700, true, 'シャンプー・ブロー・ブラッシング',            'medium', 90,  4),
    (13, 2, 'フルコース（大型）',        12500, true, 'カット・シャンプー・ブロー・爪切り・耳掃除', 'large',  180, 5),
    (14, 3, 'シャンプー&ブロー（大型）', 6800, true, 'シャンプー・ブロー・ブラッシング',            'large',  120, 4),
    (15, 3, 'フルコース（中型）',        9200, true, 'カット・シャンプー・ブロー・爪切り・耳掃除', 'medium', 150, 5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('trimming_courses', 'id'), (SELECT MAX(id) FROM trimming_courses));

-- -----------------------------------------------------------------------------
-- J-9. trimming_options
-- 城東: IDs 12-13 (+2 → 計5件), 敷島: IDs 14-15 (+2 → 計5件)
-- -----------------------------------------------------------------------------
INSERT INTO trimming_options (id, clinic_id, name, price, is_active, description, duration, is_combinable, sort_order) VALUES
    (12, 2, '歯磨き',     520, true, '歯ブラシによるデンタルケア',    15, true, 4),
    (13, 2, 'リボン装着', 210, true, '仕上げのアクセサリー装着',       5, true, 5),
    (14, 3, '肛門腺絞り', 310, true, '肛門嚢の分泌液除去',             5, true, 4),
    (15, 3, 'リボン装着', 200, true, '仕上げのアクセサリー装着',       5, true, 5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('trimming_options', 'id'), (SELECT MAX(id) FROM trimming_options));

-- -----------------------------------------------------------------------------
-- J-10. diagnosis_types
-- 城東: IDs 19-21 (+3 → 計8件), 敷島: IDs 22-24 (+3 → 計8件)
-- -----------------------------------------------------------------------------
INSERT INTO diagnosis_types (id, clinic_id, name, is_active, description, sort_order) VALUES
    (19, 2, '腫瘍',           true, '良性・悪性腫瘍（がん）',               6),
    (20, 2, '外傷・骨格',     true, '骨折・咬傷・関節疾患など',             7),
    (21, 2, '神経系',         true, '脳・脊髄・末梢神経などの神経系疾患',   8),
    (22, 3, '呼吸器系',       true, '肺・気管・鼻腔などの呼吸器系疾患',     6),
    (23, 3, '神経系',         true, '脳・脊髄・末梢神経などの神経系疾患',   7),
    (24, 3, '感染症・寄生虫', true, '細菌・ウイルス・寄生虫感染症',         8)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('diagnosis_types', 'id'), (SELECT MAX(id) FROM diagnosis_types));

-- -----------------------------------------------------------------------------
-- J-11. diagnosis_names
-- 城東: IDs 43-52 (+10), 敷島: IDs 53-62 (+10)
-- -----------------------------------------------------------------------------
INSERT INTO diagnosis_names (id, clinic_id, name, is_active, description, diagnosis_type_id, sort_order) VALUES
    -- 城東: 既存カテゴリへの追加
    (43, 2, '肝疾患',               true, '肝炎・肝不全・脂肪肝など',           9,  3),
    (44, 2, '腸重積',               true, '腸管の重積（緊急外科適応）',         9,  4),
    (45, 2, '肺炎',                 true, '肺の感染性・非感染性炎症',           10, 3),
    (46, 2, '喉頭炎',               true, '喉頭部の炎症',                       10, 4),
    (47, 2, '外耳炎',               true, '外耳道の炎症・細菌/真菌/寄生虫感染', 11, 3),
    (48, 2, '腎不全',               true, '急性・慢性腎不全',                   12, 3),
    (49, 2, '尿路結石',             true, '腎結石・膀胱結石・尿道結石',         12, 4),
    -- 城東: 新カテゴリ
    (50, 2, '肥満細胞腫',           true, '皮膚または内臓の肥満細胞腫瘍',       19, 1),
    (51, 2, '骨折',                 true, '各部位の骨折',                       20, 1),
    (52, 2, 'てんかん',             true, '反復性の痙攣発作',                   21, 1),
    -- 敷島: 既存カテゴリへの追加
    (53, 3, '膵炎',                 true, '膵臓の炎症',                         14, 3),
    (54, 3, '腸重積',               true, '腸管の重積（緊急外科適応）',         14, 4),
    (55, 3, '外耳炎',               true, '外耳道の炎症・細菌/真菌/寄生虫感染', 15, 4),
    (56, 3, '脂漏症',               true, '皮脂分泌異常による皮膚疾患',         15, 5),
    (57, 3, '腎不全',               true, '急性・慢性腎不全',                   16, 3),
    (58, 3, 'FLUTD',               true, '猫下部尿路疾患',                      16, 4),
    (59, 3, 'リンパ腫',             true, '悪性リンパ腫',                       17, 2),
    -- 敷島: 新カテゴリ
    (60, 3, '気管支炎',             true, '気管支の炎症',                       22, 1),
    (61, 3, 'てんかん',             true, '反復性の痙攣発作',                   23, 1),
    (62, 3, 'フィラリア症',         true, '犬糸状虫による心肺疾患',             24, 1)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('diagnosis_names', 'id'), (SELECT MAX(id) FROM diagnosis_names));

-- -----------------------------------------------------------------------------
-- J-12. checkup_types
-- 城東: ID 11 (+1 → 計4件), 敷島: ID 12 (+1 → 計4件)
-- -----------------------------------------------------------------------------
INSERT INTO checkup_types (id, clinic_id, name, price, is_active, description, interval, target_age, sort_order) VALUES
    (11, 2, '歯科検診', 3100, true, '歯周病チェック・歯石付着度の確認', '1年', '成犬', 4),
    (12, 3, 'フィラリア検査', 2600, true, 'フィラリア抗原検査（予防シーズン前）', '1年', '成犬', 4)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('checkup_types', 'id'), (SELECT MAX(id) FROM checkup_types));

-- -----------------------------------------------------------------------------
-- J-13. chief_complaint_types
-- 城東: IDs 15-16 (+2 → 計6件), 敷島: IDs 17-18 (+2 → 計6件)
-- -----------------------------------------------------------------------------
INSERT INTO chief_complaint_types (id, clinic_id, name, is_active, sort_order) VALUES
    (15, 2, '外傷・骨折',   true, 5),
    (16, 2, '呼吸困難',     true, 6),
    (17, 3, '皮膚・被毛異常', true, 5),
    (18, 3, '排尿・排泄異常', true, 6)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('chief_complaint_types', 'id'), (SELECT MAX(id) FROM chief_complaint_types));

-- -----------------------------------------------------------------------------
-- J-14. merchandise_items
-- 城東: IDs 16-18 (+3 → 計7件), 敷島: IDs 19-21 (+3 → 計7件)
-- -----------------------------------------------------------------------------
INSERT INTO merchandise_items (id, clinic_id, name, category, unit_price, tax_rate, sort_order) VALUES
    (16, 2, 'エリザベスカラー（M）',             'goods', 1050, 0.10, 5),
    (17, 2, 'ノミ・ダニ予防首輪',               'goods', 1650, 0.10, 6),
    (18, 2, '時間外診療費',                     'other', 5200, 0.10, 7),
    (19, 3, 'ヒルズ i/d（犬用・消化器ケア）',    'food',  3200, 0.10, 5),
    (20, 3, 'ペット用歯ブラシセット',            'goods', 1200, 0.10, 6),
    (21, 3, '時間外診療費',                     'other', 5100, 0.10, 7)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('merchandise_items', 'id'), (SELECT MAX(id) FROM merchandise_items));

-- =============================================================================
-- K. トランザクションデータ — 城東センター病院 (clinic_id=2)
-- doctors: staff 16(木村), 17(佐々木), 18(高橋)
-- owners: 23-30 / pets: 29-38
-- =============================================================================

-- -----------------------------------------------------------------------------
-- K-1. medical_records（城東: 8件、IDs 21-28）
-- -----------------------------------------------------------------------------
INSERT INTO medical_records (id, clinic_id, record_no, date, owner_id, pet_id, doctor_id, status) VALUES
    (21, 2, 'C-2025-001', '2025-11-10', 23, 29, 16, 'finalized'),
    (22, 2, 'C-2025-002', '2025-10-20', 24, 31, 17, 'finalized'),
    (23, 2, 'C-2025-003', '2025-12-15', 25, 32, 16, 'finalized'),
    (24, 2, 'C-2026-001', '2026-01-08', 26, 33, 18, 'finalized'),
    (25, 2, 'C-2025-004', '2025-12-01', 27, 34, 16, 'finalized'),
    (26, 2, 'C-2025-005', '2025-09-05', 28, 35, 17, 'finalized'),
    (27, 2, 'C-2026-002', '2026-01-15', 29, 37, 16, 'finalized'),
    (28, 2, 'C-2026-003', '2026-02-20', 30, 38, 18, 'draft')
ON CONFLICT (id) DO UPDATE SET updated_at = now();


-- -----------------------------------------------------------------------------
-- K-2. inquiries（城東: 8件、IDs 25-32）
-- -----------------------------------------------------------------------------
INSERT INTO inquiries (id, medical_record_id, chief_complaint_type_id, chief_complaint, notes, staff_id) VALUES
    (25, 21, NULL,  '定期健診・フィラリア検査',           '体調良好。フィラリア陰性。',           16),
    (26, 22, NULL,  '3種混合ワクチン接種',                 'ワクチン接種。体調問題なし。',         17),
    (27, 23, NULL,  '血液検査・一般健診',                   '年次健診。',                           16),
    (28, 24, 9,     '皮膚の痒み・発赤',                     '2週間前から痒がる。アトピー疑い。',   18),
    (29, 25, 7,     '食欲低下',                             '3日前から食欲落ちている。',             16),
    (30, 26, NULL,  '3種混合ワクチン接種（猫）',           'ワクチン接種。良好。',                 17),
    (31, 27, NULL,  '再診（足の不調）',                     '前回から改善傾向。',                   16),
    (32, 28, 7,     '食欲不振・嘔吐',                       '初診。昨日から嘔吐1回。',             18)
ON CONFLICT (id) DO UPDATE SET medical_record_id = EXCLUDED.medical_record_id, updated_at = now();


-- -----------------------------------------------------------------------------
-- K-3. clinical_plans（城東: 8件、IDs 21-28）
-- -----------------------------------------------------------------------------
INSERT INTO clinical_plans (id, medical_record_id, physical_exam, diagnosis_type_id, diagnosis_name_id, diagnosis_details, treatment_policy) VALUES
    (21, 21, '体重28kg。体温38.4℃。心肺音正常。', NULL, NULL, 'フィラリア陰性。健康良好。', 'フィラリア検査・次回予防薬処方。'),
    (22, 22, '体重3.8kg。良好。',                 NULL, NULL, 'ワクチン適応あり。',         'ワクチン接種実施。'),
    (23, 23, 'シニア期。体重変化なし。',           NULL, NULL, '経過良好。',                 '血液検査実施。結果説明。'),
    (24, 24, '全身に搔痒。皮膚発赤あり。',         11,   47,   'アトピー性皮膚炎疑い。',    '抗ヒスタミン薬処方。薬用シャンプー推奨。'),
    (25, 25, '腹部軽度緊張。体温38.8℃。',         9,    21,   '急性胃腸炎疑い。',           '絶食指示・皮下補液実施。'),
    (26, 26, '体重3.0kg。体調良好。',             NULL, NULL, 'ワクチン適応あり。',         'ワクチン接種実施。'),
    (27, 27, '跛行改善傾向。',                     20,   51,   '前肢骨折（回復期）。',       '運動制限継続。次回X線確認。'),
    (28, 28, '腹部触診で軽度抵抗感。',             9,    21,   '急性胃腸炎。',               '補液・制吐剤投与。食事制限。')
ON CONFLICT (id) DO UPDATE SET updated_at = now();


-- -----------------------------------------------------------------------------
-- K-4. vaccinations（城東: 3件、IDs 6-8）
-- -----------------------------------------------------------------------------
INSERT INTO vaccinations (id, clinic_id, medical_record_id, pet_id, vaccine_id, doctor_id, date, lot1, next_schedule_type, next_date, remarks) VALUES
    (6, 2, 22, 31, 13, 17, '2025-10-20', 'LOT-2025-J001', '1year',  '2026-10-20', '3種混合ワクチン接種。体調良好。'),
    (7, 2, 26, 35, 13, 17, '2025-09-05', 'LOT-2025-J002', '1year',  '2026-09-05', '3種混合ワクチン接種。'),
    (8, 2, 21, 29, 11, 16, '2025-11-10', 'LOT-2025-J003', '1year',  '2026-11-10', '5種混合ワクチン接種。体調良好。')
ON CONFLICT (id) DO NOTHING;


-- -----------------------------------------------------------------------------
-- K-5. exams（城東: 2件、IDs 4-5）
-- -----------------------------------------------------------------------------
INSERT INTO exams (id, clinic_id, medical_record_id, exam_type_id, doctor_id, date, result_summary, status) VALUES
    (4, 2, 23, 6, 16, '2025-12-15', 'CBC全項目正常範囲内。', 'completed'),
    (5, 2, 25, 8, 16, '2025-12-01', '腹部正面：腸管ガス像あり。他異常なし。', 'completed')
ON CONFLICT (id) DO NOTHING;


INSERT INTO exam_results (id, exam_id, exam_type_field_id, inspection_value, status) VALUES
    -- Exam 4 (CBC for MR23, 城東)
    (46, 2, 18, '8.5',  'normal'),
    (47, 2, 19, '6.8',  'normal'),
    (48, 2, 20, '46',   'normal'),
    -- Exam 5 (Xray for MR25, 城東)
    (49, 3, 24, '腸管ガス像あり', 'high'),
    (50, 3, 25, '異常なし',       'normal')
ON CONFLICT (id) DO NOTHING;


-- -----------------------------------------------------------------------------
-- K-6. treatments（城東: 5件、IDs 9-13）
-- -----------------------------------------------------------------------------
INSERT INTO treatments (id, medical_record_id, item_type, consultation_id, procedure_id, medicine_id, inventory_id, is_selected, status, content, unit_price, quantity, sort_order) VALUES
    (9,  21, 'consultation', 7,    NULL, NULL, NULL, true, 'completed', '再診料',                      900,  1, 1),
    (10, 22, 'consultation', 6,    NULL, NULL, NULL, true, 'completed', '初診料',                     2200,  1, 1),
    (11, 24, 'medicine',     NULL, NULL, 104,  NULL, true, 'completed', 'プレドニゾロン 5mg x 7日分',  420,  7, 2),
    (12, 25, 'consultation', 7,    NULL, NULL, NULL, true, 'completed', '再診料',                      900,  1, 1),
    (13, 25, 'medicine',     NULL, NULL, 101,  NULL, true, 'completed', 'アモキシシリン 50mg x 5日分', 520,  5, 2)
ON CONFLICT (id) DO UPDATE SET updated_at = now();


-- =============================================================================
-- L. トランザクションデータ — 敷島医院 (clinic_id=3)
-- doctors: staff 26(藤原), 27(松本)
-- owners: 31-38 / pets: 39-48
-- =============================================================================

-- -----------------------------------------------------------------------------
-- L-1. medical_records（敷島: 8件、IDs 29-36）
-- -----------------------------------------------------------------------------
INSERT INTO medical_records (id, clinic_id, record_no, date, owner_id, pet_id, doctor_id, status) VALUES
    (29, 3, 'S-2025-001', '2025-10-05', 31, 39, 26, 'finalized'),
    (30, 3, 'S-2025-002', '2025-12-18', 32, 40, 27, 'finalized'),
    (31, 3, 'S-2026-001', '2026-01-08', 33, 41, 26, 'finalized'),
    (32, 3, 'S-2026-002', '2026-02-15', 34, 42, 27, 'finalized'),
    (33, 3, 'S-2025-003', '2025-09-20', 35, 44, 26, 'finalized'),
    (34, 3, 'S-2026-003', '2026-03-01', 36, 45, 26, 'draft'),
    (35, 3, 'S-2025-004', '2025-11-25', 37, 46, 27, 'finalized'),
    (36, 3, 'S-2026-004', '2026-02-10', 38, 47, 26, 'finalized')
ON CONFLICT (id) DO UPDATE SET updated_at = now();

SELECT setval(pg_get_serial_sequence('medical_records', 'id'), (SELECT MAX(id) FROM medical_records));

-- -----------------------------------------------------------------------------
-- L-2. inquiries（敷島: 8件、IDs 33-40）
-- -----------------------------------------------------------------------------
INSERT INTO inquiries (id, medical_record_id, chief_complaint_type_id, chief_complaint, notes, staff_id) VALUES
    (33, 29, NULL,  '年次健診・ワクチン接種',           '体調良好。体重24kg。',                 26),
    (34, 30, NULL,  '3種混合ワクチン接種（猫）',         'ワクチン接種。体調良好。',             27),
    (35, 31, NULL,  '血液検査（年次健診）',               'シニア健診。',                         26),
    (36, 32, 12,    '嘔吐・下痢',                         '昨日から嘔吐2回。軟便あり。',         27),
    (37, 33, 17,    '皮膚の痒み・かさぶた',               'アレルギー疑い。フリーズドライ投与。', 26),
    (38, 34, NULL,  '初診・一般診察',                     '初めての来院。',                       26),
    (39, 35, NULL,  '耳を痒がる',                         '外耳炎疑い。耳垢あり。',               27),
    (40, 36, NULL,  '3種混合ワクチン接種（猫）',         'ワクチン接種。体重4.0kg。',            26)
ON CONFLICT (id) DO UPDATE SET medical_record_id = EXCLUDED.medical_record_id, updated_at = now();

SELECT setval(pg_get_serial_sequence('inquiries', 'id'), (SELECT MAX(id) FROM inquiries));

-- -----------------------------------------------------------------------------
-- L-3. clinical_plans（敷島: 8件、IDs 29-36）
-- -----------------------------------------------------------------------------
INSERT INTO clinical_plans (id, medical_record_id, physical_exam, diagnosis_type_id, diagnosis_name_id, diagnosis_details, treatment_policy) VALUES
    (29, 29, '体温38.3℃。体重24kg。心肺音正常。', NULL, NULL, '健康良好。ワクチン接種可。', '5種混合ワクチン接種実施。フィラリア予防薬処方。'),
    (30, 30, '体重4.8kg。良好。',                 NULL, NULL, 'ワクチン適応あり。',         'ワクチン接種実施。'),
    (31, 31, 'シニア期。体重31kg。',               NULL, NULL, '経過良好。',                 '血液化学検査実施。結果説明。'),
    (32, 32, '脱水傾向あり。腹部緊張。',           14,   31,   '急性胃腸炎。',               '補液・絶食・制吐剤処方。'),
    (33, 33, '体幹・四肢に紅斑。搔痒感強。',       15,   33,   'アトピー性皮膚炎。',        'ステロイド処方。薬用シャンプー推奨。'),
    (34, 34, '一般状態良好。', NULL, NULL, '初診。特異所見なし。', '経過観察。次回再診。'),
    (35, 35, '耳道内に暗褐色の耳垢。臭気あり。',  NULL, NULL, '外耳炎。',                   '耳洗浄・点耳薬処方。'),
    (36, 36, '体重4.0kg。体調良好。',             NULL, NULL, 'ワクチン適応あり。',         'ワクチン接種実施。')
ON CONFLICT (id) DO UPDATE SET updated_at = now();

SELECT setval(pg_get_serial_sequence('clinical_plans', 'id'), (SELECT MAX(id) FROM clinical_plans));

-- -----------------------------------------------------------------------------
-- L-4. vaccinations（敷島: 4件、IDs 9-12）
-- -----------------------------------------------------------------------------
INSERT INTO vaccinations (id, clinic_id, medical_record_id, pet_id, vaccine_id, doctor_id, date, lot1, next_schedule_type, next_date, remarks) VALUES
    (9, 3, 29, 39, 16, 26, '2025-10-05', 'LOT-2025-S001', '1year',  '2026-10-05', '5種混合ワクチン接種。体調良好。'),
    (10, 3, 30, 40, 18, 27, '2025-12-18', 'LOT-2025-S002', '1year',  '2026-12-18', '3種混合ワクチン接種（猫）。'),
    (11, 3, 36, 47, 19, 26, '2026-02-10', 'LOT-2026-S001', '1year',  '2027-02-10', '5種混合ワクチン（猫）接種。'),
    (12, 3, 29, 39, 20, 26, '2025-10-05', 'LOT-2025-S003', '1year',  '2026-10-05', '狂犬病ワクチン接種。')
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('vaccinations', 'id'), (SELECT MAX(id) FROM vaccinations));

-- -----------------------------------------------------------------------------
-- L-5. exams（敷島: 3件、IDs 6-8）
-- -----------------------------------------------------------------------------
INSERT INTO exams (id, clinic_id, medical_record_id, exam_type_id, doctor_id, date, result_summary, status) VALUES
    (6, 3, 31, 9,  26, '2026-01-08', 'CBC全項目正常範囲内。加齢による軽度変化のみ。', 'completed'),
    (7, 3, 32, 10, 27, '2026-02-15', '尿比重低下（1.010）。腎機能低下疑い。',        'completed'),
    (8, 3, 33, 11, 26, '2025-09-20', '腹部エコー：異常なし。皮膚問題は別要因。',      'completed')
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('exams', 'id'), (SELECT MAX(id) FROM exams));

INSERT INTO exam_results (id, exam_id, exam_type_field_id, inspection_value, status) VALUES
    -- Exam 6 (CBC for MR31, 敷島)
    (51, 6, 26, '9.2',  'normal'),
    (52, 6, 27, '7.0',  'normal'),
    (53, 6, 28, '320',  'normal'),
    -- Exam 7 (尿検査 for MR32, 敷島)
    (54, 7, 29, '1.010',  'low'),
    (55, 7, 30, '6.5',    'normal'),
    (56, 7, 31, '陰性',   'normal'),
    -- Exam 8 (超音波 for MR33, 敷島)
    (57, 8, 32, '異常なし', 'normal'),
    (58, 8, 33, '異常なし', 'normal')
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('exam_results', 'id'), (SELECT MAX(id) FROM exam_results));

-- -----------------------------------------------------------------------------
-- L-6. treatments（敷島: 5件、IDs 14-18）
-- -----------------------------------------------------------------------------
INSERT INTO treatments (id, medical_record_id, item_type, consultation_id, procedure_id, medicine_id, inventory_id, is_selected, status, content, unit_price, quantity, sort_order) VALUES
    (14, 29, 'consultation', 9,    NULL, NULL, NULL, true, 'completed', '初診料',                      1900, 1, 1),
    (15, 30, 'consultation', 10,   NULL, NULL, NULL, true, 'completed', '再診料',                       800, 1, 1),
    (16, 32, 'consultation', 10,   NULL, NULL, NULL, true, 'completed', '再診料',                       800, 1, 1),
    (17, 32, 'medicine',     NULL, NULL, 201,  NULL, true, 'completed', 'アモキシシリン 50mg x 5日分',  510, 5, 2),
    (18, 35, 'procedure',    NULL, 27,   NULL, NULL, true, 'completed', '耳道洗浄・点耳薬処置',        2600, 1, 1)
ON CONFLICT (id) DO UPDATE SET updated_at = now();

SELECT setval(pg_get_serial_sequence('treatments', 'id'), (SELECT MAX(id) FROM treatments));

-- =============================================================================
-- SEED: ステージングデータ (004_seed_staging)
-- =============================================================================
-- =============================================================================
-- Animal Ekarte - Staging Additions (clinic_id=4,5 exclusive data)
-- PostgreSQL 18
-- 依存: 001_init.sql → 002_seed_master.sql → 003_seed_demo.sql
-- 内容: clinic_id=4,5 新規追加クリニック用の最小限データセット
-- =============================================================================

-- Staging clinics: clinic_id=4,5 (master only)
-- デモ用permission_groups・rules は GORM CreateClinic が動的生成するため手動定義不要

-- NOTE: clinic, company, permission_groups, permission_group_rules, staff_clinic_assignments
-- などは 003_seed_demo.sql で clinic_id=1,2,3 のみ処理。clinic_id=4,5 の運用データ
-- が必要な場合は、以降のマイグレーション（005以降）で段階的に追加可能。
-- 現在の004は「最小限のデータセットで動作確認可能」な状態を目指す。
