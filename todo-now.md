# いま着手可能な作業（PO判断以外）

最終更新: 2026-09-12（`main` = `origin/main` 同期前提）。  
出典: Linear Team Baritech / Project ノア動物病院電子カルテの未完了 Issue + repo `todo-*`。  
**Linear 新規 Issue は作らない**（既存更新のみ）。新規本文は [todo-issue.md](todo-issue.md)。

## 選定方針

| 含める | 含めない |
|--------|----------|
| 実装済みの人間レビュー Done | クライアント／臨床の承認・決裁待ち |
| USER/ops が手順どおり実行できる作業 | 先方 enable・契約記入・実機受領待ち |
| repo 上の検証残り（前提が揃えば） | Needs Human の PO/臨床ゲート本体 |
| | Blocked / Duplicate / 案件ハブ BRT-4 |

---

## A. すぐ着手可（人間レビュー / USER・ops）

| ID | 内容 | 次アクション | 注意 |
|----|------|--------------|------|
| [BRT-226](https://linear.app/baritechllc/issue/BRT-226) | Codex Security 修正（`origin/main` 済み） | 内容確認して **Done**（人間のみ） | agent 単独 Done 禁止。再スキャンは任意・別判断 |
| [BRT-44](https://linear.app/baritechllc/issue/BRT-44) | 本番 CI/CD・backup/restore gate 残件棚卸し | Production Environment の reviewers / bypass / secrets / deploy・監視・restore 実測の棚卸しと手順突合 | 本番変更・secret 投入は USER。agent は代行しない |
| [BRT-67](https://linear.app/baritechllc/issue/BRT-67) | named env 非破壊 migrate | 対象 env・SHA・backup/rollback・承認窓を決め、必要なら承認済み手順で migrate | **agent 禁止**（STG/PROD migrate） |
| [BRT-37](https://linear.app/baritechllc/issue/BRT-37) | OPS-1 credential ローテ統合 | ローテ／revoke 計画の実行（USER） | 秘密を Linear / ログに載せない |

---

## B. 前提が揃えば着手可（repo / 検証）

| ID | 内容 | 前提 | 管理先 |
|----|------|------|--------|
| DEV-V-OWNER-DB | owner disposable DB 検証 | 安全な `TEST_DATABASE_URL`（共有DB不可・チャットに貼らない） | [todo-verification.md](todo-verification.md) |
| TASK-444-ADDENDUM-CODEGEN | カルテ追記 response 型の生成経路 | tygo/codegen 変更の明示承認 + user-run `make codegen` | [todo-issue.md](todo-issue.md) |

---

## C. 除外（PO判断・臨床・先方待ち — 参考）

着手可能リストには入れない。詳細は Linear。

| ID | 理由（要約） |
|----|----------------|
| BRT-39 / 40 / 41 / 51 | 臨床記入・承認・hub 決裁 |
| BRT-43 / 46 / 57 | 締め時間・スタッフ発行・PO-10 等の方針／PO |
| BRT-45 / 47 / 48 / 68 | UAT close / disposition / go-live / 人間レーン |
| BRT-49 / 50 / 52 / 42 | 契約記入・先方 enable・フォントQA受領・producer bundle 待ち |
| BRT-94 | 子実装は Done。残は実機 UAT・配布 gate・人間受入 |
| BRT-106 | 価値実測（業務責任者）が揃うまで実装禁止 |
| BRT-7 / 8 / 66 | Blocked |
| BRT-4 | 案件ハブ（作業単位ではない） |

---

## 参照

| 文書 | 役割 |
|------|------|
| [todo.md](todo.md) | 全体入口・残タスク要約 |
| [todo-issue.md](todo-issue.md) | 新規 Issue 本文 SoT |
| [todo-verification.md](todo-verification.md) | 検証・外部境界 |
| [todo-operations.md](todo-operations.md) | 運用TODO |
| [docs/work/linear-f1-f6-mapping.md](docs/work/linear-f1-f6-mapping.md) | Linear 照合メモ |
