-- #112 audit_logs の clinic_id / actor_id 整合性を強化する。
-- 既存の不完全ログは制約追加前に可能な範囲で補正する。

-- actor_id が staff を参照していない既存ログは system actor に退避し、元 actor_id を metadata に残す。
UPDATE audit_logs
SET
    metadata = COALESCE(metadata, '{}'::jsonb)
        || jsonb_build_object(
            'migration_note', 'actor_id was not a valid staff id before #112 constraint migration',
            'migration_original_actor_id', actor_id
        ),
    actor_type = 'system',
    actor_id = NULL
WHERE actor_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1
      FROM staffs
      WHERE staffs.id = audit_logs.actor_id
  );

-- clinic_id が欠落している既存ログは、staff のメイン所属クリニックから補完する。
UPDATE audit_logs al
SET clinic_id = (
    SELECT sca.clinic_id
    FROM staff_clinic_assignments sca
    WHERE sca.staff_id = al.actor_id
      AND sca.deleted_at IS NULL
    ORDER BY sca.is_main DESC, sca.id ASC
    LIMIT 1
)
WHERE al.clinic_id IS NULL
  AND al.actor_id IS NOT NULL
  AND EXISTS (
      SELECT 1
      FROM staff_clinic_assignments sca
      WHERE sca.staff_id = al.actor_id
        AND sca.deleted_at IS NULL
  );

-- それでも tenant を復元できない旧ログは、metadata に補正理由を残して最小 clinic に紐づける。
-- 以後は service validation と DB 制約で同種ログの作成を禁止する。
UPDATE audit_logs
SET
    metadata = COALESCE(metadata, '{}'::jsonb)
        || jsonb_build_object('migration_note', 'clinic_id was backfilled from the first clinic because original tenant was unknown'),
    clinic_id = (SELECT id FROM clinics ORDER BY id ASC LIMIT 1),
    actor_type = 'system',
    actor_id = NULL
WHERE clinic_id IS NULL;

ALTER TABLE audit_logs
    ALTER COLUMN clinic_id SET NOT NULL;

ALTER TABLE audit_logs
    DROP CONSTRAINT IF EXISTS audit_logs_clinic_id_fkey,
    ADD CONSTRAINT audit_logs_clinic_id_fkey
        FOREIGN KEY (clinic_id) REFERENCES clinics(id) ON DELETE RESTRICT;

ALTER TABLE audit_logs
    DROP CONSTRAINT IF EXISTS audit_logs_actor_id_fkey,
    ADD CONSTRAINT audit_logs_actor_id_fkey
        FOREIGN KEY (actor_id) REFERENCES staffs(id) ON DELETE RESTRICT;

ALTER TABLE audit_logs
    DROP CONSTRAINT IF EXISTS audit_logs_actor_type_check,
    ADD CONSTRAINT audit_logs_actor_type_check
        CHECK (actor_type IN ('staff', 'system'));

ALTER TABLE audit_logs
    DROP CONSTRAINT IF EXISTS audit_logs_actor_consistency_check,
    ADD CONSTRAINT audit_logs_actor_consistency_check
        CHECK (
            (actor_type = 'system' AND actor_id IS NULL)
            OR
            (actor_type = 'staff' AND actor_id IS NOT NULL)
        );
