# FE Master Code Quality Task Assignment — 2026-04-21

## Executive Summary

Comprehensive code quality scan of 48 FE master files (API 24, Routes 19, Components 5) identified **65 violations across 29 failing files**. 8 actionable task files created with prioritized execution order.

- **Total Files Affected**: 48
- **Total Violations**: 65
- **Estimated Effort**: 10-14 hours
- **Start Date**: 2026-04-21
- **Target Completion**: 2026-04-25

---

## Priority 1: Type Safety Foundation (3 tasks, ~3.5 hours)

### TASK-484: FA7 Request Type Derivation
**Severity**: 🔴 HIGH  
**Files**: 14 (animal-species, cages, chief-complaint-types, hospitalization-plans, inquiry-templates, insurances, merchandise-items, occupations, permission-groups, staffs, trimming, payment-method-master, company, consultations)  
**Violations**: 20 (CreateXxx, UpdateXxx interfaces hand-written instead of Omit<Model, ...>)  
**Effort**: 2-3 hours  
**Status**: Ready for implementation

**What to fix**: Replace all hand-written request interfaces with type derivation using `Omit<ModelXxx, 'id' | 'created_at' | 'updated_at'>` or `Partial<CreateXxxRequest>`

**Example**:
```typescript
// Before
export interface CreateAnimalSpeciesRequest {
  name: string;
  is_active: boolean;
  sort_order: number;
}

// After
export type CreateAnimalSpeciesRequest = Omit<ModelAnimalSpecies, 'id' | 'created_at' | 'updated_at'>;
```

---

### TASK-489: FA1 Query Key Const Assertion ✨ NEW
**Severity**: 🟡 MEDIUM  
**Files**: 11 (cages, checkup-types, consultations, exam-types-master, insurances, medicines, occupations, payment-method-master, procedures, vaccines-master, staffs)  
**Violations**: 11 + 48 inline query keys (missing `as const` and named constants)  
**Effort**: 1-2 hours  
**Status**: Ready for implementation

**What to fix**: Extract inline query keys to named constants with `as const` assertion for type safety and cache stability.

**Example**:
```typescript
// Before
useQuery({
  queryKey: ['masters', 'cages'],  // New reference per call = cache miss
  queryFn: getCages,
})

// After
const CAGES_QUERY_KEY = ['masters', 'cages'] as const;
useQuery({
  queryKey: CAGES_QUERY_KEY,
  queryFn: getCages,
})
```

---

### TASK-490: FA5 Exported Request Types ✨ NEW
**Severity**: 🟡 MEDIUM  
**Files**: 2 (cages.ts, reservation-type-occupations.ts)  
**Violations**: 2 (ReorderCagesRequest, LinkOccupationRequest missing exports)  
**Effort**: 30 minutes  
**Status**: Ready for implementation

**What to fix**: Export formal request interfaces instead of inline anonymous types.

**Example**:
```typescript
// Before
export function reorderCages(ids: number[]) { /* ... */ }

// After
export interface ReorderCagesRequest {
  ids: number[];
}
export function reorderCages(req: ReorderCagesRequest) { /* ... */ }
```

---

## Priority 2: Hook Standardization (3 tasks, ~4.5 hours)

### TASK-485: FA4 Missing Reorder/Update Hooks
**Severity**: 🔴 HIGH  
**Files**: 10 (chief-complaint-types, hospitalization-plans, inquiry-templates, insurances, occupations, reservation-type-groups, staffs, trimming, payment-method-master, company)  
**Violations**: 8 (9 missing `useReorderXxx` hooks + 1 missing `useUpdateCompany`)  
**Effort**: 2-3 hours  
**Status**: Ready for implementation (depends on TASK-484)

**What to fix**: Add mutation hook wrappers with query invalidation and error handling for reorder/update operations.

**Example**:
```typescript
export function useReorderChiefComplaintTypes() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: reorderChiefComplaintTypes,
    onSuccess: () => {
      queryClient.invalidateQueries({ 
        queryKey: CHIEF_COMPLAINT_TYPES_QUERY_KEY 
      });
    },
    onError: (error) => handleApiError(error, "並び替え"),
  });
}
```

---

### TASK-491: FR1 useMasterCRUD Hook Integration ✨ NEW
**Severity**: 🟡 MEDIUM  
**Files**: 2 (DiagnosisSettings.tsx, ReservationTypeSettings.tsx)  
**Violations**: 2 (manual state management instead of hook)  
**Effort**: 1 hour  
**Status**: Ready for implementation

**What to fix**: Replace manual `useState` CRUD state with `useMasterCRUD` hook for consistency and maintainability.

**Example**:
```typescript
// Before
const [editTarget, setEditTarget] = useState<DiagnosisModel | null>(null);
const handleEdit = useCallback((item) => setEditTarget(item), []);

// After
const { editTarget, setEditTarget, clearEdit } = useMasterCRUD(ResourceMasterDiagnosis);
```

---

### TASK-492: FR2 useMasterSave Hook Integration ✨ NEW
**Severity**: 🟡 MEDIUM  
**Files**: 3 (DiagnosisSettings.tsx, MedicineSettings.tsx, ReservationTypeSettings.tsx)  
**Violations**: 3 (direct mutation calls instead of hook)  
**Effort**: 1.5 hours  
**Status**: Ready for implementation

**What to fix**: Replace direct `useMutation` calls with `useMasterSave` hook for consistent error handling and loading state.

**Example**:
```typescript
// Before
const mutation = useMutation({
  mutationFn: updateDiagnosis,
  onSuccess: () => { /* ... */ },
});
mutation.mutate(data);

// After
const { mutate: save } = useMasterSave(ResourceMasterDiagnosis);
save(data);
```

---

## Priority 3: UI/Permissions (2 tasks, ~1.5 hours)

### TASK-487: FR3 usePermission Missing
**Severity**: 🟡 MEDIUM  
**Files**: 6 (ChiefComplaintSettings, HospitalizationSettings, InsuranceSettings, InterviewTemplateSettings, OccupationSettings, PaymentMethodSettings)  
**Violations**: 5 (usePermission not called explicitly)  
**Effort**: 30 minutes  
**Status**: Ready for implementation

**What to fix**: Add explicit `usePermission(ResourceXxx)` calls in Routes pages for permission caching and visibility.

**Example**:
```typescript
import { usePermission } from "@/hooks/use-permission";
import { ResourceMasterChiefComplaint } from "@/types/generated/models";

export function ChiefComplaintSettings() {
  const { canEdit, canCreate, canDelete } = usePermission(ResourceMasterChiefComplaint);
  // ... component code
}
```

---

### TASK-488: FG1 Design Tokens Hardcoding
**Severity**: 🟢 LOW  
**Files**: 6 (PermissionGroupSettings, ReservationTypeSettings, ReservationTypeSidePanel, ReservationTypeGroupSidePanel, StaffSettings, TrimmingSettings)  
**Violations**: 6 (input layout classes hardcoded as Tailwind instead of design tokens)  
**Effort**: 1 hour  
**Status**: Ready for implementation  
**⚠️ Caveat**: Not verified in partial Routes rescan — recommend full 19-file scan before implementation

**What to fix**: Replace hardcoded Tailwind input layout classes with `LAYOUT.*` design tokens.

**Example**:
```typescript
// Before
<input className="w-7 h-7 rounded cursor-pointer border-0 bg-transparent p-0" />

// After
<input className={LAYOUT.colorInputSmall} />
```

---

## Priority 4: Architecture Decision Required (Deferred)

### TASK-486: FA3/FA6 Query Key + Cache Config
**Status**: 🔴 BLOCKED (pending architectural decision)  
**Files**: 5 (company.ts, reservation-type-occupations.ts, reservation-type-unavailable-times.ts)  
**Issue**: Sub-resource query key naming pattern (hierarchical vs flat) needs architect approval  
**Effort**: 1 hour (once decision made)

**Action Required**: Clarify with PO/Architect:
- Should sub-resources use flat pattern: `["masters", "reservation-type-occupations"]`?
- Or hierarchical: `["reservation-types", reservationTypeId, "occupations"]`?

---

## Implementation Notes

### Dependency Graph
```
TASK-484 (FA7) 
  ↓ (prerequisite for)
TASK-485 (FA4) + TASK-489 (FA1) + TASK-490 (FA5)  
  ↓
TASK-491 (FR1) + TASK-492 (FR2)
  ↓ (independent)
TASK-487 (FR3) + TASK-488 (FG1)

TASK-486 ← BLOCKED (awaiting architecture decision)
```

### Batch Implementation Strategy
- **Type safety tasks (TASK-484/489/490)**: Can be implemented in parallel after understanding patterns
- **Hook tasks (TASK-485/491/492)**: Implement in order (TASK-485 must complete before TASK-491/492)
- **UI tasks (TASK-487/488)**: Independent, can be done in parallel

### Quality Checks Post-Implementation
1. Run `pnpm type-check` to verify TypeScript compliance
2. Run `pnpm lint` to check code style
3. Spot-check Routes layer mutations to ensure hook integration works
4. Verify query cache behavior (no cache misses from new const-asserted keys)

---

## Scan Metadata

- **Scan Date**: 2026-04-21
- **Scan Tool**: Parallel agent teams (Team-API, Team-Routes, Team-Components)
- **Files Scanned**: 42 (API 24/24, Routes 13/19*, Components 5/5)
- **Patterns Assessed**: FA1-FA7 (API), FR1-FR5 + FG1-FG3 (Routes), FG1-FG3 (Components)
- **Violations Found**: 65 total (API 54, Routes 11, Components 0)
- **Memory Reference**: `/memory/fe_master_rescan_20260421.md`

*Routes: Incomplete scan (13 of 19 files). Recommend full scan to verify TASK-488 FG1 violations.

---

## Questions?

All task files include:
- Detailed violation examples with line numbers
- Fix strategy with before/after code
- Acceptance criteria
- Notes on dependencies and trade-offs

Review task files in `docs/tasks/open/code-quality/` for detailed implementation guidance.
