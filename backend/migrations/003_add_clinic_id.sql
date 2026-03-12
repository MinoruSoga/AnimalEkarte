-- 003_add_clinic_id.sql
-- 医院ごとの独立管理に対応するため、clinic_id を追加する
-- 対象: owners, staffs, inventory_items, 全マスタテーブル（15種）

-- ============================================================
-- Step 1: clinic_id カラム追加（nullable）
-- ============================================================

ALTER TABLE owners               ADD COLUMN IF NOT EXISTS clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE;
ALTER TABLE staffs               ADD COLUMN IF NOT EXISTS clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE;
ALTER TABLE inventory_items      ADD COLUMN IF NOT EXISTS clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE;

-- マスタテーブル
ALTER TABLE cages                ADD COLUMN IF NOT EXISTS clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE;
ALTER TABLE service_types        ADD COLUMN IF NOT EXISTS clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE;
ALTER TABLE consultations        ADD COLUMN IF NOT EXISTS clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE;
ALTER TABLE procedures           ADD COLUMN IF NOT EXISTS clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE;
ALTER TABLE hospitalization_plans ADD COLUMN IF NOT EXISTS clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE;
ALTER TABLE trimming_courses     ADD COLUMN IF NOT EXISTS clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE;
ALTER TABLE trimming_options     ADD COLUMN IF NOT EXISTS clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE;
ALTER TABLE examination_types    ADD COLUMN IF NOT EXISTS clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE;
ALTER TABLE vaccines             ADD COLUMN IF NOT EXISTS clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE;
ALTER TABLE medicines            ADD COLUMN IF NOT EXISTS clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE;
ALTER TABLE insurances           ADD COLUMN IF NOT EXISTS clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE;
ALTER TABLE diagnosis_categories ADD COLUMN IF NOT EXISTS clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE;
ALTER TABLE checkup_types        ADD COLUMN IF NOT EXISTS clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE;

-- ============================================================
-- Step 2: 既存シードデータを最初の医院に紐付け
-- （開発環境初期化時に002_seed_master.sql で clinics が挿入済み前提）
-- ============================================================

DO $$
DECLARE
    v_clinic_id UUID;
BEGIN
    SELECT id INTO v_clinic_id FROM clinics ORDER BY created_at LIMIT 1;

    IF v_clinic_id IS NULL THEN
        RAISE EXCEPTION '003_add_clinic_id: clinics テーブルにデータがありません。002_seed_master.sql を先に実行してください。';
    END IF;

    UPDATE owners               SET clinic_id = v_clinic_id WHERE clinic_id IS NULL;
    UPDATE staffs               SET clinic_id = v_clinic_id WHERE clinic_id IS NULL;
    UPDATE inventory_items      SET clinic_id = v_clinic_id WHERE clinic_id IS NULL;
    UPDATE cages                SET clinic_id = v_clinic_id WHERE clinic_id IS NULL;
    UPDATE service_types        SET clinic_id = v_clinic_id WHERE clinic_id IS NULL;
    UPDATE consultations        SET clinic_id = v_clinic_id WHERE clinic_id IS NULL;
    UPDATE procedures           SET clinic_id = v_clinic_id WHERE clinic_id IS NULL;
    UPDATE hospitalization_plans SET clinic_id = v_clinic_id WHERE clinic_id IS NULL;
    UPDATE trimming_courses     SET clinic_id = v_clinic_id WHERE clinic_id IS NULL;
    UPDATE trimming_options     SET clinic_id = v_clinic_id WHERE clinic_id IS NULL;
    UPDATE examination_types    SET clinic_id = v_clinic_id WHERE clinic_id IS NULL;
    UPDATE vaccines             SET clinic_id = v_clinic_id WHERE clinic_id IS NULL;
    UPDATE medicines            SET clinic_id = v_clinic_id WHERE clinic_id IS NULL;
    UPDATE insurances           SET clinic_id = v_clinic_id WHERE clinic_id IS NULL;
    UPDATE diagnosis_categories SET clinic_id = v_clinic_id WHERE clinic_id IS NULL;
    UPDATE checkup_types        SET clinic_id = v_clinic_id WHERE clinic_id IS NULL;
END $$;

-- ============================================================
-- Step 3: NOT NULL 制約追加
-- ============================================================

ALTER TABLE owners               ALTER COLUMN clinic_id SET NOT NULL;
ALTER TABLE staffs               ALTER COLUMN clinic_id SET NOT NULL;
ALTER TABLE inventory_items      ALTER COLUMN clinic_id SET NOT NULL;
ALTER TABLE cages                ALTER COLUMN clinic_id SET NOT NULL;
ALTER TABLE service_types        ALTER COLUMN clinic_id SET NOT NULL;
ALTER TABLE consultations        ALTER COLUMN clinic_id SET NOT NULL;
ALTER TABLE procedures           ALTER COLUMN clinic_id SET NOT NULL;
ALTER TABLE hospitalization_plans ALTER COLUMN clinic_id SET NOT NULL;
ALTER TABLE trimming_courses     ALTER COLUMN clinic_id SET NOT NULL;
ALTER TABLE trimming_options     ALTER COLUMN clinic_id SET NOT NULL;
ALTER TABLE examination_types    ALTER COLUMN clinic_id SET NOT NULL;
ALTER TABLE vaccines             ALTER COLUMN clinic_id SET NOT NULL;
ALTER TABLE medicines            ALTER COLUMN clinic_id SET NOT NULL;
ALTER TABLE insurances           ALTER COLUMN clinic_id SET NOT NULL;
ALTER TABLE diagnosis_categories ALTER COLUMN clinic_id SET NOT NULL;
ALTER TABLE checkup_types        ALTER COLUMN clinic_id SET NOT NULL;

-- ============================================================
-- Step 4: インデックス追加（検索パフォーマンス）
-- ============================================================

CREATE INDEX IF NOT EXISTS idx_owners_clinic_id               ON owners               (clinic_id);
CREATE INDEX IF NOT EXISTS idx_staffs_clinic_id               ON staffs               (clinic_id);
CREATE INDEX IF NOT EXISTS idx_inventory_items_clinic_id      ON inventory_items      (clinic_id);
CREATE INDEX IF NOT EXISTS idx_cages_clinic_id                ON cages                (clinic_id);
CREATE INDEX IF NOT EXISTS idx_service_types_clinic_id        ON service_types        (clinic_id);
CREATE INDEX IF NOT EXISTS idx_consultations_clinic_id        ON consultations        (clinic_id);
CREATE INDEX IF NOT EXISTS idx_procedures_clinic_id           ON procedures           (clinic_id);
CREATE INDEX IF NOT EXISTS idx_hospitalization_plans_clinic_id ON hospitalization_plans (clinic_id);
CREATE INDEX IF NOT EXISTS idx_trimming_courses_clinic_id     ON trimming_courses     (clinic_id);
CREATE INDEX IF NOT EXISTS idx_trimming_options_clinic_id     ON trimming_options     (clinic_id);
CREATE INDEX IF NOT EXISTS idx_examination_types_clinic_id    ON examination_types    (clinic_id);
CREATE INDEX IF NOT EXISTS idx_vaccines_clinic_id             ON vaccines             (clinic_id);
CREATE INDEX IF NOT EXISTS idx_medicines_clinic_id            ON medicines            (clinic_id);
CREATE INDEX IF NOT EXISTS idx_insurances_clinic_id           ON insurances           (clinic_id);
CREATE INDEX IF NOT EXISTS idx_diagnosis_categories_clinic_id ON diagnosis_categories (clinic_id);
CREATE INDEX IF NOT EXISTS idx_checkup_types_clinic_id        ON checkup_types        (clinic_id);
