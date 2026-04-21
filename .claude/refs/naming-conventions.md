---
description: Naming conventions for DB / Go / API (tables, columns, enums, endpoints)
alwaysApply: false
globs: ["backend/migrations/**", "backend/internal/model/**", "backend/internal/repository/**", "backend/internal/handler/**", "backend/internal/service/**"]
---

# Naming Conventions — DB / Go / API

Project-wide naming conventions. Ensures consistency across tables, columns, enums, Go structs, and API endpoints.

---

## 1. Table Names

### Mandatory Rules

| Rule | Example | Prohibited |
|------|---------|-----------|
| **Always plural** | `companies`, `clinics`, `owners` | `company` |
| **snake_case** | `medical_records`, `billing_items` | `medicalRecords` |
| **No synonym duplication** | `appointments` (booking) | `reservation_appointments` |
| **No negation in names** | `staff_reservation_exclusions` | `staff_excluded_reservation_categories` |
| **Max 30 chars (guideline)** | `staff_reservation_exclusions` (30) | `staff_excluded_reservation_categories` (38) |

### Prefix Rules

| Condition | Prefix | Example |
|-----------|--------|---------|
| LINE-exclusive table | `line_` | `line_customers`, `line_reservation_settings` |
| Junction table | `{parent}_{child}` | `staff_clinic_assignments`, `staff_permission_groups` |
| Child table (line items) | `{parent}_items` | `billing_items`, `estimate_items` |

### Classification Master Suffix

**Use `_types` consistently. Never use `_categories`.**

```sql
-- ✅ Consistent
exam_types, checkup_types, reservation_types, diagnosis_types, chief_complaint_types

-- ❌ Prohibited: Mixed _categories
reservation_categories, diagnosis_categories, chief_complaint_categories
```

Exception: Use `_groups` for grouping tables.
```sql
reservation_type_groups  -- grouping of reservation types
```

### Table → Go struct → API path Correspondence

Struct name is singular PascalCase from table name.

| Table | Go struct | API path |
|-------|-----------|----------|
| `appointments` | `Appointment` | `/reservations` |
| `medical_records` | `MedicalRecord` | `/medical-records` |
| `billing_items` | `BillingItem` | `/billings/{id}/items` |
| `line_customers` | `LineCustomer` | `/line-customers` |
| `exam_results` | `ExamResult` | (exams child resource) |
| `billing_confirmations` | `BillingConfirmation` | (medical-records child resource) |
| `reservation_types` | `ReservationType` | `/reservation-types` |
| `care_logs` | `CareLog` | (daily-records child resource) |

---

## 2. Column Names

### Mandatory Rules

| Rule | Correct | Incorrect |
|------|---------|-----------|
| **Don't repeat table name** | `owners.name` | `owners.owner_name` |
| **Boolean: `is_`/`has_`/`can_` prefix** | `is_active`, `is_selected`, `has_insurance` | `active`, `selected`, `insurance` |
| **No abbreviations** | `body_weight`, `body_temperature`, `reference_value` | `bw`, `bt`, `ref` |
| **Plural for text columns** | `notes` | `note` |

### Text Column Naming (by use case)

| Use Case | Column | Context |
|----------|--------|---------|
| Master description | `description` | `consultations.description`, `procedures.description` |
| Operational free-text | `memo` | `hospitalizations.memo`, `billings.memo` |
| Notes/remarks | `notes` | `appointments.notes`, `vital_records.notes` (existing) |
| Owner/pet remarks | `remarks` | `owners.remarks`, `pets.remarks` (existing) |

New tables: use `memo`. Existing `notes`/`remarks`: maintain (low ROI to change).

### Foreign Key Column Names

| Reference | Role | Column |
|-----------|------|--------|
| `staffs` | Attending physician | `doctor_id` (when "doctor" is clear in domain) |
| `staffs` | Operator/recorder | `staff_id` |
| `staffs` | Approver | `confirmed_by` (allowed exception) |
| `staffs` | Creator | `created_by` (allowed exception) |
| `staffs` | Rejecter | `returned_by` (allowed exception) |

`confirmed_by` etc. are self-explanatory (context is FK to `staffs`), so `_staff_id` suffix can be omitted.

### Date/Time Columns

| Use | Type | Pattern | Example |
|-----|------|---------|---------|
| System management | `timestamptz` | `created_at`, `updated_at`, `deleted_at` | All tables |
| Operation timestamp | `timestamptz` | `{verb}_at` | `confirmed_at`, `refunded_at`, `completed_at` |
| Business date | `date` | `date` (self-evident in context) | `medical_records.date`, `vaccinations.date` |
| Period | `date` | `start_date`, `end_date` | `hospitalizations.start_date` |
| Time of day | `time` | `start_time`, `end_time` | `appointments.start_time` |

### Amount Columns

| Use | Column | Table |
|-----|--------|-------|
| Master unit price | `price` | `exam_types`, `vaccines`, `medicines` etc. |
| Line item unit price | `unit_price` | `billing_items`, `estimate_items`, `treatments` |
| Subtotal | `subtotal` | `billings`, `estimates`, `payments` |
| Total | `total_amount` | `billings`, `estimates` |
| Tax | `tax_total` | `billings`, `estimates` |
| Discount amount | `discount_amount` | `billing_items`, `estimate_items` |
| Discount rate | `discount_rate` | `billing_items`, `estimate_items`, `owners` |

---

## 3. ENUM Type Names

### Naming Pattern

```
{table singular}_{column name}
```

| Table | Column | ENUM Name | Notes |
|-------|--------|-----------|-------|
| `appointments` | `status` | `reservation_status` | Use domain concept |
| `exams` | `status` | `exam_status` | — |
| `exam_results` | `status` | `exam_result_status` | — |
| `billings` | `status` | `billing_status` | — |
| `billing_confirmations` | `status` | `confirmation_status` | — |
| `pets` | `gender` | `pet_gender` | — |
| `medicines` | `dosage_form` | `dosage_form` | Shared concept, no prefix needed |

**Prohibited**: Don't make ENUM type name identical to table name.

```sql
-- ❌ Unclear in SQL
cages.cage_type cage_type

-- ✅ Prefix with table singular
cages.cage_type cage_type  -- Table cages ≠ ENUM name cage_type
```

### ENUM Values

- All **snake_case**
- Use affirmative (not negative)
- Order by time sequence

```sql
CREATE TYPE exam_status AS ENUM (
    'pending',        -- Not yet done
    'in_progress',    -- In progress
    'result_entered', -- Result entered
    'completed',      -- Completed
    'confirmed'       -- Confirmed
);
```

---

## 4. Go Layer Naming

### Model (struct)

Singular PascalCase from table name.

| DB Table | Go struct | Go File |
|----------|-----------|---------|
| `appointments` | `Appointment` | `reservation.go` |
| `medical_records` | `MedicalRecord` | `medical_record.go` |
| `line_customers` | `LineCustomer` | `line_customer.go` |
| `exam_results` | `ExamResult` | `examination_record.go` |
| `billing_confirmations` | `BillingConfirmation` | `billing_confirmation.go` |
| `reservation_types` | `ReservationType` | `reservation_type.go` |
| `care_logs` | `CareLog` | `hospitalization.go` |
| `staff_notes` | `StaffNote` | `hospitalization.go` |

### ENUM Constants

```go
type ExaminationStatus string  // SQL: exam_status

const (
    ExaminationStatusPending       ExaminationStatus = "pending"
    ExaminationStatusInProgress    ExaminationStatus = "in_progress"
    ExaminationStatusResultEntered ExaminationStatus = "result_entered"
    ExaminationStatusCompleted     ExaminationStatus = "completed"
    ExaminationStatusConfirmed     ExaminationStatus = "confirmed"
)
```

Pattern: `{GoStructName}{PascalCaseValue}`. Go type and SQL ENUM names can differ (GORM tags bridge them).

### Repository / Service / Handler

| Layer | Pattern | Example |
|-------|---------|---------|
| Repository interface | `{Entity}Repository` | `AppointmentRepository` |
| Repository impl | `appointment_repository.go` | — |
| Service interface | `{Entity}Service` | `AppointmentService` |
| Service impl | `appointment_service.go` | — |
| Handler | `{Entity}Handler` | `AppointmentHandler` |
| Handler impl | `appointment_handler.go` | — |
| Routes | `appointment_routes.go` | — |

**Principle: 1 entity = 1 file set**. Exception: When same table is exposed via different APIs (admin `/v1/masters/` vs. LINE LIFF `/api/clinics/`), separate by purpose is OK, but **API path must match table name**.

```
# Same reservation_types table:
/v1/masters/reservation-types          ← admin (reservation_type_handler.go)
/api/clinics/{id}/reservation-types    ← LINE (reservation_type_liff_handler.go)
```

---

## 5. API Endpoints

### Path Naming

```
/api/clinics/{clinicId}/{resource}           -- List, Create
/api/clinics/{clinicId}/{resource}/{id}      -- Get, Update, Delete
/api/{parent-resource}/{parentId}/{child}    -- Child resources
```

| Rule | Correct | Incorrect |
|------|---------|-----------|
| **kebab-case** | `/medical-records` | `/medicalRecords` |
| **Plural** | `/appointments` | `/appointment` |
| **1:1 with table** | `/appointments` (← `appointments` table) | `/reservation-appointments` |
| **LINE resources use `line-` prefix** | `/line-customers` | `/reservation-customers` |

---

## Checklist (Adding New Table)

- [ ] Table: plural snake_case, ≤30 chars
- [ ] Table: no synonym duplication
- [ ] Table: add `line_` if LINE-exclusive
- [ ] Table: append `_types` for classification masters
- [ ] Columns: don't repeat table name
- [ ] Columns: Boolean has `is_`/`has_`/`can_` prefix
- [ ] Columns: use `description` (master) or `memo` (operational)
- [ ] FK: column name clarity from reference and role
- [ ] ENUM type: `{table singular}_{column name}` pattern
- [ ] ENUM values: snake_case, affirmative, time-ordered
- [ ] Go struct: table singular PascalCase
- [ ] API path: 1:1 with table, kebab-case
