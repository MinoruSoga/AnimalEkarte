"""Mock executable only; does not run Claude, a model, hooks, DB or shell payloads."""
from pathlib import Path
import json
import os
import shutil
import subprocess
import tempfile
import unittest

SOURCE = Path(__file__).resolve().parents[1] / "tools/claude-backend.py"
REPO_ROOT = Path(__file__).resolve().parents[2]
SIX_DISABLED = ["github", "chrome-devtools", "context7-mcp", "postgresql", "figma", "notion"]
SHARED = ".claude/settings.json"
OVERLAY = "backend/.claude/settings.local.json"


class LaunchTests(unittest.TestCase):
    def fixture(self, parent):
        root = Path(parent) / "worktree with spaces"
        (root / "scripts/tools").mkdir(parents=True)
        (root / ".claude").mkdir()
        (root / "backend/.claude").mkdir(parents=True)
        (root / "backend/pkg/sub").mkdir(parents=True)
        shutil.copyfile(SOURCE, root / "scripts/tools/claude-backend.py")
        (root / SHARED).write_text('{"permissions":{"deny":["Bash(sudo:*)"]}}')
        (root / OVERLAY).write_text(json.dumps({"disabledMcpjsonServers": SIX_DISABLED}))
        bindir = Path(parent) / "bin"
        bindir.mkdir()
        stub = bindir / "claude"
        stub.write_text("#!/usr/bin/env python3\nimport os,sys,json\nprint(json.dumps({'cwd':os.getcwd(),'argv':sys.argv[1:]}))\n")
        stub.chmod(0o700)
        return root.resolve(), {**os.environ, "PATH": str(bindir) + os.pathsep + os.environ["PATH"]}

    def launch(self, root, env, args=(), cwd=None, extra_env=None):
        run_env = dict(env)
        if extra_env:
            run_env.update(extra_env)
        return subprocess.run(["python3", str(root / "scripts/tools/claude-backend.py"), *args],
                              cwd=cwd or root / "backend", env=run_env, capture_output=True, text=True)

    def expect_not_launched(self, result, *unexpected):
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(result.stdout, "")
        self.assertIn("not launched", result.stderr)
        for text in unexpected:
            self.assertNotIn(text, result.stderr)

    def test_backend_launch_keeps_overlay_and_cli_arguments(self):
        with tempfile.TemporaryDirectory() as parent:
            root, env = self.fixture(parent)
            args = ["--model", "fixture-model", "--permission-mode", "plan", "review backend/internal/fixture"]
            result = self.launch(root, env, args)
            self.assertEqual(result.returncode, 0)
            payload = json.loads(result.stdout)
            self.assertEqual(payload["cwd"], str(root))
            self.assertEqual(payload["argv"], ["--settings", str(root / "backend/.claude/settings.local.json"), *args])

    def test_missing_shared_file_does_not_launch_stub(self):
        with tempfile.TemporaryDirectory() as parent:
            root, env = self.fixture(parent)
            (root / SHARED).unlink()
            self.expect_not_launched(self.launch(root, env))

    def test_missing_backend_file_does_not_launch_stub(self):
        with tempfile.TemporaryDirectory() as parent:
            root, env = self.fixture(parent)
            (root / OVERLAY).unlink()
            self.expect_not_launched(self.launch(root, env))

    def test_missing_both_files_does_not_launch_stub(self):
        with tempfile.TemporaryDirectory() as parent:
            root, env = self.fixture(parent)
            (root / SHARED).unlink()
            (root / OVERLAY).unlink()
            self.expect_not_launched(self.launch(root, env))

    def test_invalid_json_does_not_launch(self):
        for target in (SHARED, OVERLAY):
            with self.subTest(file=target), tempfile.TemporaryDirectory() as parent:
                root, env = self.fixture(parent)
                (root / target).write_text("{not valid json")
                self.expect_not_launched(self.launch(root, env), "{not valid json")

    def test_non_object_root_does_not_launch(self):
        for target in (SHARED, OVERLAY):
            with self.subTest(file=target), tempfile.TemporaryDirectory() as parent:
                root, env = self.fixture(parent)
                (root / target).write_text('["a", "b"]')
                self.expect_not_launched(self.launch(root, env))

    def test_bad_permissions_shape_does_not_launch(self):
        for content in ('{"permissions":"allow-all"}',
                        '{"permissions":{"allow":"Bash(git:*)"}}',
                        '{"permissions":{"deny":[1]}}',
                        '{"permissions":{"ask":{"Bash":true}}}'):
            with self.subTest(content=content), tempfile.TemporaryDirectory() as parent:
                root, env = self.fixture(parent)
                (root / SHARED).write_text(content)
                self.expect_not_launched(self.launch(root, env))

    def test_bad_hooks_shape_does_not_launch(self):
        for content in ('{"hooks":[]}',
                        '{"hooks":{"PreToolUse":{"matcher":"Bash"}}}'):
            with self.subTest(content=content), tempfile.TemporaryDirectory() as parent:
                root, env = self.fixture(parent)
                (root / SHARED).write_text(content)
                self.expect_not_launched(self.launch(root, env))

    def test_bad_disabled_list_does_not_launch(self):
        for content in ('{"disabledMcpjsonServers":"github"}',
                        '{"disabledMcpjsonServers":[1]}',
                        '{"disabledMcpjsonServers":{"github":true}}'):
            with self.subTest(content=content), tempfile.TemporaryDirectory() as parent:
                root, env = self.fixture(parent)
                (root / OVERLAY).write_text(content)
                self.expect_not_launched(self.launch(root, env))

    def test_present_null_fields_do_not_launch(self):
        for content in ('{"permissions":null}',
                        '{"permissions":{"allow":null}}',
                        '{"permissions":{"deny":null}}',
                        '{"permissions":{"ask":null}}',
                        '{"hooks":null}',
                        '{"disabledMcpjsonServers":null}'):
            with self.subTest(content=content), tempfile.TemporaryDirectory() as parent:
                root, env = self.fixture(parent)
                (root / SHARED).write_text(content)
                self.expect_not_launched(self.launch(root, env))

    def test_non_finite_json_constants_do_not_launch(self):
        for content in ('{"x":NaN}', '{"x":Infinity}', '{"x":-Infinity}',
                        '{"hooks":{"PreToolUse":[{"hooks":[{"type":"command","command":"x","timeout":NaN}]}]}}'):
            with self.subTest(content=content), tempfile.TemporaryDirectory() as parent:
                root, env = self.fixture(parent)
                (root / SHARED).write_text(content)
                self.expect_not_launched(self.launch(root, env))

    def test_bad_hook_entry_shape_does_not_launch(self):
        for content in (
            '{"hooks":"bad"}',
            '{"hooks":{"PreToolUse":[42]}}',
            '{"hooks":{"PreToolUse":[null]}}',
            '{"hooks":{"PreToolUse":[{"matcher":1,"hooks":[{"type":"command","command":"x"}]}]}}',
            '{"hooks":{"PreToolUse":[{"matcher":"Bash"}]}}',
            '{"hooks":{"PreToolUse":[{"hooks":"bad"}]}}',
            '{"hooks":{"PreToolUse":[{"hooks":[null]}]}}',
            '{"hooks":{"PreToolUse":[{"hooks":[{"command":"x"}]}]}}',
            '{"hooks":{"PreToolUse":[{"hooks":[{"type":"socket","command":"x"}]}]}}',
            '{"hooks":{"PreToolUse":[{"hooks":[{"type":"command"}]}]}}',
            '{"hooks":{"PreToolUse":[{"hooks":[{"type":"command","command":1}]}]}}',
            '{"hooks":{"PreToolUse":[{"hooks":[{"type":"command","command":"  "}]}]}}',
            '{"hooks":{"PreToolUse":[{"hooks":[{"type":"prompt","prompt":""}]}]}}',
            '{"hooks":{"PreToolUse":[{"hooks":[{"type":"http","command":"x"}]}]}}',
            '{"hooks":{"PreToolUse":[{"hooks":[{"type":"command","command":"x","timeout":true}]}]}}',
            '{"hooks":{"PreToolUse":[{"hooks":[{"type":"command","command":"x","timeout":-1}]}]}}',
            '{"hooks":{"PreToolUse":[{"hooks":[{"type":"command","command":"x","timeout":"5"}]}]}}',
        ):
            with self.subTest(content=content), tempfile.TemporaryDirectory() as parent:
                root, env = self.fixture(parent)
                (root / SHARED).write_text(content)
                self.expect_not_launched(self.launch(root, env))

    def test_unhashable_hook_type_does_not_traceback(self):
        for content in (
            '{"hooks":{"PreToolUse":[{"hooks":[{"type":[],"command":"x"}]}]}}',
            '{"hooks":{"PreToolUse":[{"hooks":[{"type":{},"command":"x"}]}]}}',
            '{"hooks":{"PreToolUse":[{"hooks":[{"type":null,"command":"x"}]}]}}',
        ):
            with self.subTest(content=content), tempfile.TemporaryDirectory() as parent:
                root, env = self.fixture(parent)
                (root / SHARED).write_text(content)
                result = self.launch(root, env)
                self.expect_not_launched(result)
                self.assertNotIn("Traceback", result.stderr)

    def test_timeout_numeric_bounds(self):
        big_int = json.dumps({"hooks": {"PreToolUse": [{"hooks": [
            {"type": "command", "command": "x", "timeout": 10 ** 400}]}]}})
        with tempfile.TemporaryDirectory() as parent:
            root, env = self.fixture(parent)
            (root / SHARED).write_text(big_int)
            result = self.launch(root, env)
            self.assertEqual(result.returncode, 0)
        for content in (
            '{"hooks":{"PreToolUse":[{"hooks":[{"type":"command","command":"x","timeout":0}]}]}}',
            '{"hooks":{"PreToolUse":[{"hooks":[{"type":"command","command":"x","timeout":1e400}]}]}}',
        ):
            with self.subTest(content=content), tempfile.TemporaryDirectory() as parent:
                root, env = self.fixture(parent)
                (root / SHARED).write_text(content)
                self.expect_not_launched(self.launch(root, env))

    def test_deeply_nested_json_does_not_traceback(self):
        nested = '{"a":' * 20000 + '{}' + '}' * 20000
        with tempfile.TemporaryDirectory() as parent:
            root, env = self.fixture(parent)
            (root / SHARED).write_text(nested)
            result = self.launch(root, env)
            self.expect_not_launched(result)
            self.assertNotIn("Traceback", result.stderr)

    def test_valid_hook_entry_variants_launch(self):
        content = json.dumps({"hooks": {"PreToolUse": [
            {"matcher": "Bash", "hooks": [
                {"type": "command", "command": "node .claude/hooks/fixture.js", "timeout": 5},
                {"type": "prompt", "prompt": "fixture prompt"},
                {"type": "agent", "prompt": "fixture prompt"},
                {"type": "http", "url": "https://fixture.invalid/hook"},
            ]}
        ]}})
        with tempfile.TemporaryDirectory() as parent:
            root, env = self.fixture(parent)
            (root / SHARED).write_text(content)
            result = self.launch(root, env)
            self.assertEqual(result.returncode, 0)

    def test_actual_repo_settings_pass_validation(self):
        for target, source in ((SHARED, REPO_ROOT / ".claude/settings.json"),
                               (OVERLAY, REPO_ROOT / "backend/.claude/settings.local.json")):
            with self.subTest(file=target), tempfile.TemporaryDirectory() as parent:
                root, env = self.fixture(parent)
                (root / target).write_bytes(source.read_bytes())
                result = self.launch(root, env)
                self.assertEqual(result.returncode, 0)

    def test_caller_cwds_resolve_to_root(self):
        with tempfile.TemporaryDirectory() as parent:
            root, env = self.fixture(parent)
            for cwd in (root, root / "backend", root / "backend/pkg/sub"):
                with self.subTest(cwd=str(cwd)):
                    payload = json.loads(self.launch(root, env, cwd=cwd).stdout)
                    self.assertEqual(payload["cwd"], str(root))
                    self.assertEqual(payload["argv"][0], "--settings")
                    self.assertTrue(Path(payload["argv"][1]).is_absolute())

    def test_root_settings_bytes_preserved(self):
        with tempfile.TemporaryDirectory() as parent:
            root, env = self.fixture(parent)
            before = (root / SHARED).read_bytes()
            result = self.launch(root, env)
            self.assertEqual(result.returncode, 0)
            self.assertEqual((root / SHARED).read_bytes(), before)

    def test_overlay_has_six_disabled_names(self):
        with tempfile.TemporaryDirectory() as parent:
            root, env = self.fixture(parent)
            payload = json.loads(self.launch(root, env).stdout)
            overlay = json.loads(Path(payload["argv"][1]).read_text())
            self.assertEqual(overlay["disabledMcpjsonServers"], SIX_DISABLED)

    def test_repo_backend_settings_lists_six_names(self):
        overlay = json.loads((REPO_ROOT / "backend/.claude/settings.local.json").read_text())
        self.assertEqual(overlay["disabledMcpjsonServers"], SIX_DISABLED)

    def test_denied_launch_arguments(self):
        variants = [
            ["--settings", "/tmp/SENTINEL9X"],
            ["--settings=/tmp/SENTINEL9X"],
            ["--setting-sources", "user"],
            ["--setting-sources=user"],
            ["--bare"],
            ["--safe-mode"],
            ["--restricted"],
            ["--dangerously-skip-permissions"],
            ["--allow-dangerously-skip-permissions"],
            ["--permission-mode", "bypassPermissions"],
            ["--permission-mode=bypassPermissions"],
        ]
        with tempfile.TemporaryDirectory() as parent:
            root, env = self.fixture(parent)
            for args in variants:
                with self.subTest(args=args):
                    self.expect_not_launched(self.launch(root, env, args), "SENTINEL9X", "bypassPermissions")

    def test_prompt_arguments_after_separator_allowed(self):
        with tempfile.TemporaryDirectory() as parent:
            root, env = self.fixture(parent)
            args = ["--model", "fixture-model", "--", "--bare", "--safe-mode", "explain --restricted"]
            result = self.launch(root, env, args)
            self.assertEqual(result.returncode, 0)
            payload = json.loads(result.stdout)
            self.assertEqual(payload["argv"], ["--settings", str(root / "backend/.claude/settings.local.json"), *args])

    def test_prompt_mentioning_flag_word_allowed(self):
        with tempfile.TemporaryDirectory() as parent:
            root, env = self.fixture(parent)
            result = self.launch(root, env, ["explain --bare handling"])
            self.assertEqual(result.returncode, 0)

    def test_option_value_matching_flag_is_conservatively_rejected(self):
        with tempfile.TemporaryDirectory() as parent:
            root, env = self.fixture(parent)
            for args in (["--model", "--bare"], ["--permission-mode", "--bare"]):
                with self.subTest(args=args):
                    self.expect_not_launched(self.launch(root, env, args))

    def test_equals_value_matching_flag_allowed(self):
        with tempfile.TemporaryDirectory() as parent:
            root, env = self.fixture(parent)
            for args in (["--model=--bare"], ["--append-system-prompt=--bare"]):
                with self.subTest(args=args):
                    result = self.launch(root, env, args)
                    self.assertEqual(result.returncode, 0)

    def test_environment_toggle_rejection(self):
        for var in ("CLAUDE_CODE_SIMPLE", "CLAUDE_CODE_SAFE_MODE"):
            for value in ("1", "true"):
                with self.subTest(var=var, value=value), tempfile.TemporaryDirectory() as parent:
                    root, env = self.fixture(parent)
                    self.expect_not_launched(self.launch(root, env, extra_env={var: value}))
        with tempfile.TemporaryDirectory() as parent:
            root, env = self.fixture(parent)
            result = self.launch(root, env, extra_env={"CLAUDE_CODE_SIMPLE": "0"})
            self.assertEqual(result.returncode, 0)


if __name__ == "__main__":
    unittest.main()
