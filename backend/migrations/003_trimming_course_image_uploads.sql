-- EMR-457: provenance is never reconstructed from reference absence or age.
-- Technical ledger intentionally has no deleted_at: removing proof is forbidden.
CREATE TABLE trimming_course_image_uploads (
    clinic_id BIGINT NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    object_key TEXT NOT NULL,
    uploader_staff_id BIGINT NOT NULL REFERENCES staffs(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    uploaded_at TIMESTAMPTZ,
    submitted_at TIMESTAMPTZ,
    PRIMARY KEY (clinic_id, object_key),
    CHECK (clinic_id > 0 AND uploader_staff_id > 0),
    CHECK (object_key LIKE 'trimming-images/' || clinic_id::text || '/%'
           AND object_key ~ '^trimming-images/[1-9][0-9]*/[0-9]+_[0-9a-f]{32}\.(jpg|jpeg|png|gif)$')
);
SELECT app_private.apply_rls_policy(
    'trimming_course_image_uploads', 'tenant_clinic_id_isolation',
    'app_private.has_clinic_access(clinic_id)', 'app_private.has_clinic_access(clinic_id)'
);
CREATE FUNCTION app_private.guard_trimming_course_image_uploads() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP IN ('DELETE', 'TRUNCATE') THEN
        RAISE EXCEPTION 'course image provenance cannot be removed';
    END IF;
    IF ROW(NEW.clinic_id, NEW.object_key, NEW.uploader_staff_id, NEW.created_at)
       IS DISTINCT FROM ROW(OLD.clinic_id, OLD.object_key, OLD.uploader_staff_id, OLD.created_at)
       OR (OLD.uploaded_at IS NOT NULL AND NEW.uploaded_at IS DISTINCT FROM OLD.uploaded_at)
       OR (OLD.submitted_at IS NOT NULL AND NEW.submitted_at IS DISTINCT FROM OLD.submitted_at) THEN
        RAISE EXCEPTION 'course image provenance is irreversible';
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER trimming_course_image_uploads_immutable
BEFORE UPDATE OR DELETE ON trimming_course_image_uploads
FOR EACH ROW EXECUTE FUNCTION app_private.guard_trimming_course_image_uploads();
CREATE TRIGGER trimming_course_image_uploads_no_truncate
BEFORE TRUNCATE ON trimming_course_image_uploads
FOR EACH STATEMENT EXECUTE FUNCTION app_private.guard_trimming_course_image_uploads();
