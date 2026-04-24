# BE Non-Master Code Quality Implementation Status

**Date**: 2026-04-21  
**Scope**: TASK-489～497 (9 tasks, 129 violations → fix implementation)

## Execution Plan by Priority Tier

### P0/P1 (Critical Data Integrity + API Contract)

| Task | Pattern | Files | Agent | Status |
|------|---------|-------|-------|--------|
| **TASK-489** | Service P1: FindByID先頭 | 22 | impl-service-p1 | IN PROGRESS |
| **TASK-496/497** | Handler P7/P15/P18: Response + Location | 16 | impl-handler-p7-p15-p18 | IN PROGRESS |

### P2 (Data Consistency)

| Task | Pattern | Files | Agent | Status |
|------|---------|-------|-------|--------|
| **TASK-493** | Repository P2: CountBy deleted_at | 5 | impl-repo-softdelete | IN PROGRESS |
| **TASK-494** | Repository P3: Preload deleted_at | 1 | impl-repo-softdelete | IN PROGRESS |
| **TASK-492** | Repository P4: clinicScope | 7 | — | **CLOSED** (設計制約: JOIN/サブクエリ方式は clinicScope 不適用) |

### P3 (Convention/Naming/Readability) — Pending

| Task | Pattern | Files | Agent | Status | Dependent On |
|------|---------|-------|-------|--------|--------------|
| **TASK-491** | Service P13: const/build ordering | 16 | — | PENDING | P0/P1 tier completion |
| **TASK-495** | Repository P16: method naming | 8 | — | PENDING | P2 tier completion |

### P1 (Observability) — Blocked on Requirement Review

| Task | Pattern | Files | Status | Issue |
|------|---------|-------|--------|-------|
| **TASK-490** | Service P11: slog.ErrorContext | 35 | **BLOCKED** | Scope undefined (all repo calls? Delete only? Excluding validation?) — Requires PM/lead decision |

---

## Agent Status

### CLOSED
- **impl-repo-scope** (TASK-492): ✅ Completed — clinicScope design constraint confirmed, no changes needed

### Tier 1 (Active - Target: 2-3h)

**Agents spawned with explicit first-file targets:**

1. **impl-service-p1**
   - Task: TASK-489 (22 files)
   - Files: accounting_service.go, appointment_admin_service.go, checkup_service.go, clinic_service.go, ... (18 more)
   - Action: Add FindByID at method start in Delete/Update
   - Status: IN PROGRESS (awaiting first report)

2. **impl-handler-p7-p15-p18**
   - Task: TASK-496/497 (16 files)
   - Files: auth_handler.go, cash_register_handler.go, examination_handler.go, hospitalization_handler.go, ... (12 more)
   - Action: (P7) toXxxResponse transforms + (P15) 201+Location headers + (P18) rename buildXxx → toXxx
   - Status: IN PROGRESS (awaiting first report)

3. **impl-repo-softdelete**
   - Task: TASK-493/494 (6 files)
   - Files: estimate_repository.go, examination_repository.go (Preload), medical_record_repository.go, owner_repository.go, pet_repository.go, + vital/treatment/care_plan via joins
   - Action: Add `deleted_at IS NULL` to CountBy/Preload
   - Status: IN PROGRESS (awaiting first report)

### Tier 2 (Not Started)

- **impl-repo-scope**: TASK-492 → **CLOSED** (설計상 불가능)

### Tier 3 (Standby - Ready for activation)

- **impl-service-ordering**: TASK-491 (16 files) — Spawned, awaiting signal after Tier 1 completion
- **impl-repo-naming**: TASK-495 (8 files) — Spawned, awaiting signal after Tier 1 completion

### Tier 4 (Blocked)

- **impl-service-logging**: TASK-490 (35 files) — Requires requirement clarification before spawn

---

## Expected Timeline

- **Tier 1**: 2-3h (parallel execution of P0/P1 fixes)
- **Tier 2**: Skipped (TASK-492 closed)
- **Tier 3**: 1-2h (sequential after Tier 1)
- **Tier 4**: TBD (requires decision on P11 scope)

**Total Estimate**: 4-5h (with TASK-490 decision included)

---

## Blockers & Decisions

1. **TASK-492 (clinicScope)**: Design constraint prevents implementation. Status: CLOSED.
2. **TASK-490 (error logging)**: P11 scope undefined. Awaiting PM/lead decision:
   - Option A: All repository error calls → slog.ErrorContext
   - Option B: Delete/Update methods only
   - Option C: Exclude validation errors (InvalidInput, NotFound, Conflict)
   - Option D: DB/infrastructure errors only (connection, timeout, constraint)

---

## Next Steps

1. Monitor Tier 1 agents (impl-service-p1, impl-handler-p7-p15-p18, impl-repo-softdelete) for completion
2. Upon completion, spawn Tier 3 agents (impl-service-ordering, impl-repo-naming)
3. Resolve TASK-490 requirement before implementation
4. Final verification: Run full test suite (`docker compose exec backend go test ./...`)
5. Create PR for all fixes combined or by priority tier

