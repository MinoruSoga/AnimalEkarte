-- Lステップ 動的タグ設定テーブル群
-- B / C1 / C2 / C3 カテゴリのプレフィックスをコード固定から DB 管理へ移行

-- ----------------------------------------------------------------
-- 1. 自動管理プレフィックス (B / C1 / C2 / C3)
-- ----------------------------------------------------------------
CREATE TABLE lstep_auto_managed_prefixes (
    id          BIGSERIAL    PRIMARY KEY,
    prefix      VARCHAR(100) NOT NULL UNIQUE,
    category    VARCHAR(20)  NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_lstep_auto_managed_prefixes_category ON lstep_auto_managed_prefixes (category);

-- ----------------------------------------------------------------
-- 2. 慢性疾患コード → タグ名マッピング
-- ----------------------------------------------------------------
CREATE TABLE lstep_condition_tag_mappings (
    id             BIGSERIAL    PRIMARY KEY,
    condition_code VARCHAR(50)  NOT NULL UNIQUE,
    tag_name       VARCHAR(100) NOT NULL,
    description    TEXT,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ----------------------------------------------------------------
-- 3. LINE 送信目的 → タグプレフィックスマッピング
-- ----------------------------------------------------------------
CREATE TABLE lstep_send_purpose_tag_prefixes (
    id          BIGSERIAL    PRIMARY KEY,
    purpose     VARCHAR(100) NOT NULL UNIQUE,
    tag_prefix  VARCHAR(100) NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ----------------------------------------------------------------
-- Seed: B カテゴリ（予約・キャンセル・未来院）
-- ----------------------------------------------------------------
INSERT INTO lstep_auto_managed_prefixes (prefix, category, description) VALUES
    ('reserved_',      'B', '予約タグ'),
    ('canceled_visit', 'B', 'キャンセルタグ'),
    ('no_show_',       'B', '未来院タグ');

-- ----------------------------------------------------------------
-- Seed: C1 カテゴリ（ペット基本情報）
-- ----------------------------------------------------------------
INSERT INTO lstep_auto_managed_prefixes (prefix, category, description) VALUES
    ('breed_',         'C1', '犬種・猫種タグ'),
    ('sex_',           'C1', '性別タグ'),
    ('pet_birthday_',  'C1', '誕生日タグ'),
    ('birth_year_',    'C1', '生年タグ'),
    ('spay_neutered',  'C1', '避妊去勢済みタグ'),
    ('intact',         'C1', '未避妊去勢タグ');

-- ----------------------------------------------------------------
-- Seed: C2 カテゴリ（診療・ケアスケジュール）
-- ----------------------------------------------------------------
INSERT INTO lstep_auto_managed_prefixes (prefix, category, description) VALUES
    ('vaccine_',       'C2', 'ワクチンタグ'),
    ('checkup_',       'C2', '健診タグ'),
    ('next_checkup_',  'C2', '次回健診タグ'),
    ('refill_due_',    'C2', '処方箋補充期限タグ'),
    ('next_visit_',    'C2', '次回来院タグ'),
    ('chronic_',       'C2', '慢性疾患タグ');

-- ----------------------------------------------------------------
-- Seed: C3 カテゴリ（証明書・術後・退院・死亡）
-- ----------------------------------------------------------------
INSERT INTO lstep_auto_managed_prefixes (prefix, category, description) VALUES
    ('cert_sent_',         'C3', '証明書送付タグ'),
    ('post_surgery_',      'C3', '術後フォロータグ'),
    ('post_discharge_',    'C3', '退院フォロータグ'),
    ('pet_deceased_',      'C3', 'ペット死亡タグ');

-- ----------------------------------------------------------------
-- Seed: 慢性疾患コードマッピング
-- ----------------------------------------------------------------
INSERT INTO lstep_condition_tag_mappings (condition_code, tag_name, description) VALUES
    ('ckd',      'chronic_ckd',      '慢性腎臓病'),
    ('heart',    'chronic_heart',    '心臓病'),
    ('skin',     'chronic_skin',     '皮膚疾患'),
    ('diabetes', 'chronic_diabetes', '糖尿病'),
    ('liver',    'chronic_liver',    '肝疾患'),
    ('thyroid',  'chronic_thyroid',  '甲状腺疾患'),
    ('other',    'chronic_other',    'その他慢性疾患');

-- ----------------------------------------------------------------
-- Seed: LINE 送信目的タグプレフィックス
-- ----------------------------------------------------------------
INSERT INTO lstep_send_purpose_tag_prefixes (purpose, tag_prefix, description) VALUES
    ('vaccine_cert',       'cert_sent_vaccine_',        'ワクチン証明書送付'),
    ('inspection_result',  'cert_sent_inspection_',     '検査結果送付'),
    ('post_surgery',       'post_surgery_followup_',    '術後フォローアップ'),
    ('post_discharge',     'post_discharge_followup_',  '退院フォローアップ');
