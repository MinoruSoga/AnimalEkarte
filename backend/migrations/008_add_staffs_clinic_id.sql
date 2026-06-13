-- #110 staffs に主所属 clinic_id を追加する。
-- staff_clinic_assignments は複数所属の正規データとして維持し、staffs.clinic_id は直接参照時のテナント境界用に非正規化する。

ALTER TABLE staffs
    ADD COLUMN IF NOT EXISTS clinic_id bigint;

UPDATE staffs s
SET clinic_id = (
    SELECT sca.clinic_id
    FROM staff_clinic_assignments sca
    WHERE sca.staff_id = s.id
      AND sca.deleted_at IS NULL
    ORDER BY sca.is_main DESC, sca.id ASC
    LIMIT 1
)
WHERE s.clinic_id IS NULL
  AND EXISTS (
      SELECT 1
      FROM staff_clinic_assignments sca
      WHERE sca.staff_id = s.id
        AND sca.deleted_at IS NULL
  );

-- 所属が復元できない旧データは最小 clinic に紐づける。
-- 以後の create / assignment 更新では service 層が主所属を同期する。
UPDATE staffs
SET clinic_id = (SELECT id FROM clinics ORDER BY id ASC LIMIT 1)
WHERE clinic_id IS NULL;

ALTER TABLE staffs
    ALTER COLUMN clinic_id SET NOT NULL;

ALTER TABLE staffs
    DROP CONSTRAINT IF EXISTS staffs_clinic_id_fkey,
    ADD CONSTRAINT staffs_clinic_id_fkey
        FOREIGN KEY (clinic_id) REFERENCES clinics(id) ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_staffs_clinic
    ON staffs(clinic_id)
    WHERE deleted_at IS NULL;
