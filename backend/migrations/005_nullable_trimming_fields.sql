-- trimming_records の staff_id, course_id を NULL 許可に変更
-- フロントエンドからコース・スタッフなしで登録できるようにするため
ALTER TABLE trimming_records
  DROP CONSTRAINT IF EXISTS trimming_records_staff_id_fkey,
  DROP CONSTRAINT IF EXISTS trimming_records_course_id_fkey,
  ALTER COLUMN staff_id DROP NOT NULL,
  ALTER COLUMN course_id DROP NOT NULL;

-- FK制約を再追加（NULL を許可）
ALTER TABLE trimming_records
  ADD CONSTRAINT trimming_records_staff_id_fkey
    FOREIGN KEY (staff_id) REFERENCES staffs(id) ON DELETE SET NULL,
  ADD CONSTRAINT trimming_records_course_id_fkey
    FOREIGN KEY (course_id) REFERENCES trimming_courses(id) ON DELETE SET NULL;
