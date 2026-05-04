CREATE TABLE lstep_friend_attribute_snapshots (
    id                  bigserial PRIMARY KEY,
    clinic_id           bigint NOT NULL REFERENCES clinics(id),
    line_user_id        varchar(50) NOT NULL,
    display_name        varchar(255),
    registered_at       timestamptz,
    tags                jsonb,
    scenarios           jsonb,
    traffic_source      varchar(100),
    block_status        varchar(20),
    last_message_at     timestamptz,
    snapshot_taken_at   timestamptz NOT NULL,
    csv_import_id       uuid REFERENCES lstep_csv_imports(id),
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    UNIQUE (clinic_id, line_user_id, snapshot_taken_at)
);
CREATE INDEX idx_lstep_friend_attr_clinic_user ON lstep_friend_attribute_snapshots (clinic_id, line_user_id);
CREATE INDEX idx_lstep_friend_attr_clinic_taken ON lstep_friend_attribute_snapshots (clinic_id, snapshot_taken_at DESC);
