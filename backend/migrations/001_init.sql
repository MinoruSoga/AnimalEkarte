-- =============================================================================
-- Animal Ekarte - 初期スキーマ定義 v19.0
-- PostgreSQL 18
-- テーブル数: 59
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. 拡張機能
-- -----------------------------------------------------------------------------
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- -----------------------------------------------------------------------------
-- 2. ENUM型定義（全56テーブル対応）
-- -----------------------------------------------------------------------------

-- 認証関連
CREATE TYPE account_status AS ENUM ('active', 'inactive', 'locked');

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
CREATE TYPE examination_status AS ENUM ('pending', 'in_progress', 'result_entered', 'completed', 'confirmed');
CREATE TYPE examination_result_status AS ENUM ('normal', 'high', 'low');
CREATE TYPE next_schedule_type AS ENUM ('3weeks', '4weeks', '1year', 'other');
CREATE TYPE appetite_level AS ENUM ('normal', 'increased', 'decreased', 'none');
CREATE TYPE water_intake_level AS ENUM ('normal', 'increased', 'decreased', 'none');
CREATE TYPE medical_image_type AS ENUM ('xray', 'echo', 'photo', 'endoscope', 'ct', 'mri', 'microscope', 'other');
CREATE TYPE estimate_status AS ENUM ('draft', 'sent', 'approved', 'rejected');
CREATE TYPE billing_review_status AS ENUM ('pending', 'confirmed', 'returned');
CREATE TYPE item_category AS ENUM ('examination', 'test', 'procedure', 'surgery', 'medicine', 'food', 'goods', 'other');
CREATE TYPE item_source AS ENUM ('medical_record', 'manual', 'hospitalization');

-- 予約・会計・入院関連
CREATE TYPE visit_type AS ENUM ('first', 'revisit');
CREATE TYPE reservation_status AS ENUM (
    'confirmed', 'pending', 'cancelled', 'checked_in',
    'in_consultation', 'accounting', 'completed'
);
CREATE TYPE staff_type AS ENUM ('doctor', 'nurse', 'resource');
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
CREATE TYPE trimming_status AS ENUM ('completed', 'reserved', 'in_progress');
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
-- 1. company（シングルトン: 本部情報）
-- ------------------------------------
CREATE TABLE company (
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
    company_id          bigint      NOT NULL REFERENCES company(id),
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
    updated_at  timestamptz NOT NULL DEFAULT now()
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
-- 5b. staffs（スタッフマスタ）
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
    owner_name       text            NOT NULL,
    owner_name_kana  text            NOT NULL DEFAULT '',
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
    created_at       timestamptz     NOT NULL DEFAULT now(),
    updated_at       timestamptz     NOT NULL DEFAULT now(),
    deleted_at       timestamptz
);

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
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- ------------------------------------
-- 10. exam_type_items（検査項目定義）
-- ------------------------------------
CREATE TABLE exam_type_items (
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
    updated_at   timestamptz     NOT NULL DEFAULT now()
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
    updated_at       timestamptz   NOT NULL DEFAULT now()
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
    updated_at    timestamptz NOT NULL DEFAULT now()
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
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- ------------------------------------
-- 15. reservation_category_groups（予約区分グループマスタ）
-- ------------------------------------
CREATE TABLE reservation_category_groups (
    id         BIGSERIAL   PRIMARY KEY,
    clinic_id  bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name       text        NOT NULL,
    color      text        NOT NULL DEFAULT '#3B82F6',
    sort_order integer              DEFAULT 0,
    is_active  boolean     NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_rcg_clinic ON reservation_category_groups(clinic_id);

-- ------------------------------------
-- 16. reservation_categories（予約区分マスタ）
-- ------------------------------------
CREATE TABLE reservation_categories (
    id                       BIGSERIAL   PRIMARY KEY,
    clinic_id                bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name                     text        NOT NULL,
    is_active                boolean     NOT NULL DEFAULT true,
    description              text        NOT NULL DEFAULT '',
    color                    text        NOT NULL DEFAULT '#3B82F6',
    sort_order               integer              DEFAULT 0,
    group_id                 bigint               REFERENCES reservation_category_groups(id) ON DELETE SET NULL,
    reservation_display_name text        NOT NULL DEFAULT '',
    duration_minutes         int         NOT NULL DEFAULT 15,
    short_name               text        NOT NULL DEFAULT '',
    show_short_name          boolean     NOT NULL DEFAULT false,
    reservation_visible      boolean     NOT NULL DEFAULT true,
    reservation_comment      text        NOT NULL DEFAULT '',
    reservation_image_url    text        NOT NULL DEFAULT '',
    reservation_day_option   text        NOT NULL DEFAULT 'none',
    is_internal              boolean     NOT NULL DEFAULT false,
    created_at               timestamptz NOT NULL DEFAULT now(),
    updated_at               timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_reservation_categories_group_id ON reservation_categories(group_id);

-- ------------------------------------
-- 16. consultations（診察項目マスタ）
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
    updated_at     timestamptz NOT NULL DEFAULT now()
);

-- ------------------------------------
-- 17. procedures（処置項目マスタ）
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
    updated_at  timestamptz     NOT NULL DEFAULT now()
);

-- ------------------------------------
-- 18. hospitalization_plans（入院プランマスタ）
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
    updated_at   timestamptz  NOT NULL DEFAULT now()
);

-- ------------------------------------
-- 19. trimming_courses（トリミングコースマスタ）
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
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- ------------------------------------
-- 20. trimming_options（トリミングオプションマスタ）
-- ------------------------------------
CREATE TABLE trimming_options (
    id          BIGSERIAL   PRIMARY KEY,
    clinic_id   bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name        text        NOT NULL,
    price       bigint,
    is_active   boolean     NOT NULL DEFAULT true,
    description text        NOT NULL DEFAULT '',
    duration    integer,
    combinable  boolean     NOT NULL DEFAULT true,
    sort_order  integer              DEFAULT 0,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- ------------------------------------
-- 21. diagnosis_categories（診断カテゴリマスタ）
-- ------------------------------------
CREATE TABLE diagnosis_categories (
    id          BIGSERIAL   PRIMARY KEY,
    clinic_id   bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name        text        NOT NULL,
    is_active   boolean     NOT NULL DEFAULT true,
    description text        NOT NULL DEFAULT '',
    sort_order  integer              DEFAULT 0,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- ------------------------------------
-- 22. diagnosis_names（診断病名マスタ）
-- ------------------------------------
CREATE TABLE diagnosis_names (
    id                    BIGSERIAL   PRIMARY KEY,
    clinic_id             bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name                  text        NOT NULL,
    is_active             boolean     NOT NULL DEFAULT true,
    description           text        NOT NULL DEFAULT '',
    diagnosis_category_id bigint      NOT NULL REFERENCES diagnosis_categories(id) ON DELETE CASCADE,
    sort_order            integer              DEFAULT 0,
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now()
);

-- ------------------------------------
-- 23. checkup_types（健診種別マスタ）
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
-- 24. chief_complaint_categories（主訴区分マスタ）
-- ------------------------------------
CREATE TABLE chief_complaint_categories (
    id          BIGSERIAL   PRIMARY KEY,
    clinic_id   bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name        text        NOT NULL,
    description text        NOT NULL DEFAULT '',
    is_active   boolean     NOT NULL DEFAULT true,
    sort_order  integer              DEFAULT 0,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- ------------------------------------
-- 25. inquiry_templates（問診定型文マスタ）
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
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- ==========================================================================
-- レイヤー3: owners/staffs/animal_species等依存
-- ==========================================================================

-- ------------------------------------
-- 26. pets（ペット情報）
-- ------------------------------------
CREATE TABLE pets (
    id                BIGSERIAL       PRIMARY KEY,
    clinic_id         bigint          NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    owner_id          bigint          NOT NULL REFERENCES owners(id) ON DELETE RESTRICT,
    pet_number        text            NOT NULL DEFAULT '',
    name              text            NOT NULL,
    pet_name_kana     text            NOT NULL DEFAULT '',
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
    created_at        timestamptz     NOT NULL DEFAULT now(),
    updated_at        timestamptz     NOT NULL DEFAULT now(),
    deleted_at        timestamptz
);

-- ------------------------------------
-- 27. user_clinic_memberships（ユーザー所属クリニック）
-- ------------------------------------
-- 6a. staff_clinic_assignments（スタッフ-クリニック中間テーブル）
-- ------------------------------------
CREATE TABLE staff_clinic_assignments (
    id             BIGSERIAL   PRIMARY KEY,
    staff_id       bigint      NOT NULL REFERENCES staffs(id) ON DELETE CASCADE,
    clinic_id      bigint      NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    is_main        boolean     NOT NULL DEFAULT false,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uk_staff_clinic UNIQUE (staff_id, clinic_id)
);

CREATE INDEX idx_staff_clinic_staff ON staff_clinic_assignments(staff_id);
CREATE INDEX idx_staff_clinic_clinic ON staff_clinic_assignments(clinic_id);
CREATE INDEX idx_staff_clinic_main ON staff_clinic_assignments(staff_id, is_main);

-- ------------------------------------
-- 28. permission_groups（権限グループマスタ）
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
CREATE INDEX idx_permission_groups_clinic ON permission_groups(clinic_id) WHERE deleted_at IS NULL;

-- ------------------------------------
-- 28b. permission_group_rules（権限グループ-リソース×CRUD権限）
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
    CONSTRAINT uk_permission_group_rules UNIQUE (group_id, resource)
);

CREATE INDEX idx_permission_group_rules_group ON permission_group_rules(group_id);

-- ------------------------------------
-- 28c. staff_permission_groups（スタッフ-権限グループ中間テーブル）
-- ------------------------------------
CREATE TABLE staff_permission_groups (
    staff_id  bigint NOT NULL REFERENCES staffs(id) ON DELETE CASCADE,
    group_id  bigint NOT NULL REFERENCES permission_groups(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (staff_id, group_id)
);

CREATE INDEX idx_staff_permission_groups_staff ON staff_permission_groups(staff_id);
CREATE INDEX idx_staff_permission_groups_group ON staff_permission_groups(group_id);

-- ==========================================================================
-- レイヤー4: pets依存
-- ==========================================================================

-- ------------------------------------
-- 29. reservation_appointments（予約）
-- ------------------------------------
CREATE TABLE reservation_appointments (
    id                 BIGSERIAL            PRIMARY KEY,
    clinic_id          bigint               NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    start_time         timestamptz          NOT NULL,
    end_time           timestamptz          NOT NULL,
    owner_id           bigint                        REFERENCES owners(id) ON DELETE SET NULL,
    pet_id             bigint                        REFERENCES pets(id) ON DELETE SET NULL,
    visit_type         visit_type           NOT NULL DEFAULT 'revisit',
    reservation_category_id    bigint               NOT NULL REFERENCES reservation_categories(id) ON DELETE RESTRICT,
    doctor_id          bigint                        REFERENCES staffs(id) ON DELETE SET NULL,
    is_designated      boolean                       DEFAULT false,
    status             reservation_status            DEFAULT 'pending',
    notes              text                 NOT NULL DEFAULT '',
    source             reservation_source   NOT NULL DEFAULT 'manual',
    is_staff_delegated boolean              NOT NULL DEFAULT false,
    customer_fields    jsonb                NOT NULL DEFAULT '{}',
    created_at         timestamptz          NOT NULL DEFAULT now(),
    updated_at         timestamptz          NOT NULL DEFAULT now(),
    deleted_at         timestamptz,
    CONSTRAINT chk_reservation_times CHECK (end_time >= start_time)
);

-- ------------------------------------
-- 30. hospitalizations（入院/ホテル管理）
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
-- 31. trimming_records（トリミング記録）
-- ------------------------------------
CREATE TABLE trimming_records (
    id              BIGSERIAL        PRIMARY KEY,
    clinic_id       bigint           NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    date            date             NOT NULL,
    pet_id          bigint                    REFERENCES pets(id) ON DELETE RESTRICT,
    style_request   text             NOT NULL DEFAULT '',
    staff_id        bigint                    REFERENCES staffs(id) ON DELETE SET NULL,
    status          trimming_status           DEFAULT 'reserved',
    course_id       bigint                    REFERENCES trimming_courses(id) ON DELETE SET NULL,
    bw              numeric(6,2),             -- 体重（body weight）
    bw_unit         body_weight_unit          DEFAULT 'Kg',
    bt              numeric(4,1),             -- 体温（body temperature, ℃）
    used_shampoo    text             NOT NULL DEFAULT '',
    used_ribbon     text             NOT NULL DEFAULT '',
    remarks         text             NOT NULL DEFAULT '',
    style_image     text             NOT NULL DEFAULT '',
    completed_image text             NOT NULL DEFAULT '',
    created_at      timestamptz      NOT NULL DEFAULT now(),
    updated_at      timestamptz      NOT NULL DEFAULT now(),
    deleted_at      timestamptz
);

-- ------------------------------------
-- 32. medical_records（電子カルテ）
-- ------------------------------------
CREATE TABLE medical_records (
    id                         BIGSERIAL             PRIMARY KEY,
    clinic_id                  bigint                NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    record_no                  text                  NOT NULL,
    date                       date                  NOT NULL,
    owner_id                   bigint                         REFERENCES owners(id) ON DELETE RESTRICT,
    pet_id                     bigint                         REFERENCES pets(id) ON DELETE RESTRICT,
    doctor_id                  bigint                         REFERENCES staffs(id) ON DELETE SET NULL,
    reservation_appointment_id bigint                         REFERENCES reservation_appointments(id) ON DELETE SET NULL,
    status                     medical_record_status          DEFAULT 'draft',
    version                    INTEGER               NOT NULL DEFAULT 1,
    created_at                 timestamptz           NOT NULL DEFAULT now(),
    updated_at                 timestamptz           NOT NULL DEFAULT now(),
    deleted_at                 timestamptz
);

-- ------------------------------------
-- 33. vaccinations（予防接種記録）
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
-- 34. checkups（定期健診記録）
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
-- 35. exams（検査記録）
-- ------------------------------------
CREATE TABLE exams (
    id                BIGSERIAL          PRIMARY KEY,
    medical_record_id bigint                      REFERENCES medical_records(id) ON DELETE CASCADE,
    clinic_id         bigint             NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    pet_id            bigint                      REFERENCES pets(id) ON DELETE RESTRICT,
    date              date               NOT NULL,
    exam_type_id      bigint             NOT NULL REFERENCES exam_types(id) ON DELETE RESTRICT,
    doctor_id         bigint                      REFERENCES staffs(id) ON DELETE SET NULL,
    status            examination_status          DEFAULT 'pending',
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
-- 36. inquiries（問診タブ: medical_recordsと1:1）
-- ------------------------------------
CREATE TABLE inquiries (
    id                          BIGSERIAL          PRIMARY KEY,
    medical_record_id           bigint             NOT NULL UNIQUE REFERENCES medical_records(id) ON DELETE CASCADE,
    chief_complaint_category_id bigint                      REFERENCES chief_complaint_categories(id) ON DELETE SET NULL,
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
-- 37. clinical_plans（診察/治療タブ: medical_recordsと1:1）
-- ------------------------------------
CREATE TABLE clinical_plans (
    id                    BIGSERIAL   PRIMARY KEY,
    medical_record_id     bigint      NOT NULL UNIQUE REFERENCES medical_records(id) ON DELETE CASCADE,
    physical_exam         text        NOT NULL DEFAULT '',
    diagnosis_category_id bigint               REFERENCES diagnosis_categories(id) ON DELETE SET NULL,
    diagnosis_name_id     bigint               REFERENCES diagnosis_names(id) ON DELETE SET NULL,
    diagnosis_2_category_id bigint             REFERENCES diagnosis_categories(id) ON DELETE SET NULL,
    diagnosis_2_name_id   bigint               REFERENCES diagnosis_names(id) ON DELETE SET NULL,
    diagnosis_details     text        NOT NULL DEFAULT '',
    treatment_policy      text        NOT NULL DEFAULT '',
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now()
);

-- ------------------------------------
-- 38. (vital_records は daily_records 依存のため 45 の後に定義)

-- ------------------------------------
-- 39. treatments（治療明細）
-- ------------------------------------
CREATE TABLE treatments (
    id                BIGSERIAL           PRIMARY KEY,
    medical_record_id bigint              NOT NULL REFERENCES medical_records(id) ON DELETE CASCADE,
    item_type         treatment_item_type NOT NULL DEFAULT 'other',
    consultation_id   bigint                       REFERENCES consultations(id) ON DELETE SET NULL,
    procedure_id      bigint                       REFERENCES procedures(id) ON DELETE SET NULL,
    medicine_id       bigint                       REFERENCES medicines(id) ON DELETE SET NULL,
    selected          boolean                      DEFAULT false,
    status            treatment_status             DEFAULT 'pending',
    content           text                NOT NULL DEFAULT '',
    memo              text                NOT NULL DEFAULT '',
    admin_route       varchar(50)         NOT NULL DEFAULT '',
    insurance         boolean                      DEFAULT false,
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
-- 40. treatment_plans（治療プラン: 外来・入院共用）
-- ------------------------------------
CREATE TABLE treatment_plans (
    id                 BIGSERIAL   PRIMARY KEY,
    medical_record_id  bigint               REFERENCES medical_records(id) ON DELETE CASCADE,
    hospitalization_id bigint               REFERENCES hospitalizations(id) ON DELETE CASCADE,
    treatment_content  text        NOT NULL DEFAULT '',
    memo               text        NOT NULL DEFAULT '',
    insurance          boolean              DEFAULT false,
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
-- 41. record_images（画像タブ）
-- ------------------------------------
CREATE TABLE record_images (
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
-- 42. billing_reviews（会計医師確認タブ: medical_recordsと1:1）
-- ------------------------------------
CREATE TABLE billing_reviews (
    id                BIGSERIAL             PRIMARY KEY,
    medical_record_id bigint                NOT NULL UNIQUE REFERENCES medical_records(id) ON DELETE CASCADE,
    status            billing_review_status          DEFAULT 'pending',
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
-- 43. estimates（見積書）
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
-- 44. exam_items（検査結果明細）
-- ------------------------------------
CREATE TABLE exam_items (
    id                BIGSERIAL                  PRIMARY KEY,
    exam_id           bigint                     NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    exam_type_item_id bigint                              REFERENCES exam_type_items(id) ON DELETE SET NULL,
    name              text                       NOT NULL DEFAULT '',
    inspection_value  text                       NOT NULL DEFAULT '',
    normal_value      text                       NOT NULL DEFAULT '',
    result            text                       NOT NULL DEFAULT '',
    unit              text                       NOT NULL DEFAULT '',
    ref               text                       NOT NULL DEFAULT '',
    ref_min           decimal(10,4),
    ref_max           decimal(10,4),
    is_abnormal       boolean                             DEFAULT false,
    status            examination_result_status           DEFAULT 'normal',
    sort_order        integer                             DEFAULT 0,
    created_at        timestamptz                NOT NULL DEFAULT now(),
    updated_at        timestamptz                NOT NULL DEFAULT now()
);

-- ------------------------------------
-- 45. daily_records（入院日次記録）
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
-- 45b. vital_records（バイタル記録: 外来・入院統合）
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
-- 46. care_plan_items（ケアプラン項目）
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
-- 47. estimate_items（見積明細）
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
    merchandise_item_id     bigint,
    sort_order              integer                DEFAULT 0,
    created_at              timestamptz   NOT NULL DEFAULT now(),
    updated_at              timestamptz   NOT NULL DEFAULT now(),
    CONSTRAINT chk_estimate_item_quantity CHECK (quantity > 0)
);

-- ------------------------------------
-- 48. care_log_records（ケアログ）
-- ------------------------------------
CREATE TABLE care_log_records (
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
-- 49. (vital_records は 38 に統合済み)

-- ------------------------------------
-- 50. staff_note_records（スタッフノート）
-- ------------------------------------
CREATE TABLE staff_note_records (
    id              BIGSERIAL   PRIMARY KEY,
    daily_record_id bigint      NOT NULL REFERENCES daily_records(id) ON DELETE CASCADE,
    time            time        NOT NULL,
    content         text        NOT NULL DEFAULT '',
    staff_id        bigint               REFERENCES staffs(id) ON DELETE SET NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

-- ------------------------------------
-- 51. trimming_record_options（トリミングオプション適用）
-- ------------------------------------
CREATE TABLE trimming_record_options (
    id                 BIGSERIAL PRIMARY KEY,
    trimming_record_id bigint    NOT NULL REFERENCES trimming_records(id) ON DELETE CASCADE,
    option_id          bigint    NOT NULL REFERENCES trimming_options(id) ON DELETE RESTRICT,
    sort_order         integer            DEFAULT 0,
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now()
);

-- ==========================================================================
-- レイヤー7: billings
-- ==========================================================================

-- ------------------------------------
-- 52. billings（会計）
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
-- 53. billing_items（会計明細）
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
    merchandise_item_id     bigint,
    sort_order              integer                DEFAULT 0,
    created_at              timestamptz   NOT NULL DEFAULT now(),
    updated_at              timestamptz   NOT NULL DEFAULT now(),
    deleted_at              timestamptz,
    CONSTRAINT chk_billing_item_quantity CHECK (quantity > 0)
);

-- ------------------------------------
-- 54. payments（支払い: billingsと1:1）
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
    created_at       timestamptz    NOT NULL DEFAULT now(),
    updated_at       timestamptz    NOT NULL DEFAULT now(),
    deleted_at       timestamptz
);

-- ------------------------------------
-- 55. billing_refunds（返金レコード）
-- ------------------------------------
CREATE TABLE billing_refunds (
    id           BIGSERIAL   PRIMARY KEY,
    clinic_id    bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    billing_id   bigint      NOT NULL REFERENCES billings(id) ON DELETE CASCADE,
    amount       bigint      NOT NULL CHECK (amount > 0),
    reason       text        NOT NULL DEFAULT '',
    refunded_at  timestamptz NOT NULL DEFAULT now(),
    created_at   timestamptz NOT NULL DEFAULT now()
);

-- ------------------------------------
-- 56. shift_entries（シフト管理）
-- ------------------------------------
CREATE TABLE shift_entries (
    id         BIGSERIAL   PRIMARY KEY,
    clinic_id  bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    staff_id   bigint      NOT NULL REFERENCES staffs(id) ON DELETE RESTRICT,
    date       date        NOT NULL,
    shift_type shift_type  NOT NULL,
    start_time time,
    end_time   time,
    note       text        NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uk_shift_staff_date UNIQUE (clinic_id, staff_id, date)
);

-- ------------------------------------
-- 57. merchandise_items（物販・フード・その他マスタ）
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

-- merchandise_item_id FK（merchandise_items テーブル作成後に ALTER TABLE で追加）
ALTER TABLE billing_items ADD CONSTRAINT fk_billing_items_merchandise
    FOREIGN KEY (merchandise_item_id) REFERENCES merchandise_items(id) ON DELETE SET NULL;
CREATE INDEX idx_billing_items_merchandise_item_id ON billing_items(merchandise_item_id) WHERE deleted_at IS NULL;

ALTER TABLE estimate_items ADD CONSTRAINT fk_estimate_items_merchandise
    FOREIGN KEY (merchandise_item_id) REFERENCES merchandise_items(id) ON DELETE SET NULL;
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

-- Deleted: Indexes for user_clinic_memberships and permission_groups
-- These were removed in the auth refactor (Account-based authentication)
-- New indexes for staff_clinic_assignments are defined in section 4.2

-- 飼主: clinic内でemail重複不可（論理削除を除く・空文字除く）
CREATE UNIQUE INDEX uk_owners_clinic_email ON owners(clinic_id, email) WHERE deleted_at IS NULL AND email IS NOT NULL AND email != '';

-- Deleted: user_permission_groups index (auth refactor)

-- トリミングオプション: 重複防止
CREATE UNIQUE INDEX idx_trimming_record_options_unique ON trimming_record_options(trimming_record_id, option_id);

-- billings: medical_record_idがある場合は1対1
CREATE UNIQUE INDEX idx_billings_medical_record_id_unique ON billings(medical_record_id) WHERE medical_record_id IS NOT NULL;

-- -----------------------------------------------------------------------------
-- 4.4 基本FKインデックス
-- -----------------------------------------------------------------------------

-- マスタテーブル clinic_id
-- Deleted: idx_staffs_clinic_id (staffs now uses account_id; clinic membership tracked via staff_clinic_assignments)
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
CREATE INDEX idx_reservation_categories_clinic_id ON reservation_categories(clinic_id);
CREATE INDEX idx_consultations_clinic_id ON consultations(clinic_id);
CREATE INDEX idx_consultations_parent_id ON consultations(parent_id);
CREATE INDEX idx_procedures_clinic_id ON procedures(clinic_id);
CREATE INDEX idx_procedures_parent_id ON procedures(parent_id);
CREATE INDEX idx_hospitalization_plans_clinic_id ON hospitalization_plans(clinic_id);
CREATE INDEX idx_trimming_courses_clinic_id ON trimming_courses(clinic_id);
CREATE INDEX idx_trimming_options_clinic_id ON trimming_options(clinic_id);
CREATE INDEX idx_diagnosis_categories_clinic_id ON diagnosis_categories(clinic_id);
CREATE INDEX idx_diagnosis_names_clinic_id ON diagnosis_names(clinic_id);
CREATE INDEX idx_checkup_types_clinic_id ON checkup_types(clinic_id);
CREATE INDEX idx_checkup_types_parent_id ON checkup_types(parent_id);
CREATE INDEX idx_checkup_types_deleted_at ON checkup_types(deleted_at);
CREATE INDEX idx_chief_complaint_categories_clinic_id ON chief_complaint_categories(clinic_id);
CREATE INDEX idx_inquiry_templates_clinic_id ON inquiry_templates(clinic_id);
CREATE INDEX idx_inquiry_templates_clinic_category ON inquiry_templates(clinic_id, category);

-- コアテーブル clinic_id
CREATE INDEX idx_owners_clinic_id ON owners(clinic_id);
CREATE INDEX idx_pets_clinic_id ON pets(clinic_id);

-- 診療テーブル clinic_id
CREATE INDEX idx_medical_records_clinic_id ON medical_records(clinic_id);
CREATE INDEX idx_reservation_appointments_clinic_id ON reservation_appointments(clinic_id);
CREATE INDEX idx_hospitalizations_clinic_id ON hospitalizations(clinic_id);
CREATE INDEX idx_trimming_records_clinic_id ON trimming_records(clinic_id);
CREATE INDEX idx_billings_clinic_id ON billings(clinic_id);
CREATE INDEX idx_shift_entries_clinic_id ON shift_entries(clinic_id);
CREATE INDEX idx_estimates_clinic_id ON estimates(clinic_id);

-- merchandise_items インデックス
CREATE INDEX idx_merchandise_items_clinic ON merchandise_items(clinic_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_merchandise_items_category ON merchandise_items(clinic_id, category) WHERE deleted_at IS NULL;
CREATE INDEX idx_merchandise_items_sort ON merchandise_items(clinic_id, sort_order);

-- 予約 FK インデックス
CREATE INDEX idx_reservation_appointments_owner_id ON reservation_appointments(owner_id);
CREATE INDEX idx_reservation_appointments_pet_id ON reservation_appointments(pet_id);
CREATE INDEX idx_reservation_appointments_reservation_category_id ON reservation_appointments(reservation_category_id);
CREATE INDEX idx_reservation_appointments_doctor_id ON reservation_appointments(doctor_id);

-- medical_records 子テーブル FK インデックス
CREATE INDEX idx_treatments_medical_record_id ON treatments(medical_record_id);
CREATE INDEX idx_vital_records_medical_record_id ON vital_records(medical_record_id);
CREATE INDEX idx_vital_records_daily_record_id ON vital_records(daily_record_id);
CREATE INDEX idx_vital_records_pet_id ON vital_records(pet_id);
CREATE INDEX idx_exams_medical_record_id ON exams(medical_record_id);
CREATE INDEX idx_exams_pet_id ON exams(pet_id);
CREATE INDEX idx_vaccinations_clinic_id ON vaccinations(clinic_id);
CREATE INDEX idx_vaccinations_medical_record_id ON vaccinations(medical_record_id);
CREATE INDEX idx_vaccinations_pet_id ON vaccinations(pet_id);
CREATE INDEX idx_checkups_medical_record_id ON checkups(medical_record_id);
CREATE INDEX idx_checkups_pet_id ON checkups(pet_id);
CREATE INDEX idx_clinical_plans_medical_record_id ON clinical_plans(medical_record_id);
CREATE INDEX idx_inquiries_medical_record_id ON inquiries(medical_record_id);
CREATE INDEX idx_record_images_medical_record_id ON record_images(medical_record_id);
CREATE INDEX idx_treatment_plans_medical_record_id ON treatment_plans(medical_record_id);
CREATE INDEX idx_treatment_plans_hospitalization_id ON treatment_plans(hospitalization_id);

-- hospitalization 子テーブル FK インデックス
CREATE INDEX idx_hospitalizations_pet_id ON hospitalizations(pet_id);
CREATE INDEX idx_hospitalizations_owner_id ON hospitalizations(owner_id);
CREATE INDEX idx_hospitalizations_cage_id ON hospitalizations(cage_id);
CREATE INDEX idx_care_plan_items_hospitalization_id ON care_plan_items(hospitalization_id);
CREATE INDEX idx_daily_records_hospitalization_id ON daily_records(hospitalization_id);

-- billing 子テーブル FK インデックス
CREATE INDEX idx_billing_items_billing_id ON billing_items(billing_id);
CREATE INDEX idx_billing_items_deleted_at ON billing_items(deleted_at);
CREATE INDEX idx_billings_pet_id ON billings(pet_id);
CREATE INDEX idx_billings_owner_id ON billings(owner_id);
CREATE INDEX idx_billings_medical_record_id ON billings(medical_record_id);

CREATE INDEX idx_billing_refunds_billing ON billing_refunds(billing_id);
CREATE INDEX idx_billing_refunds_clinic_billing ON billing_refunds(clinic_id, billing_id);

-- 担当医 FK インデックス（staffs）
CREATE INDEX idx_vital_records_staff_id ON vital_records(staff_id);
CREATE INDEX idx_trimming_records_staff_id ON trimming_records(staff_id);

-- record_images インデックス
CREATE INDEX idx_record_images_image_type ON record_images(image_type);
CREATE INDEX idx_record_images_taken_at ON record_images(taken_at DESC);
CREATE INDEX idx_record_images_exam_id ON record_images(exam_id) WHERE exam_id IS NOT NULL;

-- estimates インデックス
CREATE INDEX idx_estimates_medical_record_id ON estimates(medical_record_id);
CREATE INDEX idx_estimates_status ON estimates(status);
CREATE INDEX idx_estimates_owner_id ON estimates(owner_id);

-- estimate_items インデックス
CREATE INDEX idx_estimate_items_estimate_id ON estimate_items(estimate_id);

-- billing_reviews インデックス
CREATE INDEX idx_billing_reviews_status ON billing_reviews(status);

-- -----------------------------------------------------------------------------
-- 4.5 全文検索インデックス（pg_trgm GIN）
-- -----------------------------------------------------------------------------
CREATE INDEX idx_owners_name_trgm ON owners USING gin (owner_name gin_trgm_ops) WHERE deleted_at IS NULL;
CREATE INDEX idx_owners_name_kana_trgm ON owners USING gin (owner_name_kana gin_trgm_ops) WHERE deleted_at IS NULL;
CREATE INDEX idx_pets_name_trgm ON pets USING gin (name gin_trgm_ops) WHERE deleted_at IS NULL;

-- -----------------------------------------------------------------------------
-- 4.6 パフォーマンス最適化インデックス（論理削除考慮）
-- -----------------------------------------------------------------------------

-- ダッシュボード・カレンダー（最高頻度）
CREATE INDEX idx_reservation_appointments_clinic_date
  ON reservation_appointments(clinic_id, start_time)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_reservation_appointments_clinic_status
  ON reservation_appointments(clinic_id, status)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_reservation_appointments_pet_date
  ON reservation_appointments(pet_id, start_time)
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

-- トリミング一覧
CREATE INDEX idx_trimming_records_clinic_date
  ON trimming_records(clinic_id, date DESC)
  WHERE deleted_at IS NULL;

-- BE-033: 追加インデックス（検索パフォーマンス改善）
CREATE INDEX idx_owners_phone_trgm ON owners USING gin (phone gin_trgm_ops) WHERE deleted_at IS NULL;
CREATE INDEX idx_inventory_items_category ON inventory_items(category) WHERE deleted_at IS NULL;

-- 追加FKインデックス
-- Deleted: idx_user_accounts_staff_id, idx_user_accounts_job_title_id (user_accounts table removed)
CREATE INDEX idx_staffs_occupation_id ON staffs(occupation_id);
CREATE INDEX idx_pets_animal_species_id ON pets(animal_species_id);
CREATE INDEX idx_pets_insurance_id ON pets(insurance_id) WHERE insurance_id IS NOT NULL;
CREATE INDEX idx_diagnosis_names_category_id ON diagnosis_names(diagnosis_category_id);
CREATE INDEX idx_medical_records_doctor_id ON medical_records(doctor_id) WHERE doctor_id IS NOT NULL;
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
-- Deleted: idx_staffs_clinic_name (staffs no longer has clinic_id directly)
CREATE UNIQUE INDEX idx_exam_types_clinic_name ON exam_types(clinic_id, name) WHERE is_active = true;
CREATE UNIQUE INDEX idx_vaccines_clinic_name ON vaccines(clinic_id, name) WHERE is_active = true;
CREATE UNIQUE INDEX idx_medicines_clinic_name ON medicines(clinic_id, name) WHERE is_active = true;
CREATE UNIQUE INDEX idx_consultations_clinic_name ON consultations(clinic_id, name) WHERE is_active = true;
CREATE UNIQUE INDEX idx_procedures_clinic_name ON procedures(clinic_id, name) WHERE is_active = true;
CREATE UNIQUE INDEX idx_cages_clinic_name ON cages(clinic_id, name) WHERE is_active = true;
CREATE UNIQUE INDEX idx_reservation_categories_clinic_name ON reservation_categories(clinic_id, name) WHERE is_active = true;
CREATE UNIQUE INDEX idx_diagnosis_categories_clinic_name ON diagnosis_categories(clinic_id, name) WHERE is_active = true;
CREATE UNIQUE INDEX idx_trimming_courses_clinic_name ON trimming_courses(clinic_id, name) WHERE is_active = true;
CREATE UNIQUE INDEX idx_trimming_options_clinic_name ON trimming_options(clinic_id, name) WHERE is_active = true;
CREATE UNIQUE INDEX idx_insurance_clinic_name ON insurances(clinic_id, name) WHERE is_active = true;
CREATE UNIQUE INDEX idx_checkup_types_clinic_name ON checkup_types(clinic_id, name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_hospitalization_plans_clinic_name ON hospitalization_plans(clinic_id, name) WHERE is_active = true;
CREATE UNIQUE INDEX idx_occupations_clinic_name ON occupations(clinic_id, name) WHERE is_active = true;
CREATE UNIQUE INDEX idx_chief_complaint_categories_clinic_name ON chief_complaint_categories(clinic_id, name) WHERE is_active = true;
CREATE UNIQUE INDEX idx_animal_species_name ON animal_species(name) WHERE is_active = true;
CREATE UNIQUE INDEX idx_merchandise_items_clinic_name ON merchandise_items(clinic_id, name) WHERE is_active = true AND deleted_at IS NULL;

-- =============================================================================
-- 5. テーブルコメント
-- =============================================================================

COMMENT ON TABLE company IS '法人情報（シングルトン）';
COMMENT ON TABLE clinics IS '医院情報';
COMMENT ON TABLE animal_species IS 'ペット種類マスタ（システム共通）';
COMMENT ON TABLE accounts IS '認証用アカウント';
COMMENT ON TABLE occupations IS '職種マスタ';
COMMENT ON TABLE staffs IS 'スタッフマスタ';
COMMENT ON TABLE owners IS '飼主情報';
COMMENT ON TABLE inventory_items IS '在庫アイテム';
COMMENT ON TABLE exam_types IS '検査種別マスタ';
COMMENT ON TABLE exam_type_items IS '検査項目定義マスタ';
COMMENT ON TABLE vaccines IS 'ワクチンマスタ';
COMMENT ON TABLE medicines IS '薬剤マスタ';
COMMENT ON TABLE insurances IS '保険マスタ';
COMMENT ON TABLE cages IS 'ケージマスタ';
COMMENT ON TABLE reservation_categories IS '予約区分マスタ';
COMMENT ON TABLE consultations IS '診察項目マスタ';
COMMENT ON TABLE procedures IS '処置項目マスタ';
COMMENT ON TABLE hospitalization_plans IS '入院プランマスタ';
COMMENT ON TABLE trimming_courses IS 'トリミングコースマスタ';
COMMENT ON TABLE trimming_options IS 'トリミングオプションマスタ';
COMMENT ON TABLE diagnosis_categories IS '診断カテゴリマスタ';
COMMENT ON TABLE diagnosis_names IS '診断病名マスタ';
COMMENT ON TABLE checkup_types IS '健診種別マスタ';
COMMENT ON TABLE chief_complaint_categories IS '主訴区分マスタ';
COMMENT ON TABLE inquiry_templates IS '問診定型文マスタ';
COMMENT ON TABLE pets IS 'ペット情報';
-- Deleted: COMMENT ON TABLE user_clinic_memberships, permission_groups, permission_group_rules, user_permission_groups
COMMENT ON TABLE reservation_appointments IS '予約';
COMMENT ON TABLE hospitalizations IS '入院・ホテル管理';
COMMENT ON TABLE trimming_records IS 'トリミング記録';
COMMENT ON TABLE medical_records IS '電子カルテ（診療記録）';
COMMENT ON TABLE vaccinations IS 'ワクチン接種記録';
COMMENT ON TABLE checkups IS '定期健診記録';
COMMENT ON TABLE exams IS '検査記録';
COMMENT ON TABLE inquiries IS '問診情報';
COMMENT ON TABLE clinical_plans IS '診察所見・診断・治療方針';
COMMENT ON TABLE vital_records IS 'バイタル記録（外来・入院統合）';
COMMENT ON TABLE treatments IS '治療明細（処置・診察・薬剤）';
COMMENT ON TABLE treatment_plans IS '治療プラン（外来・入院共用）';
COMMENT ON TABLE record_images IS '診療画像';
COMMENT ON TABLE billing_reviews IS '会計医師確認';
COMMENT ON TABLE estimates IS '見積書';
COMMENT ON TABLE exam_items IS '検査結果項目';
COMMENT ON TABLE daily_records IS '入院日次記録';
COMMENT ON TABLE care_plan_items IS 'ケアプラン項目';
COMMENT ON TABLE estimate_items IS '見積書明細';
COMMENT ON TABLE care_log_records IS 'ケアログ';
COMMENT ON TABLE staff_note_records IS 'スタッフノート';
COMMENT ON TABLE trimming_record_options IS 'トリミングオプション適用';
COMMENT ON TABLE billings IS '会計';
COMMENT ON TABLE billing_items IS '会計明細';
COMMENT ON TABLE payments IS '支払い情報';
COMMENT ON TABLE billing_refunds IS '返金レコード（Stripe モデル）';
COMMENT ON TABLE shift_entries IS 'スタッフシフト';
COMMENT ON TABLE merchandise_items IS '物販・フード・その他マスタ';

-- ------------------------------------
-- 50. audit_logs（権限変更・認証操作の監査ログ）
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
-- レイヤー8: LINE予約システム
-- ==========================================================================

-- ------------------------------------
-- 63. reservation_settings（LINE予約基本設定 — クリニック単位 1:1）
-- ------------------------------------
CREATE TABLE reservation_settings (
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

-- ------------------------------------
-- 64. reservation_customers（LINE予約顧客）
-- ------------------------------------
CREATE TABLE reservation_customers (
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
CREATE INDEX idx_res_customers_owner
    ON reservation_customers(owner_id) WHERE owner_id IS NOT NULL;

-- ------------------------------------
-- 65. staff_excluded_reservation_categories（スタッフ × 非対応予約区分 M:N）
-- ------------------------------------
CREATE TABLE staff_excluded_reservation_categories (
    id              BIGSERIAL PRIMARY KEY,
    staff_id        bigint    NOT NULL REFERENCES staffs(id) ON DELETE CASCADE,
    reservation_category_id bigint    NOT NULL REFERENCES reservation_categories(id) ON DELETE CASCADE,
    UNIQUE(staff_id, reservation_category_id)
);

-- ------------------------------------
-- 66. shift_entry_breaks（シフト中断時間 — shift_entries の子テーブル）
-- ------------------------------------
CREATE TABLE shift_entry_breaks (
    id             BIGSERIAL PRIMARY KEY,
    shift_entry_id bigint    NOT NULL REFERENCES shift_entries(id) ON DELETE CASCADE,
    break_start    time      NOT NULL,
    break_end      time      NOT NULL
);
CREATE INDEX idx_shift_entry_breaks_entry ON shift_entry_breaks(shift_entry_id);

-- ------------------------------------
-- reservation_appointments に line_customer_id FK を追加
-- （reservation_customers テーブル作成後に実行）
-- ------------------------------------
ALTER TABLE reservation_appointments
    ADD COLUMN line_customer_id bigint REFERENCES reservation_customers(id) ON DELETE SET NULL;

CREATE INDEX idx_res_appt_line_customer ON reservation_appointments(line_customer_id)
    WHERE line_customer_id IS NOT NULL AND deleted_at IS NULL;

-- ------------------------------------
-- 予約時間枠の重複防止（部分ユニークインデックス）
-- ------------------------------------
CREATE UNIQUE INDEX uk_appointment_staff_time
    ON reservation_appointments (clinic_id, doctor_id, start_time)
    WHERE deleted_at IS NULL AND status != 'cancelled';

