-- Diagnosis 2 関連フィールドの追加 (v9.1)
ALTER TABLE clinical_plans ADD COLUMN diagnosis_2_category_id bigint;
ALTER TABLE clinical_plans ADD COLUMN diagnosis_2_name_id bigint;

-- 外部キー制約の追加
ALTER TABLE clinical_plans 
    ADD CONSTRAINT fk_clinical_plans_diagnosis_2_category 
    FOREIGN KEY (diagnosis_2_category_id) REFERENCES diagnosis_categories(id);

ALTER TABLE clinical_plans 
    ADD CONSTRAINT fk_clinical_plans_diagnosis_2_name 
    FOREIGN KEY (diagnosis_2_name_id) REFERENCES diagnosis_names(id);
