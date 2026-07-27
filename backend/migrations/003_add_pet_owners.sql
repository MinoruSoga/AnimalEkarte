CREATE TABLE pet_owners (
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

CREATE INDEX idx_pet_owners_clinic_pet ON pet_owners (clinic_id, pet_id);
CREATE INDEX idx_pet_owners_clinic_owner ON pet_owners (clinic_id, owner_id);

SELECT app_private.apply_rls_policy(
    'pet_owners',
    'tenant_clinic_id_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);
