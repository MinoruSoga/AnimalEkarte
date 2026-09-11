#!/usr/bin/env python3
"""Repository scope regression checks; no user configuration or services read."""
import json
from pathlib import Path
import tomllib
import unittest

ROOT = Path(__file__).resolve().parents[1]


class AgentScopeContracts(unittest.TestCase):
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
