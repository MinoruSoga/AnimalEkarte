ALTER TABLE billing_refunds
    DROP CONSTRAINT IF EXISTS billing_refunds_billing_id_fkey;

ALTER TABLE billing_refunds
    ADD CONSTRAINT billing_refunds_billing_id_fkey
    FOREIGN KEY (billing_id)
    REFERENCES billings(id)
    ON DELETE RESTRICT;

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

-- #113 DB-M2/M3: FK の ON DELETE 規則を明示し、clinic/owner 由来の業務データ削除を RESTRICT に統一する。

ALTER TABLE prescriptions
    DROP CONSTRAINT IF EXISTS prescriptions_clinic_id_fkey,
    ADD CONSTRAINT prescriptions_clinic_id_fkey
        FOREIGN KEY (clinic_id) REFERENCES clinics(id) ON DELETE RESTRICT;

ALTER TABLE prescriptions
    DROP CONSTRAINT IF EXISTS prescriptions_owner_id_fkey,
    ADD CONSTRAINT prescriptions_owner_id_fkey
        FOREIGN KEY (owner_id) REFERENCES owners(id) ON DELETE RESTRICT;

ALTER TABLE prescriptions
    DROP CONSTRAINT IF EXISTS prescriptions_pet_id_fkey,
    ADD CONSTRAINT prescriptions_pet_id_fkey
        FOREIGN KEY (pet_id) REFERENCES pets(id) ON DELETE RESTRICT;

ALTER TABLE prescriptions
    DROP CONSTRAINT IF EXISTS prescriptions_medical_record_id_fkey,
    ADD CONSTRAINT prescriptions_medical_record_id_fkey
        FOREIGN KEY (medical_record_id) REFERENCES medical_records(id) ON DELETE RESTRICT;

ALTER TABLE lstep_tag_cache
    DROP CONSTRAINT IF EXISTS lstep_tag_cache_owner_id_fkey,
    ADD CONSTRAINT lstep_tag_cache_owner_id_fkey
        FOREIGN KEY (owner_id) REFERENCES owners(id) ON DELETE RESTRICT;

ALTER TABLE line_link_tokens
    DROP CONSTRAINT IF EXISTS line_link_tokens_clinic_id_fkey,
    ADD CONSTRAINT line_link_tokens_clinic_id_fkey
        FOREIGN KEY (clinic_id) REFERENCES clinics(id) ON DELETE RESTRICT;

ALTER TABLE line_link_tokens
    DROP CONSTRAINT IF EXISTS line_link_tokens_owner_id_fkey,
    ADD CONSTRAINT line_link_tokens_owner_id_fkey
        FOREIGN KEY (owner_id) REFERENCES owners(id) ON DELETE RESTRICT;

ALTER TABLE lstep_migration_progress
    DROP CONSTRAINT IF EXISTS lstep_migration_progress_clinic_id_fkey,
    ADD CONSTRAINT lstep_migration_progress_clinic_id_fkey
        FOREIGN KEY (clinic_id) REFERENCES clinics(id) ON DELETE RESTRICT;

ALTER TABLE lstep_migration_progress
    DROP CONSTRAINT IF EXISTS lstep_migration_progress_owner_id_fkey,
    ADD CONSTRAINT lstep_migration_progress_owner_id_fkey
        FOREIGN KEY (owner_id) REFERENCES owners(id) ON DELETE RESTRICT;

ALTER TABLE lstep_trigger_priorities
    DROP CONSTRAINT IF EXISTS lstep_trigger_priorities_clinic_id_fkey,
    ADD CONSTRAINT lstep_trigger_priorities_clinic_id_fkey
        FOREIGN KEY (clinic_id) REFERENCES clinics(id) ON DELETE RESTRICT;

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

-- #95 vital_records / treatment_plans / care_logs に直接 clinic_id を追加する。
-- 既存データは親 medical_records / hospitalizations / daily_records から復元する。

ALTER TABLE vital_records
    ADD COLUMN IF NOT EXISTS clinic_id bigint;

UPDATE vital_records vr
SET clinic_id = mr.clinic_id
FROM medical_records mr
WHERE vr.medical_record_id = mr.id
  AND vr.clinic_id IS NULL;

UPDATE vital_records vr
SET clinic_id = dr.clinic_id
FROM daily_records dr
WHERE vr.daily_record_id = dr.id
  AND vr.clinic_id IS NULL;

ALTER TABLE vital_records
    ALTER COLUMN clinic_id SET NOT NULL;

ALTER TABLE vital_records
    DROP CONSTRAINT IF EXISTS vital_records_clinic_id_fkey,
    ADD CONSTRAINT vital_records_clinic_id_fkey
        FOREIGN KEY (clinic_id) REFERENCES clinics(id) ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_vital_records_clinic_id
    ON vital_records(clinic_id)
    WHERE deleted_at IS NULL;

ALTER TABLE treatment_plans
    ADD COLUMN IF NOT EXISTS clinic_id bigint;

UPDATE treatment_plans tp
SET clinic_id = mr.clinic_id
FROM medical_records mr
WHERE tp.medical_record_id = mr.id
  AND tp.clinic_id IS NULL;

UPDATE treatment_plans tp
SET clinic_id = h.clinic_id
FROM hospitalizations h
WHERE tp.hospitalization_id = h.id
  AND tp.clinic_id IS NULL;

ALTER TABLE treatment_plans
    ALTER COLUMN clinic_id SET NOT NULL;

ALTER TABLE treatment_plans
    DROP CONSTRAINT IF EXISTS treatment_plans_clinic_id_fkey,
    ADD CONSTRAINT treatment_plans_clinic_id_fkey
        FOREIGN KEY (clinic_id) REFERENCES clinics(id) ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_treatment_plans_clinic_id
    ON treatment_plans(clinic_id)
    WHERE deleted_at IS NULL;

ALTER TABLE care_logs
    ADD COLUMN IF NOT EXISTS clinic_id bigint;

UPDATE care_logs cl
SET clinic_id = dr.clinic_id
FROM daily_records dr
WHERE cl.daily_record_id = dr.id
  AND cl.clinic_id IS NULL;

ALTER TABLE care_logs
    ALTER COLUMN clinic_id SET NOT NULL;

ALTER TABLE care_logs
    DROP CONSTRAINT IF EXISTS care_logs_clinic_id_fkey,
    ADD CONSTRAINT care_logs_clinic_id_fkey
        FOREIGN KEY (clinic_id) REFERENCES clinics(id) ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_care_logs_clinic_id
    ON care_logs(clinic_id);

-- #93: PostgreSQL Row Level Security のテナント境界を定義する。
--
-- この migration は RLS を ENABLE するが FORCE はしない。
-- 理由:
--   - 現状のアプリケーション接続ユーザーは migration 実行ユーザーと同一で、テーブル owner の可能性が高い。
--   - FORCE RLS には全 repository 呼び出しを同一 transaction/context DB に統一し、SET LOCAL app.current_clinic_ids を必ず流す改修が先に必要。
--   - ここでは DB 直接アクセス用の非 owner ロールに対して RLS を効かせる、破壊性の低い baseline を構築する。
--
-- 運用時は対象 DB ロールに以下のような設定を付与する:
--   ALTER ROLE clinic_reader_1 SET app.current_clinic_ids = '1';
--   ALTER ROLE clinic_reader_all SET app.bypass_rls = 'on';

CREATE SCHEMA IF NOT EXISTS app_private;

CREATE OR REPLACE FUNCTION app_private.current_clinic_ids()
RETURNS bigint[]
LANGUAGE sql
STABLE
AS $$
    SELECT COALESCE(
        string_to_array(NULLIF(current_setting('app.current_clinic_ids', true), ''), ',')::bigint[],
        ARRAY[]::bigint[]
    );
$$;

CREATE OR REPLACE FUNCTION app_private.bypass_rls()
RETURNS boolean
LANGUAGE sql
STABLE
AS $$
    SELECT COALESCE(NULLIF(current_setting('app.bypass_rls', true), '')::boolean, false);
$$;

CREATE OR REPLACE FUNCTION app_private.has_clinic_access(row_clinic_id bigint)
RETURNS boolean
LANGUAGE sql
STABLE
AS $$
    SELECT app_private.bypass_rls()
        OR row_clinic_id = ANY(app_private.current_clinic_ids());
$$;

CREATE OR REPLACE FUNCTION app_private.apply_rls_policy(
    target_table regclass,
    policy_name text,
    using_expr text,
    check_expr text
)
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    EXECUTE format('ALTER TABLE %s ENABLE ROW LEVEL SECURITY', target_table);
    EXECUTE format('DROP POLICY IF EXISTS %I ON %s', policy_name, target_table);
    EXECUTE format(
        'CREATE POLICY %I ON %s FOR ALL USING (%s) WITH CHECK (%s)',
        policy_name,
        target_table,
        using_expr,
        check_expr
    );
END;
$$;

GRANT USAGE ON SCHEMA app_private TO PUBLIC;
GRANT EXECUTE ON FUNCTION app_private.current_clinic_ids() TO PUBLIC;
GRANT EXECUTE ON FUNCTION app_private.bypass_rls() TO PUBLIC;
GRANT EXECUTE ON FUNCTION app_private.has_clinic_access(bigint) TO PUBLIC;
REVOKE ALL ON FUNCTION app_private.apply_rls_policy(regclass, text, text, text) FROM PUBLIC;

-- clinic_id を直接持つ public テーブルは同一 policy で保護する。
DO $$
DECLARE
    target_table regclass;
BEGIN
    FOR target_table IN
        SELECT c.oid::regclass
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        JOIN pg_attribute a ON a.attrelid = c.oid
        WHERE n.nspname = 'public'
          AND c.relkind = 'r'
          AND a.attname = 'clinic_id'
          AND NOT a.attisdropped
        ORDER BY c.relname
    LOOP
        PERFORM app_private.apply_rls_policy(
            target_table,
            'tenant_clinic_id_isolation',
            'app_private.has_clinic_access(clinic_id)',
            'app_private.has_clinic_access(clinic_id)'
        );
    END LOOP;
END;
$$;

-- clinics は自身の id を tenant key として扱う。
SELECT app_private.apply_rls_policy(
    'clinics',
    'tenant_clinics_isolation',
    'app_private.has_clinic_access(id)',
    'app_private.has_clinic_access(id)'
);

-- accounts は staffs.account_id 経由で tenant 境界を判定する。
SELECT app_private.apply_rls_policy(
    'accounts',
    'tenant_accounts_isolation',
    'EXISTS (SELECT 1 FROM staffs s WHERE s.account_id = accounts.id AND app_private.has_clinic_access(s.clinic_id))',
    'EXISTS (SELECT 1 FROM staffs s WHERE s.account_id = accounts.id AND app_private.has_clinic_access(s.clinic_id))'
);

-- clinic_id を直接持たない子テーブルは、親テーブルの clinic_id 経由で保護する。
SELECT app_private.apply_rls_policy(
    'exam_type_fields',
    'tenant_exam_type_fields_isolation',
    'EXISTS (SELECT 1 FROM exam_types et WHERE et.id = exam_type_fields.exam_type_id AND app_private.has_clinic_access(et.clinic_id))',
    'EXISTS (SELECT 1 FROM exam_types et WHERE et.id = exam_type_fields.exam_type_id AND app_private.has_clinic_access(et.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'permission_group_rules',
    'tenant_permission_group_rules_isolation',
    'EXISTS (SELECT 1 FROM permission_groups pg WHERE pg.id = permission_group_rules.group_id AND app_private.has_clinic_access(pg.clinic_id))',
    'EXISTS (SELECT 1 FROM permission_groups pg WHERE pg.id = permission_group_rules.group_id AND app_private.has_clinic_access(pg.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'staff_permission_groups',
    'tenant_staff_permission_groups_isolation',
    'EXISTS (SELECT 1 FROM permission_groups pg WHERE pg.id = staff_permission_groups.group_id AND app_private.has_clinic_access(pg.clinic_id))',
    'EXISTS (SELECT 1 FROM permission_groups pg WHERE pg.id = staff_permission_groups.group_id AND app_private.has_clinic_access(pg.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'appointment_trimming_options',
    'tenant_appointment_trimming_options_isolation',
    'EXISTS (SELECT 1 FROM appointments a WHERE a.id = appointment_trimming_options.appointment_id AND app_private.has_clinic_access(a.clinic_id))',
    'EXISTS (SELECT 1 FROM appointments a WHERE a.id = appointment_trimming_options.appointment_id AND app_private.has_clinic_access(a.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'inquiries',
    'tenant_inquiries_isolation',
    'EXISTS (SELECT 1 FROM medical_records mr WHERE mr.id = inquiries.medical_record_id AND app_private.has_clinic_access(mr.clinic_id))',
    'EXISTS (SELECT 1 FROM medical_records mr WHERE mr.id = inquiries.medical_record_id AND app_private.has_clinic_access(mr.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'clinical_plans',
    'tenant_clinical_plans_isolation',
    'EXISTS (SELECT 1 FROM medical_records mr WHERE mr.id = clinical_plans.medical_record_id AND app_private.has_clinic_access(mr.clinic_id))',
    'EXISTS (SELECT 1 FROM medical_records mr WHERE mr.id = clinical_plans.medical_record_id AND app_private.has_clinic_access(mr.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'treatments',
    'tenant_treatments_isolation',
    'EXISTS (SELECT 1 FROM medical_records mr WHERE mr.id = treatments.medical_record_id AND app_private.has_clinic_access(mr.clinic_id))',
    'EXISTS (SELECT 1 FROM medical_records mr WHERE mr.id = treatments.medical_record_id AND app_private.has_clinic_access(mr.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'medical_record_images',
    'tenant_medical_record_images_isolation',
    'EXISTS (SELECT 1 FROM medical_records mr WHERE mr.id = medical_record_images.medical_record_id AND app_private.has_clinic_access(mr.clinic_id))',
    'EXISTS (SELECT 1 FROM medical_records mr WHERE mr.id = medical_record_images.medical_record_id AND app_private.has_clinic_access(mr.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'billing_confirmations',
    'tenant_billing_confirmations_isolation',
    'EXISTS (SELECT 1 FROM medical_records mr WHERE mr.id = billing_confirmations.medical_record_id AND app_private.has_clinic_access(mr.clinic_id))',
    'EXISTS (SELECT 1 FROM medical_records mr WHERE mr.id = billing_confirmations.medical_record_id AND app_private.has_clinic_access(mr.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'exam_results',
    'tenant_exam_results_isolation',
    'EXISTS (SELECT 1 FROM exams e WHERE e.id = exam_results.exam_id AND app_private.has_clinic_access(e.clinic_id))',
    'EXISTS (SELECT 1 FROM exams e WHERE e.id = exam_results.exam_id AND app_private.has_clinic_access(e.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'care_plan_items',
    'tenant_care_plan_items_isolation',
    'EXISTS (SELECT 1 FROM hospitalizations h WHERE h.id = care_plan_items.hospitalization_id AND app_private.has_clinic_access(h.clinic_id))',
    'EXISTS (SELECT 1 FROM hospitalizations h WHERE h.id = care_plan_items.hospitalization_id AND app_private.has_clinic_access(h.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'estimate_items',
    'tenant_estimate_items_isolation',
    'EXISTS (SELECT 1 FROM estimates e WHERE e.id = estimate_items.estimate_id AND app_private.has_clinic_access(e.clinic_id))',
    'EXISTS (SELECT 1 FROM estimates e WHERE e.id = estimate_items.estimate_id AND app_private.has_clinic_access(e.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'staff_notes',
    'tenant_staff_notes_isolation',
    'EXISTS (SELECT 1 FROM daily_records dr WHERE dr.id = staff_notes.daily_record_id AND app_private.has_clinic_access(dr.clinic_id))',
    'EXISTS (SELECT 1 FROM daily_records dr WHERE dr.id = staff_notes.daily_record_id AND app_private.has_clinic_access(dr.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'billing_items',
    'tenant_billing_items_isolation',
    'EXISTS (SELECT 1 FROM billings b WHERE b.id = billing_items.billing_id AND app_private.has_clinic_access(b.clinic_id))',
    'EXISTS (SELECT 1 FROM billings b WHERE b.id = billing_items.billing_id AND app_private.has_clinic_access(b.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'payments',
    'tenant_payments_isolation',
    'EXISTS (SELECT 1 FROM billings b WHERE b.id = payments.billing_id AND app_private.has_clinic_access(b.clinic_id))',
    'EXISTS (SELECT 1 FROM billings b WHERE b.id = payments.billing_id AND app_private.has_clinic_access(b.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'staff_reservation_exclusions',
    'tenant_staff_reservation_exclusions_isolation',
    'EXISTS (SELECT 1 FROM reservation_types rt WHERE rt.id = staff_reservation_exclusions.reservation_type_id AND app_private.has_clinic_access(rt.clinic_id))',
    'EXISTS (SELECT 1 FROM reservation_types rt WHERE rt.id = staff_reservation_exclusions.reservation_type_id AND app_private.has_clinic_access(rt.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'shift_entry_breaks',
    'tenant_shift_entry_breaks_isolation',
    'EXISTS (SELECT 1 FROM shift_entries se WHERE se.id = shift_entry_breaks.shift_entry_id AND app_private.has_clinic_access(se.clinic_id))',
    'EXISTS (SELECT 1 FROM shift_entries se WHERE se.id = shift_entry_breaks.shift_entry_id AND app_private.has_clinic_access(se.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'shift_template_breaks',
    'tenant_shift_template_breaks_isolation',
    'EXISTS (SELECT 1 FROM shift_templates st WHERE st.id = shift_template_breaks.shift_template_id AND app_private.has_clinic_access(st.clinic_id))',
    'EXISTS (SELECT 1 FROM shift_templates st WHERE st.id = shift_template_breaks.shift_template_id AND app_private.has_clinic_access(st.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'campaign_target_categories',
    'tenant_campaign_target_categories_isolation',
    'EXISTS (SELECT 1 FROM campaigns c WHERE c.id = campaign_target_categories.campaign_id AND app_private.has_clinic_access(c.clinic_id))',
    'EXISTS (SELECT 1 FROM campaigns c WHERE c.id = campaign_target_categories.campaign_id AND app_private.has_clinic_access(c.clinic_id))'
);

SELECT app_private.apply_rls_policy(
    'campaign_target_items',
    'tenant_campaign_target_items_isolation',
    'EXISTS (SELECT 1 FROM campaigns c WHERE c.id = campaign_target_items.campaign_id AND app_private.has_clinic_access(c.clinic_id))',
    'EXISTS (SELECT 1 FROM campaigns c WHERE c.id = campaign_target_items.campaign_id AND app_private.has_clinic_access(c.clinic_id))'
);

-- 公開/システム共通/認証補助テーブルは clinic_id を持たないため RLS 対象外:
-- companies, animal_species, token_blacklist,
-- lstep_auto_managed_prefixes, lstep_condition_tag_mappings, lstep_send_purpose_tag_prefixes,
-- password_reset_tokens, manual_articles, manual_article_versions
