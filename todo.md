# Remaining work ledger (open residual only)

最終整理: **2026-08-06**（main / agent ledger reconciliation）。  
**製品コードで AI が着手可能な open TASK は 0 件**。残は USER / ops / 臨床ゲート / ブラウザ実測。

> **ID namespace**: `TASK-*` = 本ファイル。受入バグは `bug.md` の `BUG-*`（**32/32 IMPLEMENTED_UNVERIFIED**。ブラウザは `reports/BROWSER_VERIFICATION_BACKLOG.md`）。

## 索引 / サマリー（open residual）

| Inv / TASK | 内容 | 処置 | Owner |
|------------|------|------|-------|
| R4 / TASK-004 | screens-drift 意図変更のコミット隔離 | land 都度 ops 手順 | USER |
| R5 / TASK-005 | closed pack 回帰ゲート | land 都度 ops 手順 | USER |
| R6 / SCEN-OPS-TREE-001 | 共有 tree thrash | worktree 隔離 | ops |
| R7 / ARCH-R2 | empty-diff COMPLETE 禁止 | harness 規律 | ops |
| SCEN-SEED-001 / TASK-009 | 003_demo clinical seed 適用 | static GREEN 済・**DB 適用のみ** | USER |
| SCEN-BROWSER-001 / TASK-010 | scenarios【要実測】 | batch5 残・browser backlog と統合可 | USER |
| SCEN-OPS-CLAIM-001 | claim 解放 | 統合後 | USER |
| SCEN-OPS-COMMIT-001 | mixed commit 説明 | rewrite 禁止 | ops |
| POST-PULL / MIGRATE-006 | migrations 適用 | `make migrate` | USER |
| TASK-020 | Playwright E2E runtime | E2E_LOGIN_* 注入後 | USER |
| TASK-021 Stage A | exclusion 破壊削除 | external use UNREPORTED・破壊承認後 | USER+PO |
| TASK-022 human | #239 S13 手動 correction + RLS proof | human residual | USER |
| TASK-023 human | #254 5 フロー UAT | E2E_LOGIN + 人証跡 | USER |
| TASK-024 human | #256 screenshot/FAQ sign-off | visual sign-off | USER |
| TASK-032 residual | lab import migration 適用・claim 解放 | code DONE | USER |
| TASK-033 | #201 救急投薬 fail-closed cutover | **BLOCKED 臨床入力 + DB review** | 臨床+USER |
| TASK-374 residual | checkup package import migration 適用 | code DONE (synthetic) | USER |
| TASK-378 residual | DB_RESET / seed 列遅れ follow-up | 001 統合 DONE | USER |
| LINE R-05 residual | production rollout + column DROP | HOLD | USER/PO |

### 対応済み（製品コード / docs — 再掲しない・git 正本）

| TASK | 要約 | 証拠（main 到達） |
|------|------|-------------------|
| TASK-025 | dose silent fallback 停止 | `eaa608b6a` + `db8387035` |
| TASK-026 | confirmed exam lock/audit | `2a8aca33c` |
| TASK-027 | manual exam lifecycle / revision | `046615f4b`〜`dfd653eaa` 系 |
| TASK-028 | closing settings PATCH | `bbf82e2b8` |
| TASK-029 | L-step disabled 文書 | `9fc5b9ffb` |
| TASK-030 | trimming deceased 回帰 | `6e5a945ef` |
| TASK-031 | print-snapshot | 2026-08-04 land |
| TASK-032 | lab import compensating revert（code） | 2026-08-04 land・**migrate 未適用** |
| TASK-374 | checkup package import（synthetic） | 2026-08-04・**migrate 未適用** |
| TASK-375 | go-live runbook replan | docs 2026-08-04 |
| TASK-376 | delivery U1–U12 / U13 境界 | docs 2026-08-04 |
| TASK-377 | dose deviation reason | 2026-08-04 |
| TASK-378 | migration 002–008 → 001 統合 | 2026-08-04・**DB_RESET は USER** |
| TASK-019 | line deep audit + PO FINAL | reports 2026-07-31 |

### Ops 規律（継続）

- 並行編集は worktree 隔離（R6）。
- COMPLETE は実 diff 必須（R7）。
- land 前 `git status --porcelain` で foreign 定義（ARCH-R3 / TASK-004）。
- agent は `make migrate` / seed apply / force-push / claim 解放をしない。

### 推奨 USER 順

1. **TASK-009** seed 適用（+ 必要なら BUG-003 `exam_reference_ranges` 含む 003_demo 再投入）
2. 未適用 migration 群（TASK-032 / 374 / POST-PULL）を環境ごとに `make migrate`（破壊的 001 再構築は TASK-378 handoff）
3. **E2E_LOGIN_*** 設定 → TASK-020 / 023
4. **TASK-010 / browser backlog**（`reports/BROWSER_VERIFICATION_BACKLOG.md` + bug.md IU 再検証）
5. **TASK-022/024** human sign-off
6. **TASK-033** 臨床 SoT 揃い後のみ READY_AGENT
7. **TASK-021** external use 調査 + 破壊承認後のみ

### prompt-craft-graph / マルチエージェント方針（2026-08-06）

- 本整理時点で **v2 feature graph の対象となる製品コード unit は 0**。
- 次に agent 実装を再開する最初の候補は **TASK-033**（臨床 gate 解除後）または **TASK-021**（PO 破壊承認後）。
- それ以外の residual は USER 実行。グラフ生成は gate 解除後に 1 unit = 1 graph で行う。

---

## 個別 residual 詳細（open only）

### TASK-004: screens-drift 意図変更セットのコミット隔離（ops）

- **状態**: ops 手順 open（次 intentional land 時）
- **Owner**: USER
- **参照**: `reports/2026-07-31-task-004-005-land-proc.md`
- **Exit**: land セットが intentional / foreign 分離され、screens-drift が混在しない

### TASK-005: closed packs 回帰ゲート（ops）

- **状態**: ops 手順 open（land 都度）
- **Owner**: USER
- **Exit**: land 前に closed pack 回帰を再実行し記録

### TASK-009: 003_demo clinical seed 適用

- **状態**: CSV/static GREEN・**DB 適用のみ USER**
- **Owner**: USER
- **参照**: `reports/2026-07-31-task-009-reseed-ops.md`、`scripts/verify_seed.py`
- **Non-actions**: agent は seed apply / DB_RESET しない
- **Note**: BUG-003 用 `exam_reference_ranges.csv` も 003_demo に追加済み（2026-08-06）— 適用はこの TASK と同時が望ましい

### TASK-010: scenarios【要実測】backlog

- **状態**: env READY / batch5 partial / 残 DEFER・BLOCKED
- **Owner**: USER（ブラウザ）
- **参照**: `reports/2026-08-01-task-010-batch5.md`、`reports/BROWSER_VERIFICATION_BACKLOG.md`
- **Exit**: census 0 または全件 PASS/明示 BLOCKED 記録

### TASK-020: Playwright runtime

- **状態**: env-forward done / **credentials BLOCKED**（E2E_LOGIN_* UNSET）
- **Owner**: USER
- **参照**: `reports/2026-07-31-task-020-env-forward.md`

### TASK-021 Stage A: exclusion 破壊削除

- **状態**: Phase2 inventory COMPLETE・in-repo FE ZERO・**external use UNREPORTED / 破壊承認未**
- **Owner**: USER+PO
- **Ready**: BLOCKED until 承認
- **参照**: `reports/2026-07-31-task-021-phase2-slice2.md`

### TASK-022 / 023 / 024: human residual

- **TASK-022**: S13 手動 correction + named signer + RLS runtime proof
- **TASK-023**: 5 業務フロー UAT（auth 後）
- **TASK-024**: screenshot/FAQ visual sign-off
- **Owner**: USER

### TASK-032 / 374 residual: migration apply + claim 解放

- code DONE。**migrate と claim 解放は USER**

### TASK-033: 救急投薬 fail-closed cutover

- **状態**: BLOCKED_CLINICAL_INPUT_AND_DECISION_SOT_RECONCILIATION_AND_DATABASE_REVIEW
- **Owner**: 臨床責任者 + USER（decision SoT 補正後のみ agent）
- **Non-actions**: gate 前の実装開始禁止

### TASK-378 residual: DB_RESET

- 001 統合 DONE。checksum 変更のため **環境ごとの volume 再構築は USER**（`docs/ops/deploy/LOCAL_DB_RESET.md`）

---

## 完了定義（この整理の DoD）

- [x] open residual と DONE を分離
- [x] agent-implementable open = 0 を明記
- [x] USER 推奨順を 1 ページに集約
- [ ] USER residual の消化（別セッション）
