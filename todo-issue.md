# 新規 Issue 台帳（repo 正本）

作成日: 2026-09-11。最終更新: 2026-09-11 / `origin/main` = `aeecf8f85`。  
理由: Linear workspace が **free issue limit** のため、**新規チケット作成は行わない**。  
エージェント／人間とも、これから発生する実装単位は本ファイルへ記載する。

| 項目 | 値 |
|------|-----|
| **新規タスク SoT** | 本ファイル `todo-issue.md` |
| **既存チケット更新** | Linear のみ（コメント・状態変更）。新規 Issue 作成禁止 |
| **案件ハブ（既存）** | [BRT-4](https://linear.app/baritechllc/issue/BRT-4) |
| **入口との関係** | [todo.md](todo.md) は開発キュー・FAIL・人間レーンの入口。新規 Issue 本文は本ファイル |
| **エージェント READY** | なし |
| **Open** | `TASK-444-ADDENDUM-CODEGEN`（deferred）のみ |

行値・秘密・PHI は書かない。完了した項目は本ファイルから除き、Git 履歴を参照する。

## Linear 運用（必須）

- **してよいこと**: 既存 Issue（例: BRT-4、BRT-226、BRT-45…）へのコメント・状態更新（明示承認がある場合の Done 含む）
- **してはいけないこと**: Linear での Issue 新規作成（プラン上限）
- 専用 Linear ID が無い完了作業は、必要なら既存ハブ（原則 BRT-4）へ証跡コメントし、詳細は commit を正とする

---

## Open

| ID | 状態 | 内容 |
|----|------|------|
| TASK-444-ADDENDUM-CODEGEN | deferred | カルテ追記 response 型の生成経路変更（tygo 対象外）。user-run `make codegen` を含む別スコープが起票・承認されたときのみ着手 |

（現時点のその他 Open なし。追加時は上表へ行を足し、下に詳細節を書く。）

---

## 新規 Issue の書き方

索引表へ1行追加し、必要なら詳細節を続ける（目安15行）。

```markdown
### <ID> — タイトル

- **問題**: …
- **根拠**: ファイル・現状（推測禁止）
- **修正方針**: …
- **受け入れ条件**: 検証可能な条件＋ scoped 検証コマンド
- **状態**: open | ready | blocked | deferred | done
- **Linear**: 既存チケット更新のみ（ID があれば記載。無ければ「BRT-4 コメント予定」等）
```

ID は既存 claim / 台帳と衝突しないこと（例: `TASK-…` / `BE-RC-…` / `BUG-…`）。

## 参照

| 文書 | 役割 |
|------|------|
| [todo.md](todo.md) | 開発キュー入口・FAIL・人間レーン |
| [docs/work/development-task-decisions.md](docs/work/development-task-decisions.md) | READY/却下の裁定 |
| [docs/work/linear-f1-f6-mapping.md](docs/work/linear-f1-f6-mapping.md) | Linear 既存照合 |
| [todo-verification.md](todo-verification.md) | 検証TODO |
| [todo-operations.md](todo-operations.md) | 運用TODO |
