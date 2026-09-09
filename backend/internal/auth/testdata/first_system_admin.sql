\set ON_ERROR_STOP on
\set ECHO none
\set VERBOSITY terse
BEGIN ISOLATION LEVEL READ COMMITTED;
SET LOCAL search_path = public, pg_temp;
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';
SET LOCAL idle_in_transaction_session_timeout = '60s';

CREATE TEMP TABLE bootstrap_input (
    database_name text, database_role text, staff_id text, clinic_id text,
    email text, password_hash text, approval_ref text, operator_ref text
) ON COMMIT DROP;
\copy pg_temp.bootstrap_input FROM '/secure/first-system-admin/input.csv' WITH (FORMAT csv)

-- Normal account writers take ROW EXCLUSIVE, which conflicts with this lock.
-- Lock memberships as well, including INSERTs that do not exist yet.
LOCK TABLE public.accounts, public.clinics, public.staffs,
    public.staff_clinic_assignments IN SHARE ROW EXCLUSIVE MODE;

DO $bootstrap$
DECLARE
    i record;
    new_account_id bigint;
BEGIN
    IF (SELECT count(*) FROM pg_temp.bootstrap_input) <> 1 THEN
        RAISE EXCEPTION 'bootstrap requires exactly one input row';
    END IF;
    SELECT * INTO STRICT i FROM pg_temp.bootstrap_input;
    IF i.database_name IS DISTINCT FROM current_database()
       OR i.database_role IS DISTINCT FROM session_user THEN
        RAISE EXCEPTION 'bootstrap target mismatch';
    END IF;
    IF NOT coalesce(i.staff_id ~ '^[1-9][0-9]{0,17}$', false)
       OR NOT coalesce(i.clinic_id ~ '^[1-9][0-9]{0,17}$', false)
       OR NOT coalesce(i.email ~ '^[^[:space:]@]+@[^[:space:]@]+\.[^[:space:]@]+$', false)
       OR NOT coalesce(length(i.email) <= 254, false)
       OR NOT coalesce(i.password_hash ~ '^\$2[ab]\$12\$[./A-Za-z0-9]{53}$', false)
       OR NOT coalesce(i.approval_ref ~ '^[A-Za-z0-9._:-]{1,100}$', false)
       OR NOT coalesce(i.operator_ref ~ '^[A-Za-z0-9._:-]{1,100}$', false) THEN
        RAISE EXCEPTION 'bootstrap input invalid';
    END IF;
    -- Intentionally includes inactive and soft-deleted administrators/accounts.
    IF EXISTS (SELECT 1 FROM public.accounts WHERE is_system_admin)
       OR EXISTS (SELECT 1 FROM public.accounts WHERE lower(email) = lower(i.email)) THEN
        RAISE EXCEPTION 'bootstrap account conflict; do not retry';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM public.staffs s
        JOIN public.clinics c ON c.id = s.clinic_id AND c.is_active
        JOIN public.staff_clinic_assignments a
          ON a.staff_id = s.id AND a.clinic_id = c.id
         AND a.is_main AND a.deleted_at IS NULL
        WHERE s.id = i.staff_id::bigint AND s.clinic_id = i.clinic_id::bigint
          AND s.is_active AND s.deleted_at IS NULL AND s.account_id IS NULL
    ) THEN
        RAISE EXCEPTION 'bootstrap staff or active main membership unavailable';
    END IF;

    INSERT INTO public.accounts (email, password_hash, is_active, is_system_admin)
    VALUES (i.email, i.password_hash, true, true)
    RETURNING id INTO new_account_id;
    UPDATE public.staffs SET account_id = new_account_id, updated_at = now()
    WHERE id = i.staff_id::bigint AND account_id IS NULL;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'bootstrap staff attachment failed';
    END IF;
    INSERT INTO public.audit_logs (
        clinic_id, actor_id, actor_type, action, resource, resource_id,
        old_value, new_value, metadata
    ) VALUES (
        i.clinic_id::bigint, NULL, 'system', 'account.bootstrap.create',
        'account', new_account_id, NULL,
        jsonb_build_object('staff_id', i.staff_id::bigint,
                           'is_active', true, 'is_system_admin', true),
        jsonb_build_object('approval_ref', i.approval_ref,
                           'operator_ref', i.operator_ref,
                           'procedure', 'first-system-admin-v1')
    );
END;
$bootstrap$;
COMMIT;

-- Only a committed account with its committed audit is a success receipt.
SELECT a.id AS account_id, s.id AS staff_id, s.clinic_id, l.id AS audit_id,
       l.metadata ->> 'approval_ref' AS approval_ref,
       l.metadata ->> 'operator_ref' AS operator_ref
FROM public.accounts a
JOIN public.staffs s ON s.account_id = a.id
JOIN public.audit_logs l ON l.resource = 'account' AND l.resource_id = a.id
WHERE a.is_system_admin AND a.is_active AND a.deleted_at IS NULL
  AND l.action = 'account.bootstrap.create';
