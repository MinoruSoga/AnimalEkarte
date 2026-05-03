ALTER TABLE appointments
    ADD COLUMN IF NOT EXISTS reservation_route varchar(20);
ALTER TABLE appointments
    ADD COLUMN IF NOT EXISTS actual_reservation_at timestamptz;
