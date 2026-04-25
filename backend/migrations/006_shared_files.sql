-- 006_shared_files.sql
-- LINE個別送信用ファイルストレージ

CREATE TABLE shared_files (
    id          BIGSERIAL PRIMARY KEY,
    clinic_id   BIGINT       NOT NULL REFERENCES clinics(id),
    owner_id    BIGINT       REFERENCES owners(id),
    uploaded_by BIGINT       NOT NULL REFERENCES staffs(id),
    file_type   VARCHAR(50)  NOT NULL,
    file_name   VARCHAR(255) NOT NULL,
    file_key    VARCHAR(500) NOT NULL,
    file_size   BIGINT       NOT NULL,
    purpose     VARCHAR(50)  NOT NULL,
    expires_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_shared_files_clinic_owner
    ON shared_files (clinic_id, owner_id);

CREATE INDEX idx_shared_files_expires_at
    ON shared_files (expires_at)
    WHERE deleted_at IS NULL;
