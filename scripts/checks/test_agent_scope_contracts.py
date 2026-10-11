#!/usr/bin/env python3
"""Repository scope regression checks; no user configuration or services read."""
import json
import re
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

    def test_operational_roles_keep_shell_access(self):
        for name in ('build-error-resolver', 'clinic-isolation-auditor',
                     'code-simplifier', 'debugger', 'go-build-resolver',
                     'go-reviewer', 'implementer', 'react-reviewer',
                     'refactor-cleaner', 'security-analyst', 'tdd-guide',
                     'typescript-reviewer'):
            with self.subTest(role=name):
                self.assertIn('Bash', self.role_tools(name))

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
