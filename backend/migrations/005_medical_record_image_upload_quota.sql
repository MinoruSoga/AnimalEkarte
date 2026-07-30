-- SEC-CS-F08-R1: authoritative medical-record image upload quota leases.
-- Shared across processes/replicas for concurrency, rate, and byte-budget gates.
-- Agents must not auto-apply; run `make migrate` after pull when this is present.

CREATE TABLE IF NOT EXISTS medical_record_image_upload_quota (
  id BIGSERIAL PRIMARY KEY,
  clinic_id BIGINT NOT NULL,
  staff_id BIGINT NOT NULL,
  declared_bytes BIGINT NOT NULL CHECK (declared_bytes >= 0),
  acquired_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  released_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_mri_upload_quota_clinic_acquired
  ON medical_record_image_upload_quota (clinic_id, acquired_at);

CREATE INDEX IF NOT EXISTS idx_mri_upload_quota_staff_acquired
  ON medical_record_image_upload_quota (clinic_id, staff_id, acquired_at);

CREATE INDEX IF NOT EXISTS idx_mri_upload_quota_inflight
  ON medical_record_image_upload_quota (clinic_id, staff_id)
  WHERE released_at IS NULL;
