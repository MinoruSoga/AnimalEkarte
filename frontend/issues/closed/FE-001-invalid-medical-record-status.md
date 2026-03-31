# FE-001: Invalid Medical Record Status Value

## Issue
Medical record form sends invalid status value `"作成中"` when creating a new record, but backend only accepts `"draft"` or `"finalized"`.

## Error
```
{"error":"invalid status: invalid value \"作成中\""}
```

## Root Cause
File: `src/features/medical-records/hooks/useMedicalRecordForm.ts` (line 87)
```typescript
const req: CreateMedicalRecordRequest = {
  status: "作成中",  // ❌ Invalid - Backend rejects this
  ...
};
```

## Solution
Change status value to valid backend constant:
- `"draft"` (MedicalRecordStatusDraft) - for new/draft records
- `"finalized"` (MedicalRecordStatusFinalized) - for completed records

For new record creation, default should be `"draft"`.

## Backend Definitions
File: `internal/model/medical_record.go`
```go
const (
  MedicalRecordStatusDraft     MedicalRecordStatus = "draft"
  MedicalRecordStatusFinalized MedicalRecordStatus = "finalized"
)
```

## Files to Update
- `src/features/medical-records/hooks/useMedicalRecordForm.ts` (lines 87, 110)

## Testing
After fix, verify medical record creation/update operations work without 400 errors.
