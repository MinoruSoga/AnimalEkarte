---
description: "タスク実装（todo.md の TASK-XXX / bug.md の BUG-XXX）→ セルフレビュー → クローズ"
argument-hint: "TASK-XXX | BUG-XXX"
---

# タスク実装

指定されたタスクIDを実装する。台帳は repo 直下の 2 ファイル（いずれも git 追跡・ローカル連番）:

- `TASK-XXX` → `todo.md`（残タスク台帳。索引/サマリー表 + `## 個別タスク詳細` の `### TASK-XXX:` 節）
- `BUG-XXX` → `bug.md`（受入テストバグ台帳。`## BUG-XXX:` 節 + `### 実装計画`）

タスクが GitHub Issue（`#NNN`）を参照する場合、仕様・受け入れ条件の正本は該当 Issue 本文とコメント。`3-session-agent.html` は Issue 分類ビューであり台帳ではない（旧 `#ledger` は 2026-07-31 廃止・経緯は git 履歴）。

スキル `implement-issue` を使用して、以下の5フェーズを実行:
1. タスク読み込み・claim 確認（AGENTS.md packet claim protocol）・依存関係チェック
2. コンテキスト収集（対象ファイル・参照実装・コーディングルール）
3. コード規約準拠の実装
4. セルフレビュー（reviewer エージェント + スコープ限定検証）
5. クローズ処理（台帳の該当節を更新・claim ブランチは USER が解放）

## 使い方

```
/implement TASK-027   # todo.md のタスクを実装
/implement BUG-009    # bug.md のバグを実装
/implement            # 両台帳の open ID 一覧を表示して選択

# タスクlookup / ID一覧
grep -n '<タスクID>' todo.md bug.md
grep -oE '^### TASK-[0-9A-Za-z-]+' todo.md
grep -oE '^## BUG-[0-9]+' bug.md
```

## 引数

$ARGUMENTS — タスクID（例: `TASK-027`, `BUG-009`）。省略時は `todo.md` 索引表と `bug.md` 対応状況サマリの open ID 一覧を表示。

> 旧 `3-session-agent.html#ledger`・`BE-XXX` / `FE-XXX`・`docs/tasks/` 体系はすべて廃止済み（経緯は git 履歴参照）。新規実装には使用しない。
