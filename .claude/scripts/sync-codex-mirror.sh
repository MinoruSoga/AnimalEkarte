#!/bin/bash
# .claude/agents + .claude/commands から .codex/agents (*.toml) と .codex/commands (*.md) を再生成する。
# .codex/agents, .codex/commands は untracked (gitignore の "Codex local state")。
# .claude/agents または .claude/commands を変更したら本スクリプトを再実行すること
# （通常は commit 時に pre-bash-commit-quality.js が自動再生成する。手動実行は任意）。
#
# .codex/agents/*.toml は Codex CLI の project-scoped custom agent 機構が実際に読み込む
# 生成物 — プロジェクトが trusted の場合にランタイムで消費される（手書きコピーではなく
# ここから常に再生成する）。TOML の developer_instructions フィールドに .claude/agents/*.md の
# 本文をそのまま埋め込む。commands は無変換の verbatim コピー。
#
# TOML 生成はエスケープが絡むため python3 に委譲する（bash/sed の入れ子クォートは壊れやすい）。
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"

python3 "$(dirname "$0")/test_sync_codex_mirror.py"
python3 "$(dirname "$0")/sync-codex-mirror.py" "$ROOT"
