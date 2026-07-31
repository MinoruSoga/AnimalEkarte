# TASK-009 — 003_demo clinical seed design (design only)

> **Date**: 2026-07-31  
> **Residual**: SCEN-SEED-001  
> **Status**: Design document only  
> **Agent actions**: **Did not** write CSV data rows, **did not** run `make migrate`, **did not** apply seeds.

---

## 1. Problem

`backend/migrations/seeds/003_demo/` keeps many **clinical / transactional** tables as **header-only CSV** (schema shape only). Scenario packs assume either:

- in-scenario create (e.g. S07 estimates), or  
- pre-seeded clinical graph (hospitalization / medical record / billing chains).

After a fresh seed apply, list/search demos that depend on clinical rows fail or force every scenario to hand-build data.

Full historical dump (~529MB, possible PHI) is **local-only** and must not be committed (`docs/ops/deploy/ANIMALEKARTE_CSV_IMPORT_COMPLETION.md`). This design targets **minimal synthetic clinic-scoped fixtures**, not paper/Excel fidelity.

---

## 2. Inventory — 29 header-only clinical / transactional CSVs

Verified 2026-07-31 by reading first lines under `backend/migrations/seeds/003_demo/` (header present, no data rows):

| # | CSV | Role in demo graph |
|:--|:---|:---|
| 1 | `appointments.csv` | Reception / reservation lifecycle |
| 2 | `appointment_trimming_options.csv` | Trimming appointment options |
| 3 | `billing_confirmations.csv` | Billing confirmation workflow |
| 4 | `billing_items.csv` | Billing line items |
| 5 | `billing_refunds.csv` | Refunds |
| 6 | `billings.csv` | Billing headers |
| 7 | `care_logs.csv` | Hospitalization care execution |
| 8 | `care_plan_items.csv` | Hospitalization care plan |
| 9 | `checkups.csv` | Checkup instances |
| 10 | `clinical_plans.csv` | Medical-record clinical plan |
| 11 | `daily_records.csv` | Hospitalization daily sheets |
| 12 | `exam_results.csv` | Exam result lines |
| 13 | `exams.csv` | Exam orders |
| 14 | `estimate_items.csv` | Estimate lines |
| 15 | `estimates.csv` | Estimates |
| 16 | `hospitalizations.csv` | Hospitalization / hotel |
| 17 | `inquiries.csv` | Record inquiry / chief complaint block |
| 18 | `medical_record_addenda.csv` | Finalized-record addenda |
| 19 | `medical_record_images.csv` | Record images |
| 20 | `medical_records.csv` | Medical records core |
| 21 | `payments.csv` | Payments |
| 22 | `payment_splits.csv` | Payment splits |
| 23 | `pet_chronic_conditions.csv` | Pet chronic conditions |
| 24 | `prescriptions.csv` | Prescriptions |
| 25 | `staff_notes.csv` | Daily-record staff notes |
| 26 | `treatment_plans.csv` | Treatment plans (record + hospitalization) |
| 27 | `treatments.csv` | Treatment rows on record |
| 28 | `vaccinations.csv` | Vaccination events |
| 29 | `vital_records.csv` | Vitals |

### Related files that **already have** demo rows (not in the 29)

Examples (not exhaustive): `clinics.csv`, `owners.csv`, `pets.csv`, masters (`medicines`, `vaccines`, `cages`, …), `payment_methods.csv`, `cash_register_closes.csv`, `audit_logs.csv`, `line_customers.csv`, `line_send_logs.csv`, `shared_files.csv`, `appointments` is header-only (listed above).

**Do not treat master/config CSVs as “fill all 29”.** Prefer a thin clinical chain.

---

## 3. Design principles

1. **Minimal rows** — enough for S05/S08-style demos and board visibility; not a full year of ops.  
2. **Clinic-scoped IDs** — use existing demo clinics (`clinics.csv`: id `1` 八王子, `2` 城東, `3` …). Prefer **clinic_id=1** for the primary chain unless a scenario needs multi-clinic isolation proof.  
3. **Reuse existing owners/pets/staff** — e.g. owner `1` / pets already seeded; doctors from `staffs.csv` / assignments. **Do not invent orphan FKs.**  
4. **FK audit before any CSV write** — next implementer must re-read live `owners.csv` / `pets.csv` / `staffs.csv` / `cages.csv` / master IDs before inserting rows. This design intentionally avoids inventing numeric FKs without that audit.  
5. **No PHI** — synthetic names only; no production dump import into git.  
6. **Agent must not apply** — USER runs migrate/seed.  
7. **Optional CSV write cap** — if implementation chooses to fill CSVs in-repo, touch **≤ 8 files** in the first slice (see §5). Prefer design-doc-only until FK audit is done (this packet is design-doc-only).

---

## 4. Proposed minimal clinical graph (logical)

Primary clinic: **clinic_id = 1**.

| Step | Entity | Count (target) | Notes |
|:---|:---|:---|:---|
| G1 | `medical_records` | 1–2 | One `draft`, optional one `finalized` for addenda demos |
| G2 | `treatment_plans` and/or `treatments` | 1–3 rows | Tied to G1; row discounts demo |
| G3 | `hospitalizations` | 1 active (+ optional 1 discharged) | Valid `cage_id`, `doctor_id`, owner/pet |
| G4 | `treatment_plans` (hospitalization) | 1–2 | Nested under G3 |
| G5 | `daily_records` + optional `care_plan_items` / `vital_records` | 1 day sheet | Unlocks detail tabs |
| G6 | `billings` + `billing_items` | 0–1 unpaid or completed | Only if S08/S10 need pre-seed; else scenario-create remains OK |
| G7 | `appointments` | 0–2 same-day | Reception board smoke; status set carefully |

**Explicitly defer (leave header-only until a scenario fails without them):**

- `estimates` / `estimate_items` — S07 creates in-scenario  
- `exam_results` / large exam matrices — size + complexity  
- `payments` / `payment_splits` — require completed billing graph + system_key methods  
- `medical_record_images` — binary/object storage  
- `billing_refunds`, `billing_confirmations`  
- `pet_chronic_conditions`, `prescriptions`, `vaccinations`, `checkups`, `inquiries`, `clinical_plans`, `staff_notes`, `appointment_trimming_options` — add only when a named scenario residual requires them

---

## 5. First implementation slice (≤ 8 files if writing CSVs)

Recommended order when USER approves data authoring:

1. `medical_records.csv`  
2. `hospitalizations.csv`  
3. `treatment_plans.csv` (record and/or hospitalization parents)  
4. `daily_records.csv`  
5. `care_plan_items.csv` (optional)  
6. `vital_records.csv` (optional)  
7. `billings.csv` (optional)  
8. `billing_items.csv` (optional; same slice as billings)

Stop at the smallest set that unblocks the highest-value scenarios. Do **not** bulk-fill all 29 in one PR.

---

## 6. Consistency / safety checklist (pre-write)

- [ ] `clinic_id` matches parent owner/pet/staff  
- [ ] Soft-delete columns empty for active demo rows  
- [ ] Status enums match model (`hospitalizations.status`, `medical_records.status`, billing status)  
- [ ] No cross-clinic FKs  
- [ ] `manifest.json` already lists tables — no reorder without seedloader rules  
- [ ] checksum / seed bundle process per `migration-seed-safety` skill and seed README  
- [ ] Avoid real phone/email beyond existing synthetic demo data  
- [ ] Foreign WIP: **do not edit** `line_reservation_settings.csv` in parallel agent work

---

## 7. USER apply steps (when ready)

Agents **must not** run these automatically. When design is approved and CSV rows (if any) are committed:

```bash
# From repo root, with Docker stack available:
# 1) Ensure backend/db containers are up (USER-managed):
docker compose ps

# 2) Apply migrations + seed bundles (project Makefile target):
make migrate

# If your environment uses DB reset for full re-seed, that is a USER-owned
# high side-effect path — do not agent-auto-run DB_RESET / make db.
```

After apply:

1. Log in as demo staff for clinic 1.  
2. Smoke: medical records list shows G1; hospitalization board shows G3.  
3. Re-run scenario prerequisites that previously noted “header-only clinical CSV”.  
4. Record results under `reports/YYYY-MM-DD-<env>.md` (not in scenario body).

---

## 8. Acceptance mapping (from todo TASK-009)

| Acceptance | How this design meets it |
|:---|:---|
| ① Target clinical CSVs no longer header-only (for the **chosen slice**) | §5 first slice; not all 29 at once |
| ② S-series findable on 003_demo | Minimal graph G1–G5 |
| ③ Apply procedure in one place | §7 + this report path |
| ④ Apply is USER manual | Explicit; agent did not apply |

---

## 9. Out of scope / non-actions this packet

- No CSV data rows written  
- No `make migrate` / seed apply  
- No product code  
- No PHI dump import  
- No inventing large master ID sets without audit  

**Next owner**: implementer with FK audit → optional ≤8 CSV files → USER `make migrate` → TASK-010 browser measurements that depend on seed.
