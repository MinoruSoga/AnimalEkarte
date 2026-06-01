-- 010: トリミングコース種別マスタを追加する (issue #73)。
--
-- 背景: トリミングコースを「シャンプー」「シャンプー＆カット」等の種別で整理する。
-- payment_methods (FEAT-368) と同型の拡張可能マスタとし、病院ごとに種別を追加できる
-- ようにする。trimming_courses.course_type_id (nullable FK) で参照する。
-- 既存コースは全件「シャンプー」種別に紐付け、運用でマスタ編集画面から手動再分類する。
-- 並び順カラムは reorderByClinicID ヘルパー(sort_order 固定更新)に合わせ sort_order とする。

-- 1. 種別マスタ
CREATE TABLE trimming_course_types (
    id          BIGSERIAL    PRIMARY KEY,
    clinic_id   bigint       NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    name        varchar(50)  NOT NULL,
    sort_order  integer      NOT NULL DEFAULT 0,
    is_active   boolean      NOT NULL DEFAULT true,
    created_at  timestamptz  NOT NULL DEFAULT now(),
    updated_at  timestamptz  NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

CREATE UNIQUE INDEX idx_trimming_course_types_clinic_name
    ON trimming_course_types(clinic_id, name) WHERE deleted_at IS NULL;
CREATE INDEX idx_trimming_course_types_clinic_order
    ON trimming_course_types(clinic_id, sort_order) WHERE deleted_at IS NULL;

COMMENT ON TABLE trimming_course_types IS 'トリミングコース種別マスタ (issue #73)';

-- 2. trimming_courses に種別 FK を追加（nullable・未分類可）
ALTER TABLE trimming_courses
    ADD COLUMN course_type_id bigint REFERENCES trimming_course_types(id);
CREATE INDEX idx_trimming_courses_course_type_id
    ON trimming_courses (course_type_id) WHERE course_type_id IS NOT NULL;

-- 3. 全 clinic にデフォルト2種別を投入（clinic_id ハードコード禁止）
INSERT INTO trimming_course_types (clinic_id, name, sort_order, is_active)
SELECT c.id, t.name, t.sort_order, true
FROM clinics c
CROSS JOIN (VALUES ('シャンプー', 1), ('シャンプー＆カット', 2)) AS t(name, sort_order);

-- 4. 既存 trimming_courses 全件を各 clinic の「シャンプー」種別に紐付ける
UPDATE trimming_courses tc
SET course_type_id = tct.id
FROM trimming_course_types tct
WHERE tct.clinic_id = tc.clinic_id
  AND tct.name = 'シャンプー'
  AND tct.deleted_at IS NULL
  AND tc.course_type_id IS NULL
  AND tc.deleted_at IS NULL;

-- 5. 新規 clinic 作成時の自動投入トリガー（create_default_payment_methods と同パターン）
CREATE OR REPLACE FUNCTION create_default_trimming_course_types()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO trimming_course_types (clinic_id, name, sort_order, is_active)
    VALUES
        (NEW.id, 'シャンプー',      1, true),
        (NEW.id, 'シャンプー＆カット', 2, true);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_create_default_trimming_course_types
    AFTER INSERT ON clinics
    FOR EACH ROW
    EXECUTE FUNCTION create_default_trimming_course_types();
