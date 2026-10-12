-- Codex #7 (PR #606 remediation): appointments.source_medical_record_id の
-- 単一列 FK（028 で追加）は clinic_id を含まないため、他医院の
-- medical_records.id を provenance として参照し得る。medical_records には
-- uq_medical_records_id_clinic (id, clinic_id) があるため、composite FK に
-- 差し替えて医院スコープを DB 制約で強制する。
--
-- 契約:
--   * source_medical_record_id IS NULL は従来どおり許容（MATCH SIMPLE:
--     複合 FK はいずれかの列が NULL なら検査しない）。
--   * ON DELETE SET NULL (source_medical_record_id) — 列指定 SET NULL は
--     001_init.sql の medicine/consultation 等の複合 FK と同じ先例形式。
--   * 既存行のうち医院越境の参照がある場合は fail-loud で中断する。
--     provenance を黙って NULL に潰さない（調査・整合後に再実行する）。
--
-- cmd/migrate wraps this file in one transaction — no BEGIN/COMMIT here.
-- Apply is manual (make migrate). Do not edit committed migrations.

DO $$
DECLARE
    violating_count integer;
BEGIN
    SELECT COUNT(*) INTO violating_count
    FROM appointments AS a
    JOIN medical_records AS mr
      ON mr.id = a.source_medical_record_id
    WHERE a.source_medical_record_id IS NOT NULL
      AND mr.clinic_id <> a.clinic_id;
    IF violating_count > 0 THEN
        RAISE EXCEPTION 'Codex#7: % appointments row(s) reference medical_records of another clinic; reconcile provenance before applying the clinic-scoped FK', violating_count;
    END IF;
END $$;

ALTER TABLE appointments
    DROP CONSTRAINT appointments_source_medical_record_id_fkey,
    ADD CONSTRAINT fk_appointments_source_medical_record_clinic
        FOREIGN KEY (source_medical_record_id, clinic_id)
        REFERENCES medical_records (id, clinic_id)
        ON DELETE SET NULL (source_medical_record_id);
