# Remaining work ledger (open residual only)

**状況ハブ:** [`STATUS.md`](STATUS.md)（全体の入口）

最終整理: **2026-08-06**。  
製品コードの agent 実装 open は **0**。本ファイルは **未完了 residual のみ**（完了 TASK の一覧・証拠表は置かない。git 履歴が正本）。

> **ID**: `TASK-*` = 本ファイル / `BUG-*` = `bug.md`（実装は 32/32 IU。ブラウザは `reports/BROWSER_VERIFICATION_BACKLOG.md`）

## 索引

| ID | 内容 | Owner |
|----|------|-------|
| TASK-004 | land 時 screens-drift 隔離 | USER |
| TASK-005 | land 前 closed-pack 回帰 | USER |
| TASK-009 | 003_demo seed の **DB 適用**（static は済） | USER |
| TASK-010 | scenarios 要実測の残 | USER |
| TASK-020 | Playwright runtime（要 E2E_LOGIN_*） | USER |
| TASK-021 | exclusion 破壊削除（PO 承認後） | USER+PO |
| TASK-022 | #239 S13 手動 correction / RLS 証跡 | USER |
| TASK-023 | #254 5 フロー UAT | USER |
| TASK-024 | #256 screenshot / FAQ sign-off | USER |
| TASK-032-apply | lab import migration 適用 + claim 解放 | USER |
| TASK-033 | #201 救急投薬 cutover（**臨床 gate 未**） | 臨床+USER |
| TASK-374-apply | checkup package import migration 適用 | USER |
| TASK-378-reset | 001 統合後の環境 DB_RESET | USER |
| POST-PULL | 各環境 `make migrate` | USER |
| SCEN-OPS-CLAIM | claim ブランチ解放 | USER |
| LINE-R05 | production rollout + column DROP（HOLD） | USER/PO |
| R6/R7 | worktree 隔離 / empty-diff COMPLETE 禁止 | ops（継続規律） |

## 推奨 USER 順

1. **TASK-009** — 003_demo 適用（`exam_reference_ranges` 含む）
2. **POST-PULL / TASK-032-apply / TASK-374-apply** — migrate
3. **TASK-378-reset** — 必要な環境だけ volume 再構築
4. **E2E_LOGIN_*** → TASK-020 / TASK-023
5. **TASK-010** + `reports/BROWSER_VERIFICATION_BACKLOG.md`
6. **TASK-022 / TASK-024** 人証跡
7. **TASK-033** 臨床 SoT 揃い後のみ agent 再開可
8. **TASK-021** 破壊承認後のみ

## 詳細（open only）

### TASK-004 / TASK-005（ops・land 都度）

- land セットの intentional/foreign 分離と closed-pack 回帰。
- 参照: `reports/2026-07-31-task-004-005-land-proc.md`

### TASK-009 — seed DB 適用

- static verifier GREEN。残は USER による DB 適用のみ。
- agent は seed apply / DB_RESET しない。
- 参照: `reports/2026-07-31-task-009-reseed-ops.md`

### TASK-010 — 要実測

- Owner: USER（ブラウザ）。
- 参照: `reports/2026-08-01-task-010-batch5.md`、`reports/BROWSER_VERIFICATION_BACKLOG.md`

### TASK-020 — Playwright

- 要 host `E2E_LOGIN_EMAIL` / `E2E_LOGIN_PASSWORD`。
- 参照: `reports/2026-07-31-task-020-env-forward.md`

### TASK-021 — exclusion 破壊削除

- 破壊承認と external use 確認が揃うまで着手しない。
- 参照: `reports/2026-07-31-task-021-phase2-slice2.md`

### TASK-022 / 023 / 024 — human residual

- 022: S13 手動 correction + signer + RLS runtime
- 023: 5 業務フロー UAT（認証後）
- 024: screenshot / FAQ visual sign-off

### TASK-032-apply / TASK-374-apply

- migration 適用と claim 解放のみ残（製品コードは main 済み）。

### TASK-033 — 救急投薬 cutover

- **BLOCKED**: 臨床入力 + decision SoT + DB review まで実装開始禁止。

### TASK-378-reset

- 001 統合後 checksum 変更。必要環境のみ `docs/ops/deploy/LOCAL_DB_RESET.md` に従う。

### LINE-R05 residual

- production rollout + column DROP は HOLD。

## Agent 規律

- migrate / seed apply / force-push / claim 解放 / VERIFIED_FIXED 付与はしない。
- 次の実装 unit は TASK-033 または TASK-021 の gate 解除後（1 unit = 1 graph）。
