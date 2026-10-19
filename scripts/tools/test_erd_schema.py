"""Offline contract tests for the bounded migration interpreter."""
import dataclasses
from pathlib import Path
import tempfile
import unittest

from erd_schema import SchemaError, load_schema, statements, _sql, _implicit_name


def consolidated_source(name):
    """Read the historical fixture from the current consolidated baseline."""
    migration = Path(__file__).resolve().parents[2] / "backend/migrations/001_init.sql"
    marker = f"-- ===== consolidated from {name} =====\n\n"
    return migration.read_text().split(marker, 1)[1].split("\n-- ===== consolidated from ", 1)[0]


class SchemaTests(unittest.TestCase):
    def schema(self, *sql):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            for number, content in enumerate(sql):
                (root / f'{number:03}.sql').write_text(content)
            return load_schema(root)

    def test_lexical_boundaries_and_sources(self):
        sql = "-- ; ignored\n/* outer /* ; */ */ CREATE TABLE \"Odd\" (\"ID\" bigint PRIMARY KEY, note text DEFAULT E'it\\\'s;fine', value numeric(8,2) DEFAULT (round(1.25, 2)));"
        table = self.schema(sql).tables['Odd']
        self.assertEqual(table.columns[2].type_sql, 'numeric(8,2)')
        self.assertEqual(table.columns[1].default_sql, "E'it\\\'s;fine'")
        self.assertEqual(table.columns[0].source, '000.sql:2')
        self.assertFalse(table.columns[0].nullable)
        with self.assertRaises(dataclasses.FrozenInstanceError):
            table.name = 'changed'

    def test_inline_and_table_constraints(self):
        schema = self.schema('CREATE TABLE parents(id bigint PRIMARY KEY, clinic_id bigint NOT NULL, UNIQUE(id,clinic_id)); CREATE TABLE children(id bigint PRIMARY KEY, parent_id bigint REFERENCES parents(id) ON DELETE SET NULL, clinic_id bigint, CONSTRAINT fk_pair FOREIGN KEY(parent_id,clinic_id) REFERENCES parents(id,clinic_id) ON DELETE SET NULL(parent_id), CHECK(id>0));')
        constraints = schema.tables['children'].constraints
        self.assertEqual([x.kind for x in constraints], ['primary', 'foreign', 'foreign', 'check'])
        self.assertEqual(constraints[2].on_delete, 'SET NULL(parent_id)')
        self.assertEqual(constraints[2].target_columns, ('id', 'clinic_id'))

    def test_ordered_alters_and_drop(self):
        schema = self.schema('CREATE TABLE t(id bigint PRIMARY KEY, old text, x int DEFAULT 2); CREATE TABLE gone(id int); CREATE INDEX old_idx ON t(old);', 'ALTER TABLE t ADD COLUMN IF NOT EXISTS extra text NOT NULL, DROP COLUMN old, ALTER COLUMN x TYPE numeric(10,2), ALTER COLUMN x SET NOT NULL, ALTER COLUMN x SET DEFAULT (3+4); ALTER TABLE t ALTER COLUMN extra DROP NOT NULL, ALTER COLUMN x DROP DEFAULT; DROP TABLE IF EXISTS gone;')
        table = schema.tables['t']
        self.assertEqual([c.name for c in table.columns], ['id', 'x', 'extra'])
        self.assertEqual(table.columns[1].type_sql, 'numeric(10,2)')
        self.assertIsNone(table.columns[1].default_sql)
        self.assertTrue(table.columns[2].nullable)
        self.assertEqual(table.indexes, ())
        self.assertNotIn('gone', schema.tables)

    def test_constraint_names_drop_and_replace(self):
        schema = self.schema('CREATE TABLE p(id int PRIMARY KEY); CREATE TABLE c(id int, p_id int REFERENCES p(id));', 'ALTER TABLE c DROP CONSTRAINT c_p_id_fkey, ADD CONSTRAINT fk FOREIGN KEY(id) REFERENCES p(id) ON DELETE RESTRICT;')
        self.assertEqual(schema.tables['c'].constraints[0].name, 'fk')

    def test_indexes_preserve_partial_and_expression(self):
        schema = self.schema('CREATE TABLE t(id int, note text); CREATE UNIQUE INDEX uq ON t(id DESC) WHERE id IS NOT NULL; CREATE INDEX expr ON t USING gin ((lower(note)) gin_trgm_ops); CREATE INDEX removed ON t(note); DROP INDEX removed;')
        indexes = schema.tables['t'].indexes
        self.assertEqual(indexes[0].columns, ('id DESC',))
        self.assertEqual(indexes[0].predicate, 'id IS NOT NULL')
        self.assertTrue(indexes[0].unique)
        self.assertIn('gin_trgm_ops', indexes[1].sql)
        self.assertIsNone(indexes[1].predicate)

    def test_function_body_is_opaque(self):
        schema = self.schema("CREATE FUNCTION f() RETURNS void AS $fn$ BEGIN EXECUTE 'CREATE TABLE hidden(x int)'; END; $fn$ LANGUAGE plpgsql; CREATE TABLE visible(x int);")
        self.assertEqual(set(schema.tables), {'visible'})

    def test_unknown_statements_and_do_fail_closed(self):
        for sql in ['DO $$ BEGIN CREATE TABLE hidden(x int); END $$;', 'CREATE VIEW v AS SELECT 1;', 'CREATE TABLE t(x int); ALTER TABLE t RENAME COLUMN x TO y;', 'CREATE TABLE t(x int GENERATED ALWAYS AS (2) STORED);', 'CREATE TABLE p(id int PRIMARY KEY); CREATE TABLE c(x int REFERENCES p(id) ON DELETE EXPLODE);', 'ALTER TYPE example RENAME TO changed;']:
            with self.subTest(sql=sql), self.assertRaises(SchemaError):
                self.schema(sql)

    def test_invalid_references_and_syntax(self):
        for sql in ['CREATE TABLE t(x int REFERENCES absent(id));', 'CREATE TABLE p(id int); CREATE TABLE c(x int REFERENCES p(absent));', "CREATE TABLE t(x text DEFAULT 'unterminated);", 'CREATE TABLE t(x int;', 'CREATE TABLE t(x int); ALTER TABLE t DROP CONSTRAINT absent;']:
            with self.subTest(sql=sql), self.assertRaises(SchemaError):
                self.schema(sql)

    def test_seeds_excluded_and_migration_hashes(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            (root / '001.sql').write_text('CREATE TABLE t(x int);')
            (root / 'seeds').mkdir()
            (root / 'seeds' / '002.sql').write_text('CREATE TABLE unwanted(x int);')
            schema = load_schema(root)
            self.assertEqual(set(schema.tables), {'t'})
            self.assertEqual(schema.migrations[0].name, '001.sql')
            self.assertEqual(len(schema.migrations[0].sha256), 64)

    def test_implicit_constraint_name_collisions(self):
        schema = self.schema('CREATE TABLE p(id int PRIMARY KEY); CREATE TABLE c(id int, p_id int, CONSTRAINT c_p_id_fkey CHECK(p_id>0)); ALTER TABLE c ADD FOREIGN KEY(p_id) REFERENCES p(id); ALTER TABLE c DROP CONSTRAINT c_p_id_fkey1;')
        self.assertEqual([c.name for c in schema.tables['c'].constraints], ['c_p_id_fkey'])
        schema = self.schema('CREATE TABLE p(id int PRIMARY KEY); CREATE TABLE c(p_id int REFERENCES p(id) REFERENCES p(id));')
        self.assertEqual([c.name for c in schema.tables['c'].constraints], ['c_p_id_fkey', 'c_p_id_fkey1'])

    def test_long_implicit_names_preserve_postgresql_suffix(self):
        examples = [
            ('owner_identity_group_members', ('group_created_clinic_id', 'group_id'), 'fkey', 'owner_identity_group_members_group_created_clinic_id_group_fkey'),
            ('pet_identity_group_members', ('group_created_clinic_id', 'group_id'), 'fkey', 'pet_identity_group_members_group_created_clinic_id_group_i_fkey'),
            ('pet_identity_groups', ('owner_group_created_clinic_id', 'owner_group_id'), 'fkey', 'pet_identity_groups_owner_group_created_clinic_id_owner_gr_fkey'),
            ('lstep_friend_attribute_snapshots', ('clinic_id', 'line_user_id', 'snapshot_taken_at'), 'key', 'lstep_friend_attribute_snapsh_clinic_id_line_user_id_snapsh_key'),
            ('reservation_type_occupations', ('reservation_type_id', 'occupation_id'), 'key', 'reservation_type_occupations_reservation_type_id_occupation_key'),
        ]
        for table, columns, suffix, expected in examples:
            with self.subTest(table=table):
                self.assertEqual(_implicit_name(table, columns, suffix, ()), expected)
                self.assertLessEqual(len(expected.encode('utf-8')), 63)
        # Byte clipping must neither split UTF-8 nor consume the suffix.
        name = _implicit_name('表' * 20, ('列' * 20,), 'fkey', ())
        self.assertEqual(name, '表' * 9 + '_' + '列' * 9 + '_fkey')
        self.assertLessEqual(len(name.encode('utf-8')), 63)

    def test_long_implicit_names_rebalance_on_collision(self):
        table = 't' * 40
        column = 'c' * 40
        schema = self.schema(f'CREATE TABLE {table}({column} int UNIQUE UNIQUE);')
        self.assertEqual([c.name for c in schema.tables[table].constraints], [
            't' * 29 + '_' + 'c' * 29 + '_key',
            't' * 29 + '_' + 'c' * 28 + '_key1',
        ])

    def test_check_dependencies_and_names_use_expression(self):
        schema = self.schema('CREATE TABLE t(a int, b int CHECK(a>b), c int CHECK(b>c));')
        checks = [c for c in schema.tables['t'].constraints if c.kind == 'check']
        self.assertEqual([c.name for c in checks], ['t_check', 't_check1'])
        self.assertEqual(checks[0].columns, ('a', 'b'))
        table = self.schema('CREATE TABLE t(a int,b int CHECK(a>b)); ALTER TABLE t DROP COLUMN a;').tables['t']
        self.assertEqual(table.constraints, ())

    def test_check_forward_reference_and_actual_single_dependency(self):
        schema = self.schema('CREATE TABLE t(a int CHECK(b>0), b int CHECK(b+b>0), c int CHECK(1>0));')
        checks = schema.tables['t'].constraints
        self.assertEqual([c.name for c in checks], ['t_b_check', 't_b_check1', 't_check'])
        self.assertEqual(checks[0].columns, ('b',))
        table = self.schema('CREATE TABLE t(a int CHECK(b>0), b int); ALTER TABLE t DROP COLUMN b;').tables['t']
        self.assertEqual(table.constraints, ())

    def test_with_select_into_refuses_unmodeled_table_creation(self):
        with self.assertRaisesRegex(SchemaError, 'INTO'):
            self.schema('CREATE TABLE visible(id int); WITH x AS (SELECT 1 AS id) SELECT id INTO hidden FROM x;')
        schema = self.schema("CREATE TABLE visible(id int); WITH x AS (SELECT 'INTO' AS word) SELECT word FROM x;")
        self.assertEqual(set(schema.tables), {'visible'})

    def test_drop_column_removes_own_dependents(self):
        schema = self.schema('CREATE TABLE t(id int PRIMARY KEY, p_id int REFERENCES t(id), UNIQUE(p_id), CHECK(p_id>0)); CREATE UNIQUE INDEX uq ON t(id) WHERE p_id IS NOT NULL; ALTER TABLE t DROP COLUMN p_id;')
        self.assertEqual([c.kind for c in schema.tables['t'].constraints], ['primary'])
        self.assertEqual(schema.tables['t'].indexes, ())

    def test_index_name_does_not_create_column_dependency(self):
        schema = self.schema('CREATE TABLE t(id int, x int); CREATE INDEX x ON t(id); ALTER TABLE t DROP COLUMN x;')
        self.assertEqual([index.name for index in schema.tables['t'].indexes], ['x'])

    def test_implicit_reference_uses_final_primary_key_columns(self):
        schema = self.schema('CREATE TABLE p(id int PRIMARY KEY); CREATE TABLE c(p_id int REFERENCES p);')
        self.assertEqual(schema.tables['c'].constraints[0].target_columns, ('id',))

    def test_drop_external_dependency_fails_immediately(self):
        for action in ['DROP TABLE p; CREATE TABLE p(id int PRIMARY KEY);', 'ALTER TABLE p DROP COLUMN id; ALTER TABLE p ADD id int PRIMARY KEY;']:
            with self.subTest(action=action), self.assertRaises(SchemaError):
                self.schema('CREATE TABLE p(id int PRIMARY KEY); CREATE TABLE c(p_id int REFERENCES p(id));' + action)

    def test_reviewed_do_changes_are_rejected(self):
        migration = Path(__file__).resolve().parents[2] / 'backend/migrations/001_init.sql'
        for body in (_sql(t) for t in statements(migration.read_text()) if t[0].upper == 'DO' and "rel.relname = 'checkup_types'" in _sql(t)):
            with self.subTest(body=body[:50]), self.assertRaisesRegex(SchemaError, 'unreviewed DO'):
                self.schema(body.replace("rel.relname = 'checkup_types'", "rel.relname = 'other_table'") + ';')

    def test_reviewed_appointment_provenance_assertion_preserves_schema(self) -> None:
        migration_sql = consolidated_source('031_appointments_source_medical_record_clinic_fk.sql')
        body = _sql(next(t for t in statements(migration_sql) if t[0].upper == 'DO'))
        ddl = 'CREATE TABLE medical_records(id bigint PRIMARY KEY, clinic_id bigint NOT NULL, UNIQUE(id, clinic_id)); CREATE TABLE appointments(id bigint PRIMARY KEY, clinic_id bigint NOT NULL, source_medical_record_id bigint REFERENCES medical_records(id));'
        before = self.schema(ddl)
        after = self.schema(ddl, body + ';')
        self.assertEqual(after.tables, before.tables)
        self.assertEqual(after.reviewed_do_blocks, (
            '001.sql:1 sha256=a2a0d54165c328b2ce6e41ba40d202cad9168941086cc888bccaadda1d27cedb data_assertion',
        ))

    def test_appointment_provenance_clinic_fk_replaces_single_column_fk(self) -> None:
        migration_sql = consolidated_source('031_appointments_source_medical_record_clinic_fk.sql')
        ddl = 'CREATE TABLE medical_records(id bigint PRIMARY KEY, clinic_id bigint NOT NULL, UNIQUE(id, clinic_id)); CREATE TABLE appointments(id bigint PRIMARY KEY, clinic_id bigint NOT NULL, source_medical_record_id bigint REFERENCES medical_records(id));'
        schema = self.schema(ddl, migration_sql)
        table = schema.tables['appointments']
        provenance = [c for c in table.constraints if c.kind == 'foreign' and 'source_medical_record_id' in c.columns]
        self.assertEqual(len(provenance), 1)
        fk = provenance[0]
        self.assertEqual(fk.name, 'fk_appointments_source_medical_record_clinic')
        self.assertEqual(fk.columns, ('source_medical_record_id', 'clinic_id'))
        self.assertEqual(fk.target_table, 'medical_records')
        self.assertEqual(fk.target_columns, ('id', 'clinic_id'))
        self.assertEqual(fk.on_delete, 'SET NULL (source_medical_record_id)')
        self.assertEqual(fk.source, '001.sql:35')
        self.assertNotIn('appointments_source_medical_record_id_fkey', {c.name for c in table.constraints})
        columns = {c.name: c for c in table.columns}
        self.assertTrue(columns['source_medical_record_id'].nullable)
        self.assertFalse(columns['clinic_id'].nullable)

    def test_changed_appointment_provenance_do_requires_review(self) -> None:
        migration_sql = consolidated_source('031_appointments_source_medical_record_clinic_fk.sql')
        body = _sql(next(t for t in statements(migration_sql) if t[0].upper == 'DO'))
        for changed in (
            body.replace('mr.clinic_id <> a.clinic_id', 'mr.clinic_id = a.clinic_id'),
            body.replace('END $$', 'CREATE TABLE hidden(id integer);\nEND $$'),
        ):
            with self.subTest(body=changed), self.assertRaisesRegex(SchemaError, 'unreviewed DO'):
                self.schema(changed + ';')

    def test_duplicate_objects_and_conditional_creation(self):
        schema = self.schema('CREATE TABLE t(x int); CREATE TABLE IF NOT EXISTS t(x int); ALTER TABLE t ADD COLUMN IF NOT EXISTS x text; CREATE INDEX idx ON t(x); CREATE INDEX IF NOT EXISTS idx ON t(x); DROP INDEX IF EXISTS absent; ALTER TABLE t DROP CONSTRAINT IF EXISTS absent;')
        self.assertEqual(schema.tables['t'].columns[0].type_sql, 'int')
        self.assertEqual(len(schema.tables['t'].indexes), 1)
        for action in ['CREATE TABLE t(x int);', 'ALTER TABLE t ADD COLUMN x text;', 'CREATE INDEX idx ON t(x);']:
            with self.subTest(action=action), self.assertRaises(SchemaError):
                self.schema('CREATE TABLE t(x int); CREATE INDEX idx ON t(x);' + action)

    def test_lexer_rejects_unfinished_and_unbalanced_tokens(self):
        for sql in ['/* never closed', 'DO $fn$ never closed;', 'CREATE TABLE t(x int));', 'CREATE TABLE "never closed(x int);']:
            with self.subTest(sql=sql), self.assertRaises(SchemaError):
                self.schema(sql)

    def test_named_check_unique_primary_and_default_null(self):
        schema = self.schema('CREATE TABLE t(x int DEFAULT NULL, y text CONSTRAINT yy UNIQUE, CONSTRAINT pk PRIMARY KEY(x), CONSTRAINT checked CHECK(x>0)); ALTER TABLE t ADD CONSTRAINT another UNIQUE(x,y);')
        table = schema.tables['t']
        self.assertEqual(table.columns[0].default_sql, 'NULL')
        self.assertFalse(table.columns[0].nullable)
        self.assertEqual({c.name for c in table.constraints}, {'yy', 'pk', 'checked', 'another'})

    def test_empty_migration_directory(self):
        with tempfile.TemporaryDirectory() as temp, self.assertRaisesRegex(SchemaError, 'no migrations'):
            load_schema(Path(temp))

    def test_revisit_intake_catalog_and_saved_s2_binding(self):
        schema = load_schema(Path(__file__).resolve().parents[2] / 'backend/migrations')
        generation = next(c for c in schema.tables['appointments'].columns if c.name == 'revisit_intake_generation')
        self.assertFalse(generation.nullable)
        self.assertEqual(generation.default_sql, '1')
        self.assertIn('revisit_intake_target_id', {c.name for c in schema.tables['medical_records'].columns})
        answer = schema.tables['revisit_intake_answers']
        self.assertIn(('clinic_id', 'target_id', 'revision'), {c.columns for c in answer.constraints if c.kind == 'unique'})
        self.assertIn(('clinic_id', 'author_customer_id', 'operation_key'), {c.columns for c in answer.constraints if c.kind == 'unique'})
        for name in ('onset', 'energy', 'appetite', 'current_symptoms', 'improvement', 'examination_preference', 'main_concern'):
            field = next(c for c in answer.columns if c.name == name)
            self.assertTrue(field.nullable)
            self.assertIsNone(field.default_sql)
        for name in ('revisit_intake_targets', 'revisit_intake_answers', 'revisit_intake_submit_receipts'):
            self.assertTrue(all(c.on_delete == 'RESTRICT' for c in schema.tables[name].constraints if c.kind == 'foreign'))

    def test_item_insurance_missing_is_not_a_database_default(self):
        schema = load_schema(Path(__file__).resolve().parents[2] / 'backend/migrations')
        for name in ('consultations', 'procedures', 'vaccines', 'checkup_types'):
            with self.subTest(table=name):
                field = next(c for c in schema.tables[name].columns if c.name == 'is_non_insurance')
                self.assertEqual(field.type_sql.lower(), 'boolean')
                self.assertTrue(field.nullable, 'legacy missing must remain representable')
                self.assertIsNone(field.default_sql, 'new-entry UI defaults must not backfill old writers')
        for name in ('medicines', 'exam_types'):
            with self.subTest(existing_table=name):
                field = next(c for c in schema.tables[name].columns if c.name == 'is_non_insurance')
                self.assertFalse(field.nullable)
                self.assertEqual(field.default_sql.lower(), 'false')

    def test_repository_final_schema(self):
        schema = load_schema(Path(__file__).resolve().parents[2] / 'backend/migrations')
        self.assertEqual(len(schema.tables), 164)
        self.assertEqual([m.name for m in schema.migrations], ['001_init.sql'])
        embeddings = schema.tables['support_manual_embeddings']
        self.assertEqual({c.name for c in embeddings.columns}, {'cache_key', 'model', 'task_type', 'chunker_version', 'dimensions', 'embedding', 'created_at'})
        self.assertTrue(all(not c.nullable for c in embeddings.columns))
        self.assertEqual(next(c for c in embeddings.constraints if c.kind == 'primary').columns, ('cache_key',))
        self.assertEqual(next(c for c in embeddings.columns if c.name == 'embedding').type_sql, 'real[]')
        self.assertEqual({c.name for c in embeddings.constraints if c.kind == 'check'}, {
            'support_manual_embeddings_dimensions_chk', 'support_manual_embeddings_length_chk',
            'support_manual_embeddings_task_chk', 'support_manual_embeddings_version_chk',
        })
        budgets = schema.tables['support_llm_budgets']
        self.assertEqual({c.name for c in budgets.columns}, {'scope', 'model', 'next_allowed_at'})
        self.assertTrue(all(not c.nullable for c in budgets.columns))
        self.assertEqual(next(c for c in budgets.constraints if c.kind == 'primary').columns, ('scope', 'model'))
        for table in (embeddings, budgets):
            self.assertFalse(any(c.kind == 'foreign' for c in table.constraints))
        deletions = schema.tables['trimming_image_deletions']
        self.assertEqual({c.name for c in deletions.columns}, {'clinic_id', 'object_key', 'actor_staff_id', 'created_at', 'storage_deleted_at', 'last_attempt_at'})
        self.assertEqual({c.name for c in deletions.columns if not c.nullable}, {'clinic_id', 'object_key', 'actor_staff_id', 'created_at'})
        self.assertTrue(next(c for c in deletions.columns if c.name == 'storage_deleted_at').nullable)
        self.assertTrue(next(c for c in deletions.columns if c.name == 'last_attempt_at').nullable)
        self.assertEqual(next(c for c in deletions.constraints if c.kind == 'primary').columns, ('clinic_id', 'object_key'))
        pending = next(i for i in deletions.indexes if i.name == 'idx_trimming_image_deletions_pending')
        self.assertEqual(pending.columns, ('clinic_id', 'last_attempt_at', 'object_key'))
        self.assertIsNotNone(pending.predicate)
        self.assertIn('storage_deleted_at IS NULL', pending.predicate)
        self.assertNotIn('prescriptions', schema.tables)
        self.assertNotIn('next_visit_recommended_date', {c.name for c in schema.tables['medical_records'].columns})
        for table_name, suffix, column_name, delete_action in [
            ('appointments', 'doctor', 'doctor_id', 'SET NULL'),
            ('hospitalizations', 'doctor', 'doctor_id', 'SET NULL'),
            ('medical_records', 'doctor', 'doctor_id', 'SET NULL'),
            ('cash_register_close_adjustments', 'actor', 'actor_id', 'RESTRICT'),
            ('medical_record_image_upload_quota', 'staff', 'staff_id', 'RESTRICT'),
            ('lab_device_waits', 'staff', 'staff_id', 'RESTRICT'),
        ]:
            with self.subTest(table=table_name):
                name = f'fk_{table_name}_{suffix}'
                constraints = schema.tables[table_name].constraints
                staff = next(c for c in constraints if c.name == name)
                self.assertEqual(staff.columns, (column_name,))
                self.assertEqual(staff.target_table, 'staffs')
                self.assertEqual(staff.target_columns, ('id',))
                self.assertEqual(staff.on_delete, delete_action)
                self.assertNotIn(name + '_clinic', {c.name for c in constraints})
        cage = next(c for c in schema.tables['cages'].constraints if c.name == 'fk_cages_merged_into_cage')
        self.assertEqual(cage.columns, ('merged_into_cage_id', 'clinic_id'))
        draft = next(c for c in schema.tables['lab_device_mapping_draft_items'].constraints if c.name == 'fk_lab_device_mapping_draft_items_draft_clinic')
        self.assertEqual(draft.columns, ('clinic_id', 'draft_id'))
        appointment_columns = {c.name for c in schema.tables['appointments'].columns}
        self.assertTrue({'intended_date', 'source_medical_record_id'} <= appointment_columns)
        provenance = next(c for c in schema.tables['appointments'].constraints if c.name == 'fk_appointments_source_medical_record_clinic')
        self.assertEqual(provenance.columns, ('source_medical_record_id', 'clinic_id'))
        self.assertEqual(provenance.target_table, 'medical_records')
        self.assertEqual(provenance.target_columns, ('id', 'clinic_id'))
        self.assertEqual(provenance.on_delete, 'SET NULL (source_medical_record_id)')
        self.assertEqual(provenance.source, '001_init.sql:7466')
        self.assertNotIn('appointments_source_medical_record_id_fkey', {c.name for c in schema.tables['appointments'].constraints})
        parent = [c for c in schema.tables['checkup_types'].constraints if c.kind == 'foreign' and 'parent_id' in c.columns]
        self.assertEqual(len(parent), 1)
        self.assertEqual(parent[0].columns, ('parent_id', 'clinic_id'))
        self.assertIn('SET NULL (parent_id)', parent[0].on_delete)
        self.assertEqual(len(schema.reviewed_do_blocks), 13)
        expected_names = {
            'owner_identity_group_members': 'owner_identity_group_members_group_created_clinic_id_group_fkey',
            'pet_identity_group_members': 'pet_identity_group_members_group_created_clinic_id_group_i_fkey',
            'pet_identity_groups': 'pet_identity_groups_owner_group_created_clinic_id_owner_gr_fkey',
            'lstep_friend_attribute_snapshots': 'lstep_friend_attribute_snapsh_clinic_id_line_user_id_snapsh_key',
            'reservation_type_occupations': 'reservation_type_occupations_reservation_type_id_occupation_key',
        }
        for table_name, expected in expected_names.items():
            self.assertIn(expected, {c.name for c in schema.tables[table_name].constraints})
        checks = {c.name: c.columns for c in schema.tables['reservation_type_unavailable_times'].constraints if c.kind == 'check'}
        self.assertEqual(checks['reservation_type_unavailable_times_check'], ('unavailable_type', 'day_of_week'))
        self.assertEqual(checks['reservation_type_unavailable_times_check1'], ('unavailable_type', 'specific_date'))


if __name__ == '__main__':
    unittest.main()
