# Claude Code 使用ガイド — AnimalEkarte

## モデル選択

| タスク | 推奨モデル | 理由 |
|--------|-----------|------|
| アーキテクチャ設計・セキュリティ分析・大規模リファクタ計画 | opus | 複雑な多層判断が必要 |
| 機能実装・コードレビュー・テスト作成・デバッグ | sonnet | 日常作業の最適コスパ |
| ファイル検索・軽微な修正・フォーマット確認・質問応答 | haiku | 低コスト・高速 |

## エージェント × モデル対応

| エージェント | 用途 | モデル |
|------------|------|-------|
| `planner` | 実装計画・リスク分析 | opus |
| `architect` | システム設計判断 | opus |
| `security-analyst` | 脆弱性分析 | opus |
| `healthcare-reviewer` | 臨床データ安全性・clinic_id隔離・患者記録保護 | opus |
| `go-expert` | Go idiom・パフォーマンス | sonnet |
| `go-reviewer` | Go コードレビュー | sonnet |
| `go-build-resolver` | Go ビルドエラー・go vet 解決 | sonnet |
| `typescript-reviewer` | TS/React コードレビュー | sonnet |
| `build-error-resolver` | TypeScript ビルドエラー・型エラー解決 | sonnet |
| `implementer` | 機能実装 | sonnet |
| `refactor-cleaner` | リファクタリング | sonnet |
| `database-reviewer` | DB スキーマレビュー | sonnet |
| `performance-optimizer` | パフォーマンス改善 | sonnet |
| `tdd-guide` | TDD ワークフロー | sonnet |
| `test-strategist` | テスト戦略 | sonnet |
| `silent-failure-hunter` | 暗黙的バグ検出 | sonnet |
| `debugger` | 軽微なデバッグ | haiku |
| `formatter` | コードフォーマット | haiku |
| `researcher` | ファイル検索・情報収集 | haiku |
| `reviewer` | 軽量レビュー | haiku |

## `/think` 判断基準

**使う（高コスト・高価値）:**
- アーキテクチャ設計、大規模リファクタ
- セキュリティ設計、脆弱性分析
- 原因不明バグの調査・デバッグ
- 複数トレードオフを伴う技術選定

**スキップ（低コスト・明確）:**
- ファイル読み込み・検索・調査
- 簡単なタイポ修正、コメント更新
- 既知パターンの実装
- 質問への回答・説明

> 迷ったら **スキップ**。Extended Thinking は 3〜5x のトークン消費。

## コンテキスト管理

| ctx% 残量 | アクション |
|----------|-----------|
| > 50% | 通常作業 |
| 20〜50% | 大きなタスクを開始しない。区切りで `/compact` |
| < 20% | 即座に `/compact` 実行 |

- タスク切り替え前に `/compact` を検討する
- `pre-compact-save-state.js` がコンパクト前に git 状態・進捗を自動保存
- `stop-save-progress.js` がセッション終了時に進捗スナップショットを保存

## 読み取り効率ルール（トークン削減）

**ファイル読み取り:**
- 大きなファイルは `limit`/`offset` で必要部分だけ読む（2000行全読みはしない）
- 編集直後のファイルは再読み禁止 — Edit/Write は成功すれば変更済みが保証される
- シンボルやパターン検索は `grep` (Bash) を使う。Read でファイル全体を読んでから探すな

**エージェント使用:**
- Explore エージェントは thoroughness を明示する: `"quick"` / `"medium"` / `"very thorough"`
- 独立したクエリは単一メッセージで並列実行する（逐次は2〜5倍のコスト）
- 既知パスへのアクセスは直接 Read — Explore/researcher エージェントは不要

**コンテキスト節約:**
- `refs/*.md` は必要な時だけ読む（テーブルに従って選択的に）
- CLAUDE.md は既にロード済みなので再読不要
- grep の結果が十分なら、そのファイルを全読みしない

## MCP 管理

- アクティブ MCP は同時 8 個以下を推奨
- 不要な MCP は `settings.local.json` の `disabledMcpServers` で無効化
- chrome-devtools は UI デバッグ時のみ使用（常時起動はリソース消費大）

## Docker ファースト再確認

npm/go コマンドはローカル実行禁止。必ず `docker compose exec frontend|backend` 経由で実行。
