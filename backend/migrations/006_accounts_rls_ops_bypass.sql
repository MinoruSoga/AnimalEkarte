-- Allow ops bypass sessions (app.bypass_rls) to INSERT/SELECT accounts.
--
-- tenant_accounts_isolation required an existing staffs.account_id back-reference,
-- which makes account INSERT structurally impossible for non-owner roles: the
-- row being checked does not exist yet, so no staff can point at it. Every other
-- tenant policy already honors app_private.bypass_rls() through
-- has_clinic_access(); accounts was the outlier. Ops tooling that runs as a
-- limited PlanetScale role (e.g. stg-uat-staff-attach) depends on that bypass
-- contract, so align accounts with it.
SELECT app_private.apply_rls_policy(
    'accounts',
    'tenant_accounts_isolation',
    'app_private.bypass_rls() OR EXISTS (SELECT 1 FROM staffs s WHERE s.account_id = accounts.id AND app_private.has_clinic_access(s.clinic_id))',
    'app_private.bypass_rls() OR EXISTS (SELECT 1 FROM staffs s WHERE s.account_id = accounts.id AND app_private.has_clinic_access(s.clinic_id))'
);
