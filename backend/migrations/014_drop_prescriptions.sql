-- EMR-236: drop the unused prescriptions resource.
-- Chart medicine lines stay on treatments. This table had no screen.
-- Existing rows are removed with the table. Do not apply from an agent session.

DROP TABLE IF EXISTS prescriptions;
