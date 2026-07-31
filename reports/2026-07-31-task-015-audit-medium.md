# TASK-015 — MEDIUM audit residual re-correlation with scenarios

> **Date**: 2026-07-31  
> **Residual**: SCEN-AUDIT-MED-001  
> **Severity**: Low / optional gate  
> **Mode**: Document-only re-correlation (no product code, no full re-audit)

---

## 1. Purpose

Wave1 DOC_ALIGNED closed major scenario↔code gaps for S/V packs. TASK-015 is a **short gate** to ensure MEDIUM audit leftovers are either:

- already covered by an open residual/TASK, or  
- explicitly marked **残差なし**, or  
- promoted to a new residual ID / BUG / 要PO.

This is **not** a second full audit of the product.

---

## 2. Methodology

1. **Source A — scenario marks**:  
   `rg '【要実測】' docs/ops/testing/scenarios`  
   Result on 2026-07-31: **78 matching lines** (includes scenario bodies, README policy line, and historical notes under `scenarios/reports/`). Treat **scenario body marks** as the executable backlog for TASK-010; do not treat old report prose as new open items.

2. **Source B — ledger residuals** (`todo.md` SCEN / SPEC / ARCH table):  
   Tracked IDs that absorb product/policy decisions vs pure measurement.

3. **Source C — prior wave claim**:  
   Main S/V packs marked DOC_ALIGNED in SCENARIOS-ALIGN-REMEDIATE-WAVE1; Class E code-determined items elevated in-wave.

4. **Correlation rule**:  
   - If a MEDIUM topic is already a TASK/residual → **link only** (no duplicate TASK).  
   - If measurement-only → **TASK-010**.  
   - If PO/policy → existing 要PO TASK.  
   - If nothing open and no mark → **残差なし**.

---

## 3. Tracked MEDIUM (and related) items — correlation matrix

| Residual / theme | Ledger TASK | Correlation outcome | Notes |
|:---|:---|:---|:---|
| SCEN-BROWSER-001 【要実測】 backlog | **TASK-010** | **Open — measurement** | 78 marks; primary consumer of this gate’s “next steps” |
| SCEN-SEED-001 clinical header-only seed | **TASK-009** | **Open — design** | Design in `reports/2026-07-31-task-009-seed-design.md`; apply USER |
| SCEN-PROD-ESTIMATE-001 estimate unlock | **TASK-012** | **Open — PO** | Packet: `2026-07-31-todo-po-decisions.md` |
| SCEN-PROD-CLOSING-001 closing reverse | **TASK-013** | **Open — PO** | Same PO packet |
| SCEN-PROD-PAYMENT-KEY-001 system_key | **TASK-014** | **Open — PO** | Same PO packet; V04 mark remains until measured/decided |
| R2 treatment plan edit RO | **TASK-002** | **Open — PO** | Hospitalization form |
| R3 bulk discount non-persist | **TASK-003** | **Open — PO** | Hospitalization form |
| SPEC-TOP-CAPABILITIES-CRUD dual surface | **TASK-021** | **Open — PO** | Capabilities vs exclusion |
| SPEC-TOP-SCREENS-AUDIT | **TASK-018** | **Open — optional audit** | This wave’s screens report |
| SPEC-TOP-LINE-AUDIT | **TASK-019** | **Open — optional audit** | LINE report |
| SPEC-TOP-E2E-RUNTIME-84 | **TASK-020** | **Open — runtime blocked** | See runtime-blocked report |
| SCEN-S11-COPY-001 S11 copy | **TASK-011** | **Open — docs Low** | Not MEDIUM product risk |
| SCEN-AUDIT-MED-001 (this gate) | **TASK-015** | **This report** | Close when matrix accepted |
| pre-WAVE1 V05 full form audit | → TASK-010 / V05 marks | **Absorbed** | No separate MEDIUM residual ID; marks live in V05 |
| available-staffs | WONTFILE | **残差なし (product)** | Must not re-open under TASK-021 |
| capabilities SoT doc | already done | **残差なし (doc)** | Write dual surface remains TASK-021 |
| ARCH Mode 3 body | ARCH-DONE | **残差なし (arch body)** | ARCH-R1/R4 are Low docs, not scenario MEDIUM |

### Explicit “残差なし” for MEDIUM product themes outside the matrix

Within the **tracked MEDIUM scenario residuals** listed in `todo.md` for SCEN-*, every MEDIUM product theme maps to TASK-009–014, 018–021, or 010.  
**No orphan MEDIUM residual ID** was found that lacks a TASK or disposition.

Low/docs items (TASK-011, 016, 017) are **out of MEDIUM scope** for this gate.

---

## 4. 【要実測】 mark census (input to TASK-010)

| Bucket | Approx. location | Disposition |
|:---|:---|:---|
| Clinical forms | `V01-clinical-forms.md` | TASK-010 |
| Accounting / reservation forms | `V02-*.md` | TASK-010 |
| Owner/pet/staff forms | `V03-*.md` | TASK-010 |
| Settings / masters | `V04-*.md` (incl. payment system_key) | TASK-010 + link TASK-014 if policy |
| Auth / LINE forms | `V05-*.md` | TASK-010 + TASK-019 for doc drift |
| Journey S01–S12 family | various `S*.md` | TASK-010; seed-dependent ones prefer post TASK-009 |
| Policy line in README | `scenarios/README.md` | Not a case |
| Historical `scenarios/reports/*` | old execution notes | Do not re-count as open marks |

**Total rg hits: 78.** Exact per-file split should be re-run at TASK-010 start (`rg -n '【要実測】' docs/ops/testing/scenarios --glob '!reports/**'` recommended for executable count).

---

## 5. Security / codex MEDIUM note (boundary)

Codex security output under `codex-security-output/` includes treatment-plan discount TOCTOU narratives. Those are **security findings**, not scenario residual IDs.  
If still open in security process, track under security ledger — **do not** silently merge into SCEN-AUDIT-MED-001. Optional cross-link only if a scenario needs a regression check after fix.

---

## 6. Gate result

| Check | Result |
|:---|:---|
| Re-correlation executed once | **Yes** (this report) |
| Orphan MEDIUM without TASK | **残差なし** |
| Measurement backlog | **TASK-010** (78 marks) |
| PO backlog | TASK-002/003/012/013/014/021 |
| Seed blocker for some marks | **TASK-009** |

**TASK-015 disposition recommendation**: mark **done** after PO/ledger acknowledges this report path (no code). Re-open only if a new audit wave produces unmapped MEDIUM themes.

---

## 7. Next steps → TASK-010

1. Prefer seed slice (TASK-009) for scenarios that need clinical search hits.  
2. Executable mark list:  
   `rg -n '【要実測】' docs/ops/testing/scenarios --glob '!**/reports/**'`  
3. browser-test skill granularity: measure → elevate expected result **or** BUG **or** 要PO.  
4. Write evidence to `reports/YYYY-MM-DD-<env>.md` only (not scenario file bodies for run logs).  
5. Runtime env must be verified first — see `reports/2026-07-31-task-010-020-runtime-blocked.md`.

---

## 8. Non-actions

- No full re-audit of all screens  
- No product code  
- No silent removal of 【要実測】 without measurement  
- No auto browser run in this packet  
