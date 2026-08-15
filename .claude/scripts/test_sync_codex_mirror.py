#!/usr/bin/env python3
"""Regression tests for the Codex mirror generator."""

import subprocess
import sys
import tempfile
import tomllib
import unittest
from pathlib import Path


SCRIPT = Path(__file__).with_name("sync-codex-mirror.py")


class SyncCodexMirrorTest(unittest.TestCase):
    def test_generated_agent_toml_preserves_backslashes(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            agents = root / ".claude" / "agents"
            commands = root / ".claude" / "commands"
            agents.mkdir(parents=True)
            commands.mkdir(parents=True)

            source_body = """Run this command:

```bash
grep -rn "console\\.\\(log\\|warn\\|error\\)" frontend/src/
```
"""
            agents.joinpath("regex-agent.md").write_text(
                f"""---
description: Finds console calls.
---

{source_body}""",
                encoding="utf-8",
            )

            subprocess.run(
                [sys.executable, str(SCRIPT), str(root)],
                check=True,
                capture_output=True,
                text=True,
            )

            generated = root / ".codex" / "agents" / "regex-agent.toml"
            with generated.open("rb") as file:
                agent = tomllib.load(file)

            self.assertEqual(f"{source_body}\n", agent["developer_instructions"])


if __name__ == "__main__":
    unittest.main()
