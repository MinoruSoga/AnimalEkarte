-- Durable work marker, committed before provider writes. It contains no desired
-- tag snapshot: retries read the current appointment, consent and LINE linkage.
-- Successful reconciliation acknowledges only its own revision in the cache
-- transaction. BIGSERIAL also prevents revision reuse after delete/reinsert.
-- Technical pending work is hard-deleted on acknowledgement, not clinical data.
CREATE TABLE lstep_next_visit_sync_retries (
    clinic_id BIGINT NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    owner_id BIGINT NOT NULL,
    revision BIGSERIAL NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (clinic_id, owner_id),
    CONSTRAINT fk_lstep_next_visit_sync_retries_owner_clinic
        FOREIGN KEY (clinic_id, owner_id) REFERENCES owners(clinic_id, id) ON DELETE RESTRICT
);

SELECT app_private.apply_rls_policy(
    'lstep_next_visit_sync_retries',
    'tenant_clinic_id_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);
