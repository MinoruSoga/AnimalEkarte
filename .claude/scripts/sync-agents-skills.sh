#!/bin/bash
# .claude/skills + .claude/commands から .agents/skills (agy 等の他エージェント向けミラー) を再生成する。
# .agents/ は untracked の生成物。.claude/skills または .claude/commands を変更したら本スクリプトを再実行すること
# （通常は commit 時に pre-bash-commit-quality.js が自動再生成する。手動実行は任意）。
#
# 変換規則:
#   - commands/X.md → skills/source-command-X/SKILL.md ラッパー
#   - パス参照・固有名詞は無変換（.claude/ のまま）— ルート AGENTS.md・.codex/CODEX.md が
#     既に .claude/CLAUDE.md を直接正本指定する慣行を確立しており、sed によるパス翻訳
#     （.claude/→.Codex/ 等）は実在しない参照を量産するだけだったため廃止した。
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SRC_SKILLS="$ROOT/.claude/skills"
SRC_COMMANDS="$ROOT/.claude/commands"
DEST="$ROOT/.agents/skills"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

# 1. skills をコピー
for dir in "$SRC_SKILLS"/*/; do
    cp -R "$dir" "$TMP/$(basename "$dir")"
done

# 2. commands を source-command-* スキルラッパーに変換
for f in "$SRC_COMMANDS"/*.md; do
    cmd="$(basename "$f" .md)"
    desc="$(awk -F': ' '/^description:/{print $2; exit}' "$f")"
    mkdir -p "$TMP/source-command-$cmd"
    {
        echo '---'
        echo "name: \"source-command-$cmd\""
        echo "description: \"$desc\""
        echo '---'
        echo
        echo "# source-command-$cmd"
        echo
        echo "Use this skill when the user asks to run the migrated source command \`$cmd\`."
        echo
        echo '## Command Template'
        # frontmatter (1つ目の --- ペア) を除いた本文
        awk 'BEGIN{fm=0} /^---$/{ if (fm<2) {fm++; next} } fm>=2{print}' "$f"
    } > "$TMP/source-command-$cmd/SKILL.md"
done

# 3. スワップ（無変換 — .claude/ 参照はそのまま。sed による固有名詞/パス翻訳はしない）
rm -rf "$DEST"
mkdir -p "$(dirname "$DEST")"
mv "$TMP" "$DEST"
trap - EXIT

echo "regenerated: $DEST"
echo "  skills:   $(find "$DEST" -maxdepth 1 -type d ! -name 'source-command-*' | tail -n +2 | wc -l | tr -d ' ')"
echo "  commands: $(find "$DEST" -maxdepth 1 -type d -name 'source-command-*' | wc -l | tr -d ' ')"
