# PO Decision Packet — TASK-002 / 003 / 012 / 013 / 014 / 021

> **Date**: 2026-07-31  
> **Scope**: Product-Owner decisions only. **No product code changes in this packet.**  
> **Source ledger**: `todo.md` (local TASK IDs; residual IDs as cited)  
> **Authoring rule**: recommend an option; do not implement.

---

## Summary table

| TASK | Residual | Severity | Recommended | One-line rationale |
|:---|:---|:---|:---|:---|
| TASK-002 | R2 | Medium / PO | **B** (or **C** if a dedicated write owner is preferred) | BE PATCH/DELETE exist; FE edit is intentionally RO. Unlock only if PO names a clinical purpose and owner. |
| TASK-003 | R3 | Medium / PO | **B** | Bulk discount is display-only; row-level discounts already persist. Avoid dual accounting SoT. |
| TASK-012 | SCEN-PROD-ESTIMATE-001 | High / PO | **B** | 26§2.1 already states no unlock; S07 honesty should match. Unlock would reopen locked financial/clinical estimates. |
| TASK-013 | SCEN-PROD-CLOSING-001 | High / PO | **B** | Closing is append-only; reverse API does not exist. Soften only with explicit boundary table + audit. |
| TASK-014 | SCEN-PROD-PAYMENT-KEY-001 | Medium / PO | **A** | Document immutable reserved `system_key`s; keep UI non-exposure (current FE has no `system_key` field). |
| TASK-021 | SPEC-TOP-CAPABILITIES-CRUD | Medium / PO | **B → staged A** | Keep exclusion as facade first; staged convergence to capabilities SoT. No destructive API delete without PO. |

---

## TASK-002 — Edit-mode treatment plan unlock

### 論点

入院フォーム **編集** で治療プラン（明細）を変更できるべきか。現状は create 時のみ POST 可能で、edit は親フィールドのみ更新・プランは参照専用。

### Current code paths

| Layer | Path | Behavior |
|:---|:---|:---|
| FE form | `frontend/src/features/hospitalization/routes/HospitalizationForm.tsx` | `readOnly={isEdit}` on treatment table |
| FE save | `frontend/src/features/hospitalization/hooks/use-hospitalization-form.ts` | Edit saves parent only |
| FE write API | `frontend/src/features/hospitalization/api/treatment-plans-write.ts` | **POST only** (`createTreatmentPlanForHospitalization`) |
| BE routes | `backend/internal/medicalrecord/routes.go` | Hospitalization nested: GET/POST/**PATCH/DELETE** `/v1/hospitalizations/:id/treatment-plans[/:planId]` |
| Spec honesty | `docs/spec/screens/09-hospitalization-form.md` §2–3 | Edit = RO plans; bulk discount display-only |

### Options

| Option | Description |
|:---|:---|
| **A** | Unlock edit form: wire FE PATCH/DELETE (+ optional add POST) with GET re-sync; permission-gate `hospitalization:edit` / `delete`. |
| **B** | Keep permanent RO on form edit; document as intentional; detail/care screens remain the operational path for ongoing care. |
| **C** | Designate **another screen** (e.g. hospitalization detail / care plan) as the sole treatment-plan write owner; form stays RO forever. |

### Clinical / accounting risks

- **A**: Concurrent edits vs discharge/billing; discount race already noted in security scan on treatment-plan update path (`treatment_plan_handler.go` / `discount:edit`). Partial plan mutation without atomic parent+plans TX if combined with TASK-001 gap.
- **B/C**: Staff cannot change create-time plans from the edit form — may force delete+recreate or workarounds if no alternate write UI is clear.
- Double-entry risk if plans feed billing and are edited after partial billing.

### Recommended

**B** (honesty / WONTFIX unlock on this form) **unless PO requires unlock**, in which case prefer **C** (single write owner outside the registration form) over scattering PATCH into the create/edit form. Do not invent unlock without a named business owner and purpose.

---

## TASK-003 — Bulk discount display-only

### 論点

入院フォームの **一括割引（% / 円）** を永続化するか。現状は概算表示のみで保存されない（honesty 文言あり）。行単位の `discount_rate` / `discount_amount` は treatment plan に存在する。

### Current code paths

| Layer | Path | Behavior |
|:---|:---|:---|
| FE UI | `frontend/src/features/hospitalization/components/HospitalizationCostSummary.tsx` | Inputs `disabled` when `readOnly` (form always passes `readOnly`) |
| FE form | `HospitalizationForm.tsx` | CostSummary always `readOnly` |
| FE types / model | hospitalization form model | No parent-level bulk discount field |
| BE plan rows | `treatment_plans` + nested create request | Per-row discount fields only |
| Spec | `docs/spec/screens/09-hospitalization-form.md` §3 | Explicit non-persistence |

### Options

| Option | Description |
|:---|:---|
| **A** | Persist bulk discount: schema + API + FE + billing alignment. |
| **B** | Accept display-only; close as accepted decision; keep honesty copy. |

### Clinical / accounting risks

- **A**: Second discount SoT vs row discounts and vs billing `discount_amount` → product-philosophy **二重管理禁止**. Allocation ambiguity (% of which subtotal, tax base, insurance).
- **B**: Operators may still misread disabled controls as “saved” if honesty regresses; residual UX risk only.

### Recommended

**B — keep display-only.** Row discounts already exist on plans; bulk is a calculator UX, not a financial fact. Escalate only if PO defines a single write owner and accounting contract.

---

## TASK-012 — Estimate unlock vs 26§2.1

### 論点

承認済み / 却下の見積を「下書きへ戻す」unlock を製品として提供するか。仕様 26§2.1 は **下書きへ戻す手段なし**。S07 は運用メモで unlock 欠如を「要仕様決定」と書く。

### Current code paths

| Layer | Path | Behavior |
|:---|:---|:---|
| Spec | `docs/spec/screens/26-estimate-detail.md` §2.1 | Locked statuses: no edit/delete UI; BE atomic reject; **no unlock** |
| FE lock helper | `frontend/src/features/estimates/lib/is-estimate-locked-status.ts` (+ list/form/detail) | Hide edit/delete; edit URL redirects |
| BE | `backend/internal/billing/estimate_service.go` / `estimate_repository.go` | `isEstimateLocked` + status NOT IN predicate on update/delete |
| BE request | `backend/internal/billing/estimate_request.go` | Create status `draft|sent` only; update may set `approved|rejected` |
| Scenario | `docs/ops/testing/scenarios/S07-estimate-status-control.md` | Proves lock; notes unlock **not implemented** |

### Options

| Option | Description |
|:---|:---|
| **A** | Implement unlock (approved/rejected → draft) with permission + audit + optional reason. |
| **B** | No unlock; treat 26§2.1 as SoT; unify S07 wording with 26 (remove “要仕様決定” ambiguity). |

### Clinical / accounting risks

- **A**: Re-opening a client-facing approved estimate after clinical communication; audit/trail integrity; race with estimate→billing conversion if any.
- **B**: Correction path = new estimate only (operational cost, but fail-closed).

### Recommended

**B — no unlock.** Align S07 with 26§2.1 as the single interpretation. If a clinic later needs correction workflow, require PO-named purpose, permission, and audit before any code.

---

## TASK-013 — Closing reverse

### 論点

レジ締め（AM/PM/EMG）の **reverse / 取消** を許すか。境界（誰が・いつ・越日 EMG・二重締め UNIQUE）が未決。

### Current code paths

| Layer | Path | Behavior |
|:---|:---|:---|
| Spec | `docs/spec/cash-register.md` §3, §5 | Append-only; **no update/delete API** |
| Spec UI | `docs/spec/screens/29-closing-aggregation.md` | Close history / preview |
| BE routes | `backend/internal/billing/routes.go` | `POST /v1/cash-register/closes`, GET list/detail/preview — **no DELETE/PATCH reverse** |
| BE service | `backend/internal/billing/cash_register_service.go` | Period resolution; close create |
| Scenario residual | SCEN-PROD-CLOSING-001 / S09 family | Boundary + reverse policy open |
| Related | post-close billing edit via `accounting-post-close-edit` | Edits billings under audit; does **not** reverse the close row |

### Options

| Option | Description |
|:---|:---|
| **A** | Design reverse (soft reopen / compensating close) with full boundary matrix. |
| **B** | Keep append-only; document that reverse is out of product; corrections via post-close edit + next-period ops only. |

### Clinical / accounting risks

- **A**: Cash handoff integrity; double-count if reverse incomplete; EMG overnight attribution; UNIQUE `(clinic_id, close_date, period)` interactions.
- **B**: Wrong close requires operational workaround (memo, next day adjustment) — painful but safer default.

### Recommended

**B — append-only; reverse not implemented.** Any reverse is a new High design after PO boundary table; do not soft-delete close rows ad hoc.

---

## TASK-014 — payment `system_key` policy

### 論点

`payment_methods.system_key`（`cash` / `credit_card` / `electronic_money` / `bank_transfer`）の不変性、UI 露出、名称変更・無効化可否。

### Current code paths

| Layer | Path | Behavior |
|:---|:---|:---|
| ADR | `docs/architecture/adr/003-payment-method-identity-and-consistency.md` | system_key adopted; TRIGGER match method↔key |
| DB / seed | `payment_methods` + `003_demo/payment_methods.csv` | Reserved keys seeded per clinic |
| BE | `backend/internal/billing/accounting_service_*.go` | Resolve by system_key, not display name |
| FE | payment-method master UI | **No `system_key` reference** in frontend features (grep empty) |
| Scenario | `docs/ops/testing/scenarios/V04-settings-master-forms.md` | 【要実測】name change / deactivate of system rows |

### Options

| Option | Description |
|:---|:---|
| **A** | Document immutable reserved keys; UI never exposes/edits `system_key`; name may be localizable display-only if BE allows; deactivation policy explicit. |
| **B** | Expose key in admin UI for power users (risk of break). |
| **C** | Allow remapping keys (migration / rewrite) — high blast radius. |

### Clinical / accounting risks

- Changing keys breaks TRIGGER, splits, cash-register category mapping, import bindings (`CLINIC_CSV_IMPORT.md` cash/credit_card bindings).
- FE already does not surface keys — accidental operator remap is currently hard; keep it that way.

### Recommended

**A — document immutable reserved keys; UI non-exposure.** Optional follow-up: lock name/active on system rows if V04 real measurement shows dangerous edits.

---

## TASK-021 — Staff reservation capabilities dual surface

### 論点

Write SoT は `staff_reservation_capabilities` と doc 固定済み。一方 exclusion API/UI 名・候補フィルタ（`excluded_courses`）が残る dual surface。全面 CRUD 削除か facade 維持か。

### Current code paths

| Layer | Path | Behavior |
|:---|:---|:---|
| Doc SoT | `docs/spec/reservation-to-record-flow.md` | Capabilities SoT + dual residual narrative |
| BE | `backend/internal/staff/staff_handler.go` | GET/PUT `.../excluded-reservation-types` **and** `.../capable-reservation-types` |
| FE write | `frontend/src/features/master/api/staff-reservation-types.ts` | Capable API for write |
| FE naming | `frontend/src/features/master/components/StaffSidePanelSections.tsx` | `StaffExcludedReservationTypesSection` name residual |
| FE filter | `frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx` + `use-reservation-types.ts` | Candidate filter uses `excluded_courses` |
| Out of scope | available-staffs endpoint | **WONTFILE / do not implement** (SPEC-TOP-AVAILABLE-STAFFS) |

### Options

| Option | Description |
|:---|:---|
| **A** | Hard converge: remove exclusion write (and eventually read) APIs; rename FE; filter only from capabilities. |
| **B** | Facade: exclusion endpoints remain as derived/compat view; write path only capabilities; staged rename/filter migration. |
| **C** | Keep dual indefinitely (accepted dual) — conflicts with dual-management ban long-term. |

### Clinical / accounting risks

- **A immediate**: Breaking external/API consumers and staff master UX mid-clinic-day → wrong staff offered for course → overbook / wrong skill.
- **B**: Temporary dual read models if not derived consistently.
- available-staffs must **not** be introduced under this TASK.

### Recommended

**B first, then staged A.**  
1) Freeze write SoT on capabilities.  
2) Treat exclusion as compatibility facade (read derived or thin wrapper).  
3) Stage FE rename + candidate filter off capabilities.  
4) **No destructive API delete without PO sign-off** and consumer inventory.

---

## Decision log template (for PO)

For each TASK, PO records:

1. Chosen option (A/B/C)  
2. Business owner (person name) + purpose  
3. Follow-up implementation TASK IDs (or WONTFIX close)  
4. Date / signature  

Until that log exists, agents must **not** implement unlock/reverse/key remap/API deletion for these items.

---

## Explicit non-actions this packet

- No product code edits  
- No `todo.md` mutation  
- No touch of foreign WIP (`line_reservation_settings.csv` etc.)  
- No migration / seed apply  
