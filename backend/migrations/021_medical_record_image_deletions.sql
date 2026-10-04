-- 診療画像の削除台帳。画像行は物理削除する。この表が署名停止と storage 削除の再試行の正本。
-- medical_record_images への FK は置かない（行は消える）。medical_records からの CASCADE も置かない。
-- 既に発行済みの署名 URL は storage 削除に失敗したとき最大 15 分残る。この表はそれを取り消さない。
-- Do not edit 001_init.sql. User must run: make migrate

CREATE TABLE medical_record_image_deletions (
    id BIGSERIAL PRIMARY KEY,
    clinic_id BIGINT NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    medical_record_id BIGINT NOT NULL,
    image_id BIGINT NOT NULL,
    image_object_key TEXT NOT NULL DEFAULT '',
    thumbnail_object_key TEXT NOT NULL DEFAULT '',
    storage_status TEXT NOT NULL,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    last_error TEXT NOT NULL DEFAULT '',
    actor_staff_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT medical_record_image_deletions_status_check
        CHECK (storage_status IN ('pending_storage', 'storage_deleted', 'retained_shared')),
    CONSTRAINT medical_record_image_deletions_clinic_image_uidx
        UNIQUE (clinic_id, image_id)
);

CREATE INDEX idx_medical_record_image_deletions_clinic_record
    ON medical_record_image_deletions (clinic_id, medical_record_id);

SELECT app_private.apply_rls_policy(
    'medical_record_image_deletions',
    'tenant_clinic_id_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);
