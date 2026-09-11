#!/bin/bash
# .claude/agents + .claude/commands から .codex/agents (*.toml) と .codex/commands (*.md) を再生成する。
# .codex/agents, .codex/commands は untracked (gitignore の "Codex local state")。
# .claude/agents または .claude/commands を変更したら本スクリプトを再実行すること
# （通常は commit 時に pre-bash-commit-quality.js が自動再生成する。手動実行は任意）。
#
# .claude/codex-agent-manifest.json はプロジェクト固有 role の明示 allowlist。
# 一般 role はユーザースコープへ委譲。生成物は静的 TOML であり、実際の role
# discovery は起動中の Codex で別途確認する。commands は無変換コピー。
# 全出力を事前検証し、既知の生成 bytes/hash だけ更新。外国 WIP は上書きしない。
#
# TOML 生成はエスケープが絡むため python3 に委譲する（bash/sed の入れ子クォートは壊れやすい）。
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"

python3 -B "$(dirname "$0")/test_sync_codex_mirror.py"
python3 -B "$(dirname "$0")/sync-codex-mirror.py" "$ROOT"
