-- FEAT-379: seed default specialty checkup age thresholds per clinic
-- Runs once per clinic; guards with WHERE NOT EXISTS to stay idempotent.
-- age_min 5 = 歯科/皮膚耳, 7 = 眼科/腎臓ドック (暫定値 — SPEC-002 Q6 確定前)
INSERT INTO lstep_tag_code_mappings (clinic_id, tag_name, code_type, codes, age_min)
SELECT c.id, 'HLTH_専門検診候補', 'specialty_dental', '{}', 5
FROM clinics c
WHERE NOT EXISTS (
    SELECT 1 FROM lstep_tag_code_mappings m
    WHERE m.clinic_id = c.id
      AND m.tag_name  = 'HLTH_専門検診候補'
      AND m.code_type = 'specialty_dental'
      AND m.deleted_at IS NULL
);

INSERT INTO lstep_tag_code_mappings (clinic_id, tag_name, code_type, codes, age_min)
SELECT c.id, 'HLTH_専門検診候補', 'specialty_skin_ear', '{}', 5
FROM clinics c
WHERE NOT EXISTS (
    SELECT 1 FROM lstep_tag_code_mappings m
    WHERE m.clinic_id = c.id
      AND m.tag_name  = 'HLTH_専門検診候補'
      AND m.code_type = 'specialty_skin_ear'
      AND m.deleted_at IS NULL
);

INSERT INTO lstep_tag_code_mappings (clinic_id, tag_name, code_type, codes, age_min)
SELECT c.id, 'HLTH_専門検診候補', 'specialty_ophthalmology', '{}', 7
FROM clinics c
WHERE NOT EXISTS (
    SELECT 1 FROM lstep_tag_code_mappings m
    WHERE m.clinic_id = c.id
      AND m.tag_name  = 'HLTH_専門検診候補'
      AND m.code_type = 'specialty_ophthalmology'
      AND m.deleted_at IS NULL
);

INSERT INTO lstep_tag_code_mappings (clinic_id, tag_name, code_type, codes, age_min)
SELECT c.id, 'HLTH_専門検診候補', 'specialty_kidney', '{}', 7
FROM clinics c
WHERE NOT EXISTS (
    SELECT 1 FROM lstep_tag_code_mappings m
    WHERE m.clinic_id = c.id
      AND m.tag_name  = 'HLTH_専門検診候補'
      AND m.code_type = 'specialty_kidney'
      AND m.deleted_at IS NULL
);
