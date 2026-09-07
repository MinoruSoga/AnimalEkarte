#!/usr/bin/env python3
"""Static safety contracts for the skill-audit safety packet.

This test reads instruction text and generated mirrors only. It deliberately
does not run command snippets, inspect environment files, or contact runtime
services.
"""

from __future__ import annotations

import pathlib
import unittest


ROOT = pathlib.Path(__file__).resolve().parents[2]


def read(relative_path: str) -> str:
    return (ROOT / relative_path).read_text(encoding="utf-8")


class InstructionSafetyContractsTest(unittest.TestCase):
    def test_docker_recovery_is_non_destructive(self) -> None:
        text = read(".claude/skills/docker-patterns/SKILL.md")
        self.assertIn("通常の cache/node_modules 復旧には使わない", text)
        self.assertIn("ユーザーの明示承認", text)
        self.assertNotIn("docker compose down -v", text)
        self.assertNotIn("make clean  # または", text)
        self.assertNotIn("make clean    # キャッシュクリア + 再ビルド", text)

    def test_vite_logging_and_scan_examples_are_safe(self) -> None:
        text = read(".claude/skills/security-checklist/SKILL.md")
        self.assertIn("公開設定", text)
        self.assertIn("シークレットマネージャー", text)
        self.assertIn('"staff_id", staffID', text)
        self.assertIn("match_file_count=", text)
        self.assertNotIn("VITE_API_KEY", text)
        self.assertNotIn('"email", req.Email', text)
        self.assertNotIn("grep -rn", text)

    def test_postgres_diagnostics_are_value_safe(self) -> None:
        text = read(".claude/skills/postgres-patterns/SKILL.md")
        self.assertIn("allowlist", text)
        self.assertIn("=set", text)
        self.assertIn("=unset", text)
        self.assertNotIn("env | grep DB", text)
        self.assertNotIn("DB_PASSWORD=", text)

    def test_rollback_uses_captured_baseline_and_owned_patch(self) -> None:
        testing = read(".claude/skills/golang-testing/SKILL.md")
        refactor = read(".claude/commands/refactor.md")
        self.assertIn("byte baseline", testing)
        self.assertIn("隔離コピー", testing)
        self.assertIn("task-owned patch only", refactor)
        self.assertIn("RED before GREEN", refactor)
        self.assertNotIn("HEAD と byte-identical", testing)
        self.assertIn("`git checkout -- <file>`、HEAD 基準の巻き戻しは他者 WIP を失わせるため使用しない", refactor)
        self.assertNotIn("git checkout -- <file> で即時リバート", refactor)

    def test_fresh_db_guidance_is_disposable_and_conditional(self) -> None:
        for relative_path in (
            ".claude/skills/migration-seed-safety/SKILL.md",
            ".claude/skills/golang-testing/SKILL.md",
            ".claude/skills/scoped-verification-gates/SKILL.md",
            ".claude/skills/stg-release-readiness/SKILL.md",
        ):
            with self.subTest(relative_path=relative_path):
                text = read(relative_path)
                self.assertIn("disposable", text)
                self.assertIn("docs-only", text)
                self.assertIn("ユーザー", text)
                self.assertNotIn("docker compose down -v", text)
                self.assertNotIn("DROP DATABASE", text)

    def test_generated_safety_mirrors_match_canonical_sources(self) -> None:
        for skill in (
            "docker-patterns",
            "security-checklist",
            "postgres-patterns",
            "golang-testing",
            "migration-seed-safety",
            "scoped-verification-gates",
            "stg-release-readiness",
        ):
            with self.subTest(skill=skill):
                self.assertEqual(
                    read(f".claude/skills/{skill}/SKILL.md"),
                    read(f".agents/skills/{skill}/SKILL.md"),
                )

        source_refactor = read(".claude/commands/refactor.md")
        self.assertEqual(source_refactor, read(".codex/commands/refactor.md"))
        self.assertEqual(read(".claude/commands/harness.md"), read(".codex/commands/harness.md"))
        self.assertEqual(read(".claude/commands/implement.md"), read(".codex/commands/implement.md"))

        source_body = source_refactor.split("---\n", 2)[2]
        wrapper = read(".agents/skills/source-command-refactor/SKILL.md")
        self.assertIn('name: "source-command-refactor"', wrapper)
        self.assertIn(source_body, wrapper)


if __name__ == "__main__":
    unittest.main()
