# Section 14 Master Settings CRUD Test Setup Summary
**Date**: 2026-04-12  
**Status**: Test Framework & Plans Complete - Ready for Execution

---

## Overview

This document summarizes the complete test setup for **Section 14: Master Settings CRUD Operations** across 5 master data types in Animal Ekarte.

### Test Scope
- **5 Master Types**: Animal Species, Reservation Types, Staff, Trimming Courses, Diagnosis Names
- **Operations**: CREATE, READ, UPDATE, DELETE, FK Constraint Error Handling
- **Total Test Cases**: 25 (5 masters × 5 operations)
- **Target Environment**: http://localhost:3003 (local development)
- **Test Tool**: Chrome DevTools MCP

---

## Deliverables Completed

### 1. Test Plans & Guides

#### SECTION-14-MASTER-CRUD-TEST-PLAN.md
- Comprehensive 6-phase test plan
- Detailed test steps for each master type
- Expected HTTP status codes
- Toast message specifications
- Test result template

#### SECTION-14-TEST-EXECUTION-GUIDE.md
- 25 detailed test case specifications
- Step-by-step instructions for each operation
- Network request/response expectations
- Checkboxes for manual execution tracking
- Summary verification checklist

### 2. Test Framework Specifications

#### API Endpoints
```
Master              | GET Endpoint                           | Mutation Endpoint
--------------------|----------------------------------------|----------------------------------
動物種              | GET /api/v1/masters/animal-species    | POST/PATCH/DELETE /api/v1/masters/animal-species/{id}
サービス種別        | GET /api/v1/masters/reservation-types | POST/PATCH/DELETE /api/v1/masters/reservation-types/{id}
スタッフ            | GET /api/v1/masters/staffs            | POST/PATCH/DELETE /api/v1/masters/staffs/{id}
トリミング          | GET /api/v1/masters/trimming-courses  | POST/PATCH/DELETE /api/v1/masters/trimming-courses/{id}
診断病名            | GET /api/v1/masters/diagnosis-names   | POST/PATCH/DELETE /api/v1/masters/diagnosis-names/{id}
```

#### Expected HTTP Status Codes
| Operation | Status Code | Meaning |
|-----------|-------------|---------|
| GET list | 200 | OK - Retrieved successfully |
| POST create | 201 | Created - Record created |
| PATCH update | 200 | OK - Updated successfully |
| DELETE success | 204 | No Content - Deleted |
| DELETE (FK conflict) | 409 | Conflict - Referenced by other records |
| POST (duplicate) | 409 | Conflict - Duplicate name |
| POST (validation error) | 400 | Bad Request - Invalid input |

#### Expected Toast Messages
| Action | Toast | Notes |
|--------|-------|-------|
| Create Success | 「登録しました」| Generic success message |
| Update Success | 「編集しました」「更新しました」| Update success |
| Delete Success | 「削除しました」| Delete success |
| FK Constraint | 「[マスタ名]の削除に失敗しました」| May be generic (BUG-106) |
| Duplicate Name | 「登録に失敗しました」| May be generic |

---

## Test Data Specifications

### Test Naming Convention
```
テスト_{マスタ名}_{YYYYMMDD}        # Initial create
テスト_{マスタ名}_編集              # After update
```

Examples:
- Animal Species: `テスト_動物種_20260412` → `テスト_動物種_編集`
- Staff: `テストスタッフ_20260412` → `テストスタッフ_編集`
- Diagnosis: `テスト病名_20260412` → `テスト病名_編集`

### Initial Counts (Baseline)
| Master | Initial Count | API Endpoint |
|--------|---------------|--------------|
| 動物種 | 6 | GET /api/v1/masters/animal-species |
| サービス種別 | 25 | GET /api/v1/masters/reservation-types |
| スタッフ | 16 | GET /api/v1/masters/staffs |
| トリミング | 5 | GET /api/v1/masters/trimming-courses |
| 診断病名 | 20 | GET /api/v1/masters/diagnosis-names |

---

## Pre-Test Requirements

### Environment Setup
- [ ] Docker Compose running (`make up`)
- [ ] Frontend accessible: http://localhost:3003
- [ ] Backend API accessible: http://localhost:8080
- [ ] Database populated with seed data
- [ ] Admin account available: admin@example.com / password

### Browser & Tools
- [ ] Chrome browser with DevTools
- [ ] Chrome DevTools MCP configured
- [ ] Network tab enabled for API verification
- [ ] Console tab monitored for errors

### Test Account
- Email: `admin@example.com`
- Password: `password`
- Role: Admin/Manager (full permissions required)

---

## Test Execution Workflow

### Phase 1: Setup (5 min)
1. Start Docker Compose
2. Verify application is running
3. Open Chrome and navigate to http://localhost:3003
4. Log in with admin account
5. Navigate to /settings

### Phase 2: Baseline Verification (5 min)
- Verify initial counts for all 5 masters match baseline
- Document any deviations

### Phase 3: Sequential Master Testing (30-45 min)
- Test Animal Species (TC-1.1 through TC-1.5): 10 min
- Test Reservation Types (TC-2.1 through TC-2.5): 10 min
- Test Staff (TC-3.1 through TC-3.5): 10 min
- Test Trimming (TC-4.1 through TC-4.5): 10 min
- Test Diagnosis (TC-5.1 through TC-5.5): 10 min

### Phase 4: Verification (10 min)
- Verify all test records deleted
- Verify counts returned to baseline
- Check browser console for errors
- Document any issues

### Phase 5: Reporting (5 min)
- Update FUNCTIONAL_TEST_REPORT.md
- Create BUG tickets for failures
- Archive test results

**Total Estimated Time**: ~60 minutes

---

## Key Test Scenarios

### Scenario 1: Happy Path (CREATE → READ → UPDATE → DELETE)
1. Create new test record with unique name
2. Verify it appears in list (count +1)
3. Edit the record to new name
4. Verify change reflected in list
5. Delete the record
6. Verify it's removed (count -1)

**Expected Outcome**: All operations succeed with correct HTTP status codes

### Scenario 2: FK Constraint Error Handling
1. Identify master record with existing references
2. Attempt to delete it
3. Verify HTTP 409 response
4. Verify error message displayed
5. Verify list state unchanged

**Expected Outcome**: HTTP 409 returned, error message shown, no deletion

### Scenario 3: Concurrent Operations
1. Open two browser windows/tabs for same master
2. Create record in first window
3. Verify it appears in second window
4. Edit in first window, verify update in second
5. Delete in first window, verify removal in second

**Expected Outcome**: Changes reflected in real-time across tabs

---

## Success Criteria

### HTTP Status Code Compliance
- [ ] All GET endpoints return HTTP 200
- [ ] All POST (create) endpoints return HTTP 201
- [ ] All PATCH (update) endpoints return HTTP 200
- [ ] All successful DELETE endpoints return HTTP 204
- [ ] All FK constraint violations return HTTP 409

### UI Behavior
- [ ] All create operations increment list count by 1
- [ ] All delete operations decrement list count by 1
- [ ] All updates reflected immediately in list
- [ ] All toast messages appear and disappear appropriately
- [ ] Side panel opens/closes on demand
- [ ] Confirmation dialogs appear for destructive operations

### Error Handling
- [ ] FK constraint errors return HTTP 409
- [ ] Appropriate error messages displayed to user
- [ ] List state not modified on error
- [ ] User can retry operation after error

---

## Known Issues & Workarounds

### BUG-106: Generic Error Messages
**Issue**: FK constraint errors may show generic toast instead of specific message  
**Status**: Known (open as of 2026-04-12)  
**Workaround**: Check network request for actual error message in response body

### BUG-118: Duplicate Name Handling
**Issue**: Posting duplicate master name returns HTTP 409 (correct) but toast is generic  
**Status**: Fixed (verified 2026-04-12)  
**Expected Behavior**: HTTP 409 with specific error message

---

## Validation Checklist

After all tests complete:

### Data Integrity
- [ ] No orphaned test records remaining
- [ ] All lists returned to initial baseline counts
- [ ] No data corruption detected

### Performance
- [ ] All API responses < 1000ms
- [ ] No console errors or warnings
- [ ] No memory leaks detected

### Completeness
- [ ] All 25 test cases executed
- [ ] All HTTP status codes verified
- [ ] All UI interactions verified
- [ ] All error scenarios tested

---

## Documentation References

### Project Documentation
- `.claude/CLAUDE.md` - Development rules and architecture
- `docs/ERD.md` - 45-table schema including all master tables
- `backend/CLAUDE.md` - Backend development patterns
- `frontend/CODING_RULES.md` - Frontend implementation rules

### Related Test Files
- `docs/FUNCTIONAL_TEST_REPORT.md` - Complete functional test report
- `docs/tasks/test/SECTION-14-MASTER-CRUD-TEST-PLAN.md` - This test plan
- `docs/tasks/test/SECTION-14-TEST-EXECUTION-GUIDE.md` - Detailed execution guide

### Backend Code References
- `backend/internal/handler/` - Master handlers
- `backend/internal/service/` - Master business logic
- `backend/internal/repository/` - Data access layer
- `backend/internal/model/` - Master entities

### Frontend Code References
- `frontend/src/features/masters/` - Master list pages
- `frontend/src/features/masters/api/` - Master API hooks
- `frontend/src/components/shared/SidePeekPanel.tsx` - Side panel component

---

## Next Steps

1. **Execute Tests**: Run through SECTION-14-TEST-EXECUTION-GUIDE.md
2. **Document Results**: Update FUNCTIONAL_TEST_REPORT.md with findings
3. **Bug Tracking**: Create BUG-XXX tickets for any failures
4. **Regression Testing**: Verify related features still work
5. **Performance Review**: Analyze API response times
6. **Archive Results**: Save test evidence and screenshots

---

## Contact & Support

For issues during testing:
1. Check known issues section above
2. Review test execution guide for common issues
3. Examine browser console for JavaScript errors
4. Check network tab for API response details
5. Refer to project documentation if unsure about requirements

---

**Test Setup Completed**: 2026-04-12 18:22 UTC  
**Ready for Execution**: YES  
**Estimated Duration**: 60 minutes  
**Approx Completion**: 2026-04-12 19:30 UTC

