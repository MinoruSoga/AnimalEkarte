---
description: タスク実装（docs/tasks/open/ のタスクID）→ セルフレビュー → クローズ
argument-hint: "FEAT-XXX | PERF-XXX | BUG-XXX | SEED-XXX 等"
---

# タスク実装

指定されたタスクID（`docs/tasks/open/` 配下、前方一致検索）を実装する。

スキル `implement-issue` を使用して、以下の5フェーズを実行:
1. タスク読み込み・依存関係チェック
2. コンテキスト収集（対象ファイル・参照実装・コーディングルール）
3. コード規約準拠の実装
4. セルフレビュー（reviewer エージェント + Lint/Build）
5. クローズ処理（ファイル移動 + 親TASK更新）

## 使い方

```
/implement PERF-FOLLOWUP-01   # 特定タスクを実装（前方一致で docs/tasks/open/ から検索）
/implement FEAT-searchable-select-targets
/implement                    # open タスク一覧を表示して選択
```

## 引数

$ARGUMENTS — タスクID（例: `PERF-FOLLOWUP-01`, `FEAT-searchable-select-targets`, `BUG-XXX`, `SEED-XXX`）。省略時は `docs/tasks/open/` の一覧を表示。

> 旧 `BE-XXX` / `FE-XXX` イシュー体系は `docs/archive/{backend,frontend}-issues/` に移設済み。新規実装には使用しない。
