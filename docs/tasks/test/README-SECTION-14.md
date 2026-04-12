# Section 14: Master Settings CRUD Test Framework
**Animal Ekarte Functional Test Suite**  
**Date Created**: 2026-04-12  
**Status**: Ready for Execution

---

## Quick Start

### For Test Executors (Browser Testing)
1. Start Docker: `make up`
2. Open Chrome DevTools: F12
3. Navigate to: http://localhost:3003
4. Log in: admin@example.com / password
5. Open: `/docs/tasks/test/SECTION-14-TEST-EXECUTION-GUIDE.md`
6. Execute tests step-by-step, checking boxes as you go
7. Report results back to: `docs/FUNCTIONAL_TEST_REPORT.md`

### For Test Managers (Review & Planning)
1. Review: `SECTION-14-TEST-SETUP-SUMMARY.md` (overview & requirements)
2. Review: `SECTION-14-MASTER-CRUD-TEST-PLAN.md` (detailed test plan)
3. Assign: `SECTION-14-TEST-EXECUTION-GUIDE.md` to testers
4. Track: Test completion with checkboxes
5. Verify: All 25 test cases completed before sign-off

---

## Document Guide

### 📋 SECTION-14-TEST-SETUP-SUMMARY.md
**Purpose**: Overview, requirements, and pre-flight checklist  
**Audience**: Test managers, stakeholders  
**Length**: 1 page  
**Key Sections**:
- Test scope: 5 masters, 25 test cases
- Pre-test requirements & checklist
- Workflow phases with time estimates
- Success criteria
- Known issues & workarounds

**When to Use**: 
- Before assigning tests to team
- To understand overall test strategy
- To verify environment setup

---

### 🎯 SECTION-14-MASTER-CRUD-TEST-PLAN.md
**Purpose**: Detailed test plan with expected outcomes  
**Audience**: QA engineers, test lead  
**Length**: 2-3 pages  
**Key Sections**:
- 6-phase test execution plan
- Master endpoint reference table
- CRUD operation steps for each of 5 masters
- Expected HTTP status codes (200, 201, 204, 409)
- Toast message specifications
- Test result template

**When to Use**:
- As reference during test execution
- To verify API contract compliance
- To document expected vs actual behavior

---

### 🧪 SECTION-14-TEST-EXECUTION-GUIDE.md
**Purpose**: Step-by-step test case instructions with checkboxes  
**Audience**: Test executors (manual browser testing)  
**Length**: 4-5 pages  
**Key Sections**:
- 25 numbered test cases (TC-1.1 through TC-5.5)
- Each case has: preconditions, steps, expected results
- Checkbox format for tracking progress
- Network verification steps
- Summary verification checklist

**When to Use**:
- During actual test execution
- To mark progress as you complete each test
- As reference for each test case details

---

## Test Coverage Matrix

```
Master              READ    CREATE  UPDATE  DELETE  FK Error  TOTAL
-------------------------------------------------------------------
動物種              TC-1.1  TC-1.2  TC-1.3  TC-1.4  TC-1.5    5 tests
サービス種別        TC-2.1  TC-2.2  TC-2.3  TC-2.4  TC-2.5    5 tests
スタッフ            TC-3.1  TC-3.2  TC-3.3  TC-3.4  TC-3.5    5 tests
トリミング          TC-4.1  TC-4.2  TC-4.3  TC-4.4  TC-4.5    5 tests
診断病名            TC-5.1  TC-5.2  TC-5.3  TC-5.4  TC-5.5    5 tests
-------------------------------------------------------------------
TOTAL TEST CASES:   5       5       5       5       5         25 tests
```

---

## Test Data

### Baseline Counts (Initial State)
```
動物種:        6 items
サービス種別:  25 items
スタッフ:      16 items
トリミング:    5 items
診断病名:      20 items
```

### Test Data Format
```
Create:  テスト_{マスタ名}_{YYYYMMDD}
Update:  テスト_{マスタ名}_編集
Delete:  (same item from update)
```

Examples:
- `テスト_動物種_20260412` → `テスト_動物種_編集` → DELETE
- `テストスタッフ_20260412` → `テストスタッフ_編集` → DELETE
- `テスト病名_20260412` → `テスト病名_編集` → DELETE

---

## API Endpoints Quick Reference

| Master | GET | CREATE | UPDATE | DELETE |
|--------|-----|--------|--------|--------|
| 動物種 | `/api/v1/masters/animal-species` | `POST` | `PATCH /{id}` | `DELETE /{id}` |
| サービス | `/api/v1/masters/reservation-types` | `POST` | `PATCH /{id}` | `DELETE /{id}` |
| スタッフ | `/api/v1/masters/staffs` | `POST` | `PATCH /{id}` | `DELETE /{id}` |
| トリミング | `/api/v1/masters/trimming-courses` | `POST` | `PATCH /{id}` | `DELETE /{id}` |
| 診断病名 | `/api/v1/masters/diagnosis-names` | `POST` | `PATCH /{id}` | `DELETE /{id}` |

---

## Expected HTTP Status Codes

| Operation | Code | Meaning | Example |
|-----------|------|---------|---------|
| List/Read | 200 | OK - Success | GET /masters/* → 200 |
| Create | 201 | Created - Record created | POST /masters/* → 201 |
| Update | 200 | OK - Updated | PATCH /masters/*/{id} → 200 |
| Delete (Success) | 204 | No Content | DELETE /masters/*/{id} → 204 |
| FK Constraint Error | 409 | Conflict - Referenced | DELETE /masters/*/{id} → 409 |
| Duplicate Name | 409 | Conflict - Duplicate | POST /masters/* (dup) → 409 |
| Validation Error | 400 | Bad Request | POST /masters/* (invalid) → 400 |

---

## Test Execution Timeline

### Estimated Schedule
```
Phase 1: Setup               (5 min)   - Docker, browser, login
Phase 2: Baseline Check      (5 min)   - Verify initial counts
Phase 3: Master Testing      (40 min)  - 5 masters × 8 min each
Phase 4: Verification        (5 min)   - Cleanup, error check
Phase 5: Reporting           (5 min)   - Update results
───────────────────────────────────────
TOTAL                        60 min
```

### Sample Daily Schedule
```
10:00 - Setup & baseline (10 min)
10:10 - Animal Species tests (10 min)
10:20 - Reservation Type tests (10 min)
10:30 - BREAK
10:40 - Staff tests (10 min)
10:50 - Trimming tests (10 min)
11:00 - Diagnosis tests (10 min)
11:10 - Verification & reporting (10 min)
11:20 - COMPLETE
```

---

## Success Criteria Checklist

### Pre-Test
- [ ] Docker Compose running
- [ ] Frontend accessible (http://localhost:3003)
- [ ] Admin account works (admin@example.com)
- [ ] Chrome DevTools open with Network tab enabled
- [ ] Test guide printed or in separate monitor

### During Test
- [ ] Each test case has checkbox marked [✓] PASS or [✗] FAIL
- [ ] Network requests verified in Chrome DevTools
- [ ] Toast messages captured in notes
- [ ] Screenshots taken for any NG cases
- [ ] Console checked for JavaScript errors

### Post-Test
- [ ] All 25 test cases marked complete
- [ ] Test item counts returned to baseline
- [ ] No orphaned test data in database
- [ ] Results documented in FUNCTIONAL_TEST_REPORT.md
- [ ] BUG-XXX tickets created for any failures

---

## Troubleshooting

### Test Setup Issues

**Q: Application won't start**  
A: Check `make up` output, verify Docker running, check ports 3003 and 8080

**Q: Can't log in**  
A: Verify database seeded, check admin account exists, clear browser cookies

**Q: Network requests not showing**  
A: Open DevTools → Network tab → F5 to reload, ensure XHR filter selected

### Test Execution Issues

**Q: Create operation shows 201 but item not in list**  
A: Wait 2-3 seconds for API response, refresh page with F5, check browser console

**Q: Delete returns 409 (FK error) but list shows it was deleted**  
A: Refresh page, may be UI inconsistency, check database directly

**Q: Toast message doesn't match expected**  
A: May be a UI bug (BUG-106 generic messages), note in test results

**Q: Test item disappeared from list**  
A: Check if page was accidentally refreshed, scroll to find it, check sort order

---

## Common Issues & Workarounds

### BUG-106: Generic Error Messages
**Issue**: FK constraint errors show generic toast instead of specific message  
**Workaround**: Check DevTools Network tab for actual error message  
**Impact**: Test still passes if HTTP 409 returned, just note generic message

### BUG-118: Duplicate Name Handling
**Issue**: Duplicate master names return 409 (correct) but generic toast  
**Workaround**: Check network request for error details  
**Impact**: API works correctly, UI message is generic

### React Query Cache Issues
**Issue**: After delete, item still shows in list briefly  
**Workaround**: Wait 2-3 seconds or refresh page  
**Impact**: UI flicker, but resolves automatically

---

## Document Maintenance

### When Test Execution Completes
1. Mark completion date in this README
2. Update baseline counts if different from expected
3. Update BUG count in summary
4. Archive test evidence (screenshots, network logs)
5. Commit results to git

### When Issues Are Found
1. Create BUG-XXX ticket in docs/tasks/open/crash/
2. Reference test case (e.g., "TC-1.3 UPDATE failed")
3. Include expected vs actual results
4. Note reproduction steps
5. Update FUNCTIONAL_TEST_REPORT.md

### When Code Changes
1. Re-baseline initial counts
2. Update API endpoint references if changed
3. Update expected status codes if business logic changed
4. Re-run all 25 tests after code changes

---

## Related Documentation

### Project Standards
- `.claude/CLAUDE.md` - Development rules
- `frontend/CODING_RULES.md` - Frontend rules
- `backend/CLAUDE.md` - Backend patterns

### Architecture
- `docs/ERD.md` - Database schema (45 tables)
- `docs/architecture.md` - System architecture
- `.claude/docs/overview.md` - Layered architecture overview

### Functional Tests
- `docs/FUNCTIONAL_TEST_REPORT.md` - Master test report (all sections)
- `docs/IMPLEMENTATION_COMPLETE.md` - Feature completion status
- `docs/tasks/open/` - Open bug tickets

---

## Sign-Off Template

```
Test Execution Sign-Off
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Date: 2026-04-12
Test Executor: [Name]
Environment: http://localhost:3003 (Local)
Total Tests: 25
Results: [  ] All PASS [  ] Some NG [  ] Blocked

Test Summary:
  PASS:    __ / 25
  NG:      __ / 25
  PARTIAL: __ / 25

Known Issues:
  - [BUG-XXX] [description]

Recommendations:
  [ ] All clear, no issues
  [ ] Deploy to staging
  [ ] Do NOT deploy, bugs need fixing
  [ ] Retest after fixes

Signed: ________________  Date: __________
```

---

## Footer

**Created**: 2026-04-12 18:22 UTC  
**Framework Version**: 1.0  
**Status**: Ready for Execution  
**Estimated Tests to Execute**: 25  
**Estimated Time Required**: 60 minutes

For questions or issues, refer to `.claude/CLAUDE.md` or `docs/architecture.md`

