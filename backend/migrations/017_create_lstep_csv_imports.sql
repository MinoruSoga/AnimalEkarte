CREATE TABLE lstep_csv_imports (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id       bigint NOT NULL REFERENCES clinics(id),
    csv_type        varchar(50) NOT NULL,
    file_name       varchar(255) NOT NULL,
    uploaded_by_user_id bigint REFERENCES accounts(id),
    row_count       int NOT NULL DEFAULT 0,
    success_count   int NOT NULL DEFAULT 0,
    error_count     int NOT NULL DEFAULT 0,
    status          varchar(20) NOT NULL DEFAULT 'pending',
    error_log       jsonb,
    imported_at     timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_lstep_csv_imports_clinic_imported ON lstep_csv_imports (clinic_id, imported_at DESC);
