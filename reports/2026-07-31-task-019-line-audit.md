# TASK-019 — `docs/spec/line/**` vs BE/FE LINE surfaces (high-level)

> **Date**: 2026-07-31  
> **Residual**: SPEC-TOP-LINE-AUDIT  
> **Severity**: Medium / optional full pass  
> **Mode**: Inventory + high-level surface map. **Not** a full field-by-field re-spec.

---

## 1. Purpose

Ensure LINE / LIFF / L-step documentation has a single optional audit lane. Secrets, production webhooks, and live Messaging API calls are **out of scope**.

---

## 2. Documentation inventory — `docs/spec/line/**`

| File | Role |
|:---|:---|
| `docs/spec/line/README.md` | Hub index; admin URL quick links |
| `docs/spec/line/architecture.md` | Auth, link tokens, webhook/system shape |
| `docs/spec/line/reservation-spec.md` | Owner-facing reservation app requirements |
| `docs/spec/line/setup.md` | Console / Messaging / L-step setup |
| `docs/spec/line/lstep-integration.md` | CPM, triggers, CRM automation |
| `docs/spec/line/cost-analysis.md` | Cost notes |
| `docs/spec/line/CLAUDE.md` | Agent-local rules for this tree |

**Related (outside line/ but LINE UX):**

| File | Role |
|:---|:---|
| `docs/spec/screens/28-line-reservation.md` | Clinic LINE reservation settings UI |
| `docs/spec/screens/37-line-reserve-owner-flow.md` | Owner 13-step reserve flow |
| `docs/spec/screens/38-liff-pet-health.md` | LIFF pet health / link |
| `docs/spec/screens/31-lstep-integration.md` | L-step admin surfaces |
| `docs/spec/screens/34-lstep-delivery-monitor.md` | Delivery monitor |
| Root `01_曽我さん向け_カルテLステップ連携実装仕様書.md` | Client-received original (hub links here) |

**Count**: 7 markdown files under `docs/spec/line/` (including CLAUDE.md).

---

## 3. Backend surfaces (high-level)

Primary package: `backend/internal/lstep/`.

| Surface | Evidence path | Notes |
|:---|:---|:---|
| LINE webhook | `routes.go` → `POST /api/line/webhook` (`ReceiveLineWebhook`) | Public entry; rate-limit tests exist |
| Owner LINE send | `POST /api/v1/owners/:id/line/send`, send-logs; L-step aliases `/lstep/send*` | Permissioned |
| Link token | `POST /api/v1/owners/:id/line/link-token` | LIFF binding |
| Unlink | `DELETE /api/v1/owners/:id/line` | |
| Line customers | `/api/v1/clinics/.../line-customers` list + link-owner | |
| L-step settings / tags / CSV / triggers | Additional handlers in `internal/lstep` (large package ~260 Go files) | Admin CRM |
| Reservation public/LIFF APIs | Overlap with `internal/reservation` for slots/booking | Cross-domain; reservation write owner remains reservation package |

**Seed**: `003_demo` has partial LINE demo rows (`line_customers.csv`, `line_send_logs.csv`); `line_reservation_settings.csv` may be foreign WIP — **do not edit** in parallel agent sessions.

---

## 4. Frontend / client surfaces (high-level)

| Surface | Location | Spec anchor |
|:---|:---|:---|
| Clinic admin LINE reservation | `frontend/src` features for `/line-reservation/*`, settings integrations | screens 28 |
| L-step admin | `/settings/integrations/lstep`, tags, analytics, checkup-sync | screens 31, 34; line README URLs |
| Owner reserve app | `frontend/line-reserve/` | screens 37; `reservation-spec.md` |
| LIFF pet health | `frontend/liff/` | screens 38; architecture link-token |
| Shared axios API | feature APIs calling `/v1/.../line*` and public reserve endpoints | |

---

## 5. Correlation result (high-level)

| Area | Doc coverage | Code presence | Residual |
|:---|:---|:---|:---|
| Webhook entry | architecture | BE present | **partial** — signature/ops details need setup.md discipline; no full re-verify this packet |
| Link token / LIFF bind | architecture + 38 | BE + liff app | **partial** — S12/V05 still carry 【要実測】 |
| Owner reserve journey | reservation-spec + 37 | line-reserve app | **partial** — S04 marks |
| L-step triggers/CPM | lstep-integration | large BE + admin FE | **partial** — V05 marks; not field-audited here |
| Delivery monitor | 34 | FE + BE logs | **partial** — S01 observation marks |
| Cost analysis | cost-analysis.md | n/a product | docs-only; **残差なし** for product drift |
| available-staffs | n/a | WONTFILE | **残差なし** (do not invent) |
| Secrets in docs | setup | — | Must stay non-secret; **no residual ID** if clean |

### Overall disposition

**partial** — documentation tree and major surfaces **exist and map**, but this packet did **not** prove field-level parity for every trigger, LIFF step, or admin form.

**Not** 「残差なし」 for full LINE audit.  
**Not** a new critical BUG ID without measurement.

| Residual ID | Meaning |
|:---|:---|
| **SPEC-TOP-LINE-AUDIT** | Remains open until a deeper pass or PO WONTFILE of full pass |
| **SPEC-TOP-LINE-AUDIT-DEEP** (follow-up label) | Optional: webhook contract + line-reserve step matrix + L-step trigger table vs code |
| Measurement marks in V05/S04/S12 | **TASK-010** |
| Seed for LINE-heavy demos | Prefer TASK-009 only where clinical; LINE settings may already seed |

---

## 6. Recommended deep-pass checklist (future)

1. Diff `reservation-spec.md` steps ↔ `frontend/line-reserve/src` routes.  
2. Diff architecture webhook events ↔ `ReceiveLineWebhook` handlers.  
3. Diff lstep-integration trigger list ↔ worker/scheduler + BE trigger code.  
4. Confirm screens 28/31/34/37/38 API tables still match handlers (symbol-drift helps only named tokens).  
5. Never log channel secrets; setup.md remains ops-only.

---

## 7. Acceptance vs todo TASK-019

| Acceptance | This report |
|:---|:---|
| ① Result recorded | **Yes** |
| ② Opens with IDs or 残差なし | **partial** + SPEC-TOP-LINE-AUDIT open; cost-analysis product drift **残差なし** |
| ③ Product code only after decision | Honored — no code |

---

## 8. Non-actions

- No live webhook calls  
- No secret inspection beyond confirming policy  
- No product code  
- No edit to `line_reservation_settings.csv`  
