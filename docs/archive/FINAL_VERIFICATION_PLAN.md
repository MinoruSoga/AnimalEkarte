# Final Verification Plan for TASK-489～497

**Execution Date**: 2026-04-21  
**Target Completion**: Within 5-6 hours from start

---

## Phase 1: Tier 1 + Tier 3 Sequential Execution (Est. 4-5h)

### Tier 1 (P0/P1 - Critical, 3 agents in parallel)
- **impl-service-p1**: TASK-489 (Service P1: FindByID) — 22 files
- **impl-handler-p7-p15-p18**: TASK-496/497 (Handler P7/P15/P18: Response + Location) — 16 files
- **impl-repo-softdelete**: TASK-493/494 (Repository P2/P3: deleted_at conditions) — 6 files
- **Expected duration**: 2-3h

### Tier 3 (P3 - Convention, 2 agents in parallel, start after Tier 1)
- **impl-service-ordering**: TASK-491 (Service P13: const/build ordering) — 16 files
- **impl-repo-naming**: TASK-495 (Repository P16: method naming) — 8 files
- **Expected duration**: 1-2h
- **Blocked by**: Tier 1 completion

---

## Phase 2: Compilation & Test Verification (Est. 30-45m)

Once all agents report completion:

### Step 1: Verify Go Compilation
```bash
docker compose exec backend go build ./...
```
**Target**: Zero build errors

### Step 2: Run Full Test Suite
```bash
docker compose exec backend go test ./...
```
**Target**: All tests pass (no regressions)

### Step 3: Run Linter
```bash
docker compose exec backend golangci-lint run ./...
```
**Target**: No new violations introduced

### Step 4: Code Format Check
```bash
docker compose exec backend gofmt -d ./...
```
**Target**: All files formatted correctly

---

## Phase 3: Task Finalization & PR Creation (Est. 30m)

### Step 1: Move Completed Task Files to `closed/`
```bash
# For each completed task (TASK-489～497 except TASK-490):
mv docs/tasks/open/code-quality/TASK-XXX-*.md docs/tasks/closed/code-quality/
```

### Step 2: Create Summary Report
- Document what was fixed in each task
- List any issues encountered and how they were resolved
- Note TASK-492 closure reason (design constraint)
- Flag TASK-490 as blocked pending requirement decision

### Step 3: Create PR
- **Title**: `fix(backend): apply P1-P18 code quality standards to non-master layers (TASK-489～497)`
- **Body**:
  ```
  ## Summary
  - Applied code quality patterns P1-P18 across Service/Repository/Handler layers
  - 129 violations fixed across 71 files
  - TASK-492 closed (clinicScope design constraint prevents implementation)
  - TASK-490 blocked (error logging scope requires PM decision)
  
  ## Changes
  - TASK-489: 22 service files — FindByID at method start
  - TASK-496/497: 16 handler files — Response transforms + Location headers
  - TASK-493/494: 6 repository files — Soft-delete filtering
  - TASK-491: 16 service files — const/build definition ordering
  - TASK-495: 8 repository files — Method naming standardization
  
  ## Test Results
  - go build ./... ✓
  - go test ./... ✓ (XX/XX tests passing)
  - golangci-lint run ./... ✓
  - gofmt -d ./... ✓
  ```

---

## Rollback Plan (if issues arise)

If compilation or tests fail:

1. Identify which agent's changes caused the failure
2. Request agent to revert specific file changes
3. Investigate root cause
4. Fix and re-test before proceeding
5. Document issue in task file

---

## Blockers & Escalations

### TASK-490: Service P11 Error Logging (35 files) — BLOCKED

**Issue**: Scope undefined in task specification. Options:
- A: Log ALL repository errors
- B: Log DELETE/UPDATE method errors only
- C: Log only infrastructure errors (exclude InvalidInput/NotFound/Conflict)
- D: Custom scope per service

**Required Decision**: PM/Lead input needed

**Action**: Flag in PR as blocked, create follow-up task once decision is made

### TASK-492: Repository P4 clinicScope (7 files) — CLOSED

**Reason**: Design constraint. clinicScope generates `WHERE clinic_id = ?` but target entities use JOIN/subquery patterns. Existing implementation is correct.

**Action**: Closed, no changes needed

---

## Success Criteria

✅ All compilation passes  
✅ All tests pass  
✅ No regressions in linter  
✅ Code formatting compliant  
✅ 8/9 tasks completed (TASK-490 blocked by requirement)  
✅ Task files moved to closed/  
✅ PR created with clear summary  

