-- Supports cutover rollback/maintenance checks and normal appointment-linked
-- medical-record lookups without holding avoidable long FK scans.
CREATE INDEX IF NOT EXISTS idx_medical_records_appointment_id
    ON medical_records (appointment_id)
    WHERE appointment_id IS NOT NULL;
