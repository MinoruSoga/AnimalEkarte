# BE-015: Medical Record Creation API Contract Mismatch

**Status**: Open
**Priority**: High
**Affects**: Medical Record feature creation flow
**Date Created**: 2026-03-16

## Summary
The backend's createMedicalRecordRequest struct doesn't match what the frontend expects. This blocks the medical record creation feature.

## Current Backend Contract
`backend/internal/handler/medical_record_request.go`:
```go
type createMedicalRecordRequest struct {
    RecordNo                 string    `json:"record_no" binding:"required"`
    Date                     time.Time `json:"date" binding:"required"`
    OwnerID                  *uint64   `json:"owner_id"`
    PetID                    *uint64   `json:"pet_id"`
    DoctorID                 *uint64   `json:"doctor_id"`
    ReservationAppointmentID *uint64   `json:"reservation_appointment_id"`
    Status                   string    `json:"status"`
}
```

## What Frontend is Trying to Send
```typescript
CreateMedicalRecordRequest {
    pet_id: string            // ← Backend expects uint64
    owner_id: string          // ← Backend expects uint64
    visit_date: string        // ← Backend field is "date", not "visit_date"
    visit_type: string        // ← Not in backend struct
    status: string
    chief_complaint?: string        // ← Not in backend struct
    chief_complaint_category_id?: number  // ← Not in backend struct
    plan?: string             // ← Not in backend struct
    assessment?: string       // ← Not in backend struct
    notes?: string            // ← Not in backend struct
    diagnosis_1_category_id?: number  // ← Not in backend struct
    diagnosis_1_name_id?: number      // ← Not in backend struct
    diagnosis_2_category_id?: number  // ← Not in backend struct
    diagnosis_2_name_id?: number      // ← Not in backend struct
}
```

## Issues

1. **Type Mismatch**: Frontend sends `pet_id` and `owner_id` as strings, backend expects uint64
2. **Field Name Mismatch**: Frontend sends `visit_date`, backend expects `date`
3. **Missing Backend Field**: Frontend doesn't send `record_no`, but backend requires it
4. **Missing Extended Fields**: Frontend wants to send chief_complaint, plan, assessment, etc., but backend struct doesn't have them
5. **Architectural Problem**: Inquiry and ClinicalPlan fields are in separate entities, but frontend wants to create them atomically with MedicalRecord

## Root Cause
- MedicalRecord model has only basic fields (RecordNo, Date, OwnerID, PetID, Status)
- Detailed medical information (chief_complaint, plan, assessment) is stored in Inquiry and ClinicalPlan relation entities
- Backend and frontend have different assumptions about the data structure

## Required Changes

### Option A: Flatten API (Recommended)
Modify `createMedicalRecordRequest` to accept flattened fields and handle creation of related Inquiry/ClinicalPlan records atomically:
```go
type createMedicalRecordRequest struct {
    RecordNo                 string    `json:"record_no,omitempty"`  // Make optional, auto-generate if empty
    Date                     string    `json:"visit_date" binding:"required"`  // Accept string, parse to time.Time
    OwnerID                  string    `json:"owner_id"`             // Accept string, parse to uint64
    PetID                    string    `json:"pet_id"`               // Accept string, parse to uint64
    DoctorID                 *string   `json:"doctor_id"`
    ReservationAppointmentID *string   `json:"reservation_appointment_id"`
    Status                   string    `json:"status"`

    // Inquiry fields
    ChiefComplaint           string `json:"chief_complaint,omitempty"`
    ChiefComplaintCategoryID *uint64 `json:"chief_complaint_category_id"`
    Notes                    string `json:"notes,omitempty"`

    // ClinicalPlan fields
    Plan                     string  `json:"plan,omitempty"`
    Assessment               string  `json:"assessment,omitempty"`
    Diagnosis1CategoryID     *uint64 `json:"diagnosis_1_category_id"`
    Diagnosis1NameID        *uint64 `json:"diagnosis_1_name_id"`
    Diagnosis2CategoryID     *uint64 `json:"diagnosis_2_category_id"`
    Diagnosis2NameID        *uint64 `json:"diagnosis_2_name_id"`
}
```

Then in handler, create MedicalRecord + Inquiry + ClinicalPlan atomically (in a transaction).

### Option B: Sequence API Calls (Less Preferred)
Require frontend to:
1. Create MedicalRecord with basic fields (RecordNo, Date, PetID, OwnerID, Status)
2. Create Inquiry separately (POST /v1/medical-records/:id/inquiries)
3. Create ClinicalPlan separately (POST /v1/medical-records/:id/clinical-plans)

This works but requires 3 API calls and has race condition risk.

## Frontend Dependency
The medical record creation feature is blocked on this. Currently:
- `frontend/src/features/medical-records/hooks/useMedicalRecordForm.ts` line 94-110 tries to send fields the backend doesn't accept
- This will cause validation errors when frontend tries to create a new medical record

## Verification
After fixing:
1. Backend should accept CreateMedicalRecordRequest with all flattened fields
2. Inquiry and ClinicalPlan should be created atomically with MedicalRecord
3. Frontend CreateMedicalRecordRequest type should match the backend contract
4. Create a new medical record from the frontend and verify all fields save correctly
