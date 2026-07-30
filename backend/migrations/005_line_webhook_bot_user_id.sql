-- SEC-CS-F05-R1: LINE webhook signature routing key.
-- destination in webhook body is the LINE Messaging API bot user ID.
-- Lookup is O(1) via this column; empty means not yet provisioned (excluded from unique index).

ALTER TABLE line_reservation_settings
    ADD COLUMN line_bot_user_id text NOT NULL DEFAULT '';

-- Only provisioned bot IDs must be unique. Unprovisioned rows share ''.
CREATE UNIQUE INDEX uq_line_reservation_settings_line_bot_user_id
    ON line_reservation_settings (line_bot_user_id)
    WHERE line_bot_user_id <> '';

COMMENT ON COLUMN line_reservation_settings.line_bot_user_id IS
    'LINE Messaging API bot user ID (webhook destination). Used for fixed-work signature routing (SEC-CS-F05-R1). Empty until provisioned.';
