#!/usr/bin/env python3
"""Isolated regression fixtures; no application, network, or shared outputs."""
import hashlib
import json
import subprocess
import sys
import tempfile
import tomllib
import unittest
from pathlib import Path

SCRIPT = Path(__file__).with_name("sync-codex-mirror.py")
sys.path.insert(0, str(SCRIPT.parent))

class SyncCodexMirrorTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        (self.root / ".claude/agents").mkdir(parents=True)
        (self.root / ".claude/commands").mkdir()
        self.body = 'Run "regex\\.test" and \\ paths.\nTriple """ quotes.\n'
        self.source("planner", self.body)
        self.source("architect", "Generic instructions.\n")
        self.manifest([{"name": "animalekarte-planner", "source": "planner.md"}])
        (self.root / ".claude/commands/check.md").write_text("exact command\n")

    def source(self, name, body):
        (self.root / f".claude/agents/{name}.md").write_text("---\ndescription: Test role\n---\n\n" + body)

    def manifest(self, roles):
        (self.root / ".claude/codex-agent-manifest.json").write_text(json.dumps({"version": 1, "roles": roles}))

    def run_sync(self, ok=True):
        result = subprocess.run([sys.executable, "-B", str(SCRIPT), str(self.root)], capture_output=True, text=True)
        self.assertEqual(result.returncode == 0, ok, result.stdout + result.stderr)
        return result

    def snapshot(self):
        return {str(p.relative_to(self.root)): p.read_bytes() for p in self.root.rglob("*") if p.is_file()}

    def test_exact_body_sources_commands_and_idempotence(self):
        before = self.snapshot()
        self.run_sync()
        role = tomllib.loads((self.root / ".codex/agents/animalekarte-planner.toml").read_text())
        self.assertEqual(role["developer_instructions"], self.body)
        self.assertNotIn("sandbox_mode", role)
        self.assertFalse((self.root / ".codex/agents/architect.toml").exists())
        self.assertEqual((self.root / ".codex/commands/check.md").read_text(), "exact command\n")
        for path, data in before.items():
            self.assertEqual((self.root / path).read_bytes(), data)
        first = self.snapshot()
        self.run_sync()
        self.assertEqual(first, self.snapshot())

    def test_read_only_role(self):
        self.manifest([{"name": "clinic-isolation-auditor", "source": "planner.md", "sandbox_mode": "read-only"}])
        self.run_sync()
        role = tomllib.loads((self.root / ".codex/agents/clinic-isolation-auditor.toml").read_text())
        self.assertEqual(role["sandbox_mode"], "read-only")

    def test_invalid_manifest_is_non_mutating(self):
        cases = [
            [{"name": "planner", "source": "planner.md"}],
            [{"name": "../escape", "source": "planner.md"}],
            [{"name": "valid", "source": "../planner.md"}],
            [{"name": "valid", "source": "missing.md"}],
            [{"name": "valid", "source": "planner.md", "sandbox_mode": "danger-full-access"}],
            [{"name": "valid", "source": "planner.md", "sandbox_mode": "workspace-write"}],
            [{"name": "valid", "source": "planner.md", "extra": True}],
            [{"name": "valid", "source": "planner.md"}] * 2,
            [{"name": "one", "source": "planner.md"}, {"name": "two", "source": "planner.md"}],
        ]
        for roles in cases:
            with self.subTest(roles=roles):
                self.manifest(roles)
                before = self.snapshot()
                self.run_sync(False)
                self.assertEqual(before, self.snapshot())

    def test_foreign_files_survive_and_changed_owned_files_block_all_writes(self):
        self.run_sync()
        foreign = self.root / ".codex/agents/foreign.toml"
        foreign.write_text("foreign WIP")
        self.run_sync()
        self.assertEqual(foreign.read_text(), "foreign WIP")
        (self.root / ".codex/commands/check.md").write_text("foreign edit")
        self.source("planner", "New source")
        before = self.snapshot()
        self.run_sync(False)
        self.assertEqual(before, self.snapshot())

    def test_foreign_target_blocks_before_any_output_written(self):
        dest = self.root / ".codex/commands/check.md"
        dest.parent.mkdir(parents=True)
        dest.write_text("WIP")
        before = self.snapshot()
        self.run_sync(False)
        self.assertEqual(before, self.snapshot())

    def test_legacy_outputs_migrate_only_if_exactly_owned(self):
        import importlib.util
        spec = importlib.util.spec_from_file_location("mirror", SCRIPT)
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)
        dest = self.root / ".codex/agents"
        dest.mkdir(parents=True)
        for name in ("planner", "architect"):
            desc, body = module.read_frontmatter_and_body(self.root / f".claude/agents/{name}.md")
            (dest / f"{name}.toml").write_text(module.legacy_agent(name, desc, body))
        self.run_sync()
        self.assertFalse((dest / "planner.toml").exists())
        self.assertFalse((dest / "architect.toml").exists())
        self.assertTrue((dest / "animalekarte-planner.toml").exists())

    def test_changed_legacy_reserved_output_blocks_migration(self):
        dest = self.root / ".codex/agents/architect.toml"
        dest.parent.mkdir(parents=True)
        dest.write_text("foreign edit")
        before = self.snapshot()
        self.run_sync(False)
        self.assertEqual(before, self.snapshot())

    def test_owned_obsolete_output_removed_foreign_untouched(self):
        self.run_sync()
        foreign = self.root / ".codex/agents/foreign.toml"
        foreign.write_text("WIP")
        self.manifest([])
        self.run_sync()
        self.assertFalse((self.root / ".codex/agents/animalekarte-planner.toml").exists())
        self.assertEqual(foreign.read_text(), "WIP")

    def test_source_update_replaces_hash_owned_output(self):
        self.run_sync()
        self.source("planner", "Updated instructions\n")
        self.run_sync()
        role = tomllib.loads((self.root / ".codex/agents/animalekarte-planner.toml").read_text())
        self.assertEqual(role["developer_instructions"], "Updated instructions\n")

    def test_modified_retired_output_blocks_without_deleting(self):
        self.run_sync()
        self.manifest([])
        role = self.root / ".codex/agents/animalekarte-planner.toml"
        role.write_text("edited by another agent")
        before = self.snapshot()
        self.run_sync(False)
        self.assertEqual(before, self.snapshot())

    def test_symlink_file_target_blocks_all_writes(self):
        outside = self.root / "foreign"
        outside.write_text("WIP")
        target = self.root / ".codex/commands/check.md"
        target.parent.mkdir(parents=True)
        target.symlink_to(outside)
        before = self.snapshot()
        self.run_sync(False)
        self.assertEqual(before, self.snapshot())

    def test_repository_manifest_fixture(self):
        import shutil
        project = SCRIPT.parents[2]
        for relative in (".claude/agents", ".claude/commands"):
            # Fixture setup only: copy canonical inputs, never real destinations.
            for path in (project / relative).glob("*.md"):
                shutil.copyfile(path, self.root / relative / path.name)
        shutil.copyfile(project / ".claude/codex-agent-manifest.json", self.root / ".claude/codex-agent-manifest.json")
        self.run_sync()
        initial = self.snapshot()
        self.run_sync()
        self.assertEqual(initial, self.snapshot())
        roles = list((self.root / ".codex/agents").glob("*.toml"))
        self.assertEqual(len(roles), 16)
        self.assertTrue((self.root / ".codex/agents/animalekarte-implementer.toml").exists())

    def test_symlink_destination_rejected(self):
        outside = self.root / "outside"
        outside.mkdir()
        (self.root / ".codex").symlink_to(outside, target_is_directory=True)
        self.run_sync(False)
        self.assertEqual(list(outside.iterdir()), [])

    def test_invalid_ownership_path_rejected(self):
        self.run_sync()
        ledger = self.root / ".codex/.mirror-ownership.json"
        ledger.write_text(json.dumps({"version": 1, "files": {"../outside": hashlib.sha256(b"x").hexdigest()}}))
        before = self.snapshot()
        self.run_sync(False)
        self.assertEqual(before, self.snapshot())

if __name__ == "__main__":
    unittest.main()
