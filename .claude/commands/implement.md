---
description: "Plane Issue実装 → セルフレビュー → Plane証跡更新"
argument-hint: "EMR-XXX"
---

# タスク実装

指定されたタスクIDを実装する。実行状態の正本は Plane（workspace `baritechllc` / project `EMR`）。repo 入口は `todo.md`。確認済み製品 FAIL 台帳は `todo.md#product-bugs` と Plane `bug` ラベル（`bug.md` は空ファイル）。

- 実行タスク、仕様、受け入れ条件、状態、依存関係のSoTはPlane Issue。
- `docs/work/phase2-deferred.md`は今期外・見送りの索引であり実装台帳ではない。
- `STATUS.md` と旧二台帳は復活させない。

タスクが GitHub Issue（`#NNN`）を参照する場合、該当 Issue 本文とコメントは出典・補助仕様として扱う。正本は Plane Issue の本文・コメントであり、矛盾がある場合は Plane を優先し、相違点を `needs-human` で人間に確認する。`3-session-agent.html` は Issue 分類ビューであり台帳ではない（旧 `#ledger` は 2026-07-31 廃止・経緯は git 履歴）。

手順の正本はスキル `implement-issue`。このコマンドは起動入口だけを持つ。以下の5フェーズを実行:
1. タスク読み込み・claim 確認（AGENTS.md packet claim protocol）・依存関係チェック
2. コンテキスト収集（対象ファイル・参照実装・コーディングルール）
3. コード規約準拠の実装
4. セルフレビュー（reviewer エージェント + スコープ限定検証）
5. クローズ処理（Plane 証跡更新・claim 解放は AGENTS.md の作成者別規則に従う）

## 使い方

```
/implement EMR-123   # Plane Issueを実装
/implement            # Plane の対象 Issue を確認してから選ぶ

# タスクlookup / ID一覧
Plane で Issue 本文・コメント・依存関係を確認する
```

## 引数

$ARGUMENTS — Plane Issue ID（例: `EMR-123`）。省略時は実装対象の Plane ID をユーザーに求める。

> 旧 `3-session-agent.html#ledger`・`BE-XXX` / `FE-XXX`・`docs/tasks/` 体系はすべて廃止済み（経緯は git 履歴参照）。新規実装には使用しない。
