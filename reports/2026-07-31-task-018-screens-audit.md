# TASK-018 — `docs/spec/screens/**` vs product routes audit

> **Date**: 2026-07-31  
> **Residual**: SPEC-TOP-SCREENS-AUDIT  
> **Severity**: Medium / optional full pass  
> **Mode**: Counting methodology + residual list. **Full page-by-page re-spec not completed** in this packet.

---

## 1. Purpose

After screens-spec Mode 3 / Pack A–F, known residuals R1–R3 live as TASK-001–003. This TASK exists so silent drift on **other** screens has a single optional lane — without spawning 70 per-screen TASKs.

---

## 2. Counting methodology

### 2.1 Numbered screen specs (product-facing screen docs)

**Definition**: Markdown files matching `docs/spec/screens/[0-9]*.md` (includes flow doc `99-medical-record-flow.md`).

**Count**: **41** (stated in `docs/spec/screens/README.md`; verified by directory listing: `00`–`32`, `34`–`40`, `99` — **no `33`** file).

These are **logical screen/flow specs**, not 1:1 with React Router leaves.

### 2.2 Product leaf routes

**Definition**: Routes counted by frontend route inventory test.

**Evidence**: `frontend/src/app/routes/route-inventory.test.tsx`  
`expect(pages).toHaveLength(84)` — **84 product pages**, plus separate asserts for redirects/wildcard.

**Also referenced**: e2e static inventory / `docs/spec/ui-design-compliance.md` (runtime history 83→84 leaves).

### 2.3 Total markdown under `docs/spec/screens/**`

**Definition**: All `*.md` under the screens tree (numbered + settings + index/common).

| Bucket | Approx. count | Contents |
|:---|:---|:---|
| Numbered `[0-9]*.md` | 41 | Screen/flow specs |
| `settings/*.md` | 22 | 21 domain settings docs + `settings/README.md` |
| Root helpers | 3 | `README.md`, `CLAUDE.md`, `common-dialogs.md` |
| **Total md** | **66** | Matches inventory intent for this audit |
| `images/*` | (png) | Not counted as specs |

### 2.4 Why 41 ≠ 84

| Factor | Effect |
|:---|:---|
| One spec covers many routes | e.g. master portal `20` + `settings/*`; L-step `31` folds multiple admin surfaces |
| List + form share one narrative | Some domains split list/form specs; others combine |
| LIFF / external mini-apps | Documented under screens and/or `docs/spec/line/**` |
| Settings leaves | Many `/settings/...` product pages map into settings/*.md, not unique numbered files |
| Flow-only docs | `99-medical-record-flow`, `common-dialogs` are not product leaves |

**Rule for future audits**: never “fix” 41 to 84 by inventing empty numbered files; map **product leaves → nearest spec** and record gaps.

---

## 3. What this packet completed

| Activity | Done? |
|:---|:---|
| Recount 41 / 84 / 66 methodology | **Yes** |
| README vs route-inventory length gate awareness | **Yes** (static 84 already gated) |
| Full page-by-page honesty pass of all 41 + 22 settings | **No** |
| Deep re-read of every FE route module | **No** |
| High-risk sample (known open only) | **Partial** — points to existing TASK-001–003 / 006 done |

### High-risk sample (quick)

| Area | Spec | Status |
|:---|:---|:---|
| Hospitalization form atomicity / RO plans / bulk discount | `09-hospitalization-form.md` | Open **TASK-001–003** (not re-litigated here) |
| Identity links | `40-identity-links.md` | **TASK-006 done** (present on main) |
| Owners list API | `03-owners-list.md` | **TASK-007 done** |
| Hospitalization board open gate | `07-hospitalization-list.md` | **TASK-008 done** |
| Estimate lock | `26-estimate-detail.md` | Policy **TASK-012** |
| Closing | `29-closing-aggregation.md` + `cash-register.md` | Policy **TASK-013** |
| Payment methods | `settings/payment-methods.md` | Policy **TASK-014** |

---

## 4. Structural residual: specs lag product leaves

**Finding**: Spec coverage is **coarser** than 84 leaves. That is accepted architecture of the docs, but it means:

1. Drift can hide in settings leaves and multi-route features without a numbered file change.  
2. Mode 3 “COMPLETE with known residual only” is true only if optional full passes (this TASK) stay scheduled.  
3. Inventory 84 is a **code gate**; screens README 41 is a **doc index gate** — both must stay labeled distinctly (already in README).

### Follow-up list (not new product features)

| ID | Work | Priority |
|:---|:---|:---|
| **SPEC-TOP-SCREENS-AUDIT-MAP** | Produce leaf→spec matrix for all 84 paths (spreadsheet or md table); mark SPEC / SHARED / LINE / UNDOC | Med — prerequisite to full pass |
| **TASK-001–003** | Hospitalization honesty/impl (existing) | High/Med |
| **TASK-012–014 / 021** | PO decisions affecting specs | PO |
| **TASK-018-FULL** (optional follow-up name) | After matrix: pack-wise re-read settings/* + high-churn clinical forms | Low/optional |
| Missing number **33** | Confirm intentional gap vs deleted screen | Docs hygiene only |

No evidence in this packet of a **new** silent product BUG beyond already-ledgered residuals.  
**Residual IDs for this TASK**: **SPEC-TOP-SCREENS-AUDIT** remains open until matrix + at least one full pass **or** explicit WONTFILE of full pass by PO.

---

## 5. Acceptance vs todo TASK-018

| Acceptance | This report |
|:---|:---|
| ① Result recorded once | **Yes** — this file |
| ② New opens with IDs or 残差なし | **SPEC-TOP-SCREENS-AUDIT-MAP** follow-up; no orphan product BUG from sample; full pass **not** 残差なし |
| ③ 84 vs 41 counting explicit | **Yes** §2 |

---

## 6. Non-actions

- No product code  
- No mass screens md rewrite  
- No claim that all 84 leaves were manually QA’d  
- No commit  
