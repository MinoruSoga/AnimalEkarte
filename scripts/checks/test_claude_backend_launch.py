"""Mock executable only; does not run Claude, a model, hooks, DB or shell payloads."""
from pathlib import Path
import json
import os
import shutil
import subprocess
import tempfile
import unittest

SOURCE = Path(__file__).resolve().parents[1] / "tools/claude-backend.py"

class LaunchTests(unittest.TestCase):
    def fixture(self, parent):
        root = Path(parent) / "worktree with spaces"
        (root / "scripts/tools").mkdir(parents=True)
        (root / ".claude").mkdir()
        (root / "backend/.claude").mkdir(parents=True)
        shutil.copyfile(SOURCE, root / "scripts/tools/claude-backend.py")
        (root / ".claude/settings.json").write_text('{"permissions":{"deny":["Bash(sudo:*)"]}}')
        (root / "backend/.claude/settings.local.json").write_text('{"disabledMcpjsonServers":["fixture"]}')
        bindir = Path(parent) / "bin"
        bindir.mkdir()
        stub = bindir / "claude"
        stub.write_text("#!/usr/bin/env python3\nimport os,sys,json\nprint(json.dumps({'cwd':os.getcwd(),'argv':sys.argv[1:]}))\n")
        stub.chmod(0o700)
        return root.resolve(), {**os.environ, "PATH": str(bindir) + os.pathsep + os.environ["PATH"]}
    def test_backend_launch_keeps_overlay_and_cli_arguments(self):
        with tempfile.TemporaryDirectory() as parent:
            root, env = self.fixture(parent)
            args = ["--model", "fixture-model", "--permission-mode", "plan", "review backend/internal/fixture"]
            result = subprocess.run(["python3", str(root / "scripts/tools/claude-backend.py"), *args],
                                    cwd=root / "backend", env=env, capture_output=True, text=True)
            self.assertEqual(result.returncode, 0)
            payload = json.loads(result.stdout)
            self.assertEqual(payload["cwd"], str(root))
            self.assertEqual(payload["argv"], ["--settings", str(root / "backend/.claude/settings.local.json"), *args])
    def test_missing_shared_file_does_not_launch_stub(self):
        with tempfile.TemporaryDirectory() as parent:
            root, env = self.fixture(parent)
            (root / ".claude/settings.json").unlink()
            result = subprocess.run(["python3", str(root / "scripts/tools/claude-backend.py")],
                                    cwd=root / "backend", env=env, capture_output=True, text=True)
            self.assertNotEqual(result.returncode, 0)
            self.assertEqual(result.stdout, "")
            self.assertIn("not launched", result.stderr)

if __name__ == "__main__":
    unittest.main()
