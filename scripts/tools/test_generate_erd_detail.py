"""Public CLI regression tests for generated ERD documents; no database required."""

import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


CLI = Path(__file__).with_name("generate-erd-detail.py")
FIXTURE = """
CREATE TABLE parents (id bigint PRIMARY KEY, name text NOT NULL);
CREATE TABLE children (
    id bigint PRIMARY KEY,
    parent_id bigint REFERENCES parents(id) ON DELETE SET NULL,
    amount numeric(10,2) NOT NULL DEFAULT 0,
    note text DEFAULT 'a|b',
    deleted_at timestamptz
);
CREATE UNIQUE INDEX uq_children_parent_active ON children (parent_id)
    WHERE deleted_at IS NULL;
"""


class DetailedERDCLI(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.migrations = self.root / "backend/migrations"
        self.migrations.mkdir(parents=True)
        (self.migrations / "001_init.sql").write_text(FIXTURE)
        self.output = self.root / "docs/architecture/erd-detail"

    def run_cli(self, mode):
        return subprocess.run(
            [sys.executable, "-B", str(CLI), "--root", str(self.root), mode],
            text=True, capture_output=True, check=False,
        )

    def generate(self):
        result = self.run_cli("--write")
        self.assertEqual(result.returncode, 0, result.stderr + result.stdout)

    def bytes_snapshot(self):
        return {p.name: p.read_bytes() for p in self.output.iterdir()}

    def test_all_columns_fk_exact_types_and_predicates_are_documented(self):
        self.generate()
        child = (self.output / "children.md").read_text()
        self.assertIn("| `amount` | `numeric(10,2)` | NOT NULL | `0` |", child)
        self.assertIn("| `parent_id` | `bigint` | NULL |", child)
        self.assertIn("`children_parent_id_fkey`", child)
        self.assertIn("[parents](parents.md)", child)
        self.assertIn("ON DELETE SET NULL", child)
        self.assertIn("deleted_at IS NULL", child)
        self.assertIn("a&#124;b", child)
        self.assertIn('parents o|..o{ children : "parent_id"', child)
        self.assertIn("部分unique", child)
        manifest = json.loads((self.output / "manifest.json").read_text())
        self.assertEqual(manifest["counts"], {"tables": 2, "columns": 7, "foreign_keys": 1})
        self.assertEqual([m["name"] for m in manifest["migrations"]], ["001_init.sql"])

    def test_check_is_read_only_and_two_writes_are_identical(self):
        self.generate()
        snapshot = self.bytes_snapshot()
        result = self.run_cli("--check")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(self.bytes_snapshot(), snapshot)
        self.generate()
        self.assertEqual(self.bytes_snapshot(), snapshot)

    def test_check_detects_tampered_column_and_fk_without_repairing(self):
        self.generate()
        p = self.output / "children.md"
        p.write_text(p.read_text().replace("numeric(10,2)", "integer").replace("children_parent_id_fkey", "removed_fk"))
        snapshot = self.bytes_snapshot()
        result = self.run_cli("--check")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("children.md", result.stderr)
        self.assertEqual(self.bytes_snapshot(), snapshot)

    def test_check_detects_missing_and_extra_generated_pages(self):
        self.generate()
        (self.output / "parents.md").unlink()
        result = self.run_cli("--check")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("parents.md", result.stderr)
        self.generate()
        (self.output / "obsolete.md").write_text((self.output / "parents.md").read_text())
        result = self.run_cli("--check")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("obsolete.md", result.stderr)

    def test_new_migration_changes_columns_and_source_digest(self):
        self.generate()
        before = json.loads((self.output / "manifest.json").read_text())
        (self.migrations / "002_extra.sql").write_text("ALTER TABLE children ADD COLUMN extra text;")
        self.assertNotEqual(self.run_cli("--check").returncode, 0)
        self.generate()
        after = json.loads((self.output / "manifest.json").read_text())
        self.assertNotEqual(before["source_digest"], after["source_digest"])
        self.assertEqual(after["counts"]["columns"], 8)
        self.assertIn("`extra`", (self.output / "children.md").read_text())

    def test_removed_table_removes_only_stale_owned_page(self):
        self.generate()
        note = self.output / "notes.md"
        note.write_text("Human notes are preserved.\n")
        (self.migrations / "002_remove.sql").write_text("DROP TABLE children;")
        self.generate()
        self.assertFalse((self.output / "children.md").exists())
        self.assertEqual(note.read_text(), "Human notes are preserved.\n")

    def test_unsupported_schema_change_fails_before_touching_outputs(self):
        self.generate()
        snapshot = self.bytes_snapshot()
        (self.migrations / "002_unsupported.sql").write_text("ALTER TABLE children RENAME TO renamed;")
        result = self.run_cli("--write")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("002_unsupported.sql", result.stderr)
        self.assertEqual(self.bytes_snapshot(), snapshot)

    def test_does_not_overwrite_human_file_at_generated_path(self):
        self.output.mkdir(parents=True)
        p = self.output / "children.md"
        p.write_text("Human document.\n")
        result = self.run_cli("--write")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("children.md", result.stderr)
        self.assertEqual(p.read_text(), "Human document.\n")

    def test_seed_sql_is_not_schema_inventory(self):
        seeds = self.migrations / "seeds"
        seeds.mkdir()
        (seeds / "live_seed.sql").write_text("CREATE TABLE excluded (id bigint);")
        self.generate()
        self.assertFalse((self.output / "excluded.md").exists())

    def test_generated_path_symlink_is_refused_without_touching_target(self):
        self.generate()
        p = self.output / "children.md"
        target = self.root / "foreign.md"
        target.write_bytes(p.read_bytes())
        p.unlink()
        p.symlink_to(target)
        before = target.read_bytes()
        self.assertNotEqual(self.run_cli("--check").returncode, 0)
        self.assertNotEqual(self.run_cli("--write").returncode, 0)
        self.assertEqual(target.read_bytes(), before)

    def test_foreign_manifest_is_refused_before_any_document_write(self):
        self.generate()
        (self.output / "manifest.json").write_text("42\n")
        snapshot = self.bytes_snapshot()
        result = self.run_cli("--write")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("foreign manifest", result.stderr)
        self.assertEqual(self.bytes_snapshot(), snapshot)

    def test_output_directory_symlink_is_refused_without_touching_target(self):
        self.generate()
        snapshot = self.bytes_snapshot()
        foreign = self.root / "foreign-docs"
        self.output.rename(foreign)
        self.output.symlink_to(foreign, target_is_directory=True)
        (self.migrations / "002_extra.sql").write_text("ALTER TABLE children ADD COLUMN extra text;")
        self.assertNotEqual(self.run_cli("--write").returncode, 0)
        self.assertEqual({p.name: p.read_bytes() for p in foreign.iterdir()}, snapshot)
        self.assertNotEqual(self.run_cli("--check").returncode, 0)


if __name__ == "__main__":
    unittest.main()
