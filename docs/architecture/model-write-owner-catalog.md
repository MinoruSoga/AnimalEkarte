# Model write-owner catalog

> **Purpose**: Map major GORM types that live in `backend/internal/model` to their **write owner** domain package. Types may stay shared in `model` for import graph reasons; ownership is about **who may mutate the business fact**, not where the struct file sits.
> **Scope**: Documentation + review rules only. **No bulk move** of `internal/model` into domains (ARCH-A2 / ADR-006).
> **Related**: [ADR-006](adr/006-backend-domain-package-boundaries.md), [boundary map](be9-2a-boundary-map.md), [cross-domain orchestration catalog](cross-domain-orchestration-catalog.md), `appointment_write_owner_lint_test.go`.

## Rules (new types and PRs)

1. **One business fact → one write owner**  
   Name the owner package in the PR description when adding or changing a persisted fact.

2. **Where to put the Go type**  
   - Prefer owner domain only when the type is truly private and moving it does not create import cycles.  
   - Keeping the type in `internal/model` is fine; then **document the write owner here** (or update this catalog in the same PR).

3. **Command / DTO**  
   Domain-specific request/response/command types stay in the owner domain package (existing pattern). Do not grow `model` with HTTP DTOs.

4. **No “model-only” facts**  
   A new GORM row type without an owner package that implements create/update/delete (or an explicit orchestration path) is a review reject.

5. **Do not clone shared IDs / enums**  
   Reuse existing `model` enums and ID types. Split only when a real consumer boundary appears (ARCH-A8).

6. **No bulk re-home Issue**  
   Do not open a “move all of model into domains” epic. Touch-path gradual ownership only.

## Catalog table

| Fact / GORM type(s) | Typical table(s) | Write owner package | Notes / gate |
|---|---|---|---|
| `Reservation` | `appointments` | `reservation` | Sole appointment lifecycle owner; AST gate `appointment_write_owner_lint_test.go` |
| `ReservationType`, slots / occupations / unavailable | `reservation_types`, related | `reservation` | Master for booking categories |
| `LineReservationSetting` | `line_reservation_settings` | `reservation` | LIFF-facing booking settings owned with reservation |
| `AppointmentTrimmingDetail`, `AppointmentTrimmingOption` | `appointment_trimming_details`, `appointment_trimming_options` | `trimming` (+ `reservation` intents) | Detail graph written in same tx via `CreateForTrimming` / `UpdateForTrimming` / `DeleteForTrimming`; appointment row still `reservation` |
| `TrimmingCourse`, `TrimmingOption`, `TrimmingCourseType` | `trimming_courses`, `trimming_options`, `trimming_course_types` | `trimming` | Trimming masters |
| `Staff`, `ShiftEntry`, `ShiftTemplate`, breaks | `staffs`, `shift_entries`, `shift_templates`, … | `staff` | Staff/shift write owner (not reservation) |
| `StaffReservationCapability`, `StaffReservationExclusion` | capability / exclusion tables | `staff` | Capability data; reservation consumes via interface |
| `StaffClinicAssignment` | `staff_clinic_assignments` | `staff` | Membership assignment |
| `Occupation` | `occupations` | `staff` | Staff-related master |
| `Account`, `TokenBlacklist`, `PasswordResetToken` | `accounts`, `token_blacklist`, … | `auth` | Identity / session |
| `PermissionGroup`, `PermissionGroupRule`, `StaffPermissionGroup` | permission tables | `auth` | Clinic-scoped RBAC definitions |
| `Owner` | `owners` | `owner` | |
| `Pet`, `PetOwner`, `PetChronicCondition` | `pets`, `pet_owners`, `pet_chronic_conditions` | `pet` | |
| `AnimalSpecies` | `animal_species` | `pet` (or clinic master via pet flows) | Shared species master; writes via pet/clinic master paths — treat `pet` as primary app owner unless a dedicated master package is introduced |
| `MedicalRecord`, addenda, images | `medical_records`, … | `medicalrecord` | Clinical record aggregate |
| `Checkup`, checkup types/fields/results | `checkups`, … | `medicalrecord` | |
| `Examination`, `ExamResult`, exam types/fields/ranges | `exams`, `exam_results`, … | `medicalrecord` | |
| `Prescription`, `Vaccination`, `Vaccine`, `VitalRecord` | clinical child tables | `medicalrecord` | |
| `Treatment`, `ClinicalPlan`, `Procedure`, `Medicine`, dose params | treatment masters + rows | `medicalrecord` | |
| `Hospitalization`, daily/care/treatment plan rows, `Cage`, plans | hospitalization tables | `medicalrecord` | Discharge+billing orchestrates billing in same tx — see orchestration catalog |
| `Inquiry`, `InquiryTemplate`, diagnosis masters, chief complaint | inquiry / diagnosis tables | `medicalrecord` | |
| `LabImport*`, `LabExamReport*` | lab import/report tables | `medicalrecord` | Saga-style import; still medicalrecord owner |
| `LabDeviceWait`, `LabDeviceItemMaster`, `LabDeviceStationSettings`, `LabImportJobItem` | `lab_device_waits`, `lab_device_item_masters`, `lab_device_station_settings`, `lab_import_job_items` | `medicalrecord` | ADR-007。device 受信は `job_id`+`pet_id`。fixture Commit に載せない |
| `Billing`, `BillingItem`, `Payment`, `PaymentSplit` | `billings`, … | `billing` | |
| `Estimate`, `EstimateItem` | `estimates`, … | `billing` | |
| `BillingConfirmation`, `BillingRefund` | confirmation / refund tables | `billing` | |
| `Campaign*`, `CashRegisterClose*`, `PaymentMethodMaster`, `Insurance` | billing-adjacent | `billing` | |
| `InventoryItem`, `MerchandiseItem` | inventory tables | `inventory` | Co-tx with medicine from medicalrecord uses inventory owner APIs/tx scope |
| `Clinic`, `ClinicSettings`, `ClinicHoliday`, `ClosingSpecialPeriod`, `ClinicIntegration` | clinic tables | `clinic` | |
| `Company` | `companies` | `clinic` | Multi-clinic org if used |
| `LineCustomer`, `LineLinkToken`, `LineSendLog`, `SharedFile` | LINE tables | `lstep` | |
| `Lstep*` settings/tags/csv/delivery counters | lstep tables | `lstep` | |
| `OwnerIdentityGroup*`, `PetIdentityGroup*` | identity group tables | `identitylink` | No Go import of owner/pet packages |
| `ManualArticle`, `ManualArticleVersion` | manual tables | `manualarticle` | |
| `AuditLog` | `audit_logs` | `audit` (cross-cutting) | Prefer `LogEntryTx` with ambient tx when integrity requires fail-closed |

### Explicit non-owners (do not write these tables from other domains)

| Fact | Forbidden | Use instead |
|---|---|---|
| `appointments` / `Reservation` | any package other than `reservation` | typed reservation intents (`CreateForTrimming`, `CompleteForAccounting`, `BackfillForMedicalRecord`, …) |
| `staffs` / `shift_entries` | `reservation` direct mutation | staff write APIs / consumer-side writer injected into reservation |
| `billings` graph | medicalrecord independent billing repos | `DischargeWithBilling` / billing owner services in ambient tx |
| lstep tag sync tables | owner/pet/billing/medicalrecord concrete `lstep` import | consumer-side `TagSyncNotifier` (or equivalent) |

## Review checklist (copy into PR when touching model)

- [ ] Write owner package named  
- [ ] This catalog updated if a **new** major fact or owner change  
- [ ] No new generic map-based mutator for owner-owned tables exposed outside owner  
- [ ] No duplicated enum/ID type for an existing shared concept  
- [ ] Cross-domain write appears in [cross-domain-orchestration-catalog](cross-domain-orchestration-catalog.md) if multi-package  
