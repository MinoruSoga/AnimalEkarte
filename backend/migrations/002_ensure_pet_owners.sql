-- BRT-71: STG で pet_owners が欠落し shared-pets / sub-owners が 500 になっていた。
-- 001_init に定義はあるが、checksum repair 等で DDL 未適用の DB があり得るため
-- append-only で IF NOT EXISTS の安全ネットを置く（破壊的操作なし）。

CREATE TABLE IF NOT EXISTS pet_owners (
    id           BIGSERIAL PRIMARY KEY,
    clinic_id    BIGINT NOT NULL REFERENCES clinics(id),
    pet_id       BIGINT NOT NULL,
    owner_id     BIGINT NOT NULL,
    relationship TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (pet_id, owner_id),
    FOREIGN KEY (clinic_id, pet_id) REFERENCES pets (clinic_id, id),
    FOREIGN KEY (clinic_id, owner_id) REFERENCES owners (clinic_id, id)
);

CREATE INDEX IF NOT EXISTS idx_pet_owners_clinic_pet ON pet_owners (clinic_id, pet_id);
CREATE INDEX IF NOT EXISTS idx_pet_owners_clinic_owner ON pet_owners (clinic_id, owner_id);

SELECT app_private.apply_rls_policy(
    'pet_owners',
    'tenant_clinic_id_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);
