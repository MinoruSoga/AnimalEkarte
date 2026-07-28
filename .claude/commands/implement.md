---
description: "タスク実装（3-session-agent.html #ledger のタスクID）→ セルフレビュー → クローズ"
argument-hint: "TASK-XXX | FEAT-XXX | PERF-XXX | BUG-XXX | SEC-XXX"
---

# タスク実装

指定されたタスクID（repo 直下 `3-session-agent.html` の `#ledger` 節に `<section class="task" id="<タスクID>">` 形式で記載）を実装する。

スキル `implement-issue` を使用して、以下の5フェーズを実行:
1. タスク読み込み・依存関係チェック
2. コンテキスト収集（対象ファイル・参照実装・コーディングルール）
3. コード規約準拠の実装
4. セルフレビュー（reviewer エージェント + Lint/Build）
5. クローズ処理（`3-session-agent.html` の `#ledger` 節から該当 `<section>` を削除）

## 使い方

```
/implement PERF-FOLLOWUP-01   # 特定タスクを実装（`#ledger` 内の `id="<タスクID>"` を grep で検索）
/implement FEAT-searchable-select-targets
/implement                    # `#ledger` のタスクID一覧を表示して選択

# タスクlookup / ID一覧
grep -n 'id="<タスクID>"' 3-session-agent.html
grep -oE 'id="(TASK|BUG|FEAT|PERF|SEC)[A-Za-z0-9-]*"' 3-session-agent.html
```

## 引数

$ARGUMENTS — タスクID（例: `TASK-XXX`, `PERF-FOLLOWUP-01`, `FEAT-searchable-select-targets`, `BUG-XXX`, `SEC-XXX`）。省略時は `3-session-agent.html` の `#ledger` 節にあるタスクID一覧を表示。

> 旧 `BE-XXX` / `FE-XXX` イシュー体系および旧 `docs/tasks/` 体系は廃止済み（経緯は git 履歴参照）。新規実装には使用しない。
