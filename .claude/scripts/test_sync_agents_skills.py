#!/usr/bin/env python3
"""Skill/rule mirror safety fixtures, independent of application services."""
import json
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

SCRIPT = Path(__file__).with_name("sync-agents-skills.py")
PROJECT = SCRIPT.parents[2]


class SkillMirrorTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        for folder in (".claude/skills/demo/references", ".claude/commands", ".claude/rules", "scripts"):
            (self.root / folder).mkdir(parents=True)
        (self.root / ".claude/skills/demo/SKILL.md").write_text("canonical skill\n")
        (self.root / ".claude/skills/demo/references/example.md").write_bytes(b"reference\r\n")
        (self.root / ".claude/rules/safety.md").write_text("safe rule\n")
        self.command = self.root / ".claude/commands/demo.md"
        self.command.write_text('---\ndescription: "Quote \\"value\\" and backslash \\\\ with 日本語"\n---\n\nBody\n---\nend\n')
        shutil.copyfile(PROJECT / "scripts/yaml-frontmatter-description.mjs", self.root / "scripts/yaml-frontmatter-description.mjs")

    def run_sync(self, success=True):
        result = subprocess.run([sys.executable, "-B", str(SCRIPT), str(self.root)], capture_output=True, text=True)
        self.assertEqual(result.returncode == 0, success, result.stderr + result.stdout)
        return result

    def snapshot(self):
        return {str(p.relative_to(self.root)): p.read_bytes() for p in self.root.rglob("*") if p.is_file()}

    def test_exact_copies_wrapper_quoting_and_idempotence(self):
        before = self.snapshot()
        self.run_sync()
        for name in ("skills/demo/SKILL.md", "skills/demo/references/example.md", "rules/safety.md"):
            self.assertEqual((self.root / ".agents" / name).read_bytes(), (self.root / ".claude" / name).read_bytes())
        wrapper = (self.root / ".agents/skills/source-command-demo/SKILL.md").read_text()
        expected = ('---\nname: "source-command-demo"\n'
                    'description: "Quote \\"value\\" and backslash \\\\ with 日本語"\n'
                    '---\n\n# source-command-demo\n\n'
                    'Use this skill when the user asks to run the migrated source command `demo`.\n\n'
                    '## Command Template\n\nBody\n---\nend\n')
        self.assertEqual(wrapper, expected)
        for path, data in before.items():
            self.assertEqual((self.root / path).read_bytes(), data)
        first = self.snapshot()
        self.run_sync()
        self.assertEqual(first, self.snapshot())

    def test_foreign_files_preserved_and_modified_target_blocks_all(self):
        self.run_sync()
        unknown = self.root / ".agents/skills/demo/foreign.md"
        unknown.write_text("WIP")
        self.run_sync()
        self.assertEqual(unknown.read_text(), "WIP")
        (self.root / ".agents/rules/safety.md").write_text("foreign rule edit")
        (self.root / ".claude/skills/demo/SKILL.md").write_text("updated")
        before = self.snapshot()
        self.run_sync(False)
        self.assertEqual(before, self.snapshot())

    def test_legacy_exact_bytes_adopted(self):
        self.run_sync()
        (self.root / ".agents/.mirror-ownership.json").unlink()
        self.run_sync()
        self.assertTrue((self.root / ".agents/.mirror-ownership.json").exists())

    def test_unrecognized_legacy_target_never_overwritten(self):
        target = self.root / ".agents/rules/safety.md"
        target.parent.mkdir(parents=True)
        target.write_text("WIP")
        before = self.snapshot()
        self.run_sync(False)
        self.assertEqual(before, self.snapshot())

    def test_removed_owned_files_only(self):
        self.run_sync()
        (self.root / ".claude/skills/demo/references/example.md").unlink()
        foreign = self.root / ".agents/skills/demo/references/foreign.md"
        foreign.write_text("WIP")
        self.run_sync()
        self.assertFalse((self.root / ".agents/skills/demo/references/example.md").exists())
        self.assertEqual(foreign.read_text(), "WIP")

    def test_modified_removed_target_blocks(self):
        self.run_sync()
        (self.root / ".claude/skills/demo/references/example.md").unlink()
        (self.root / ".agents/skills/demo/references/example.md").write_text("WIP")
        before = self.snapshot()
        self.run_sync(False)
        self.assertEqual(before, self.snapshot())

    def test_symlink_parent_source_and_ledger_rejected(self):
        for kind in ("destination", "source", "ledger"):
            with self.subTest(kind=kind):
                outside = self.root / "outside"
                outside.mkdir(exist_ok=True)
                if kind == "destination":
                    target = self.root / ".agents"
                elif kind == "source":
                    target = self.root / ".claude/skills/demo/link"
                else:
                    (self.root / ".agents").mkdir(exist_ok=True)
                    target = self.root / ".agents/.mirror-ownership.json"
                target.symlink_to(outside, target_is_directory=True)
                self.run_sync(False)
                target.unlink()
                self.assertEqual(list(outside.iterdir()), [])

    def test_invalid_command_description_blocks_all(self):
        self.command.write_text("---\nname: missing-description\n---\nbody\n")
        before = self.snapshot()
        self.run_sync(False)
        self.assertEqual(before, self.snapshot())

    def test_duplicate_wrapper_destination_blocks_all(self):
        collision = self.root / ".claude/skills/source-command-demo"
        collision.mkdir()
        (collision / "SKILL.md").write_text("duplicate")
        before = self.snapshot()
        self.run_sync(False)
        self.assertEqual(before, self.snapshot())

    def test_repository_sources_fixture_matches_legacy_wrapper_bytes(self):
        for relative in (".claude/skills", ".claude/commands", ".claude/rules"):
            shutil.copytree(PROJECT / relative, self.root / relative, dirs_exist_ok=True, symlinks=True)
        self.command.unlink()
        self.run_sync()
        for command in (self.root / ".claude/commands").glob("*.md"):
            result = subprocess.run(["awk", 'BEGIN{fm=0} /^---$/{ if (fm<2) {fm++; next} } fm>=2{print}', str(command)], check=True, capture_output=True)
            wrapper = self.root / f".agents/skills/source-command-{command.stem}/SKILL.md"
            self.assertTrue(wrapper.read_bytes().endswith(b"## Command Template\n" + result.stdout))
        first = self.snapshot()
        self.run_sync()
        self.assertEqual(first, self.snapshot())


if __name__ == "__main__":
    unittest.main()
