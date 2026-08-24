-- AE-MIG-NEG-1: Jouto KNJO refunds/red-slips are negative billings and
-- cash/card splits. Drop the non-negative CHECK so CSV import can load them
-- as recorded. Fresh DBs still create the CHECK in 001_init.sql then drop it
-- here so existing environments can migrate without resetting 001.
ALTER TABLE billings DROP CONSTRAINT IF EXISTS chk_billings_amounts;
