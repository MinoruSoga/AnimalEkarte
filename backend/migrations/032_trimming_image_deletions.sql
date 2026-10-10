-- Permanent storage lifecycle tombstones, committed before R2/local deletion.
-- Shared keys remain attachable until a no-reference check reserves deletion.
-- No soft delete: forgetting a tombstone would allow attaching a deleted object.
-- Clinical mutation attribution remains in audit_logs; this dedicated sink records
-- the first deletion reservation's clinic, object key, staff and timestamp.
-- No dependency on medical_record_images or medical_record_image_deletions.
-- User-run migration required before deploying the image lifecycle code.
CREATE TABLE trimming_image_deletions (
    clinic_id BIGINT NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    object_key TEXT NOT NULL,
    actor_staff_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (clinic_id, object_key)
);

SELECT app_private.apply_rls_policy(
    'trimming_image_deletions',
    'tenant_clinic_id_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);
