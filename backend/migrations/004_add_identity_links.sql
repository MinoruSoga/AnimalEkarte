-- #239 Phase 1: owner/pet identity link tables (manual link/unlink only).
-- Parents already have uq_owners_clinic_id_id / uq_pets_clinic_id_id in 001_init.
-- RLS is defense-in-depth; application runtime scope is the first boundary.
-- created_clinic_id is the group insert RLS anchor (immutable after insert).

-- ---------------------------------------------------------------------------
-- Owner identity groups
-- ---------------------------------------------------------------------------
CREATE TABLE owner_identity_groups (
    id                BIGSERIAL PRIMARY KEY,
    created_clinic_id BIGINT NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    version           BIGINT NOT NULL DEFAULT 1 CHECK (version >= 1),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ,
    UNIQUE (created_clinic_id, id)
);

CREATE TABLE owner_identity_group_members (
    id                      BIGSERIAL PRIMARY KEY,
    group_created_clinic_id BIGINT NOT NULL,
    group_id                BIGINT NOT NULL,
    clinic_id               BIGINT NOT NULL,
    owner_id                BIGINT NOT NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at              TIMESTAMPTZ,
    FOREIGN KEY (group_created_clinic_id, group_id)
        REFERENCES owner_identity_groups(created_clinic_id, id) ON DELETE RESTRICT,
    FOREIGN KEY (clinic_id, owner_id)
        REFERENCES owners(clinic_id, id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX uq_owner_identity_active_member
    ON owner_identity_group_members(clinic_id, owner_id)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uq_owner_identity_active_group_member
    ON owner_identity_group_members(group_id, clinic_id, owner_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_owner_identity_group_members_group
    ON owner_identity_group_members(group_id)
    WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- Pet identity groups (must hang under an owner identity group)
-- ---------------------------------------------------------------------------
CREATE TABLE pet_identity_groups (
    id                            BIGSERIAL PRIMARY KEY,
    created_clinic_id             BIGINT NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    owner_group_created_clinic_id BIGINT NOT NULL,
    owner_group_id                BIGINT NOT NULL,
    version                       BIGINT NOT NULL DEFAULT 1 CHECK (version >= 1),
    created_at                    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at                    TIMESTAMPTZ,
    UNIQUE (created_clinic_id, id),
    FOREIGN KEY (owner_group_created_clinic_id, owner_group_id)
        REFERENCES owner_identity_groups(created_clinic_id, id) ON DELETE RESTRICT
);

CREATE TABLE pet_identity_group_members (
    id                      BIGSERIAL PRIMARY KEY,
    group_created_clinic_id BIGINT NOT NULL,
    group_id                BIGINT NOT NULL,
    clinic_id               BIGINT NOT NULL,
    pet_id                  BIGINT NOT NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at              TIMESTAMPTZ,
    FOREIGN KEY (group_created_clinic_id, group_id)
        REFERENCES pet_identity_groups(created_clinic_id, id) ON DELETE RESTRICT,
    FOREIGN KEY (clinic_id, pet_id)
        REFERENCES pets(clinic_id, id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX uq_pet_identity_active_member
    ON pet_identity_group_members(clinic_id, pet_id)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uq_pet_identity_active_group_member
    ON pet_identity_group_members(group_id, clinic_id, pet_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_pet_identity_group_members_group
    ON pet_identity_group_members(group_id)
    WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- Immutable created_clinic_id (groups only)
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION app_private.prevent_identity_group_created_clinic_id_update()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.created_clinic_id IS DISTINCT FROM OLD.created_clinic_id THEN
        RAISE EXCEPTION 'created_clinic_id is immutable'
            USING ERRCODE = 'check_violation';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_owner_identity_groups_created_clinic_immutable
    BEFORE UPDATE ON owner_identity_groups
    FOR EACH ROW
    EXECUTE FUNCTION app_private.prevent_identity_group_created_clinic_id_update();

CREATE TRIGGER trg_pet_identity_groups_created_clinic_immutable
    BEFORE UPDATE ON pet_identity_groups
    FOR EACH ROW
    EXECUTE FUNCTION app_private.prevent_identity_group_created_clinic_id_update();

-- ---------------------------------------------------------------------------
-- RLS (ENABLE only; FORCE remains out of scope — app scope is first boundary)
-- ---------------------------------------------------------------------------
SELECT app_private.apply_rls_policy(
    'owner_identity_groups',
    'tenant_owner_identity_groups_isolation',
    'app_private.has_clinic_access(created_clinic_id)',
    'app_private.has_clinic_access(created_clinic_id)'
);

SELECT app_private.apply_rls_policy(
    'owner_identity_group_members',
    'tenant_owner_identity_group_members_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);

SELECT app_private.apply_rls_policy(
    'pet_identity_groups',
    'tenant_pet_identity_groups_isolation',
    'app_private.has_clinic_access(created_clinic_id)',
    'app_private.has_clinic_access(created_clinic_id)'
);

SELECT app_private.apply_rls_policy(
    'pet_identity_group_members',
    'tenant_pet_identity_group_members_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);

COMMENT ON TABLE owner_identity_groups IS
    '#239 owner identity link group; created_clinic_id is immutable RLS anchor; last-member unlink soft-deletes; no revive';
COMMENT ON TABLE owner_identity_group_members IS
    '#239 owner identity members; active uniqueness per (clinic_id, owner_id); soft-delete unlink';
COMMENT ON TABLE pet_identity_groups IS
    '#239 pet identity link group under owner identity group; created_clinic_id immutable';
COMMENT ON TABLE pet_identity_group_members IS
    '#239 pet identity members; active uniqueness per (clinic_id, pet_id); soft-delete unlink';
