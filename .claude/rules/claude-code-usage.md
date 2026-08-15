# Claude Code 使用ガイド — AnimalEkarte

## モデル選択

正本は `~/.claude/rules/ecc/common/performance.md`（根拠ベースでのエスカレーション方針）。カテゴリ別の固定モデル指定はしない（2026-07-21 doctor監査でグローバル方針との矛盾を解消・project側の表を削除）。

## エージェント × モデル対応

正本は各エージェントの frontmatter（`.claude/agents/*.md` の `model:`）。表による二重管理はしない（対応表がエージェント追加/削除のたびにドリフトしていたため廃止）。

## コンテキスト管理

| ctx% 残量 | アクション |
|----------|-----------|
| > 50% | 通常作業 |
| 20〜50% | 大きなタスクを開始しない。区切りで `/compact` |
| < 20% | 即座に `/compact` 実行 |

- タスク切り替え前に `/compact` を検討する
- `stop-save-progress.js` がセッション終了時に進捗スナップショットを保存

## 並行エージェント / Git 安全

- 正本: `git-worktree-safety.md`
- 並行タスクは **worktree 隔離必須**（共有 working tree 禁止）
- `git reset --hard` / `git clean -fd(x)` / discard-all restore / force-push は deny + PreToolUse でブロック

## 読み取り効率ルール（トークン削減）

**ファイル読み取り:**
- シンボルやパターン検索は `grep` (Bash) を使う。Read でファイル全体を読んでから探すな

**エージェント使用:**
- Explore エージェントは thoroughness を明示する: `"quick"` / `"medium"` / `"very thorough"`
- 既知パスへのアクセスは直接 Read — Explore/researcher エージェントは不要

**コンテキスト節約:**
- `refs/*.md` は必要な時だけ読む（テーブルに従って選択的に）
- CLAUDE.md は既にロード済みなので再読不要
- grep の結果が十分なら、そのファイルを全読みしない

## MCP 管理

- アクティブ MCP は同時 8 個以下を推奨
- 不要な MCP は `settings.local.json` の `disabledMcpServers` で無効化
- chrome-devtools は UI デバッグ時のみ使用（常時起動はリソース消費大）
