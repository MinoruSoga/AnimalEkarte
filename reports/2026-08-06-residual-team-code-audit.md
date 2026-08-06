# Residual team static code audit — BUG-001..032 (IMPLEMENTED_UNVERIFIED)

| Field | Value |
|---|---|
| Date | 2026-08-06 |
| Repo | `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte` |
| Branch | `main` (no checkout / no product code change) |
| Scope | STATUS.md §3 個票 `対応状況` = IMPLEMENTED_UNVERIFIED (32) |
| Method | Static existence only: claimed implement SHA from STATUS + key code markers in tree |
| Git object check | Not available in this agent environment (no shell/`git cat-file`). Classification uses **key marker presence** (STATUS-allowed alternative). |
| Not done | migrate / seed / db_reset / browser / VERIFIED_FIXED |

## 1. Summary counts

| Classification | Count |
|---|---:|
| CODE_PRESENT | 32 |
| CODE_MISSING | 0 |
| AMBIGUOUS | 0 |
| **Total** | **32** |

- **CODE_MISSING that would reopen agent product work: none**
- **Agent product OPEN remains 0** (static tree only; runtime/browser still DEFERRED per STATUS / BROWSER_VERIFICATION_BACKLOG)
- **Do not mark VERIFIED_FIXED** (this audit is existence, not scenario re-verify)

## 2. Campaign re-check (requested)

| BUG | Claim (STATUS) | Marker evidence in tree | Class |
|---|---|---|---|
| **BUG-012** LTV hang | payments clinic_id scope; `max_single_visit` JOIN not correlated; `ListOwnerAggregation` 20s; FE aggregations/CPM 25s axios timeout; campaign SHA `85e7be513` | BE: `backend/internal/lstep/aggregation_service.go` `aggregationQueryTimeout = 20 * time.Second` + `WithTimeout`; `backend/internal/owner/ltv_repository.go` payments JOIN via `billings b0 … AND b0.clinic_id = ?`, `max_single_visit_amount` via LEFT JOIN aggregate on `billings b2` (not per-row correlated subquery). FE: `frontend/src/features/aggregation/api/get-aggregations.ts` `timeout: 25_000` + BUG-012 comment | **CODE_PRESENT** |
| **BUG-024** permission matrix | commit `ce8ae6f46`; FE full matrix + explicit false PATCH; toast suppress on rules lag; BE `replaceRules` bool Select INSERT | FE: `PermissionGroupSettings.tsx` BUG-024 toast guard; `permission-group-settings-model` builds rules with `can_view: false` etc.; tests in `PermissionGroupSettings.test.tsx`. BE: `permission_group_repository.go` `replaceRules` + `Select(… "CanView", …)` INSERT | **CODE_PRESENT** |
| **BUG-026/029** false-success guards | commit `e0c5cc5e1`; insurance coverage 0–100; payment name precheck; SidePanel/master save toast only after await success | `validateInsuranceForm` + tests (BUG-026); `validatePaymentMethodForm` duplicate precheck (BUG-029); `use-master-save.ts` `toast.success` only after successful `mutateAsync` (catch → no success toast) | **CODE_PRESENT** |
| **BUG-003** `exam_reference_ranges` seed in `003_demo` | commit `1c6395915`; CSV in demo; `assessExamResult` | File `backend/migrations/seeds/003_demo/exam_reference_ranges.csv` (dog/cat ref_min/ref_max rows); `manifest.json` lists table; BE `assessExamResult` / `computeExamResultStatus` in `examination_service.go`. **Apply still human migrate/seed (not run here)** | **CODE_PRESENT** |
| **BUG-008/014** LIFF mock fail-closed + compose | commit `1c6395915`; compose `LIFF_MOCK`/`VITE_LIFF_MOCK`; mock lookup fail → 503 no real-auth fallthrough | `docker-compose.yml` `LIFF_MOCK: "true"`, `VITE_LIFF_MOCK=true`; `backend/internal/middleware/liff_auth.go` BUG-008/014 block + `StatusServiceUnavailable`; FE `shared-liff/use-liff.ts` mock-token when `VITE_LIFF_MOCK===true` | **CODE_PRESENT** |

## 3. Full table BUG-001..032

| BUG | STATUS claim (commit / key evidence) | Tree evidence (markers) | Status |
|---|---|---|---|
| BUG-001 | `d7bf32f2214d6bb6c252b99b001d2ed2044de7c9` — `PetRepository.FindAll` space-independent owner name, owner id, pet_number, blank fail-closed | `backend/internal/pet/repository.go` regexp_replace space strip + pet_number ILIKE + blank fail-closed; `TestPetRepository_FindAll_Search` | CODE_PRESENT |
| BUG-002 | `a17d39d6f46ddaf8afcba7ed53419dbc4f92e968` — death success syncs outer pets/editingPet | `PetEditModal.tsx` / Fields BUG-002 outer sync callback; deceased dialog `onRecorded` | CODE_PRESENT |
| BUG-003 | `1c6395915` — `003_demo/exam_reference_ranges.csv` | CSV + manifest entry + `assessExamResult` (see §2) | CODE_PRESENT |
| BUG-004 | current truth `examination_parent_audit_test.go` write order items→revision→audit→status | `TestExaminationService_ConfirmWithItemsPersistsItemsBeforeStatusTransition`; service confirmed-last write comments | CODE_PRESENT |
| BUG-005 | `dfd653eaa…` active doctor filter; `staffSelectorList`; no `?? "doctor"` fail-open | `use-staffs.ts` fail-closed staff_type; `queryKeys.masters.staffSelectorList()`; tests | CODE_PRESENT |
| BUG-006 | `3db97bb19…` `formatPatientPetDetails`; PatientInfoCard default 「不明」 | `format-pet-details.ts`; VaccinationForm passes petDetails; PatientInfoCard default `"不明"`; tests reject fixed 「9才5ヶ月…」 | CODE_PRESENT |
| BUG-007 | `7f663716d…` `useGetVaccinations` pet_id + HISTORY_FETCH_LIMIT | `get-vaccinations.ts` + BUG-007 tests | CODE_PRESENT |
| BUG-008 | `1c6395915` compose LIFF mock + LiffAuth fail-closed | §2 campaign | CODE_PRESENT |
| BUG-009 | `6e9674286` + review `55125f858` tab→server status/page/limit, 不明 fail-closed | `HospitalizationList` statusFilter/serverTotal; transforms 「不明」; constants map tab→wire status | CODE_PRESENT |
| BUG-010 | `646fb4353…` + audit residuals — clinical plan fields single PATCH / controlled | `clinical-plan.ts` physical_exam/diagnosis_details/treatment_policy; ClinicalPlanSection controlled tests | CODE_PRESENT |
| BUG-011 | `b65cf69ef…` `AllocateNextEstimateNo` EST-{N} | estimate repository/service AllocateNext + concurrent/tx tests | CODE_PRESENT |
| BUG-012 | LTV hang fixes (campaign `85e7be513`) | §2 campaign | CODE_PRESENT |
| BUG-013 | `74aa3e2c6…` `GET …/unbilled-details` + vaccination_master_unbillable | routes + `UnbilledDetails` + service tests | CODE_PRESENT |
| BUG-014 | same C-LIFF-AUTH as BUG-008 | shared use-liff + compose + LiffAuth (§2) | CODE_PRESENT |
| BUG-015 | `98639b4fa` + `28539d466` `toggleWeightValueAndUnit` | VitalsTabRows toggle + round-trip tests | CODE_PRESENT |
| BUG-016 | `7ee0edbac…` entity read notFound gate on vax/exam/hosp | `entity-read-result.ts`; form hooks isReadNotFound | CODE_PRESENT |
| BUG-017 | `7f7106375…` fieldErrors + ExaminationForm.validation | ExaminationForm.validation.test.tsx Japanese alerts/ARIA | CODE_PRESENT |
| BUG-018 | `75f8912fc…` + `0b2bde815…` complete command / idempotent | `accounting_complete_test.go` CompleteAccounting + Idempotency | CODE_PRESENT |
| BUG-019 | same commit as 016 — estimate Not Found | EstimateForm `resolveEntityReadResult` + non-disclosure UI | CODE_PRESENT |
| BUG-020 | `617f6f9bf…` phone error clear on valid | ReservationFormModal BUG-020 test + isValidOwnerPhone clear | CODE_PRESENT |
| BUG-021 | `eb7db0dc9…` death date empty/future field errors | PetDeceasedDialog fieldErrors + aria-invalid tests | CODE_PRESENT |
| BUG-022 | `fc3c12b28…` deceased transform / 不明 fail-closed | `mapPetStatusLabel` deceased→死亡; unknown→不明 | CODE_PRESENT |
| BUG-023 | `41ba79b4c…` `permission_group_name_conflict` | apperrors code + FE handle-api-error map | CODE_PRESENT |
| BUG-024 | `ce8ae6f46` matrix + replaceRules | §2 campaign | CODE_PRESENT |
| BUG-025 | `4335e3a99…` create builder sends `is_active` | chief-complaint / interview-template create builders + tests | CODE_PRESENT |
| BUG-026 | `e0c5cc5e1` insurance coverage validate | §2 campaign | CODE_PRESENT |
| BUG-027 | same as 023 `animal_species_name_conflict` | apperrors + FE handle-api-error | CODE_PRESENT |
| BUG-028 | `842fe78d4…` procedure anesthesia required | CreateProcedureRequest anesthesia; treatment-item-side-panel-model options; builders | CODE_PRESENT |
| BUG-029 | `e0c5cc5e1` payment false-success | §2 campaign | CODE_PRESENT |
| BUG-030 | `37c84041b…` + `324b356e8…` Save Create(DoNothing)+Select Updates; ID re-read | `line_reservation_setting_repository.go` DoNothing + identity re-read comments/tests | CODE_PRESENT |
| BUG-031 | `5cf86efc4…` AuthProvider restore on `/login` | use-auth.tsx restoreSession; use-auth-initial-session.test.tsx BUG-031 | CODE_PRESENT |
| BUG-032 | `944f2e4dd…` PreviewCheckupSync 15s, LIMIT/cap, FE 20s | `CheckupSyncPreviewTimeout = 15s`; owner cap 100 tests; FE `timeout: 20_000` on preview | CODE_PRESENT |

## 4. Agent work implication

| Question | Answer |
|---|---|
| Any CODE_MISSING reopening implement loop? | **No** |
| Agent product OPEN after this audit? | **Remains 0** |
| Residual risk | Existence ≠ runtime green. Browser/scenario still DEFERRED (32). BUG-003 needs **human seed apply** before H/L can show in demo DB. BUG-008/014 need **compose restart** so env flags bind to running containers. |
| VERIFIED_FIXED? | **Not marked** (forbidden for this audit) |

## 5. Notes / limitations

1. Commit SHAs cited in STATUS were **not** re-checked via `git cat-file` (no shell in this subagent). Marker-based presence is the STATUS-allowed fallback and was applied uniformly.
2. Truncated SHAs in STATUS (e.g. `dfd653eaa…`, `7f7106375…`, campaign short SHAs) were not expanded; markers still matched claimed behavior.
3. No product code, migrate, seed, force-push, or STATUS status flip performed.
4. Evidence paths are absolute under `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/`.

---

**Bottom line:** 32/32 IMPLEMENTED_UNVERIFIED bugs have **CODE_PRESENT** markers in current tree. **CODE_MISSING = 0 → agent product open remains 0.** Browser verification remains out of scope for this residual static audit.
