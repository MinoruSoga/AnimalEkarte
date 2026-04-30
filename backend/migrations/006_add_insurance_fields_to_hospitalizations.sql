-- 006_add_insurance_fields_to_hospitalizations.sql
-- 入院レコードに保険会社名・保険番号フィールドを追加

ALTER TABLE hospitalizations
  ADD COLUMN insurance_company_name VARCHAR(100) NULL,
  ADD COLUMN insurance_number       VARCHAR(50)  NULL;

COMMENT ON COLUMN hospitalizations.insurance_company_name IS '保険会社名（入院単位で記録）';
COMMENT ON COLUMN hospitalizations.insurance_number       IS '保険番号（入院単位で記録）';
