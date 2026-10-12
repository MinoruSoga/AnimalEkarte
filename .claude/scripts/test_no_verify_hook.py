#!/usr/bin/env python3
"""Run hook logic on synthetic commands without executing the commands."""
import json
import pathlib
import subprocess
import unittest

HOOK = pathlib.Path(__file__).resolve().parents[1] / 'hooks/pre-bash-block-no-verify.js'


class NoVerifyHookTest(unittest.TestCase):
    def test_blocks_bypass_flags_but_allows_normal_checks_and_push_dry_run(self):
        for command, expected in (
            ('git commit --no-verify -m synthetic', 2),
            ('git commit -n -m synthetic', 2),
            ('git commit -an -m synthetic', 2),
            ('git commit --no-gpg-sign -m synthetic', 2),
            ('git push --no-verify origin synthetic', 2),
            ('git push -n origin synthetic', 0),
            ('git commit -m synthetic', 0),
            ('git diff --check', 0),
        ):
            with self.subTest(command=command):
                result = subprocess.run(['node', str(HOOK)], input=json.dumps({
                    'tool_name': 'Bash', 'tool_input': {'command': command},
                }), text=True, capture_output=True, timeout=5)
                self.assertEqual(result.returncode, expected, result.stderr)
                if command == 'git commit -n -m synthetic':
                    self.assertIn('blocked -n alias', result.stderr)


if __name__ == '__main__':
    unittest.main()
