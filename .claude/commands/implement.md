---
description: タスク実装（todo.md「個別タスク詳細」節のタスクID）→ セルフレビュー → クローズ
argument-hint: "FEAT-XXX | PERF-XXX | BUG-XXX | SEED-XXX 等"
---

# タスク実装

指定されたタスクID（repo 直下 `todo.md` の「個別タスク詳細」節に `### <タスクID>: <タイトル>` 形式で記載）を実装する。

スキル `implement-issue` を使用して、以下の5フェーズを実行:
1. タスク読み込み・依存関係チェック
2. コンテキスト収集（対象ファイル・参照実装・コーディングルール）
3. コード規約準拠の実装
4. セルフレビュー（reviewer エージェント + Lint/Build）
5. クローズ処理（todo.md から該当セクション削除 + 索引行更新）

## 使い方

```
/implement PERF-FOLLOWUP-01   # 特定タスクを実装（todo.md 内の `### <タスクID>` 見出しを grep で検索）
/implement FEAT-searchable-select-targets
/implement                    # todo.md の個別タスク詳細節の見出し一覧を表示して選択
```

## 引数

$ARGUMENTS — タスクID（例: `PERF-FOLLOWUP-01`, `FEAT-searchable-select-targets`, `BUG-XXX`, `SEED-XXX`）。省略時は `todo.md` の個別タスク詳細節の見出し一覧を表示。

> 旧 `BE-XXX` / `FE-XXX` イシュー体系および旧 `docs/tasks/` 体系は廃止済み（経緯は git 履歴参照）。新規実装には使用しない。
