# Cross-Domain Orchestration Catalog

> **Purpose**: List AnimalEkarte cross-domain write/orchestration paths with explicit same-tx vs separate-tx contracts, fail-closed vs best-effort labels, failure recovery, audit participation, and test anchors.
> **Scope**: Documentation only. Does not rewrite domain packages, revive Clean Architecture/layer-first layout, or change reservation appointment write ownership (ADR-006).
> **Related**: [ADR-006](adr/006-backend-domain-package-boundaries.md), [Architecture Overview](overview.md), [Go/Gin Backend Guidelines](../../.claude/rules/go-gin-backend-guidelines.md).

## Catalog table

| Path ID | Initiator | Owner operation | Transaction boundary | Fail-closed vs best-effort | Failure recovery | Audit participation | Test anchors |
|---|---|---|---|---|---|---|---|
| PATH-RES-MR-AUTOCREATE | `reservation` Create/Update handler after `confirmed`/`checked_in` (`ReservationHandler.CreateReservation` / `UpdateReservation`) | `medicalrecord.AutoCreateFromReservation` | **separate-tx** (post-commit call; not joined to reservation write) | **best-effort** (intentional; reservation success is not rolled back) | Skip or retry on next confirmed/checked_in write; same-day pet advisory lock serializes auto-create; subrecords failure leaves durable audit | Main create audit runs only on successful create path; subrecords failure → `medical_record.subrecords_failed` (best-effort audit) | `backend/internal/reservation/reservation_handler.go` (`shouldAutoCreateMedicalRecordForReservation`); `backend/internal/medicalrecord/medical_record_auto_create.go` (`AutoCreateFromReservation`) |
| PATH-RES-MR-CANCEL-CLEANUP | `reservation` Update when status → `cancelled` | `medicalrecord.DeleteDraftFromReservation` | **separate-tx** (detached from request ambient tx; timeout-bounded) | **best-effort** (intentional; cancel already committed) | Failure categories (dependency/state/internal) audited; draft-only delete via normal `Delete` path; non-draft left intact | Failure → `reservation.draft_cleanup_failed` with category metadata; success uses normal medical-record delete audit | `backend/internal/reservation/reservation_handler.go` (`UpdateReservation`); `backend/internal/medicalrecord/medical_record_auto_create.go` (`DeleteDraftFromReservation`) |
| PATH-MR-RES-BACKFILL | `medicalrecord` appointment-linked Create | `reservation.BackfillForMedicalRecord` | **same-tx** (ambient transaction required; medicalrecord opens/owns tx) | **fail-closed** (missing ambient tx, lock, or backfill error aborts create) | Retry whole create; no partial appointment backfill without medical-record commit | Medical-record create audit on success path; reservation intent has no independent partial commit | ADR-006 appointments intent table; `backend/internal/medicalrecord/service_deps.go` (`mrReservationRepo.BackfillForMedicalRecord`) |
| PATH-MR-RES-PREPARE-FINALIZE | `medicalrecord` finalized Create/Update | `reservation.PrepareForMedicalRecordFinalization` | **same-tx** (ambient + lifecycle advisory lock → row lock through commit) | **fail-closed** (`no_show`/`cancelled` appointment blocks finalization; lock failure fails write) | Retry finalization under same lock order; serializes with `MarkNoShow` | Finalization audit remains on medicalrecord path; reservation transition refusal is fail-closed inside same tx | ADR-006 `PrepareForMedicalRecordFinalization`; `backend/internal/medicalrecord/service_deps.go` (`mrReservationRepo`) |
| PATH-BILL-RES-COMPLETE | `billing` accounting complete path | `reservation.CompleteForAccounting` | **same-tx** (billing ambient tx required; reservation participates via `dbOrTx`) | **fail-closed** (no ambient tx → error; appointment complete rolls back with billing) | Retry complete operation; reload for response before commit (no error-after-commit invert) | Billing complete audit in billing tx; reservation participates without separate commit | ADR-006 `CompleteForAccounting`; `backend/internal/billing/accounting_repository.go` (Create ambient-tx note); `ReservationIntentRepository.CompleteForAccounting` |
| PATH-HOSP-DISCHARGE-BILL | `medicalrecord` hospitalization discharge | `hospitalizationService.DischargeWithBilling` (+ billing create/items in same call) | **same-tx** (`Transactor.WithTx`; hosp status + optional billing/items) | **fail-closed** (any step error rolls back status/billing/items) | Retry discharge; already-discharged is invalid input | When `CreateAccounting`: `AuditTxLogger.LogEntryTx` with `hospitalization.discharge_with_billing` (or model constant) in same tx — audit failure rolls back business write | `backend/internal/medicalrecord/hospitalization_discharge.go` (`DischargeWithBilling`) |
| PATH-TRIM-APPT-CREATE | `trimming` Create | `reservation.CreateForTrimming` + trimming detail/options | **same-tx** (ambient required; advisory lock → appointment → detail → options) | **fail-closed** (any write failure rolls back whole graph) | Retry create; no orphan appointment without detail when path creates both | `AuditTxLogger` trimming create audit in same tx (fail-closed if audit sink missing) | `backend/internal/trimming/trimming_service.go` (`Create`); `ReservationIntentRepository.CreateForTrimming` |
| PATH-TRIM-APPT-UPDATE (related) | `trimming` Update | `reservation.UpdateForTrimming` + detail/options | **same-tx** (ambient + booking lock + row lock) | **fail-closed** | Retry update under lock order | Trimming update audit in same tx | `trimming_service.go` (`Update`); `UpdateForTrimming` |
| PATH-TRIM-APPT-DELETE (related) | `trimming` Delete | `reservation.DeleteForTrimming` | **same-tx** (ambient + lock + medical-record dependency check) | **fail-closed** (medical-record dependency → conflict) | Resolve dependency or abandon delete | Trimming delete audit in same tx | `trimming_service.go` (`Delete`); `DeleteForTrimming` |

### Intentional best-effort contracts (do not silently upgrade)

The following paths are **documented separate-tx best-effort** contracts. Do **not** reclassify them as same-tx without a product decision, recovery redesign, and test updates:

1. **PATH-RES-MR-AUTOCREATE** — reservation confirmed/received medical-record auto-create runs after reservation commit.
2. **PATH-RES-MR-CANCEL-CLEANUP** — draft medical-record cleanup after cancel runs detached from the cancel transaction.

Both must retain failure audit/observability when side effects fail. Partial success (reservation ok, medical-record side effect missing) is expected under this contract and is recovered by retry on later writes and durable failure audits—not by silent upgrade to same-tx.

## New-path rules

Any **new** cross-domain write or orchestration path MUST follow these rules before merge:

1. **Owner typed intents only**  
   Consumer domains call reservation/billing/medicalrecord (etc.) through **owner typed intents** or a minimal consumer-side interface. Do not add generic field-update APIs or independent write implementations against another domain's tables (especially `appointments` — write owner remains `reservation`).

2. **Ambient-tx participation or explicit orchestration**  
   Multi-domain writes either:
   - participate in the initiator's **ambient transaction** (`WithTx` + `dbOrTx`/txCtx), or
   - use an **explicit orchestration** boundary with a documented owner and commit order.  
   Do not open ad-hoc nested independent transactions that can commit while the outer path rolls back.

3. **Fail-closed by default**  
   Default is **fail-closed**: missing ambient tx for locks, missing required dependencies, validation failure, or audit failure (when audit is part of the integrity contract) aborts the whole business write.

4. **Best-effort only when documented**  
   Best-effort is allowed only when the catalog (or ADR) records:
   - separate-tx boundary,
   - **failure recovery** (retry, compensation, or next-write re-entry),
   - **observability** (structured log + durable audit or equivalent),
   - explicit acceptance of partial success.  
   Undocumented best-effort is forbidden.

5. **No silent partial success**  
   If any part of a multi-write graph can succeed while another fails, the contract must label that outcome (best-effort + recovery). Operators and tests must be able to detect the failure. Never return HTTP success while leaving an unobserved partial write for fail-closed clinical/financial integrity paths.

6. **Catalog update required**  
   Adding a path requires a new row (or stable Path ID) in this catalog with initiator, owner operation, transaction boundary (same-tx vs separate-tx), fail-closed vs best-effort, failure recovery, audit participation, and test anchors.

## Out of scope for this catalog

- Changing production Go behavior or reservation write-owner lint.
- Layer-first / Clean Architecture package revival.
- Long-form task dumps (STATUS.md).
- Secrets, credentials, PII, or PHI examples.
