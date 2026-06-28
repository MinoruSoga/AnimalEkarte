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
