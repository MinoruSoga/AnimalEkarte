CREATE TABLE IF NOT EXISTS lstep_delivery_trigger_log (
    id               bigserial    PRIMARY KEY,
    owner_id         bigint       NOT NULL REFERENCES owners(id),
    clinic_id        bigint       NOT NULL REFERENCES clinics(id),
    trigger_type     varchar(50)  NOT NULL,
    scheduled_at     timestamptz  NOT NULL,
    status           varchar(20)  NOT NULL DEFAULT 'scheduled',
    fired_at         timestamptz,
    excluded_reason  varchar(100),
    created_at       timestamptz  NOT NULL DEFAULT now(),
    updated_at       timestamptz  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_lstep_delivery_trigger_log_lookup
    ON lstep_delivery_trigger_log (clinic_id, owner_id, trigger_type, scheduled_at);

CREATE INDEX IF NOT EXISTS idx_lstep_delivery_trigger_log_clinic_date
    ON lstep_delivery_trigger_log (clinic_id, scheduled_at);
