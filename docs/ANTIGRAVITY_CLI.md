# Antigravity CLI (`agy`) 運用メモ

本プロジェクトで AI エージェントを使う際の、Antigravity CLI (`agy`) の最低限の運用メモです。

## 背景: Gemini CLI からの移行

Google One / 無料枠 (Gemini Code Assist for individuals) では、Gemini CLI が Antigravity CLI に移行します。
Gemini CLI は 2026-06-18 に該当ティア向けのリクエスト提供が停止します（順次移行告知が出ます）。

このリポジトリでは、Gemini CLI 用コンテキストとして `GEMINI.md` を引き続き保持しますが、
実運用は Antigravity CLI (`agy`) を前提に進めてください。

## インストールと起動

インストール (macOS / Linux):

```sh
curl -fsSL https://antigravity.google/cli/install.sh | bash
```

PATH 反映後に起動:

```sh
agy
```

## よく出る警告と対処

### MCP issues detected

MCP (外部ツール連携) の接続や設定に問題がある合図です。
Antigravity CLI 内で以下を実行して、落ちている MCP を特定します。

```text
/mcp list
```

### Skill conflict detected (overriding ...)

同名の Skill が「グローバル」と「プロジェクトローカル」に共存している合図です。
通常はプロジェクトローカル側を優先して問題ありませんが、意図しない上書きを避けるなら重複を解消します。

確認:

```sh
ls -la ~/.agents/skills/
ls -la ./.agents/skills/
```

## このリポジトリでの作業ルール (重要)

- 規約の SSOT はプロジェクトルートの `.claude/CLAUDE.md`。
- ローカルで `pnpm` / `go` を直接叩かず、原則 `docker compose exec ...` を使う。
- 影響が大きいコマンド (テスト、lint、install、DB 操作等) は自動実行しない。必要ならコマンドを提示して人が実行する。

## エージェントに依頼する時のテンプレ

Antigravity CLI でプロンプトを作る時は、スコープと禁止事項を最初に固定してください。

```text
あなたはこのリポジトリのドキュメント更新担当です。

まず以下を読んでください:
- docs/README.md
- docs/AI_DEVELOPMENT_WORKFLOW.md
- .claude/CLAUDE.md
- GEMINI.md

やること:
- docs/ 配下のドキュメントを、現在の運用 (Antigravity CLI) に合わせて最新化する

制約:
- コード (backend/, frontend/) は変更しない
- docs/ 以外のファイルは変更しない
- 事実と異なる数字 (テーブル数/ハンドラー数等) を推測で書き換えない
- 変更したファイル一覧と、変更理由を短く報告する
```

