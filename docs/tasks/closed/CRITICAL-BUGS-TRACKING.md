# Critical & Medium Bug Tracking

**Last Updated:** 2026-04-01
**Total NG Items:** 277 (from FUNCTIONAL_TEST_REPORT.md)
**Classification:** CRITICAL (41), MEDIUM (214), LOW (19)

## CRITICAL Bugs (41 items)

### ✅ FIXED (1/41)
- [x] **BE-FK-001: Owner Deletion with Dependent Pets**
  - **Status:** FIXED (2026-04-01)
  - **Impact:** Prevents deletion of owners with associated pets
  - **Implementation:** CountPetsByOwnerID check + 409 Conflict response
  - **Files:** owner_service.go, owner_repository.go
  - **Commit:** e3fdf9b

### ⏳ TODO (40/41)

#### Data Integrity Issues (15 items)
1. **BE-FK-002: Pet Deletion with Medical Records**
   - Pets with medical records should not be deletable
   - Need: CountMedicalRecordsByPetID in medical_record_repo

2. **BE-FK-003: Clinic Deletion with Associated Data**
   - Clinic deletion should check for owners, pets, staff, appointments
   - Need: CheckDependencies method in clinic_service

3. **BE-FK-004: AnimalSpecies Deletion Edge Case**
   - Already has check, but error message uses WrapAlreadyExists (should be WrapConflict)
   - ✅ Fixed: Changed to WrapConflict

4. **BE-FK-005: Equipment/Cage Deletion Verification**
   - Needs validation of hospitalization dependencies

5. **BE-FK-006 to BE-FK-015:** Other master-detail FK relationships
   - Bill deletion with invoice items
   - Hospitalization with care plans and logs
   - Medical record with exam results
   - Treatment with items
   - Prescription with medicine records

#### Validation Issues (10 items)
6. **BE-VAL-001: Email Format Validation**
   - No regex validation for email format
   - Current: Only uniqueness check
   - Need: Add RFC 5322 email validation

7. **BE-VAL-002: Phone Number Format Validation**
   - No format validation for phone numbers
   - Need: Accept common formats (090-xxxx-xxxx, 03-xxxx-xxxx, etc.)

8. **BE-VAL-003: Postal Code Format**
   - No validation for postal code format
   - Need: Accept 7-digit format with optional hyphen

9. **BE-VAL-004: Negative Amount Validation**
   - ✅ Existing: UnitPrice, DiscountRate checks in place
   - Verify: Quantity, Weight validations

10. **BE-VAL-005: Enum Field Validation**
    - ✅ Existing: Gender, Status, DangerLevel validation in place
    - Need: Verify all enum types are validated

#### API/HTTP Issues (8 items)
11. **BE-HTTP-001: Concurrent Edit Conflict Detection**
    - Medical records have Version field for optimistic locking
    - Status: Implement in Update method
    - Need: Verify version increment on updates

12. **BE-HTTP-002: Cache Invalidation on Updates**
    - Medical records cache needs invalidation on state changes
    - Need: React Query cache busting on PATCH/DELETE

13. **BE-HTTP-003: Status Code Consistency**
    - PATCH returns 200 OK (should be 200 or 204)
    - DELETE returns 204 No Content (correct)
    - Verify: All PUT/PATCH/DELETE status codes

14. **BE-HTTP-004: Error Response Format**
    - Verify all error responses follow standard format
    - Need: Ensure error codes are consistent

15. **BE-HTTP-005 to BE-HTTP-008:** Request/Response validation issues

#### Caching Issues (7 items)
16. **BE-CACHE-001: Query Result Caching**
    - ListOwners query should cache results
    - Invalidate on Create/Update/Delete

17. **BE-CACHE-002 to BE-CACHE-007:** Cache invalidation for other entities

---

## MEDIUM Bugs (214 items)

### Categories
- **Missing Features:** 85 items (form fields, filter options, export functions)
- **UI/UX Issues:** 65 items (button placement, labels, error messages)
- **Data Display Issues:** 40 items (sorting, pagination, formatting)
- **Performance Issues:** 24 items (list loading, search performance)

### Examples
- [ ] Medical record status transitions not fully implemented
- [ ] Vaccination record archiving
- [ ] Trimming service cancellation workflow
- [ ] Hospitalization discharge procedures
- [ ] Accounting report generation
- [ ] Permission-based feature visibility

---

## LOW Bugs (19 items)

- [ ] Graph/chart generation (medical trends)
- [ ] Document printing (invoices, medical reports)
- [ ] Comment/note system
- [ ] Audit logging for record changes
- [ ] Data export (CSV, PDF)

---

## Priority Recommendations

### Phase 1: Data Integrity (Next Sprint)
1. Fix remaining FK dependency checks (15 items)
2. Add email/phone/postal code validation (4 items)
3. Implement version checking for concurrent edits

### Phase 2: API/HTTP Layer (Sprint+1)
4. Verify all HTTP status codes
5. Implement proper cache invalidation
6. Add request/response validation

### Phase 3: Features & UX (Sprint+2)
7. Implement MEDIUM priority features (85 items)
8. Fix UI/UX issues (65 items)
9. Add missing data displays (40 items)

---

## Testing Strategy

### Unit Tests
- FK dependency checks: ✅ Added
- Validation rules: ✅ Add test cases
- Concurrent edit handling: ⏳ Add tests

### Integration Tests
- End-to-end owner lifecycle (create → update → delete with pets)
- Medical record versioning
- Cache invalidation flow

### Acceptance Tests
- API contract compliance
- Error response formats
- HTTP status codes

---

## Metrics

| Category | Total | Fixed | In Progress | Backlog |
|----------|-------|-------|-------------|---------|
| CRITICAL | 41 | 1 | 0 | 40 |
| MEDIUM | 214 | 0 | 0 | 214 |
| LOW | 19 | 0 | 0 | 19 |
| **Total** | **277** | **1** | **0** | **276** |

---

## Next Steps

1. ✅ Complete: FK dependency check for owner deletion
2. → Implement remaining FK checks (BE-FK-002 to BE-FK-015)
3. → Add email/phone/postal code validation
4. → Review concurrent edit version handling
5. → Create integration tests for critical flows
