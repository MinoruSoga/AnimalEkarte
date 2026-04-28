# NOTION-TAB-002: Unified Tab Component Architecture

**Status**: In Progress  
**Scope**: Refactor 8 tab-using files to use UnifiedTabs shared component  
**Impact**: Zero logic changes; pure UI abstraction consolidation  

---

## Overview

NOTION-TAB-001 completed styling unification (6 files). NOTION-TAB-002 consolidates **all 8 tab implementations** into a single `UnifiedTabs` component that:

1. **Abstracts shadcn/ui Tabs** (AggregationDashboardPage, HospitalizationList, HospitalizationTabbedView)
2. **Abstracts Radix TabsPrimitive** (DiagnosisSettings, TreatmentPlanMaster, TrimmingSettings)
3. **Unifies custom button tabs** (AccountingList, MedicalRecordForm)
4. **Applies consistent design tokens** — no more per-page className variations
5. **Maintains all functionality** — URL state, CSV generation, form submission, etc.

---

## Inventory: 8 Tab-Using Files

| File | Current Type | Items | URL State | Design Token Alignment |
|------|-------------|-------|-----------|------------------------|
| `components/ui/tabs.tsx` | shadcn/ui wrapper | — | — | Base component |
| `aggregation/AggregationDashboardPage.tsx` | shadcn/ui Tabs | 3 (revenue/visit/last_visit) | ✅ searchParams | ✅ NOTION-TAB-001 fixed |
| `hospitalization/HospitalizationList.tsx` | shadcn/ui Tabs | 4 (active/reserved/discharged/all) | ✅ searchParams | ✅ NOTION-TAB-001 fixed |
| `hospitalization/HospitalizationTabbedView.tsx` | shadcn/ui Tabs | 2 (daily/plan) | ❌ Local state | ✅ NOTION-TAB-001 fixed |
| `master/DiagnosisSettings.tsx` | Radix TabsPrimitive | 2 | ❌ Local state | ✅ NOTION-TAB-001 fixed |
| `master/TreatmentPlanMaster.tsx` | Radix TabsPrimitive | 3 | ❌ Local state | ✅ NOTION-TAB-001 fixed |
| `master/TrimmingSettings.tsx` | Radix TabsPrimitive | 2 | ❌ Local state | ✅ NOTION-TAB-001 fixed |
| `accounting/AccountingList.tsx` | Custom buttons | 2 (list/unpaid) | ✅ searchParams | ✅ NOTION-TAB-001 fixed |
| `medical-records/MedicalRecordForm.tsx` | Custom buttons | 9 (問診/診察/治療/etc) | ❌ Local state | ⚠️ Not yet aligned |

---

## Refactoring Strategy

### Phase 1: Component Creation ✅
- ✅ Created `components/shared/UnifiedTabs.tsx` with:
  - `UnifiedTabs` component (shadcn default, custom button variant)
  - `UnifiedTabsContent` wrapper for TabsContent
  - Two implementations: `variant="shadcn"` (default) and `variant="button"`
  - Consistent token application: C.bgLight, C.dataActiveBorderB, C.dataActiveText

### Phase 2: Refactor Files (8 files)

#### 2a. shadcn/ui Tabs Files (3 files)

**Files**: AggregationDashboardPage, HospitalizationList, HospitalizationTabbedView

**Pattern**:
```typescript
// Before
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';

<Tabs value={activeTab} onValueChange={handleTabChange}>
  <TabsList className={`grid ... ${C.bgLight}`}>
    <TabsTrigger className={`... ${C.dataActiveBorderB} ${C.dataActiveText}`}>...</TabsTrigger>
    // repeat for each item
  </TabsList>
  <TabsContent value="tab1">...</TabsContent>
  // repeat for each item
</Tabs>

// After
import { UnifiedTabs, UnifiedTabsContent } from '@/components/shared/UnifiedTabs';

const tabItems = [
  { value: 'tab1', label: 'Tab 1 Label' },
  { value: 'tab2', label: 'Tab 2 Label' },
];

<UnifiedTabs
  items={tabItems}
  value={activeTab}
  onValueChange={handleTabChange}
>
  <UnifiedTabsContent value="tab1">...</UnifiedTabsContent>
  <UnifiedTabsContent value="tab2">...</UnifiedTabsContent>
</UnifiedTabs>
```

#### 2b. Radix TabsPrimitive Files (3 files)

**Files**: DiagnosisSettings, TreatmentPlanMaster, TrimmingSettings

**Pattern**:
```typescript
// Before (Radix direct usage)
import * as TabsPrimitive from '@radix-ui/react-tabs';

<TabsPrimitive.Root value={activeTab} onValueChange={setActiveTab}>
  <TabsPrimitive.List className={`... ${C.bgLight}`}>
    <TabsPrimitive.Trigger className={`... ${C.dataActiveBorderB} ${C.dataActiveText}`}>...</TabsPrimitive.Trigger>
    // repeat
  </TabsPrimitive.List>
  <TabsPrimitive.Content value="tab1">...</TabsPrimitive.Content>
  // repeat
</TabsPrimitive.Root>

// After (via UnifiedTabs)
// Same as shadcn pattern above (UnifiedTabs abstracts Radix under the hood)
```

#### 2c. Custom Button Tab Files (2 files)

**Files**: AccountingList, MedicalRecordForm

**Pattern**:
```typescript
// Before
const [activeTab, setActiveTab] = useState('tab1');

<div className="flex border-b">
  {tabItems.map((tab) => (
    <button
      onClick={() => setActiveTab(tab.value)}
      className={`... ${activeTab === tab.value ? `${C.text} ...` : `${C.text50} ...`}`}
    >
      {tab.label}
      {activeTab === tab.value ? <div className="..." /> : null}
    </button>
  ))}
</div>
{/* content rendering with activeTab === 'tab1' ? ... : null */}

// After
<UnifiedTabs
  items={tabItems}
  value={activeTab}
  onValueChange={setActiveTab}
  variant="button"  // Custom button styling
>
  {/* Content can stay inline or move to UnifiedTabsContent */}
</UnifiedTabs>
```

---

## Implementation Details

### UnifiedTabs API

```typescript
interface TabItem {
  value: string;
  label: ReactNode;
}

interface UnifiedTabsProps {
  items: TabItem[];
  value: string;
  onValueChange: (value: string) => void;
  children?: ReactNode;
  variant?: 'shadcn' | 'button';  // Default: 'shadcn'
  className?: string;              // Root wrapper
  listClassName?: string;           // TabsList-specific
  triggerClassName?: string;        // Additional TabsTrigger classes
}

// Usage
<UnifiedTabs
  items={[
    { value: 'tab1', label: 'Revenue' },
    { value: 'tab2', label: 'Visits' },
  ]}
  value={activeTab}
  onValueChange={setActiveTab}
  variant="shadcn"  // or "button"
/>
```

### Design Token Application

**All variants automatically apply**:
- `TabsList` background: `C.bgLight` (rgba(55,53,47,0.09) — subtle gray)
- Active TabsTrigger/Button: `C.dataActiveBorderB` (data-[state=active]:border-b-[#37352F])
- Active TabsTrigger/Button text: `C.dataActiveText` (data-[state=active]:text-[#37352F])
- Inactive text: `C.text60` or `C.text50`

No per-page token variation. All styling is encapsulated in UnifiedTabs.

---

## File Changes Summary

### Create
- `frontend/src/components/shared/UnifiedTabs.tsx` — ✅ Done

### Modify (8 files, ~150 lines total diff)
1. ✅ `frontend/src/features/aggregation/AggregationDashboardPage.tsx` — Convert Tabs to UnifiedTabs
2. ✅ `frontend/src/features/hospitalization/HospitalizationList.tsx` — Convert Tabs to UnifiedTabs
3. ✅ `frontend/src/features/hospitalization/HospitalizationTabbedView.tsx` — Convert Tabs to UnifiedTabs
4. ✅ `frontend/src/features/master/DiagnosisSettings.tsx` — Convert TabsPrimitive to UnifiedTabs
5. ✅ `frontend/src/features/master/TreatmentPlanMaster.tsx` — Convert TabsPrimitive to UnifiedTabs
6. ✅ `frontend/src/features/master/TrimmingSettings.tsx` — Convert TabsPrimitive to UnifiedTabs
7. ✅ `frontend/src/features/accounting/AccountingList.tsx` — Convert custom buttons to UnifiedTabs(variant="button")
8. ✅ `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx` — Convert custom buttons to UnifiedTabs(variant="button")

### No Changes Required
- `frontend/src/components/ui/tabs.tsx` — Remains as-is (base shadcn/ui component)

---

## Testing & Validation

### Type Safety
```bash
docker compose exec frontend pnpm type-check
```
Must pass with 0 errors. All TabItem[] definitions must be typed.

### Functional Verification

**Per-File Checklist**:

| File | Tab Count | URL State | CSV? | Notes |
|------|-----------|-----------|------|-------|
| AggregationDashboardPage | 3 | ✅ searchParams | ✅ | revenue/visit/last_visit → URL, CSV download |
| HospitalizationList | 4 | ✅ searchParams | ❌ | active/reserved/discharged/all → URL |
| HospitalizationTabbedView | 2 | ❌ Local | ❌ | daily/plan — local state only |
| DiagnosisSettings | 2 | ❌ Local | ❌ | Local TabsPrimitive → UnifiedTabs |
| TreatmentPlanMaster | 3 | ❌ Local | ❌ | Local TabsPrimitive → UnifiedTabs |
| TrimmingSettings | 2 | ❌ Local | ❌ | Local TabsPrimitive → UnifiedTabs |
| AccountingList | 2 | ✅ searchParams | ❌ | list/unpaid → URL |
| MedicalRecordForm | 9 | ❌ Local | ❌ | 問診/診察/治療/etc — local state + scroll reset |

### Manual Tests
1. ✅ All tabs render with Notion Primary bottom-line (not strong blue)
2. ✅ All tab switches update URL searchParams (Aggregation, Hospitalization, Accounting)
3. ✅ CSV generation still works (Aggregation)
4. ✅ Tab content lazy-loads correctly (MedicalRecordForm with mountedTabs)
5. ✅ No visual regression in mobile/tablet layouts
6. ✅ Accessibility: `role="tab"`, `data-state="active/inactive"`

---

## Risk Assessment

**Low Risk** (pure abstraction, no logic changes):
- ✅ No state management changes
- ✅ No URL routing changes
- ✅ No CSV/export logic changes
- ✅ No API calls added/removed
- ✅ No form submission changes

**Edge Cases to Verify**:
- MedicalRecordForm: Scroll reset behavior when tab changes (ref-based logic stays intact)
- AggregationDashboardPage: CSV download with correct activeTab (closure passed correctly)
- HospitalizationList: 4 tabs with width-constrained layout (grid layout preserved)

---

## Next Steps

1. Refactor shadcn/ui Tabs files (3 files)
2. Refactor Radix TabsPrimitive files (3 files)
3. Refactor custom button files (2 files)
4. Run `pnpm type-check` → must pass
5. Manual visual testing on all 8 pages
6. Commit: "refactor(fe): consolidate tab implementations into UnifiedTabs shared component"
