# Residual Agent Board — 2026-08-06

**Ops pack:** [`2026-08-06-residual-user-ops-pack.md`](./2026-08-06-residual-user-ops-pack.md)  
**Ledger:** [`STATUS.md`](../STATUS.md)  
**Agent product open:** **0**

> Kanban 風の残件表。実装 unit は gate 解除後のみ。claim 名は live `git branch --list 'claim/*'` で実測すること。

## Board

| id | owner | state | blocker |
|----|-------|-------|---------|
| TASK-378-reset | USER | **local DONE** | 2026-08-06 make reset postflight OK |
| TASK-009 | USER | **local DONE** | exam_reference_ranges=20; seed CSV col fix |
| POST-PULL | USER | READY_USER | 他 env の `make migrate` |
| TASK-032-apply | USER | local DDL 済 / 他 env 残 | claim 0; lab_import tables exist local |
| TASK-374-apply | USER | local DDL 済 / 他 env+#211 残 | checkup import receipts exist local |
| E2E_LOGIN | USER | **BLOCKED** | host/` .env.local` UNSET |
| TASK-020 | USER | BLOCKED | credential + stack up |
| TASK-023 | USER | BLOCKED | credential + 五フロー UAT（#254） |
| TASK-010 | USER | **EXCLUDED** | residual 対象外 (2026-08-06); RUNBOOK 保管のみ |
| BROWSER_BACKLOG | USER | **EXCLUDED** | 同上 · 手作業 IU 32 は residual closeout に含めない |
| TASK-022 | USER | READY_USER | S13 + RLS runtime |
| TASK-024 | USER | READY_USER | #256 visual sign-off |
| TASK-033 | 臨床+USER | **HOLD** | #201 clinical SoT |
| TASK-021 | USER+PO | **HOLD** | destructive PO approval |
| LINE-R05 | USER/PO | **HOLD** | production DROP |
| #89/#97/#98/#99 | USER | **HOLD** | credential / legacy path |
| SCEN-OPS-CLAIM | — | **DONE** | claim/* local+origin = 0 (2026-08-06) |
| TASK-004/005 | USER | OPS | land 都度 |
| R6/R7 | ops | DISCIPLINE | worktree / empty-diff ban |
| agent-code-audit | residual-team | **DONE** | CODE_PRESENT 32 / MISSING 0 |
| agent-env-inventory | residual-team | **DONE** | env report landed |
| agent-ops-pack | residual-team | **DONE** | USER pack landed |
| agent-product | — | **DONE** | open product implementation units = 0 |

## Swimlanes（推奨順）

| Lane | 順序 |
|------|------|
| **DB** | 009 → POST-PULL/032/374 → 378? |
| **Auth E2E / UAT** | E2E_LOGIN → 020 → 023 |
| **Browser** | 010 + BACKLOG |
| **Human docs** | 022 → 024 |
| **Gates** | 033 → 021（並行不可の依存は clinical/PO） |
| **Security ops** | #89/#97/#98/#99（別レーン・常時 HOLD まで） |
| **Claims** | SCEN-OPS-CLAIM（マージ後随時） |

## Legend

| state | 意味 |
|-------|------|
| READY_USER | 人が今すぐ着手可能 |
| CONDITIONAL | 条件付き（checksum 等） |
| HOLD | gate / 専権 / 決裁待ち。agent 実装禁止 |
| DISCIPLINE | 継続ルール |
| DONE | 残実装 unit なし |

---

*No secrets. Agent does not migrate / reset / delete claims / mark VERIFIED_FIXED.*
