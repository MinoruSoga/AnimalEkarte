-- Enforce tenant ownership across LSTEP friend snapshots and their CSV import.
-- Existing mismatches abort the migration before either constraint is changed.

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM lstep_friend_attribute_snapshots AS snapshot
        JOIN lstep_csv_imports AS csv_import
          ON csv_import.id = snapshot.csv_import_id
        WHERE snapshot.csv_import_id IS NOT NULL
          AND snapshot.clinic_id <> csv_import.clinic_id
    ) THEN
        RAISE EXCEPTION
            'cross-clinic lstep friend snapshot csv_import reference exists'
            USING ERRCODE = '23514';
    END IF;
END
$$;

ALTER TABLE lstep_csv_imports
    ADD CONSTRAINT uq_lstep_csv_imports_clinic_id_id
    UNIQUE (clinic_id, id);

ALTER TABLE lstep_friend_attribute_snapshots
    DROP CONSTRAINT IF EXISTS lstep_friend_attribute_snapshots_csv_import_id_fkey;

ALTER TABLE lstep_friend_attribute_snapshots
    ADD CONSTRAINT fk_lstep_snapshots_clinic_csv_import
    FOREIGN KEY (clinic_id, csv_import_id)
    REFERENCES lstep_csv_imports (clinic_id, id)
    ON DELETE RESTRICT;
