-- ============================================================
-- LINE予約システム マイグレーション
-- 既存テーブル拡張 + 新規テーブル4つ
-- ============================================================

-- ============================================================
-- A0. ENUM型の追加
-- ============================================================
CREATE TYPE staff_type AS ENUM ('doctor', 'nurse', 'resource');
CREATE TYPE reservation_source AS ENUM ('manual', 'line');

-- ============================================================
-- A1. service_types に予約用カラムを追加
-- ============================================================
ALTER TABLE service_types ADD COLUMN duration_minutes      INT NOT NULL DEFAULT 15;
ALTER TABLE service_types ADD COLUMN short_name            TEXT NOT NULL DEFAULT '';
ALTER TABLE service_types ADD COLUMN show_short_name       BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE service_types ADD COLUMN reservation_visible   BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE service_types ADD COLUMN reservation_comment   TEXT NOT NULL DEFAULT '';
ALTER TABLE service_types ADD COLUMN reservation_image_url TEXT NOT NULL DEFAULT '';
ALTER TABLE service_types ADD COLUMN reservation_day_option TEXT NOT NULL DEFAULT 'none';
ALTER TABLE service_types ADD COLUMN is_internal           BOOLEAN NOT NULL DEFAULT false;

-- ============================================================
-- A2. staffs に予約用カラムを追加
-- ============================================================
ALTER TABLE staffs ADD COLUMN staff_type            staff_type NOT NULL DEFAULT 'doctor';
ALTER TABLE staffs ADD COLUMN reservation_visible   BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE staffs ADD COLUMN reservation_comment   TEXT NOT NULL DEFAULT '';
ALTER TABLE staffs ADD COLUMN reservation_image_url TEXT NOT NULL DEFAULT '';

-- ============================================================
-- B1. LINE予約基本設定（クリニック単位 1:1）
-- ============================================================
CREATE TABLE reservation_settings (
    id                        BIGSERIAL PRIMARY KEY,
    clinic_id                 BIGINT NOT NULL UNIQUE REFERENCES clinics(id),
    status                    TEXT NOT NULL DEFAULT 'stopped',

    -- ページ編集（トップページ）
    header_text               TEXT NOT NULL DEFAULT '',
    reservation_notice        TEXT NOT NULL DEFAULT '',
    cancel_notice             TEXT NOT NULL DEFAULT '',
    privacy_policy            TEXT NOT NULL DEFAULT '',

    -- 基本設定
    closed_weekdays           JSONB NOT NULL DEFAULT '[]',
    closed_dates              JSONB NOT NULL DEFAULT '[]',
    national_holiday_closed   BOOLEAN NOT NULL DEFAULT false,
    business_hours            JSONB NOT NULL DEFAULT '{"start":"0900","end":"1900"}',
    business_hours_by_weekday JSONB,
    break_hours               JSONB NOT NULL DEFAULT '[{"start":"1200","end":"1300"}]',
    daily_limit               INT DEFAULT 1,
    monthly_limit             INT,
    booking_window_max_days   INT NOT NULL DEFAULT 30,
    booking_window_min_days   INT NOT NULL DEFAULT 2,
    calendar_months           INT NOT NULL DEFAULT 2,
    phone_number              TEXT NOT NULL DEFAULT '',
    notification_email        TEXT NOT NULL DEFAULT '',
    request_example           TEXT NOT NULL DEFAULT '',
    time_slot_mode            TEXT NOT NULL DEFAULT 'minimize_gaps',
    time_slot_interval_minutes INT NOT NULL DEFAULT 15,
    no_staff_mode             TEXT NOT NULL DEFAULT 'first_available',
    show_no_staff_option      BOOLEAN NOT NULL DEFAULT true,

    -- 追加入力フィールド定義
    additional_fields         JSONB NOT NULL DEFAULT '[
        {"key":"phone","label":"電話番号","required":true,"placeholder":"例) 090-1234-5678"},
        {"key":"owner_name","label":"飼い主名","required":true,"placeholder":""},
        {"key":"pet_info","label":"ペットの名前と種類","required":true,"placeholder":"例) ポチ（柴犬）"},
        {"key":"symptoms","label":"診察内容","required":false,"placeholder":""}
    ]',

    -- LINE連携
    line_channel_id           TEXT NOT NULL DEFAULT '',
    line_channel_secret       TEXT NOT NULL DEFAULT '',
    liff_id                   TEXT NOT NULL DEFAULT '',
    line_access_token         TEXT NOT NULL DEFAULT '',

    created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- B2. LINE予約顧客
-- ============================================================
CREATE TABLE reservation_customers (
    id                BIGSERIAL PRIMARY KEY,
    clinic_id         BIGINT NOT NULL REFERENCES clinics(id),
    line_user_id      TEXT NOT NULL,
    display_name      TEXT NOT NULL DEFAULT '',
    real_name         TEXT NOT NULL DEFAULT '',
    additional_fields JSONB NOT NULL DEFAULT '{}',
    owner_id          BIGINT REFERENCES owners(id) ON DELETE SET NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(clinic_id, line_user_id)
);
CREATE INDEX idx_res_customers_owner
    ON reservation_customers(owner_id) WHERE owner_id IS NOT NULL;

-- ============================================================
-- B3. スタッフ × 非対応サービス種別（M:N）
-- ============================================================
CREATE TABLE staff_excluded_service_types (
    id              BIGSERIAL PRIMARY KEY,
    staff_id        BIGINT NOT NULL REFERENCES staffs(id) ON DELETE CASCADE,
    service_type_id BIGINT NOT NULL REFERENCES service_types(id) ON DELETE CASCADE,
    UNIQUE(staff_id, service_type_id)
);

-- ============================================================
-- B4. シフト中断時間（shift_entries の子テーブル）
-- ============================================================
CREATE TABLE shift_entry_breaks (
    id             BIGSERIAL PRIMARY KEY,
    shift_entry_id BIGINT NOT NULL REFERENCES shift_entries(id) ON DELETE CASCADE,
    break_start    TIME NOT NULL,
    break_end      TIME NOT NULL
);
CREATE INDEX idx_shift_entry_breaks_entry ON shift_entry_breaks(shift_entry_id);

-- ============================================================
-- A3. reservation_appointments にLINE予約用カラムを追加
-- （reservation_customers が先に存在する必要があるため最後）
-- ============================================================
ALTER TABLE reservation_appointments ADD COLUMN source             reservation_source NOT NULL DEFAULT 'manual';
ALTER TABLE reservation_appointments ADD COLUMN line_customer_id   BIGINT REFERENCES reservation_customers(id) ON DELETE SET NULL;
ALTER TABLE reservation_appointments ADD COLUMN is_staff_delegated BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE reservation_appointments ADD COLUMN customer_fields    JSONB NOT NULL DEFAULT '{}';

CREATE INDEX idx_res_appt_line_customer ON reservation_appointments(line_customer_id)
    WHERE line_customer_id IS NOT NULL AND deleted_at IS NULL;

-- ============================================================
-- B5. 既存テーブルへの制約追加
-- ============================================================
ALTER TABLE shift_entries
    ADD CONSTRAINT uk_shift_staff_date UNIQUE (clinic_id, staff_id, date);
