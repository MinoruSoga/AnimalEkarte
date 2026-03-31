# FE-015: Medical Record Creation API Type Mismatch

**Status**: Blocked by Backend
**Priority**: High
**Blocked by**: BE-015
**Date Created**: 2026-03-16

## Summary
Frontend type definitions have been corrected to match the actual backend API contract, but the medical record creation feature is blocked because the backend and frontend expectations don't align.

## Current Status

### Fixed (Code Convention Compliance)
✅ `frontend/src/features/medical-records/api/types.ts`:
- Refactored `CreateMedicalRecordRequest` and `UpdateMedicalRecordRequest` to use `Omit`/`Partial` patterns derived from `models.ts`
- Removed hand-written interface definitions
- Added documentation explaining why certain fields are excluded

### Identified Problem (Backend API Contract Mismatch)
Frontend code attempts to send these fields in CreateMedicalRecordRequest:
- `pet_id` (as string)
- `owner_id` (as string)
- `visit_date` (custom field name)
- `visit_type` (custom field)
- `chief_complaint`, `chief_complaint_category_id`, `notes` (belong to Inquiry)
- `plan`, `assessment`, `diagnosis_*` (belong to ClinicalPlan)

But the backend's `createMedicalRecordRequest` struct only accepts:
- `record_no` (required, frontend doesn't send)
- `date` (time.Time, frontend sends string as "visit_date")
- `owner_id` (uint64, frontend sends string)
- `pet_id` (uint64, frontend sends string)
- `doctor_id`, `reservation_appointment_id`, `status`

## Why This Matters
1. **Type Safety**: TypeScript will now error on `useMedicalRecordForm.ts` lines 95-110 because those fields don't exist in the corrected type
2. **API Failures**: Even if TypeScript errors are suppressed, the API calls will fail with validation errors from the backend
3. **Data Integrity**: Medical record creation cannot work until backend and frontend contracts align

## Architecture Insight
The backend intentionally separates data:
- **MedicalRecord**: Basic visit info (date, patient, doctor, status)
- **Inquiry**: Chief complaint details (chief_complaint, chief_complaint_category_id, notes)
- **ClinicalPlan**: Diagnosis/treatment planning (diagnosis, assessment)

This is good design (proper separation), but the frontend expects a flattened structure for convenience.

## Required Backend Fix
See `backend/issues/open/BE-015-medical-record-create-api-contract.md` for details.

TL;DR: Backend needs to either:
1. **Option A (Recommended)**: Flatten API to accept all fields and create related entities atomically
2. **Option B**: Require frontend to sequence 3 separate API calls (create MedicalRecord, then Inquiry, then ClinicalPlan)

## Frontend Workaround (If Backend Option B)
Until backend flattens the API, modify `useMedicalRecordForm.ts` to:
1. Create MedicalRecord first with basic fields
2. Create/update Inquiry separately
3. Create/update ClinicalPlan separately

This would require:
- Changing createMutation to only pass basic fields
- Calling inquiry and clinicalPlan mutations sequentially after MedicalRecord is created
- Handling the complex state management of multi-step creation

## Files Modified
- `frontend/src/features/medical-records/api/types.ts` ✅ Fixed type derivation
- `backend/issues/open/BE-015-medical-record-create-api-contract.md` 📋 Filed backend issue

## Next Steps
1. ⏸️  **Pause FE-015**: Medical record creation blocked on backend fix
2. 🔨 **Fix BE-015**: Modify backend's createMedicalRecordRequest to accept flattened fields
3. ✅ **Unblock**: Once backend is fixed, medical record creation will work
4. 📝 **Update useMedicalRecordForm.ts**: Verify it sends correct data after backend fix
5. ✅ **Complete**: Test end-to-end new record creation flow
