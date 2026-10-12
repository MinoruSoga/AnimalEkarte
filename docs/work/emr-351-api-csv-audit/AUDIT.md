# EMR-351: API vs CSV-import validation audit (21 baseline tables)

Status: audit document, draft for Codex/Lead review. No production code changed.
Baseline: `ed04c0237f81ddc2890d39b528d3608a58180c8b`.
Canonical contract: `backend/internal/csvimport/cutover_contract.go` — 21 tables, 236 columns
(contract hash `9e02897b86c742f94162ddfd4171067b3036a3ac4f055c6d3fd4a614e0fb37c9`).
Companion matrix: [column-matrix.csv](column-matrix.csv) (one row per contract
column). Lead-produced artifacts in this directory (optional provenance, not
sole authority): [aggregate-counts.json](aggregate-counts.json),
[api-binding-catalog.json](api-binding-catalog.json),
[COUNT-SUMMARY.md](COUNT-SUMMARY.md) (see §7).

## 1. Method

For every column in `CutoverTableSpecs` (inventory mirrored in
`evidence/baseline-column-inventory.json`) we compared:

- **API surface** — request binding + service-layer create/update validation.
  DTO binding tags were filled from the mechanically extracted catalog
  [api-binding-catalog.json](api-binding-catalog.json) (236 columns / 168 bound, baseline
  `ed04c0237f81ddc2890d39b528d3608a58180c8b`). The catalog proves struct/field
  binding only — not service validation, handlers, or writeability; service facts
  below are cited from source independently.
- **Importer preflight** — manifest/digest/header/hash, ID bands, required strings,
  CSV-parent references, unavailable-parent empty rules, payment-graph arithmetic.
- **Apply/DB floor** — `001_init.sql` NOT NULL / CHECK / enum / UNIQUE / FK /
  triggers. Static-DDL facts are code-confirmed; real apply against the current
  DB is NOT RUN.

Difference classes: `OK` (no meaningful gap), `EMR` (consumer follow-up needed),
`MIG` (producer-side guarantee needed), `HIST` (intentional historical acceptance
with code+governing-reason evidence), `OWNER` (owner decision required), `UNK`
(evidence not captured — see §5).

## 2. Confirmed findings (headline)

1. **`staff_clinic_assignments` is not in the 21-table contract.** API clinical staff
   references require an active assignment row (`AssertMedicalRecordDoctorInClinic`,
   `validateClinicalStaffReference`); imported `staffs` rows get none from the import
   itself. Note: staff identity/assignment provisioning is already tracked by
   **EMR-144** (Blocked pending authoritative identity map; a STG-28 staff-attach
   receipt exists but production status is unconfirmed) and **EMR-145** (waiting on
   EMR-144 same-target login/clinic/permissions acceptance). This audit records the
   integration-receipt gap only; it does not propose a new provisioning flow. → T-A.
2. **`staffs.staff_type` is not a CSV column.** DDL defaults every imported staff row
   to `'doctor'` (`001_init.sql:277`) — a conditional risk: any actual non-doctor
   staff WOULD be mislabeled and (mis)satisfy the `staff_type='doctor'` half of the
   doctor check (the assignment half still fails until provisioning completes,
   EMR-144/145). Whether non-doctor staff exist in the data is unverified — no
   actual nurses/trimmers are proven misclassified. → T-B (EMR-376).
3. **Negative-value taxonomy is layered.** Current static DDL omits amount CHECKs
   on the historical accounting tables; this describes storage behavior, not
   blanket business acceptance. Explicit AE-MIG-NEG-1 reasoning at
   `001_init.sql:1888-1889;6257-6259` supports only `billings.total_amount` and
   cash/card `payment_splits.amount` as HIST here. Master negatives belong to
   EMR-345. `billing_items.unit_price`, `payments.total_amount`,
   `payments.discount_amount`, and negative estimate values are OWNER/EMR-376
   pending broader governing policy. Submitting such values on the relevant
   minimum-constrained DTO path WOULD fail; no actual DB store or API save was
   executed. → T-I.
4. **DB floors the preflight never checks** (WOULD fail at COPY/apply — static DDL,
   no actual apply run): `name_kana !~ '[ァ-ヶ]'` on owners+pets, `quantity > 0` on
   billing/estimate items, vital ranges (temp 30–50, HR <500, RR <200, weight >0),
   `end_time >= start_time`, per-clinic uniques (`record_no`, `estimate_no`, master
   names, owner email, non-empty owner phone `001_init.sql:4532-4534`), 1:1
   uniques (inquiries, clinical_plans, trimming details), and enum domains not
   otherwise covered by preflight. → T-F (EMR-374; producer guarantees MIG-37/41).
   Already preflighted and therefore NOT T-F gaps: `billings.medical_record_id`
   uniqueness (`cutover_payment_graph.go:12;22-31`), `payments.billing_id`
   uniqueness + completed-status (`cutover_payment_graph.go:74-87`), billing status
   domain (`cutover_payment_graph.go:37-43`), and payment-method domain.
5. **`billing_items.clinic_id` is trigger-overwritten** by `sync_billing_items_clinic_id`
   (`001_init.sql:6210-6237`) — the CSV value is ignored at apply; contract column is
   informational only. `billing_items` also lacks `tax_rate`, `source`, discount, and
   provenance-link columns → defaults (`'manual'`, `0.10`, NULL). → T-D (EMR-364 adjacency).
6. **`estimate_items` API coverage gap is narrower than the column count suggests.**
   `category`, `discount_rate`, `discount_amount`, and `is_insurance_applicable` ARE
   bound on the item request (`estimate_request.go:66,69-71`) and validated by the
   service (`estimate_service.go:114-161` — category domain, discount 0..100,
   nonnegative amount). Truly unbound on the item request: `tax_type`, `tax_rate`,
   plus the two POPULATED csv-parent links `procedure_id` and
   `merchandise_item_id` (`consultation_id` and `medicine_id` are `must-be-empty`
   UnavailableParent columns — a non-empty value fails preflight, so there is no
   populated-link edit gap; the medicine contract is the separate EMR-343
   appendix). Estimate update also recomputes parent totals from items.
   `estimate_items.unit_price` has API `min=0` (`estimate_request.go:67` +
   `estimate_service.go:124-126`); negative import vs that minimum is a
   user-scope/API-compatibility decision, not an API gap and not implicit
   historical acceptance. → T-C (EMR-375) / T-I (EMR-376).
7. **`payments.method`/`payment_splits.method` are restricted to `cash|credit_card` by
   the payment graph, while the DB enum has 4 values** (`electronic_money`,
   `bank_transfer` unreachable). Producer must map legacy methods or the contract
   decides otherwise. → T-J.
8. **Same-CSV referential pairing is unchecked.** Existence of `owner_id`/`pet_id`/
   `medical_record_id` parents is validated, but not that the referenced pet belongs to
   the referenced owner (appointments, medical_records, billings, estimates,
   vital_records, exams, vaccinations). Parent existence + band checks do not prove
   relational pairing; API update paths re-validate owner↔pet pairing
   (`ValidateReservationOwnerPetLinksWithRepo`, accounting/estimate FK checks), so a
   mispaired import is stored but edit-blocked. → T-G (real gap).
9. **Finalized/terminal imported states freeze ordinary editing — by design; the
   lock shape differs per child.** `medical_records.status='finalized'` → edits
   rejected (`medical_record_crud.go`); inquiry create/update takes the parent
   `FOR UPDATE` and rejects finalized parents (`inquiry_repository.go:63-73`);
   vital updates route through finalized handling (`vital_service.go:203-223`);
   clinical-plan update is guarded by an `UPDATE ... WHERE draft` clause, not a
   row lock (`clinical_plan_repository.go:118-137`); non-mutable appointment
   statuses freeze trimming detail (`trimming_service_update.go:59`);
   `billings.status='completed'` requires post-close reason / `'cancelled'` is
   rejected; `medical_records.appointment_id` is immutable after create. Imported
   `exams` default to `pending` (status not in CSV) → editable. How many finalized
   rows exist is informational counting, not automatically a violation; whether a
   future correction/addendum path applies to imported finalized rows is an owner
   policy decision — no such path is claimed proven here. → T-E.
10. **Actor columns (`entered_by`, `paid_by`, `created_by`) are contract columns
    with real preflight**, not free-form text: they are band columns and validated
    as csv-parents into the same clinic's `staffs` CSV (`cutover_contract.go:251;
    257-259`, `cutover_contract_validate.go:654-669`, `cutover_reference_graph.go:43;
    51-53;194-209`). Cross-clinic actors are rejected by source preflight; permissive
    DB FKs are NOT bundle eligibility. Current graph validation checks source band
    and staff existence, not staff activity.
    `cutover_entered_by_fk_contract_test.go` directly exercises inactive
    `entered_by` preflight behavior only; it is not an owner-approval receipt for
    `paid_by` or `created_by`, nor a general clinical-capability proof. These
    actor/activity columns are OWNER/EMR-376 pending an explicit policy. CSV table
    absence alone does not imply a new capability-provisioning flow.
11. **`payments.clinic_id` is resolved OK, not a defect.** Later DDL
    (`001_init.sql:4183-4248`) adds the column NOT NULL with a clinic FK and
    composite billing/payment-method clinic FKs; COPY carries the column in the
    common path.
12. **Equal-time appointments are an API-parity observation, not a DB violation.**
    DB enforces `end_time >= start_time` while `sharedkernel.ValidateTimeRange`
    (`:116-120`) requires strictly `>`. Lead's count run observed equal-time rows
    (hachioji 81712, jouto 126637, shikishima 4981, hakobuneco 26310) — they store
    fine but that field cannot be re-saved through the API. → T-L (EMR-376).

## 3. Classification tally (matrix `cls` column)

Derived from the matrix by read-back of the `cls` field (each row carries exactly
one class label; totals sum to the 236 data rows):

| cls | rows | meaning |
|-----|------|---------|
| OK | 129 | no meaningful gap |
| EMR | 40 | consumer follow-up (EMR-374/375/376 groups incl. EMR-345 scope) |
| MIG | 44 | producer guarantee (EMR-374 scope; MIG-37/41) |
| HIST | 2 | explicit acceptance with code+governing reason ONLY: `billings.total_amount` (AE-MIG-NEG-1 negative billing, `001_init.sql:1888-1889`) and `payment_splits.amount` (AE-MIG-NEG-1 negative cash-card split, `001_init.sql:6257-6259`) |
| OWNER | 20 | owner decision pending (terminal-state locks, generated timestamps, pets.status/deceased_at, appointments.source, actor columns, non-family negative amounts) — EMR-376 |
| UNK | 1 | `payments.insurance_ratio` broader service rule — T-K, mapped EMR-376 |

These are classification counts, not violation counts. Per-clinic occurrence
counts: see §7 (Lead count run completed; aggregates at
[aggregate-counts.json](aggregate-counts.json)).

## 4. Follow-up mapping (consumer gaps ≠ EMR-345)

Draft groups are anchored to actual Plane items created by Lead — no new tickets
are proposed by this audit:

| Group | Scope | Plane item | State |
|-------|-------|-----------|-------|
| T-A | staff assignment receipt gap | EMR-144 + EMR-145 | existing; EMR-144 Blocked (identity map), EMR-145 waiting on 144 |
| T-B | staffs.staff_type / non-CSV staff columns | EMR-376 | Needs Human |
| T-C | imported fields with no API write path | EMR-375 | Needs Human |
| T-D | billing_items coverage/defaults | EMR-376 | Needs Human |
| T-E | terminal/finalized imported states | EMR-376 | Needs Human |
| T-F | apply-time DB floors | EMR-374 (+ MIG-37/41 guarantees) | Backlog |
| T-G | same-CSV referential pairing | EMR-374 (+ MIG-37/41) | Backlog |
| T-H | generated identifiers / system fields | EMR-376 | Needs Human |
| T-I | negative-value taxonomy beyond EMR-345 | EMR-376 (+ existing EMR-345) | Needs Human |
| T-J | payments/splits method domain | EMR-376 (+ MIG-37/41) | Needs Human |
| T-K | residual unknowns | EMR-376 (owner investigation) | Needs Human |
| T-L | API-side value constraints w/o import parity | EMR-376 | Needs Human |

The sections below keep the per-group evidence; the draft text is superseded by
the actual items above.

### T-A — staff_clinic_assignments integration receipt gap (EMR; bounded)

- **Problem:** the 21-table contract imports `staffs` but not
  `staff_clinic_assignments` (`001_init.sql:1179-1188`, `UNIQUE(staff_id,clinic_id)`).
  `AssertMedicalRecordDoctorInClinic` requires active staff + `staff_type='doctor'` +
  active assignment; `validateClinicalStaffReference` requires the same membership.
- **Existing tickets — do NOT duplicate:** **EMR-144** owns staff
  identity/assignment provisioning (currently Blocked pending the authoritative
  identity map; a STG-28 staff-attach receipt exists but production status is
  unconfirmed). **EMR-145** owns the switch-work follow-up, waiting on EMR-144
  same-target login/clinic/permissions acceptance. This draft is only a bounded
  verification/integration-receipt gap, not a new provisioning flow.
- **Current behavior:** migrated staff exist but fail doctor/clinical-reference
  validation on API create/update paths (medical records, appointment doctor
  changes, vaccination/vital staff refs) until provisioning completes.
- **Scope:** record the dependency and add one verification receipt: after
  EMR-144/145 land, prove every imported active doctor referenced by imported
  clinical rows resolves through `AssertMedicalRecordDoctorInClinic`. If the
  check is proven missing in code, scope it narrowly; otherwise close as covered.
- **Non-scope:** no new provisioning pipeline; no change to actor/`entered_by`
  semantics (intentionally historical); EMR-144/145 and raw rosters untouched.
- **Acceptance:** integration receipt exists proving imported staff pass the
  clinical-reference assertion post-provisioning, or an exact missing-code delta
  is named; decision on `staff_type` correctness tracked under T-B.
- **Evidence:** `cutover_contract.go` (21 specs, no assignments table),
  `001_init.sql:1179-1188`, `trimming_service_update.go:83`, EMR-144/EMR-145
  tickets, stg-uat-staff-attach receipt.

### T-B — staffs.staff_type and non-CSV staff columns (EMR-376)

- **Problem:** CSV carries `id, clinic_id, name, license_number, is_active,
  reservation_visible` only. `staff_type` defaults `'doctor'` for all imported rows;
  `account_id`, `occupation_id`, `sort_order`, `reservation_display_name`,
  `reservation_comment`, `reservation_image_url` are defaulted/NULL.
- **Risk (conditional, not proven):** any actual non-doctor staff WOULD be
  mislabeled `doctor` — combined with completed provisioning (EMR-144/145) this
  could grant doctor capability incorrectly. Whether non-doctor staff exist in the
  data is unverified; no actual nurses/trimmers are proven misclassified.
- **Scope:** owner decision on adding `staff_type` (and optionally `occupation_id`,
  `sort_order`) to the contract, or producer guarantee that all migrated staff are
  doctors, or a post-import fix-up list.
- **Acceptance:** imported staff roles match source-of-truth; no nurse/trimmer passes
  the doctor check; matrix updated.
- **Evidence:** `001_init.sql:277` (default), `staff_request.go:17;70` (API enum),
  `cutover_contract.go` staffs spec.

### T-C — imported columns with no API write path (EMR-375)

- **Problem:** three clusters of imported columns cannot be written through the
  ordinary API:
  1. `estimate_items`: `tax_type`, `tax_rate`, and the POPULATED csv-parent links
     `procedure_id` / `merchandise_item_id` are not bound on the item request
     (`estimate_request.go:64-85` binds name, category, unit_price, quantity,
     discount_rate, discount_amount, is_insurance_applicable, sort_order only);
     estimate update also recomputes parent totals from items. `consultation_id`
     and `medicine_id` are `must-be-empty` UnavailableParent columns — not
     populated-link gaps; the medicine contract is the EMR-343 appendix.
  2. `inquiries`: only `chief_complaint` and `notes` are update-bound
     (`inquiry_request.go:3-18`, `inquiry_service.go:13-19,48-55`); imported
     `owner_observations`, `history`, `allergy_info`, `current_medications`, and
     `staff_id` have no update binding.
  3. `procedures.is_surgery` is absent from the API DTO and service mapping
     (`procedure_request.go:3-14;31-43`, `procedure_service.go:20-34`);
     `billing_items.sort_order` is bound on create/complete only, not update.
- **Risk:** imported values in these columns are preserved-or-overwritten
  silently depending on the update path; totals recompute may clobber imported
  `subtotal`/`tax_total`/`total_amount` on estimates.
- **Scope:** per cluster, either extend the DTO+service to the imported fields,
  or document them as import-only with explicitly locked behavior — an
  EMR consumer decision per cluster.
- **Non-scope:** `estimate_items.medicine_id` semantics — owned by EMR-342/343
  (appendix); stays `must-be-empty` until that contract lands.
- **Acceptance:** each listed column is either API-editable with validation
  equivalent to create paths, or documented import-only with locked behavior;
  recompute does not silently rewrite imported totals.
- **Evidence:** `estimate_request.go:64-85`, `estimate_service.go:114-161`,
  `inquiry_request.go:3-18`, `inquiry_service.go:13-19,48-55`,
  `procedure_request.go:3-14;31-43`, `procedure_service.go:20-34`,
  `billing_item_request.go:56`, `accounting_request.go:193`,
  `cutover_reference_graph.go:54`.

### T-D — billing_items coverage/defaults (EMR; EMR-364 adjacent)

- **Problem:** CSV lacks `tax_rate` (DB default `0.10`), `source` (default `'manual'`),
  `discount_rate`/`discount_amount`, `other_reason`, `created_by`, and provenance
  links (`merchandise_item_id`, `treatment_id`, `appointment_id`, `vaccination_id`,
  `exam_id`, `trimming_*`). `clinic_id` column is trigger-overwritten
  (`001_init.sql:6210-6243`).
- **Scope:** owner decision per column: add to contract (EMR-364 already proposes
  `tax_rate`) vs keep defaults; confirm `'manual'` source is the intended label for
  migrated lines; document trigger behavior in contract docs.
- **Acceptance:** decision per missing column recorded; if `tax_rate` is added it is
  coordinated with EMR-364 and provider output; matrix updated.
- **Evidence:** `001_init.sql:1913-1935,4048,6165-6243`, EMR-364 evidence file.

### T-E — imported terminal/finalized states freeze API editing (EMR docs + OWNER)

- **Problem:** imported `status='finalized'` medical records, non-mutable appointment
  statuses, `completed`/`cancelled` billings, and non-draft estimates lock ordinary
  edit paths. The lock shape differs per child: inquiry create/update takes the
  parent `FOR UPDATE` and rejects finalized parents
  (`inquiry_repository.go:63-73`); vital updates route through finalized handling
  (`vital_service.go:203-223`); clinical-plan updates are guarded by an
  `UPDATE ... WHERE draft` clause rather than a row lock
  (`clinical_plan_repository.go:118-137`); trimming detail requires a mutable
  appointment status (`ValidateTrimmingAppointmentMutable`,
  `trimming_service_update.go:59`); completed billings require a post-close reason
  and cancelled billings reject edits; `medical_records.appointment_id` is
  immutable after create.
- **Important framing:** the count of finalized/terminal imported rows is
  informational — not automatically a violation. Whether a correction/addendum
  path should apply to imported rows is an owner policy decision; the
  `medical_record_addenda` table exists in DDL but a usable correction flow for
  imported finalized records is not claimed proven.
- **Scope:** document the intended correction path per state and confirm owner
  acceptance that migrated terminal states are intentionally read-only; identify
  any field that has *no* correction path (e.g., `medical_records.appointment_id`).
- **Acceptance:** per-state documented path or explicit owner acceptance of
  read-only; no imported row has an unrecoverable wrong value without a defined
  route.
- **Evidence:** `inquiry_repository.go:63-73`, `vital_service.go:203-223`,
  `clinical_plan_repository.go:118-137`, `trimming_service_update.go:59`,
  `medical_record_crud.go`, `001_init.sql:1420`, `medical_record_addenda`
  (`001_init.sql:1457-1472`).

### T-F — apply-time DB floors missing from preflight (EMR-374; producer guarantees MIG-37/41)

- **Problem:** these constraints WOULD fire at COPY/apply with no importer precheck
  (static DDL; no actual apply run by this audit):
  - uniques: master names per clinic (procedures/vaccines/merchandise_items and other
    name-unique masters), `owners.email` non-empty per clinic,
    `owners.phone` non-empty per clinic (`001_init.sql:4532-4534`),
    `medical_records.record_no`, `estimates.estimate_no`;
  - 1:1 uniques: `inquiries.medical_record_id`, `clinical_plans.medical_record_id`,
    `appointment_trimming_details.appointment_id`;
  - CHECKs: `name_kana !~ '[ァ-ヶ]'` (owners/pets), `quantity > 0` (billing/estimate
    items), vital ranges, `end_time >= start_time`;
  - enum domains not otherwise covered by preflight (procedures.anesthesia/tax_type,
    pets.gender/status, owners.membership_type, appointments.visit_type/status/
    source, medical_records.status/visit_type, vital_records.weight_unit,
    billing_items.category/tax_type, estimates.status,
    estimate_items.category/tax_type, vaccinations.next_schedule_type,
    vaccines.species).
- **Already preflighted (NOT in scope):** `billings.medical_record_id` uniqueness
  (`cutover_payment_graph.go:12;22-31`), `payments.billing_id` uniqueness and
  completed-status (`cutover_payment_graph.go:74-87`), billing status domain
  (`cutover_payment_graph.go:37-43`), and payment-method domain.
- **Scope:** producer (MIG) guarantees these hold in emitted CSVs; EMR may optionally
  add cheap preflight checks for the same rules (defense in depth) — this audit does
  not implement them.
- **Acceptance:** Lead aggregate scanner reports 0 violations per rule per clinic, or
  violations triaged to producer fixes; owner confirms `name_kana` CHECK intent
  (it rejects katakana — unusual; may be a legacy artifact worth revisiting).
- **Evidence:** `001_init.sql:297,1116,1344,1763-1769,1845,1935,2194-2201,2450-2459`;
  enum list `001_init.sql:77-134`; preflight coverage
  `cutover_payment_graph.go:12;22-31;37-43;74-87`.

### T-G — same-CSV referential pairing unchecked (MIG)

- **Problem:** importer validates parent *existence*, not pairing: `pet_id` ↔
  `owner_id` (appointments, medical_records, billings, estimates), `pet_id` ↔
  `medical_record_id` pairing (vital_records, exams, vaccinations where the record's
  own pet differs). API update paths re-validate pairing → mispaired rows edit-block.
- **Scope:** producer guarantee (join-consistent emission) + scanner counts; EMR
  optional precheck.
- **Acceptance:** 0 mispaired rows per clinic, or documented owner-approved
  exceptions.
- **Evidence:** `trimming_service_update.go:66`, accounting/estimate FK validation,
  `cutover_reference_graph.go` (existence-only semantics).

### T-H — generated identifiers, actor columns, and system fields (EMR-376; owner policy pending)

- **Problem:** `record_no`, `estimate_no`, `created_at`/`recorded_at`,
  `entered_by`, `created_by`, `paid_by`, `appointments.source`, `pets.deceased_at`
  (+`pets.status` update binding), `billings.completed_at` are
  system/workflow-generated or create-only in AE but imported historically.
  Actor columns ARE contract+preflight columns — cross-clinic actors are
  rejected by source preflight (`cutover_contract_validate.go:654-669`,
  `cutover_reference_graph.go:43;51-53;194-209`). The
  `cutover_entered_by_fk_contract_test.go` test proves CURRENT preflight
  behavior for inactive `entered_by` only — it is NOT general owner acceptance
  of `paid_by`/`created_by`/other actor columns, and it does not prove all
  clinically-migrated actors are accepted. `pets.pet_number` is PATCH-editable
  (`pet_request.go:180`, max=50) — NOT an uncorrectable identifier. Identifiers
  with no write path (record_no post-create, estimate_no, deceased_at) have no
  API correction route today.
- **Scope:** policy waiting on owner — confirm per column whether each is
  intentionally historical or needs an admin correction path; actor-column
  acceptance (including inactive staff) is a separate owner/policy decision,
  not inferred from absent DB CHECKs or from the entered_by preflight test.
- **Acceptance:** per-column decision recorded; where correction path required,
  linked ticket exists.
- **Evidence:** `cutover_entered_by_fk_contract_test.go`,
  `cutover_contract_validate.go:654-669`, `cutover_reference_graph.go:43;51-53;
  194-209`, `pet_request.go:180;209-214`, `001_init.sql:1415,1688,1333`.

### T-I — negative-value taxonomy beyond EMR-345 (EMR/OWNER)

- **Problem:** API rejects negatives on `billings.total_amount` (and requires
  subtotal+tax=total), `billing_items.unit_price` (create/complete `min=0` plus
  service check), all estimate totals/amounts (create `min=0` +
  `estimate_service_tx.go` validation), and `estimate_items.unit_price`
  (`estimate_request.go:67` `min=0` + `estimate_service.go:124-126`);
  DB has no CHECKs — that is a current permissive-DB observation, not by itself
  proof of owner acceptance. Only two columns carry an EXPLICIT governing reason:
  `billings.total_amount` (AE-MIG-NEG-1 negative-billing acceptance,
  `001_init.sql:1888-1889`) and `payment_splits.amount` (AE-MIG-NEG-1 negative
  cash-card split acceptance, `001_init.sql:6257-6259`) — these stay HIST, and
  this is NOT a blanket "all billing/payment negatives accepted" policy.
  `payment_splits.amount` is legal in DB, importable (graph requires ≠0), but
  not creatable via API (`accounting_request.go:272` `min=1`). The historical
  master-only task leaves existing negative billing/estimate data untouched;
  that does not prove all clinical-migration negatives are accepted.
- **Scope:** confirm per-column: `billings.total_amount` +
  `payment_splits.amount` = HIST (exact AE-MIG-NEG-1 cites above);
  `billing_items.unit_price`, `payments.total_amount`,
  `payments.discount_amount`, and estimates totals /
  `estimate_items.unit_price`/`discount_amount` = OWNER decides
  (historical-accept vs producer sanitize — no implicit estimate acceptance);
  `payments.discount_amount` sign policy unknown → owner confirm.
- **Acceptance:** taxonomy table approved; estimate policy recorded; no ambiguity
  left on whether negative estimates may be imported.
- **Evidence:** `001_init.sql:1888-1889`, `001_init.sql:6257-6259` (AE-MIG-NEG-1),
  `estimate_request.go:53-57;67`, `estimate_service.go:124-126`,
  `estimate_service_tx.go`, `accounting_request.go:165;272;333`,
  `billing_item_request.go:40`, `cutover_payment_graph.go`.

### T-J — payments/splits method domain (MIG + OWNER)

- **Problem:** payment graph accepts only `cash|credit_card`; DB enum includes
  `electronic_money`, `bank_transfer`. Legacy rows with other methods are
  unrepresentable.
- **Scope:** producer maps legacy methods (documented mapping), or contract/payment
  graph widened by owner decision; verify counts of non-cash/card methods in source.
- **Acceptance:** 0 rows outside the agreed domain, or contract updated on both sides.
- **Evidence:** `cutover_payment_graph.go`/`cutover_payment_contract.go`,
  `001_init.sql:132`.

### T-L — API-side value constraints with no import-side parity check (EMR-376)

- **Problem:** the API enforces value rules that the importer never checks — the
  imported value is stored but cannot be re-set through the API on that column:
  - `owners.phone` max 30 + service regex
    `^0[0-9]{1,4}-?[0-9]{1,4}-?[0-9]{4}$` and `owners.postal_code` postal format —
    `owner/validators.go:31-87` covers Phone and PostalCode only;
    `owners.email` max 254 + email regex; `owners.discount_rate` service range
    0–100 (`validators`/`shared kernel :109-113`). `owners.company_phone` (max 30)
    and `owners.home_postal_code` (max 10) have NO service format check — parity
    is limited to the max-length binding.
  - `procedures.tax_rate` and `merchandise_items.tax_rate` API `min=0 max=1`
    (`procedure_request.go:13;42`, `merchandise_item_request.go:19;42`); this is a
    range-parity gap independent of the typed-decimal/precision representation
    question (EMR-365 covers vaccine precision only — do not delegate the whole
    master-tax question to it).
  - `pets.weight` service `>=0` (`pet validators :75`);
    `merchandise_items.category` API subset `{food,goods,other}` of the 12-value
    DB enum; `merchandise_items.sort_order` update `min=0`;
    `estimate_items.discount_rate` 0–100 / `discount_amount` nonnegative
    (`estimate_service.go:114-161`).
  - `appointments.end_time`: API requires strictly `>`
    (`sharedkernel.ValidateTimeRange:116-120`) vs DB `>=`; equal-time rows are
    parity observations (counts: hachioji 81712, jouto 126637, shikishima 4981,
    hakobuneco 26310), not DB violations.
  - `appointments.status`: general reservation DTO oneof covers 7 of 8 DB labels
    (includes `completed`; excludes `no_show`); the trimming-scoped DTO covers 6
    (excludes `completed` and `no_show`); DB enum has 8 labels, no `tentative`.
    Scoped-DTO counts are not proof of all appointment edit failures.
  - `procedures.duration`: no binding min and no service rule proven — do not
    invent one.
- **Scope:** EMR consumer parity decision per rule (EMR-376): add matching
  preflight checks, or document the API-narrower-than-import behavior as accepted.
- **Acceptance:** per rule, either an importer parity check exists or a documented
  consumer decision records the import-only value domain; per-clinic occurrence
  counts come from Lead's count run, not this audit.
- **Evidence:** `owner/validators.go:31-87`, `validators_contact.go:12-16;30-51`,
  `http_request.go:187-257`, `procedure_request.go:13;42`,
  `merchandise_item_request.go:16;19;21;39;42;44`, `pet_request.go:120;192`,
  `estimate_request.go:69-70`, `estimate_service.go:114-161`,
  `sharedkernel.ValidateTimeRange:116-120`, `reservation_request.go:149;261`,
  `trimming_request.go`.

### T-K — residual unknowns (named only; EMR-376 owner investigation)

- Not a defect; one residual remains after the catalog fill + re-review:
  `payments.insurance_ratio` — bound (`accounting_request.go:209;341`) and the
  builder assigns the value directly (`accounting_service_builders.go:163-164`),
  so no binding/assignment range exists; any *other* service-side rule was not
  re-verified and is not claimed.
- Resolved out of this bucket this pass: `payments.clinic_id` (later DDL adds the
  column — finding §2.11), `owners.company_phone`/`home_postal_code` (service
  format confirmed NOT applied — `validators.go:31-87`),
  `procedures.duration` (no binding min; no service rule invented).
- **Acceptance:** the residual resolves to a cited source line or a recorded
  owner decision.

## 5. Uncertainty register

Matrix `UNK` cells (1) are named in T-K (`payments.insurance_ratio` broader
service rule). Previously-unknown bindings resolved via
[api-binding-catalog.json](api-binding-catalog.json) + cited source: `pets.pet_number` (PATCH
max=50), inquiry fields (only `chief_complaint`/`notes` bound), all three
clinical-plan texts (bound), `vital_records.recorded_at` (required create,
optional PATCH) + `staff_id` (bound, assignment-validated),
`billing_items.is_insurance_applicable` (bound create/update),
`payments.insurance_name` (max 255), `vaccines.species` (`oneof=dog cat both`),
all vaccination `supplemental`/`lot1-4`/`remarks` (bound create/update),
`estimate_items.category/discount_rate/discount_amount/is_insurance_applicable`
(bound + service-validated).

## 6. Appendix — proposed `medicines` as table 22 (EMR-342/EMR-343)

Separate from the 21-table baseline. Design source of truth:
[../emr-342-medicines-contract-design/DESIGN.md](../emr-342-medicines-contract-design/DESIGN.md)
(do not fork this document). Proposed pin semantics: the bundle pin is the FULL
producer file SHA, not a header hash; the proposed-22 producer-file SHA is
UNKNOWN until a producer artifact exists.

- Proposed `medicines.csv` columns (15): `id, clinic_id, name, price, is_active,
  description, parent_id, dosage_form, medicine_unit, inventory_id, default_quantity,
  tax_type, tax_rate, sort_order, is_non_insurance` — see DESIGN §3.2.
- ID-band columns: `id`, `parent_id`; required strings: `name`, `description`.
- `estimate_items.medicine_id` flips from `cutoverUnavailableParent` (must-be-empty)
  to `cutoverCSVParent` once the table lands (`cutover_reference_graph.go:54`).
- Only `medicines.inventory_id` is empty-only (`must-be-empty`, same rule as
  `vaccines.inventory_id`); `medicine_unit` is NOT NULL-only — for agreed
  column-level rules see DESIGN §3.2/§3.3 rather than this audit.
- Contract hash/table count change: 21→22; old 21-table bundles rejected by contract
  check (current 4-clinic bundles are `REHEARSAL_ONLY`; DESIGN §3.7/V6).
- Producer coordination required with MIG-34/MIG-25 (stage output + column mirror +
  hash sync); price sign handling inherits EMR-345's resolution; `price`/`tax_type`
  mapping pending MIG-25 ②-1; decimal-precision questions pending EMR-365 — none of
  these are promoted to approved requirements by this audit.
- Dosage-form allow-list per DESIGN §3.2/§3.3 is a new enum rule the same T-F
  pattern applies to if adopted.

## 7. Aggregate counts — Lead scan completed

**No actual source CSV was read by this agent.** Codex Lead completed a
read-only, network-none Docker count run covering the 4 clinics, 21 files, 236
columns, with same-open SHA + manifest row binding. The completed scan is a
**283-rule bounded survey**: the initial 253-rule snapshot plus 23 initial-DDL
enum rules and 7 numeric overflow-after-PG-round rules. Lead-produced durable
artifacts live in this directory:

- [aggregate-counts.json](aggregate-counts.json) — per-clinic per-rule
  aggregates
- [api-binding-catalog.json](api-binding-catalog.json) — DTO binding
  extraction used in §1 (optional provenance, not sole authority — claims cite
  repo source)
- [COUNT-SUMMARY.md](COUNT-SUMMARY.md) — human-readable run summary

Key observations from the completed scan:

- The 30 selected initial-DDL enum/numeric-overflow rules returned zero hits.
  Separately, the listed partial-unique predicates and relational-pairing
  counters returned zero in this frozen cohort. These are different rule
  families, not a combined 30-rule inventory, and do not establish all DDL
  floors or API acceptance.
- The DDL identity SHA recorded after the survey's copy check matches the
  baseline DDL — the surveyed schema is the audited schema.

Scope of what the counts prove and do not prove:

- Per-rule counts are scoped to the rule's own condition: DTO `oneof`/`min`/
  `max` checks apply only if the row were submitted through that DTO
  (create vs update vs trimming-scoped DTOs differ — e.g. reservation DTO 7
  labels vs trimming DTO 6).
- Duplicate-rule counts are counts of rows sharing a key, NOT counts of
  distinct records.
- Required-string/empty-name rules are not equivalent to DB `NOT NULL` or CSV
  `ForceNotNull` semantics.
- Terminal-state counts (finalized/completed/cancelled) are informational —
  not violations by themselves.
- Unsupported / still UNKNOWN: general date/timestamp formats, text/COPY
  nullability edge cases (e.g. merchandise empty-string vs NULL vs NOT NULL),
  later-ALTER checks, API recompute behavior, permissions, correction flows,
  and any API save/capability proof — none of which a count run can show.
- No "zero violations on all floors" or "import PASS" claim is made or implied.
- Run/build IDs in the artifacts are technical provenance only; no patient IDs,
  names, or cell values appear.

The earlier task-local draft `evidence/audit-counts.py` is **superseded and
unaccepted** — it remains in place untouched as history only; it verified
nothing (known deficiencies: float parsing, missing phone/format/range rules,
hash-vs-read mismatch). Nothing in this audit depends on its output.

## 8. Privacy

This document and the matrix contain only table/column/rule coordinates. No CSV cell
values, names, or identifiers appear anywhere.
