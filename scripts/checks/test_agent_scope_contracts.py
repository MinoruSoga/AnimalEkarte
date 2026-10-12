#!/usr/bin/env python3
"""Repository scope regression checks; no user configuration or services read."""
import json
import os
import re
import subprocess
from pathlib import Path
import tomllib
import unittest

ROOT = Path(__file__).resolve().parents[2]


class AgentScopeContracts(unittest.TestCase):
    def role_tools(self, name):
        text = (ROOT / '.claude/agents' / (name + '.md')).read_text()
        frontmatter = text.split('---\n', 2)[1]
        return json.loads(re.search(r'^tools: (.+)$', frontmatter, re.MULTILINE)[1])

    def test_analysis_roles_do_not_have_shell_access(self):
        for name in ('architect', 'silent-failure-hunter', 'test-strategist'):
            with self.subTest(role=name):
                self.assertNotIn('Bash', self.role_tools(name))

    def test_test_generation_keeps_claude_edit_tools(self):
        tools = self.role_tools('test-strategist')
        for tool in ('Read', 'Edit', 'Write', 'Grep', 'Glob'):
            self.assertIn(tool, tools)

    def test_repository_roles_do_not_have_shell_access(self):
        roles = sorted((ROOT / '.claude/agents').glob('*.md'))
        self.assertTrue(roles)
        for path in roles:
            with self.subTest(role=path.stem):
                tools = self.role_tools(path.stem)
                self.assertTrue(tools)
                self.assertLessEqual(set(tools), {'Read', 'Edit', 'Write', 'Grep', 'Glob'})

    def test_backend_security_denials_match_shared_settings(self):
        shared = json.loads((ROOT / '.claude/settings.json').read_text())
        backend = json.loads((ROOT / 'backend/.claude/settings.local.json').read_text())
        self.assertTrue(shared['permissions']['deny'])
        self.assertEqual(backend['permissions']['deny'], shared['permissions']['deny'])
        self.assertFalse(backend['permissions'].get('allow'))
        self.assertEqual(backend['disabledMcpjsonServers'],
                         ['github', 'chrome-devtools', 'context7-mcp', 'postgresql', 'figma', 'notion'])

    def test_raw_device_redirect_denials_cover_direct_spellings(self):
        shared = json.loads((ROOT / '.claude/settings.json').read_text())
        deny = shared['permissions']['deny']
        for family in ('disk*', 'rdisk*', 'sd*', 'hd*', 'vd*', 'xvd*', 'nvme*',
                       'mmcblk*', 'loop*', 'dm-*', 'md*', 'ada*', 'da*', 'nvd*', 'mapper/*'):
            with self.subTest(family=family):
                self.assertIn(f'Bash(*> /dev/{family})', deny)
                self.assertIn(f'Bash(*>/dev/{family})', deny)
        self.assertNotIn('Bash(*> /dev/*)', deny)
        self.assertNotIn('Bash(*>/dev/*)', deny)

    def test_backend_security_hook_runs_from_root_or_backend(self):
        backend = json.loads((ROOT / 'backend/.claude/settings.local.json').read_text())
        entries = backend['hooks']['PreToolUse']
        self.assertEqual(len(entries), 1)
        self.assertEqual(entries[0]['matcher'], 'Bash')
        hook = entries[0]['hooks'][0]
        self.assertEqual(hook['type'], 'command')
        cases = (
            ('git status --short', 0),
            ('git reset --hard HEAD', 2),
            ('dd if=/dev/zero of=/dev/sda', 2),
            ('echo ok > /dev/null', 0),
        )
        for cwd in (ROOT, ROOT / 'backend'):
            env = {'PATH': os.environ['PATH'], 'CLAUDE_PROJECT_DIR': str(cwd),
                   'GIT_CONFIG_COUNT': '1', 'GIT_CONFIG_KEY_0': 'safe.directory',
                   'GIT_CONFIG_VALUE_0': str(ROOT)}
            for command, expected in cases:
                with self.subTest(cwd=str(cwd), command=command):
                    result = subprocess.run(['sh', '-c', hook['command']], cwd=cwd, env=env,
                                            input=json.dumps({'tool_name': 'Bash', 'tool_input': {'command': command}}),
                                            capture_output=True, text=True, timeout=15)
                    self.assertEqual(result.returncode, expected, result.stderr)
                    if expected == 2:
                        self.assertEqual(json.loads(result.stdout)['decision'], 'deny')
                        self.assertNotIn(command, result.stdout + result.stderr)
            result = subprocess.run(['sh', '-c', hook['command']], cwd=cwd, env=env,
                                    input='{invalid', capture_output=True, text=True, timeout=15)
            self.assertEqual(result.returncode, 2)
            self.assertNotIn('{invalid', result.stdout + result.stderr)

    def test_shell_removed_roles_keep_read_tools(self):
        expected = {
            'reviewer': ('Read', 'Grep', 'Glob'),
            'formatter': ('Read', 'Edit'),
            'database-reviewer': ('Read', 'Grep', 'Glob'),
        }
        for name, tools in expected.items():
            with self.subTest(role=name):
                actual = self.role_tools(name)
                self.assertNotIn('Bash', actual)
                for tool in tools:
                    self.assertIn(tool, actual)

    def test_edit_roles_keep_write_or_edit(self):
        for name in ('build-error-resolver', 'code-simplifier', 'debugger',
                     'formatter', 'go-build-resolver', 'implementer',
                     'refactor-cleaner', 'tdd-guide', 'test-strategist'):
            with self.subTest(role=name):
                self.assertTrue({'Edit', 'Write'} & set(self.role_tools(name)))

    def test_role_prohibition_prose_present(self):
        expected = {
            'reviewer': ('BLOCKED',),
            'formatter': ('BLOCKED',),
            'database-reviewer': ('BLOCKED', 'ユーザー実行例'),
            'build-error-resolver': ('ユーザー実行例', '変更ファイルスコープ'),
            'refactor-cleaner': ('make codegen', 'ユーザー実行'),
        }
        for name, markers in expected.items():
            with self.subTest(role=name):
                body = (ROOT / '.claude/agents' / f'{name}.md').read_text(encoding='utf-8')
                for marker in markers:
                    self.assertIn(marker, body)

    def test_codex_test_strategy_does_not_expand_write_scope(self):
        manifest = json.loads((ROOT / '.claude/codex-agent-manifest.json').read_text())
        role = next(role for role in manifest['roles']
                    if role['name'] == 'animalekarte-test-strategist')
        self.assertEqual(role['sandbox_mode'], 'read-only')

    def test_codex_inherits_generic_user_settings(self):
        path = ROOT / '.codex/config.toml'
        config = tomllib.loads(path.read_text()) if path.exists() else {}
        for key in ('model', 'model_reasoning_effort', 'sandbox_mode', 'approval_policy',
                    'features', 'plugins', 'shell_environment_policy', 'agents'):
            self.assertNotIn(key, config, f'{key} belongs to user defaults')
        self.assertFalse(config.get('mcp_servers'))

    def test_claude_inherits_user_preferences(self):
        config = json.loads((ROOT / '.claude/settings.json').read_text())
        for key in ('model', 'outputStyle', 'effort', 'alwaysThinkingEnabled',
                    'includeCoAuthoredBy', 'teammateMode'):
            self.assertNotIn(key, config)
        for key in ('CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS', 'MAX_THINKING_TOKENS',
                    'CLAUDE_AUTOCOMPACT_PCT_OVERRIDE', 'ENABLE_CLAUDEAI_MCP_SERVERS',
                    'ECC_GATEGUARD', 'ECC_DISABLED_HOOKS'):
            self.assertNotIn(key, config.get('env', {}))

    def test_shared_project_does_not_provision_personal_connectors(self):
        self.assertFalse(json.loads((ROOT / '.mcp.json').read_text()).get('mcpServers'))

    def test_project_does_not_blanket_allow_external_or_shell_mutation(self):
        config = json.loads((ROOT / '.claude/settings.json').read_text())
        allow = config.get('permissions', {}).get('allow', [])
        for rule in ('Bash(make:*)', 'Bash(gh:*)', 'Bash(docker compose exec:*)',
                     'Bash(git merge:*)', 'Bash(git:*)', 'Bash(curl:*)'):
            self.assertNotIn(rule, allow)

    def test_project_runtime_safety_is_retained(self):
        text = (ROOT / 'AGENTS.md').read_text()
        self.assertIn('claim/', text)
        self.assertIn('migration', text.lower())
        self.assertIn('clinic', text.lower())
        self.assertIn('Docker', text)

    def test_wrangler_does_not_require_seedlogin_operator_secrets(self):
        text = (ROOT / 'backend/wrangler.jsonc').read_text()
        self.assertIn('SEEDLOGIN_OPERATOR', text)
        self.assertNotRegex(text, r'"SEEDLOGIN_OPERATOR_(EMAIL|NAME|PASSWORD)"')


if __name__ == '__main__':
    unittest.main()
