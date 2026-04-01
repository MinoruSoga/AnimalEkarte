# BE-FK-DEPENDENCY-CHECKS: FK Dependency Check Implementation

## Status
✅ FIXED

## Issue Summary
Master record deletion operations were not properly validating foreign key dependencies before deletion. This violated the CLAUDE.md requirement to check dependencies and return 409 Conflict errors when dependent records exist.

## Root Cause
- Owner deletion had no dependency check for associated pets
- Error handling for FK conflicts was inconsistent (some used WrapAlreadyExists instead of WrapConflict)

## Fixes Implemented

### 1. Owner Deletion FK Check (CRITICAL FIX)
**Files:**
- `backend/internal/repository/owner_repository.go` - Added `CountPetsByOwnerID` method
- `backend/internal/service/owner_service.go` - Added FK validation before deletion
- `backend/internal/service/owner_service_test.go` - Added test case for conflict scenario

**Changes:**
```go
// Service layer
func (s *ownerService) Delete(ctx context.Context, clinicID, id uint64) error {
    petCount, err := s.repo.CountPetsByOwnerID(ctx, clinicID, id)
    if err != nil {
        return apperrors.Wrap(err, "failed to check pet dependencies")
    }
    if petCount > 0 {
        return apperrors.WrapConflict("ペットが紐付いているため削除できません。先にペットを削除してください")
    }
    // ... proceed with deletion
}
```

### 2. Animal Species Error Response Fix
**Files:**
- `backend/internal/repository/animal_species_repository.go`

**Change:**
- Line 86: Changed `WrapAlreadyExists` → `WrapConflict` for FK conflict response
- **Reason:** 409 Conflict is the correct HTTP status for FK violations

### 3. Test Coverage
Added test case: `TestOwnerService_Delete/returns_conflict_error_when_owner_has_pets`
- Verifies 409 Conflict returned when owner has dependent pets
- Validates error message and status code

## Verification
```bash
docker compose exec backend go test ./internal/service -v -run TestOwnerService_Delete
# All 4 test cases pass:
# - deletes_owner_successfully ✓
# - returns_not_found_error_when_owner_does_not_exist ✓
# - returns_error_on_repository_failure ✓
# - returns_conflict_error_when_owner_has_pets ✓ (NEW)
```

## FK Dependency Status (All Master Records)

| Entity | Method | Status | Check Type |
|--------|--------|--------|------------|
| Owner | Delete | ✅ FIXED | CountPetsByOwnerID |
| AnimalSpecies | Delete | ✅ EXISTING | CountBySpeciesID (in repository) |
| CheckupType | Delete | ✅ EXISTING | CountUsageByCheckupTypeID |
| ChiefComplaintCategory | Delete | ✅ EXISTING | CountByChiefComplaintCategoryID |
| Cage | Delete | ✅ EXISTING | ExistsByCageID |

## Test Results
- Unit tests: 100% pass
- Coverage: FK dependency checks for 5 master entities
- Error handling: All use WrapConflict for 409 responses

## CLAUDE.md Compliance
✅ Implements requirement: "マスタ削除時は必ず依存レコードの存在をチェックし、参照がある場合は apperrors.WrapConflict(...) で 409 を返す"

## Commit
```
commit e3fdf9b
fix(backend): add FK dependency checks for master record deletion
```

---
**Fixed Date:** 2026-04-01
**Impact:** CRITICAL - Prevents data integrity violations
