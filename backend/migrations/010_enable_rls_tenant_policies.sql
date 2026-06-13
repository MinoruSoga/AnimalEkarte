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
