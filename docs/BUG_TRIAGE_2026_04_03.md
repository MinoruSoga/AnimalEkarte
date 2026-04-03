# Bug Triage Report - Post v2.2.0 Release
> **Date**: 2026-04-03
> **Prepared**: After FUNCTIONAL_TEST_REPORT.md analysis of 8,255 test items
> **Status**: v2.2.0 deployed to production ✅

---

## Executive Summary

Total test items: **2,611**
Confirmed working: **2,111** (81%)
Unconfirmed: **1,514** (mostly UI confirmations)
Critical blockers: **7 items**

---

## TIER 1: CRITICAL - Fix Immediately (Estimated: 3-4 days)

### 1. **BUG-102: Clinical-Plan PATCH Error**
- **Module**: Medical Records / Diagnosis Tab
- **Severity**: MEDIUM
- **Endpoint**: `PATCH /api/v1/medical-records/:id/clinical-plan`
- **Issue**: When diagnosis fields contain only DEFAULT values (`plan="# 治療方針"`, `assessment="# 診断詳細"`), FE sends `undefined` → BE returns 400 "at least one field must be provided"
- **Impact**: Users cannot save diagnosis tab with default values
- **Fix Options**:
  - *Option A (Recommended)*: FE sends string value even if default → BE accepts gracefully
  - *Option B*: BE allows partial/empty updates with 200 response (no-op)
- **Effort**: SMALL (1-2 hours)
- **Files to Check**:
  - `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx` (clinical-plan save handler)
  - `backend/internal/handler/clinical_plan_handler.go` (PATCH validation)

### 2. **BUG-019: RBAC - Permission Group Visibility**
- **Module**: Permission Management / Admin Settings
- **Severity**: HIGH (Security & UX)
- **Issue**: Non-admin `staff` user can see "新規登録" (New) and edit buttons in permission group management
- **Expected**: Buttons hidden for non-admin users (role-based access control)
- **Related**: BUG-056, FE-133, FE-134
- **Impact**: Potential unauthorized permission escalation attempts
- **Effort**: SMALL-MEDIUM (2-3 hours)
- **Files to Check**:
  - `frontend/src/features/admin/components/PermissionGroupList.tsx` (visibility check)
  - `backend/internal/middleware/auth.go` (role extraction)

### 3. **BUG-109: Merchandise Item Delete - Missing FK Check**
- **Module**: Inventory / Merchandise Master
- **Severity**: CRITICAL (Data Integrity)
- **Issue**: `merchandise_item_service.Delete()` has NO FK dependency check before deletion
- **Risk**: Can delete items referenced in `billing_items` and `estimate_items`, creating orphaned records
- **Current Implementation**: Only `FindByID + Delete`, no validation
- **Fix Required**: Add dependency count check before delete (same pattern as used in treatment service)
- **Effort**: SMALL (30 minutes)
- **Files to Change**:
  - `backend/internal/service/merchandise_item_service.go:172-193` (Add FK check)
  - `backend/internal/repository/merchandise_item_repository.go:122-124` (Implement CountUsageByMerchandiseItemID)

### 4. **Doctor ID Mismatch in Medical Record Form**
- **Module**: Medical Records / PatientInfoCard
- **Severity**: MEDIUM (UX)
- **Issue**: When creating medical record from reservation, `doctor_id` initial value doesn't match the reserved doctor
- **Expected**: `PatientInfoCard` should show reserved doctor name
- **Impact**: Users see different doctor than expected (data integrity OK, UX confusing)
- **Effort**: SMALL (1 hour)
- **Files to Check**:
  - `frontend/src/features/medical-records/components/PatientInfoCard.tsx` (initial value)
  - `backend/internal/handler/medical_record_handler.go:GetMedicalRecord` (API response)

---

## TIER 2: HIGH - Sprint 1 of v2.3.0 (Estimated: 8-10 days)

### 5. **Hospitalization Management - Backend APIs**
- **Module**: Hospitalization
- **Severity**: HIGH
- **Status**: UI 100% complete, APIs 0% implemented
- **Missing Endpoints**:
  1. `PUT /v1/hospitalizations/:id/care-plan-items/:item_id` - Update care plan
  2. `DELETE /v1/hospitalizations/:id/care-plan-items/:item_id` - Delete care plan
  3. `PUT /v1/hospitalizations/:id/daily-records/:date/vitals/:vital_id` - Update vital
  4. `DELETE /v1/hospitalizations/:id/daily-records/:date/vitals/:vital_id` - Delete vital
  5. `POST /v1/hospitalizations/:id/staff-notes` - Add staff note
- **Effort**: LARGE (16 hours backend + 4 hours FE integration)
- **Files Needed**:
  - `backend/internal/handler/hospitalization_handler.go` (new handlers)
  - `backend/internal/service/hospitalization_service.go` (business logic)
  - `frontend/src/features/hospitalization/api/` (hook updates)

### 6. **Dashboard - Kanban Drag & Drop**
- **Module**: Dashboard / Reservations Kanban
- **Severity**: MEDIUM (UX)
- **Issue**: Dragging reservation card to "受付済" column doesn't update status
- **Expected**: Card moves and status changes to `reservation_status="checked_in"`
- **Impact**: Staff must use manual menu to check in patients
- **Effort**: LARGE (8-10 hours investigation + redesign)
- **Investigation Needed**:
  - Current drag-drop implementation pattern
  - API endpoint for status update
  - Whether design needs modification

### 7. **Accounting - Unit Price Auto-fill**
- **Module**: Accounting / Estimate/Invoice
- **Severity**: MEDIUM
- **Issue**: Treatment unit prices ("一般診察", "注射") show ¥0, unclear if should auto-fill from master
- **Expected Behavior**: TBD - needs product clarification
- **Impact**: Manual entry required, error-prone
- **Effort**: MEDIUM (investigation + 4-6 hours)
- **Depends On**: Product decision on master linkage vs manual entry

---

## TIER 3: MEDIUM - Sprint 2+ (Estimated: 4-6 days total)

### 8. **Master Item Delete - UI Missing**
- **Module**: Masters / Item Management
- **Severity**: MEDIUM
- **Issue**: Delete button missing from master item list UI
- **Note**: API returns 409 correctly (FK conflict check works)
- **Fix**: Add delete button UI + confirmation dialog
- **Effort**: SMALL (2-3 hours)

### 9. **Accounting - Record Reflection in List**
- **Module**: Accounting
- **Severity**: MEDIUM
- **Issue**: After confirming accounting in medical record, new record doesn't appear in accounting management list
- **Impact**: Manual page refresh needed
- **Effort**: MEDIUM (investigation + 4-5 hours)
- **Investigation**: Query cache invalidation, record creation timing

### 10. **Vaccine - Data Import Functionality**
- **Module**: Vaccinations
- **Severity**: MEDIUM
- **Issue**: Data import button exists but functionality untested
- **Current**: Button UI complete
- **Effort**: LARGE (investigation + 6-8 hours)
- **Unknowns**: Import format, validation rules, mapping

---

## TIER 4: LOW - Polish & Testing (Estimated: 2-3 days)

### Accessibility & Navigation
- Keyboard navigation (Tab, arrow keys, focus trap)
- Browser back button edge cases
- Long text display handling (overflow, truncation)
- Concurrent tab editing

---

## Recommended Development Roadmap

### **Sprint 1: v2.3.0a (Fix Critical Bugs)** - 3-4 days
**Priority**: Production stability

- [ ] BUG-102: Clinical-plan PATCH error
- [ ] BUG-019: RBAC permission visibility
- [ ] BUG-109: Merchandise delete FK check
- [ ] Doctor ID mismatch fix

**Deliverable**: v2.3.0a hotfix release

---

### **Sprint 2: v2.3.0b (Hospitalization Backend)** - 8-10 days
**Priority**: Feature completion

- [ ] Implement 5 missing hospitalization APIs
- [ ] Integrate with FE hooks
- [ ] End-to-end testing
- [ ] Deploy to staging

**Deliverable**: v2.3.0b feature release

---

### **Sprint 3: v2.3.0c (Dashboard & Accounting)** - 6-8 days
**Priority**: UX improvement

- [ ] Kanban drag & drop implementation/redesign
- [ ] Unit price auto-fill (pending product decision)
- [ ] Accounting list cache invalidation
- [ ] Master delete UI

**Deliverable**: v2.3.0c enhancement release

---

### **Sprint 4: v2.3.1 (Data Import & Polish)** - 5-7 days
**Priority**: Data operations & accessibility

- [ ] Vaccine data import functionality
- [ ] Accessibility testing & fixes
- [ ] Edge case handling
- [ ] Documentation updates

**Deliverable**: v2.3.1 full release + docs

---

## Risk Assessment

### **Blocking Risks** (Must resolve before merge)
1. BUG-102: Blocks medical record workflow
2. BUG-109: Data integrity risk
3. Hospitalization APIs: Feature incomplete

### **Medium Risks** (Can defer 1-2 sprints)
1. BUG-019: RBAC bypass (document workaround)
2. Kanban drag & drop: Manual workflow exists
3. Unit price auto-fill: Needs product clarity

### **Low Risks** (Polish items)
1. Keyboard navigation: Accessibility improvement
2. Long text display: Rare edge case

---

## Testing Strategy for Each Fix

### BUG-102: Clinical-Plan PATCH
```bash
# Test case: Save diagnosis with default values only
1. Create medical record
2. Leave Diagnosis fields as default (# 治療方針, # 診断詳細)
3. Click Save
4. Expected: HTTP 200 success (not 400)
```

### BUG-019: RBAC Permission
```bash
# Test case: Non-admin user accessing permission management
1. Login as 'vet@example.com' (staff role)
2. Navigate to admin → permission groups
3. Expected: "新規登録" button hidden
4. Expected: Edit/delete buttons grayed out or hidden
```

### BUG-109: Merchandise FK
```bash
# Test case: Delete merchandise item in use
1. Create billing item with merchandise X
2. Try to delete merchandise X
3. Expected: 409 Conflict with message: "この物販品目は請求・見積データで使用中のため削除できません"
4. Verify: Item NOT deleted
```

---

## Implementation Notes

### **Database Schema Impact**
- **No migrations required** for TIER 1-3 fixes
- Hospitalization APIs use existing schema (care_plan_items, daily_records, vital_records already exist)
- Merchandise delete fix uses existing `CountUsageByMerchandiseItemID` (currently returns 0)

### **API Contract Changes**
- Clinical-plan PATCH: Change validation logic (no breaking change to request)
- Hospitalization: 5 new endpoints (additive, no impact on existing clients)
- All other fixes: No API contract changes

### **Dependencies**
- No new external packages required
- No Go version upgrade needed
- No React/Node upgrade needed

---

## Success Criteria

### Before v2.3.0 Release
- [ ] All TIER 1 bugs fixed and tested
- [ ] All TIER 2 features implemented and passing integration tests
- [ ] TIER 3 bugs resolved or documented as deferred
- [ ] Staging deployment: all 17+ feature tests passing
- [ ] Production v2.3.0: tagged and documented

### Code Quality Gates
- [ ] Go: golangci-lint passes (0 errors)
- [ ] React: npm run lint passes (0 errors)
- [ ] TypeScript: npm run build succeeds
- [ ] Test coverage: >80% for new code
- [ ] Integration tests: All green

---

## Documentation Needed

1. **BUG-102 Resolution**: Update API docs for clinical-plan PATCH behavior
2. **BUG-019 RBAC**: Update admin documentation on role-based visibility
3. **Hospitalization APIs**: Add to API docs, swagger comments
4. **v2.3.0 Release Notes**: List all bug fixes + features

---

## References

- **FUNCTIONAL_TEST_REPORT.md**: Full test coverage and issue details
- **IMPLEMENTATION_COMPLETE_2026_04_02.md**: v2.2.0 delivery summary
- **Backend CLAUDE.md**: Error handling & API patterns
- **Frontend CODING_RULES.md**: React 19 action patterns

---

**Last Updated**: 2026-04-03
**Next Review**: After Sprint 1 completion (estimated 2026-04-06)
