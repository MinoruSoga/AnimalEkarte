# FE-071: 共有 hooks ファイル名 kebab-case 化

**Status**: Closed
**Priority**: High
**Affects**: `frontend/src/hooks/` — 全 feature から参照される共有 hooks
**Date Created**: 2026-03-18
**Related**: TASK-018, FE-072

## Summary

`frontend/src/hooks/` 内の 5 ファイルが camelCase で命名されている。プロジェクト規約（kebab-case 強制）に準拠するため、ファイル名を kebab-case にリネームし、全 import パスを更新する。

## 現状のコード

### 違反ファイル一覧（5 ファイル）

```
frontend/src/hooks/usePagination.ts        → use-pagination.ts
frontend/src/hooks/useStaffValidation.ts   → use-staff-validation.ts
frontend/src/hooks/useReducedMotion.ts     → use-reduced-motion.ts
frontend/src/hooks/useUnsavedChanges.ts    → use-unsaved-changes.ts
frontend/src/hooks/useSortableList.ts      → use-sortable-list.ts
```

### import 元ファイル一覧（21 箇所）

#### `usePagination` → `use-pagination`（3 箇所）
```typescript
// frontend/src/features/owners/routes/OwnersList.tsx:24
import { usePagination } from "@/hooks/usePagination";

// frontend/src/features/medical-records/routes/MedicalRecords.tsx:22
import { usePagination } from "@/hooks/usePagination";

// frontend/src/features/trimming/routes/TrimmingList.tsx:30
import { usePagination } from "@/hooks/usePagination";
```

#### `useStaffValidation` → `use-staff-validation`（2 箇所）
```typescript
// frontend/src/features/medical-records/routes/MedicalRecords.tsx:23
import { useStaffValidation } from "@/hooks/useStaffValidation";

// frontend/src/features/trimming/routes/TrimmingList.tsx:31
import { useStaffValidation } from "@/hooks/useStaffValidation";
```

#### `useReducedMotion` → `use-reduced-motion`（2 箇所）
```typescript
// frontend/src/features/reservations/components/WeekView.tsx:13
import { useReducedMotion } from "@/hooks/useReducedMotion";

// frontend/src/features/master/routes/MedicineSettings.tsx:63
import { useReducedMotion } from "@/hooks/useReducedMotion";
```

#### `useUnsavedChanges` → `use-unsaved-changes`（8 箇所）
```typescript
// frontend/src/features/owners/routes/OwnerForm.tsx:47
import { useUnsavedChanges } from "@/hooks/useUnsavedChanges";

// frontend/src/features/medical-records/routes/MedicalRecordForm.tsx:32
import { useUnsavedChanges } from "@/hooks/useUnsavedChanges";

// frontend/src/features/estimates/routes/EstimateForm.tsx:17
import { useUnsavedChanges } from '@/hooks/useUnsavedChanges';

// frontend/src/features/vaccinations/routes/VaccinationForm.tsx:25
import { useUnsavedChanges } from "@/hooks/useUnsavedChanges";

// frontend/src/features/examinations/routes/ExaminationForm.tsx:16
import { useUnsavedChanges } from "@/hooks/useUnsavedChanges";

// frontend/src/features/trimming/routes/TrimmingForm.tsx:30
import { useUnsavedChanges } from "@/hooks/useUnsavedChanges";

// frontend/src/features/hospitalization/routes/HospitalizationForm.tsx:26
import { useUnsavedChanges } from "@/hooks/useUnsavedChanges";

// frontend/src/features/inventory/routes/InventoryForm.tsx:23
import { useUnsavedChanges } from "@/hooks/useUnsavedChanges";
```

#### `useSortableList` → `use-sortable-list`（6 箇所）
```typescript
// frontend/src/features/master/routes/ServiceTypeSettings.tsx:4
import { useSortableList } from "@/hooks/useSortableList";

// frontend/src/features/master/routes/TreatmentPlanMaster.tsx:11
import { useSortableList } from "@/hooks/useSortableList";

// frontend/src/features/master/routes/MedicineSettings.tsx:64
import { useSortableList } from "@/hooks/useSortableList";

// frontend/src/features/master/routes/DiagnosisSettings.tsx:18
import { useSortableList } from "@/hooks/useSortableList";

// frontend/src/features/master/routes/CageSettings.tsx:4
import { useSortableList } from "@/hooks/useSortableList";

// frontend/src/features/master/routes/AnimalSpeciesSettings.tsx:4
import { useSortableList } from "@/hooks/useSortableList";
```

## 必要な変更

### 1. ファイル rename（git mv）

```bash
cd frontend/src/hooks
git mv usePagination.ts use-pagination.ts
git mv useStaffValidation.ts use-staff-validation.ts
git mv useReducedMotion.ts use-reduced-motion.ts
git mv useUnsavedChanges.ts use-unsaved-changes.ts
git mv useSortableList.ts use-sortable-list.ts
```

### 2. import パス更新（21 箇所）

全ファイルで以下の置換を実行：

| 変更前 | 変更後 |
|--------|--------|
| `@/hooks/usePagination` | `@/hooks/use-pagination` |
| `@/hooks/useStaffValidation` | `@/hooks/use-staff-validation` |
| `@/hooks/useReducedMotion` | `@/hooks/use-reduced-motion` |
| `@/hooks/useUnsavedChanges` | `@/hooks/use-unsaved-changes` |
| `@/hooks/useSortableList` | `@/hooks/use-sortable-list` |

**注意**: 関数名（`usePagination` 等）はそのまま。変更するのはファイル名とimportパスのみ。

## プロジェクトルール遵守チェック

- [ ] ファイル名が kebab-case
- [ ] import パスが新ファイル名を参照
- [ ] 関数名は変更しない（camelCase のまま正しい）

## 完了条件

- [ ] 5 ファイルが kebab-case にリネーム済み
- [ ] 21 箇所の import パスが更新済み
- [ ] `pnpm build` パス
- [ ] `pnpm lint` パス
- [ ] `grep -r "usePagination\|useStaffValidation\|useReducedMotion\|useUnsavedChanges\|useSortableList" frontend/src --include="*.ts" --include="*.tsx" -l` で import パスに旧名が残っていない（関数名は除く）

## クローズ情報

- **Closed At**: 2026-03-18
- **変更ファイル**:
  - `frontend/src/hooks/usePagination.ts` → `use-pagination.ts`
  - `frontend/src/hooks/useStaffValidation.ts` → `use-staff-validation.ts`
  - `frontend/src/hooks/useReducedMotion.ts` → `use-reduced-motion.ts`
  - `frontend/src/hooks/useUnsavedChanges.ts` → `use-unsaved-changes.ts`
  - `frontend/src/hooks/useSortableList.ts` → `use-sortable-list.ts`
  - 21箇所の import パス更新（+ PATTERNS.md 1箇所）
